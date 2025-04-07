package db

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-telemetry/v3/internal/features"
)

// FilterByDigest will use the digest to filter the query for an object type
// objType must "belong to" a Data record.  Bottle, Manifest, Event, and PublicArtifact all have this property.
func FilterByDigest(dgst digest.Digest, objType string) func(db *gorm.DB) *gorm.DB {
	return FilterByDigests([]digest.Digest{dgst}, objType)
}

// FilterByDigests will use the digests to filter the query for an object type
// objType must "belong to" a Data record.  Bottle, Manifest, Event, and PublicArtifact all have this property.
func FilterByDigests(dgsts []digest.Digest, objType string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		join := fmt.Sprintf("INNER JOIN digests ON digests.data_id = %s.data_id AND digests.digest IN ?", objType)
		tx := con.Joins(join, dgsts)
		return tx
	}
}

// SearchByAuthor will search by Author name or email.
func SearchByAuthor(author string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		if author == "" {
			return con
		}
		newquery := con.Joins("INNER JOIN authors ON authors.bottle_id = bottles.id")
		if strings.Contains(author, "@") {
			// assume we are matching on email (names should not have @ in them)
			return newquery.Where("authors.email = ?", author)
		}
		// else we name on similarity of name
		return newquery.Where("authors.name LIKE ?", "%"+author+"%")
	}
}

// SearchByRepository will search by image repository that the bottle is stored in.
func SearchByRepository(bottleRepo string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		if bottleRepo == "" {
			return con
		}
		bottleRepos := con.Session(&gorm.Session{NewDB: true}).
			Distinct("bottles.id").
			Table("events").
			Joins("INNER JOIN bottles ON events.bottle_id = bottles.id").
			Where("events.repository = ?", bottleRepo)

		return con.
			Where("bottles.id IN (?)", bottleRepos)
	}
}

// RankByDescription will use FTS if available to rank order the search by best match on description.
func RankByDescription(description string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		ctx := con.Statement.Context
		log := logger.FromContext(ctx)

		if description == "" {
			return con
		}

		switch con.Name() {
		case "postgres":
			log.InfoContext(ctx, "Using Postgres FTS")
			// '&' them together to make it work with Pgsql's tsquery format
			query := strings.ReplaceAll(description, " ", " & ")

			// Rank order the results
			// https://stackoverflow.com/questions/12933805/best-way-to-use-postgresql-full-text-search-ranking
			ranks := con.Session(&gorm.Session{NewDB: true}).
				Table("bottles").
				Select("ts_rank_cd(to_tsvector('english', bottles.description), to_tsquery(?)) AS score, bottles.id", query).
				Where("bottles.description_tsv @@ to_tsquery(?)", query)

			con.Statement.Selects = append(con.Statement.Selects, "score")
			con = con.Joins("INNER JOIN (?) AS ranked_bottles ON ranked_bottles.id = bottles.id", ranks).
				Order("score DESC").
				Group("ranked_bottles.score")

			// fallback without FTS
			// log.InfoContext(ctx, "Not using Postgres FTS")
			// con = con.Where("description ILIKE ?", "%"+description+"%")
		case "sqlite":
			if features.SqliteFTS5 {
				log.InfoContext(ctx, "Using SQLite FTS")
				ranks := con.Session(&gorm.Session{NewDB: true}).
					Table("description_fts").
					Where("description_fts MATCH ?", description)
				con = con.Joins("INNER JOIN (?) AS ranked_description ON ranked_description.description = bottles.description", ranks)
			} else {
				con = con.Where("description LIKE ?", "%"+description+"%")
			}
		}
		return con
	}
}

// RankByNumPulls orders the bottles by number of pull events each has.
func RankByNumPulls() func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		ranks := con.Session(&gorm.Session{NewDB: true}).
			Table("events").
			Select("events.bottle_id, COUNT(events.id) as pull_score").
			Where("events.action = 'pull'").
			Group("events.bottle_id")

		if con.Name() == "postgres" {
			con.Statement.Selects = append(con.Statement.Selects, "COALESCE(SUM(pull_score), 0) as pull_score")
		} else {
			con.Statement.Selects = append(con.Statement.Selects, "pull_score")
		}
		con = con.Joins("LEFT JOIN (?) AS ranked_bottles ON ranked_bottles.bottle_id = bottles.id", ranks).
			Order("pull_score DESC")

		return con
	}
}

// FilterByTrustLevel selects records that include a trust level matching the trust level string eg "verified", "trusted".
func FilterByTrustLevel(filter string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		switch filter {
		case "verified":
			return con.Where("verified = true")
		case "trusted":
			return con.Where("trusted = true")
		default:
			return con
		}
	}
}

