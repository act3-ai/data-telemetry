package db

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
)

// FindDigests returns all the digests known for a given data item.
func FindDigests(con *gorm.DB, dataID uint) ([]digest.Digest, error) {
	tx := con.Model(&Digest{}).
		Select("digest").
		Where("data_id = ?", dataID).
		Order("digest")
	var digests []digest.Digest
	if err := tx.Find(&digests).Error; err != nil {
		return nil, err
	}
	return digests, nil
}

// TODO consider using subquery
// Consider getting all the digests in batch using IN operator
// use GROUP BY and https://www.postgresql.org/docs/current/functions-aggregate.html  (string_agg for postgres)
// https://www.sqlitetutorial.net/sqlite-group_concat/  (GROUP_CONCAT for sqlite and mysql)

// Digested can be added to a query type to enable.
type Digested struct {
	DigestsCSV string `gorm:"digests_csv" json:"-"`
	// DataID  uint            `json:"-"`
	Digests []digest.Digest `gorm:"-"`
}

// AfterFind will set the Digests field.
func (e *Digested) AfterFind(tx *gorm.DB) error {
	// digests, err := FindDigests(tx, e.DataID)
	// if err != nil {
	// 	return err
	// }
	// e.Digests = digests
	digests := strings.Split(e.DigestsCSV, ",")
	e.Digests = make([]digest.Digest, len(digests))
	for i, d := range digests {
		e.Digests[i] = digest.Digest(d)
	}
	return nil
}

// TODO can we inject IncludeDigests into a BeforeFind or Query() function?
// tx.Callback().Query().Before("gorm:query").Register("telemetry:digests_query", IncludeDigests(tx.Statement.Table))

// IncludeDigests is a scope that must be used with Digested above.  It populates the DigestsCSV field.
func IncludeDigests(tableName string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		// ctx := con.Statement.Context
		// log := logger.FromContext(ctx)
		join := fmt.Sprintf("INNER JOIN digests aliases ON %s.data_id = aliases.data_id", tableName)
		tx := con.
			Joins(join).
			Group(tableName + ".id") // data_id will work as well

		// see https://learnsql.com/blog/group-concat/
		if tx.Name() == "postgres" {
			// postgres is a specific case
			// notice we cannot use DISTINCT but it seems it is always distinct
			tx.Statement.Selects = append(tx.Statement.Selects, "STRING_AGG(aliases.digest, ',') as digests_csv")
		} else {
			// GROUP_CONCAT is supported by SQLite, MySQL, HSQLDB
			tx.Statement.Selects = append(tx.Statement.Selects, "GROUP_CONCAT(DISTINCT aliases.digest) as digests_csv")
		}
		return tx
	}
}

// IncludeIsDeprecated includes an extra column "is_deprecated" if the bottle is deprecated.
func IncludeIsDeprecated() func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		tx := con.Joins("LEFT JOIN deprecates deprecation_check ON deprecation_check.deprecated_bottle_digest = aliases.digest")
		if tx.Name() == "postgres" {
			// postgres needs this select to be an aggregate function when used with other aggregate functions
			// when we are getting all digests for a bottle, we want to make sure NONE of the digests for the bottle are in the deprecated table
			// if ANY of the bottle digests give us "true" for their deprecation check, we have that whole bottle "true" for its "is_deprecated" column
			tx.Statement.Selects = append(
				tx.Statement.Selects,
				"CASE WHEN true = ANY(ARRAY_AGG(CASE WHEN deprecation_check.deprecated_bottle_digest IS NULL THEN false ELSE true END)) THEN true ELSE false END AS is_deprecated",
			)
		} else {
			tx.Statement.Selects = append(tx.Statement.Selects, "CASE WHEN deprecation_check.deprecated_bottle_digest IS NULL THEN false ELSE true END AS is_deprecated")
		}
		return tx
	}
}

// IncludeNumPulls includes an extra column "num_pulls" which is the number of pull events for the bottle.
func IncludeNumPulls() func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		tx := con.Joins("LEFT JOIN (SELECT bottle_id, COUNT(*) pull_count FROM events WHERE action='pull' GROUP BY bottle_id) pulls_table ON bottles.id = pulls_table.bottle_id")
		tx.Statement.Selects = append(
			tx.Statement.Selects,
			"MAX(COALESCE(pulls_table.pull_count, 0)) as num_pulls",
		)
		return tx
	}
}

// FindArtifact returns the artifact by bottle and path.
func FindArtifact(con *gorm.DB, bottleDigest digest.Digest, artifactPath string) (*PublicArtifact, error) {
	// Preload Tables in the Bottle struct (argument is either a table or a field name in bottle, not sure which)
	tx := con.
		Preload("Data").
		// If we Preload("Blobs") then we would want to exclude some fields with this trick
		// Preload("Blobs", func(tx *gorm.DB) *gorm.DB {
		// 	return tx.Select("id", "digest")
		// }).

		Scopes(FilterByDigest(bottleDigest, "bottles")).
		Joins("INNER JOIN bottles ON bottles.id = public_artifacts.bottle_id").
		Where(&PublicArtifact{Path: artifactPath})

	artifact := &PublicArtifact{}
	if err := tx.First(artifact).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, httputil.NewHTTPError(err, http.StatusNotFound, "Artifact not found")
		}
		return nil, err
	}
	return artifact, nil
}

// UserPulls will query the events and return a mapping of username to number of pulls for a given bottle.
func UserPulls(con *gorm.DB, dgst digest.Digest, limit int) (map[string]int, error) {
	tx := con.Table("events").
		Select("username, COUNT(*) AS pull_count").
		Where("action = 'pull' AND bottle_digest = ?", dgst).
		Limit(limit).
		Group("username")

	type result struct {
		Username  string
		PullCount int
	}

	results := []result{}
	if err := tx.Find(&results).Error; err != nil {
		return nil, err
	}

	userPulls := make(map[string]int, len(results))
	for _, r := range results {
		userPulls[r.Username] = r.PullCount
	}
	return userPulls, nil
}

// BottlePulls will query the events and return the pull Request count of the provided digest.
func BottlePulls(con *gorm.DB, dgst digest.Digest) int64 {
	var result int64
	con.Table("events").Where("bottle_digest = ? and action = ?", dgst, "pull").Count(&result)
	return result
}

// EventsCount will get total count of events in DB.
func EventsCount(con *gorm.DB) int64 {
	var result int64
	con.Table("events").Count(&result)
	return result
}

// ManifestsCount will get total count of Manifests in DB.
func ManifestsCount(con *gorm.DB) int64 {
	var result int64
	con.Table("manifests").Count(&result)
	return result
}

// BottlesCount will get total count of Bottles in DB.
func BottlesCount(con *gorm.DB) int64 {
	var result int64
	con.Table("bottles").Count(&result)
	return result
}

// ArtifactsCount will get total count of Artifacts in DB.
func ArtifactsCount(con *gorm.DB) int64 {
	var result int64
	con.Table("public_artifacts").Count(&result)
	return result
}

// SignaturesCount will get total count of Artifacts in DB.
func SignaturesCount(con *gorm.DB) int64 {
	var result int64
	con.Table("signatures").Count(&result)
	return result
}

// BlobDataBytes will get the total number of bytes for data blobs in DB.
func BlobDataBytes(con *gorm.DB) int64 {
	var result int64
	con.Table("data").Select("sum(length(raw_data))").Find(&result)
	return result
}