// FilterByPublicKeyFP scopes the request to signatures with a given fingerprint.
func FilterByPublicKeyFP(fp digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		return con.Where("public_key_finger_print = ?", fp)
	}
}

// FilterBySelectors scopes the request to a bottle selector.
func FilterBySelectors(selectors []string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		ctx := con.Statement.Context
		log := logger.FromContext(ctx)

		expr3 := con.Session(&gorm.Session{NewDB: true})
		// multiple selectors are combined in an OR fashion
		for i, selector := range selectors {
			log.DebugContext(ctx, "Applying", "selector", selector)

			// TODO this parser only supports integer valued gt and lt comparison and we want float64 support
			// We need to extend/rewrite it in the future possibly pulling the code from ace/data/tool
			sel, err := labels.Parse(selector)
			if err != nil {
				err = httputil.NewHTTPError(err, http.StatusBadRequest, err.Error())
				con.AddError(err) //nolint:errcheck
				return con
			}

			requirements, selectable := sel.Requirements()
			if !selectable || len(requirements) == 0 {
				return con
			}

			// We need a new session without any clauses/expressions
			expr2 := con.Session(&gorm.Session{NewDB: true})
			for j, requirement := range requirements {
				// see https://github.com/kubernetes/apimachinery/blob/4a9e16b3571218c2e15b8d0319f0b8e0e3fbacf5/pkg/labels/selector.go#L218
				log.DebugContext(ctx, "Processing selector requirement", "requirement", requirement.String())
				labelTable := fmt.Sprintf("label%dx%d", i, j)
				expr := conditionForRequirement(con, labelTable, requirement)
				join := fmt.Sprintf("LEFT JOIN labels %s ON %s.bottle_id = bottles.id", labelTable, labelTable)
				expr2 = expr2.Where(expr)
				con = con.Joins(join)
			}
			expr3 = expr3.Or(expr2)
		}

		return con.Where(expr3)
	}
}

func conditionForRequirement(con *gorm.DB, labelTable string, requirement labels.Requirement) *gorm.DB {
	// We need a new session without any clauses/expressions
	expr := con.Session(&gorm.Session{NewDB: true})

	switch requirement.Operator() {
	case selection.Equals, selection.DoubleEquals:
		value, ok := requirement.Values().PopAny()
		if !ok {
			// this should not be possible
			con.AddError(fmt.Errorf("value is required for key %s", requirement.Key())) //nolint:errcheck
			return con
		}
		expr = expr.Where(labelTable+".key = ?", requirement.Key()).
			Where(labelTable+".value = ?", value)
	case selection.NotEquals:
		value, ok := requirement.Values().PopAny()
		if !ok {
			// this should not be possible
			con.AddError(fmt.Errorf("value is required for key %s", requirement.Key())) //nolint:errcheck
			return con
		}
		subQuery := con.Session(&gorm.Session{NewDB: true}).
			Table("labels").
			Select("labels.bottle_id").
			Where("labels.key = ? AND labels.value = ?", requirement.Key(), value)
		expr = expr.Where("bottles.id NOT IN (?)", subQuery)
		// expr = expr.Where(labelTable+".key != ? AND "+labelTable+".value != ?", requirement.Key(), value)
	case selection.In:
		expr = expr.Where(labelTable+".key = ?", requirement.Key()).
			Where(labelTable+".value IN ?", requirement.Values().List())
	case selection.NotIn:
		// expr = expr.Where(labelTable+".key = ?", requirement.Key()).
		// 	Where(labelTable+".value NOT IN ?", requirement.Values().List())
		subQuery := con.Session(&gorm.Session{NewDB: true}).
			Table("labels").
			Select("labels.bottle_id").
			Where("labels.key = ? AND labels.value IN (?)", requirement.Key(), requirement.Values().List())
		expr = expr.Where("bottles.id NOT IN (?)", subQuery)
	case selection.Exists: // selector="key1"
		// no condition on value, since we are only checking if the key exists (with any value or no value)
		expr = expr.Where(labelTable+".key = ?", requirement.Key())
	case selection.DoesNotExist: // selector="!key1"
		subQuery := con.Session(&gorm.Session{NewDB: true}).
			Table("labels").
			Select("labels.bottle_id").
			Where("labels.key = ?", requirement.Key())
		expr = expr.Where("bottles.id NOT IN (?)", subQuery)
	case selection.GreaterThan:
		value, ok := requirement.Values().PopAny()
		if !ok {
			// this should not be possible
			con.AddError(fmt.Errorf("value is required for key %s", requirement.Key())) //nolint:errcheck
			return con
		}
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			con.AddError(fmt.Errorf("value must be a float for key %s", requirement.Key())) //nolint:errcheck
			return con
		}
		expr = expr.Where(labelTable+".key = ?", requirement.Key()).
			Where(labelTable+".numeric_value > ?", v)
	case selection.LessThan:
		value, ok := requirement.Values().PopAny()
		if !ok {
			// this should not be possible
			con.AddError(fmt.Errorf("value is required for key %s", requirement.Key())) //nolint:errcheck
			return con
		}
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			con.AddError(fmt.Errorf("value must be a float for key %s", requirement.Key())) //nolint:errcheck
			return con
		}
		expr = expr.Where(labelTable+".key = ?", requirement.Key()).
			Where(labelTable+".numeric_value < ?", v)
	}
	return expr
}

// WithSignature scopes the request to bottles with a signature that has the given fingerprint.
func WithSignature(digests []digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		bottleIDsWithSignature := con.Session(&gorm.Session{NewDB: true}).
			Distinct("signatures.bottle_id").
			Table("signatures").
			Where("signatures.public_key_finger_print IN ?", digests)

		return con.Where("bottles.id IN (?)", bottleIDsWithSignature)
	}
}

// WithSignatureAnnotations scopes the request to bottles with a signature that has the given annotation.
func WithSignatureAnnotations(annotations []string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		if len(annotations) == 0 {
			return con
		}
		bottleIDsWithSignatureWithAnnotation := con.Session(&gorm.Session{NewDB: true}).
			Distinct("signatures.bottle_id").
			Table("signature_annotations").
			Joins("INNER JOIN signatures ON signature_annotations.signature_id = signatures.id").
			Where("signature_annotations.key || '=' || signature_annotations.value IN  ?", annotations)

		return con.Where("bottles.id IN (?)", bottleIDsWithSignatureWithAnnotation)
	}
}

// ParentsOf is a scope that will the query to parents of the provided digests.
func ParentsOf(digests []digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		// ctx := con.Statement.Context
		// log := logger.FromContext(ctx)

		/*
			SELECT bottles.id, bottles.description
			FROM bottles
				INNER JOIN digests digests_ancestor ON bottles.data_id = digests_ancestor.data_id
			WHERE digests_ancestor.digest IN (
				SELECT DISTINCT sources.bottle_digest
				FROM bottles
					INNER JOIN digests digests_original ON bottles_original.data_id = digests_original.data_id
					INNER JOIN bottles bottles_original ON digests_original.data_id = bottles_original.data_id
					INNER JOIN sources ON bottles_original.id = sources.bottle_id
				WHERE digests_original.digest = "sha256:3e8e2e3db7a8e23283cd0a8bd14d697dcc55459da33c57af5148ec7923924a52"
			);
		*/

		parentDigests := con.Session(&gorm.Session{NewDB: true}).
			Distinct("sources.bottle_digest").
			Table("bottles").
			Joins("INNER JOIN digests digests_original ON bottles.data_id = digests_original.data_id").
			Joins("INNER JOIN sources ON bottles.id = sources.bottle_id").
			Where("digests_original.digest IN ?", digests)

		return con.Joins("INNER JOIN digests digests_parents ON digests_parents.data_id = bottles.data_id AND digests_parents.digest IN (?)", parentDigests)
	}
}

// ChildrenOf is a scope that will the query to children of the provided digests.
func ChildrenOf(digests []digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		// ctx := con.Statement.Context
		// log := logger.FromContext(ctx)
		/*
			SELECT bottles.id, bottles.description
			FROM bottles
				INNER JOIN sources ON sources.bottle_id = bottles.id
			WHERE bottle_digest IN (
				SELECT DISTINCT digests.digest
				FROM digests
					INNER JOIN digests digests_original ON digests.data_id = digests_original.data_id
				WHERE digests_original.digest = "sha256:93648e4272e7d8044959f7d64d75175425099f4fd9eb80f7e0ea294e0034fdef"
			);
		*/

		childDigests := con.Session(&gorm.Session{NewDB: true}).
			Distinct("digests.digest").
			Table("digests").
			Joins("INNER JOIN digests digests_original ON digests.data_id = digests_original.data_id").
			Where("digests_original.digest IN ?", digests)

		return con.Joins("INNER JOIN sources ON sources.bottle_id = bottles.id").
			Where("bottle_digest IN (?)", childDigests)
	}
}

// DeprecatedBy is a scope that will query to digests that are deprecated by the provided digest.
func DeprecatedBy(dgst digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		deprecatedDigests := con.Session(&gorm.Session{NewDB: true}).
			Select("deprecates.deprecated_bottle_digest").
			Table("bottles").
			Joins("INNER JOIN digests ON digests.data_id = bottles.data_id").
			Joins("INNER JOIN deprecates ON deprecates.bottle_id = bottles.id").
			Where("digests.digest = ?", dgst)

		return con.Joins("INNER JOIN digests ON digests.data_id = bottles.data_id").
			Where("digests.digest IN (?)", deprecatedDigests)
	}
}

// DeprecatesThis is a scope that will query to digests that deprecates the provided digest.
func DeprecatesThis(dgst digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		deprecatesDigests := con.Session(&gorm.Session{NewDB: true}).
			Select("digests.digest").
			Table("digests").
			Joins("INNER JOIN bottles ON bottles.data_id = digests.data_id").
			Joins("INNER JOIN deprecates ON deprecates.bottle_id = bottles.id").
			Where("deprecates.deprecated_bottle_digest = ?", dgst)

		return con.Joins("INNER JOIN digests ON digests.data_id = bottles.data_id").
			Where("digests.digest IN (?)", deprecatesDigests)
	}
}

// ExcludeDeprecated is a scope that will prevent bottles from being included in the query if they are deprecated by another bottle.
func ExcludeDeprecated() func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		deprecatedDigests := con.Session(&gorm.Session{NewDB: true}).
			Select("deprecates.deprecated_bottle_digest").
			Table("deprecates")

		return con.Joins("INNER JOIN digests deprecated_digests ON deprecated_digests.data_id = bottles.data_id").
			Where("deprecated_digests.digest NOT IN (?)", deprecatedDigests)
	}
}

// FilterByParts is a scope that will show bottles that have all parts with provided digests.
func FilterByParts(partDigests []digest.Digest) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		partFilteredBottleQuery := con.Joins("INNER JOIN digests ON digests.data_id = bottles.data_id")
		for _, dgst := range partDigests {
			bottleDigestsWithParts := con.Session(&gorm.Session{NewDB: true}).
				Select("digests.digest").
				Table("digests").
				Joins("INNER JOIN bottles ON bottles.data_id = digests.data_id").
				Joins("INNER JOIN parts ON parts.bottle_id = bottles.id").
				Where("parts.digest = ?", dgst)

			partFilteredBottleQuery = partFilteredBottleQuery.Where("digests.digest IN (?)", bottleDigestsWithParts)
		}

		return partFilteredBottleQuery
	}
}

// FilterByMetric is a scope that will show bottles that have metrics matching the filter
// filters may take the form of "MetricName>Value", "MetricName<Value' or just "MetricName".
func FilterByMetric(metricFilters []string) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		metricFilteredBottleQuery := con.Joins("INNER JOIN digests ON digests.data_id = bottles.data_id")
		for _, metric := range metricFilters {
			var filterValueString string
			var metricValueFilterQuery string

			metricName := strings.TrimSpace(metric)
			if strings.Contains(metric, ">") {
				metricName, filterValueString, _ = strings.Cut(metric, ">")
				metricValueFilterQuery = "metrics.value > ?"
			} else if strings.Contains(metric, "<") {
				metricName, filterValueString, _ = strings.Cut(metric, "<")
				metricValueFilterQuery = "metrics.value < ?"
			}

			bottleDigestsWithMetrics := con.Session(&gorm.Session{NewDB: true}).
				Select("digests.digest").
				Table("digests").
				Joins("INNER JOIN bottles ON bottles.data_id = digests.data_id").
				Joins("INNER JOIN metrics ON metrics.bottle_id = bottles.id").
				Where("metrics.name = ?", metricName)

			if len(filterValueString) > 0 {
				f, err := strconv.ParseFloat(filterValueString, 64)
				if err != nil {
					con.AddError(fmt.Errorf("metric filter value must be a float for %s: %s", metricName, filterValueString)) //nolint:errcheck
					return con
				}
				bottleDigestsWithMetrics = bottleDigestsWithMetrics.Where(metricValueFilterQuery, f)
			}

			metricFilteredBottleQuery = metricFilteredBottleQuery.Where("digests.digest IN (?)", bottleDigestsWithMetrics)
		}

		return metricFilteredBottleQuery
	}
}

// SortByMetric is a scope that will sort bottles by a given metric's value.
func SortByMetric(metricName string, ascending bool) func(db *gorm.DB) *gorm.DB {
	return func(con *gorm.DB) *gorm.DB {
		orderBy := "MAX(metrics.value)"
		if !ascending {
			orderBy = fmt.Sprintf("%s desc", orderBy)
		}
		return con.Joins("JOIN metrics ON metrics.bottle_id = bottles.id AND metrics.name = ?", metricName).Order(orderBy)
	}
}
