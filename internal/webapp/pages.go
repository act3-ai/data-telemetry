package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"

	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/data/schema/pkg/selectors"
	"gitlab.com/act3-ai/asce/data/schema/pkg/util"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	latest "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io/v1"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/middleware"
)

type bottleResultEntry struct {
	db.Bottle
	db.Digested
	IsDeprecated bool
	NumPulls     int
	TotalCount   int // number of bottles returned alongside this resultEntry
	MetricIdx    int `gorm:"-"`
}

func (a *WebApp) handleAbout(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	con := middleware.DatabaseFromContext(ctx)
	log := logger.FromContext(ctx)

	type TotalCount struct {
		EventsCount     int64
		ManifestsCount  int64
		BottlesCount    int64
		ArtifactsCount  int64
		SignaturesCount int64
		BlobDataBytes   int64
	}

	log.InfoContext(ctx, "Collecting counts")
	eventsCount := db.EventsCount(con)
	manifestsCount := db.ManifestsCount(con)
	bottlesCount := db.BottlesCount(con)
	artifactsCount := db.ArtifactsCount(con)
	signaturesCount := db.SignaturesCount(con)
	blobDataBytes := db.BlobDataBytes(con)

	values := TotalCount{eventsCount, manifestsCount, bottlesCount, artifactsCount, signaturesCount, blobDataBytes}

	return a.executeTemplateAsResponse(ctx, w, "documentation.html", values, "../")
}

func (a *WebApp) getPageHandler(page string) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		params := bottleRequestParams{
			Limit: 9,
		} // set default values here

		type values struct {
			Params bottleRequestParams
			Errors string
		}

		errorReply := func(err *httputil.HTTPError) error {
			v := values{
				params,
				err.Error(),
			}
			w.WriteHeader(err.StatusCode)
			// TODO better error handling for this component
			return a.executeTemplateAsResponse(ctx, w, page, v, "../")
		}

		if err := params.populateFromURLQuery(r.URL.Query()); err != nil {
			return errorReply(httputil.NewHTTPError(err, http.StatusUnprocessableEntity, "Invalid query parameters"))
		}

		// overriding limit if it's 0
		if params.Limit == 0 {
			params.Limit = 9
		}

		if time.Time(params.CreatedBefore).IsZero() {
			params.CreatedBefore = requestTimestamp(time.Now())
		}

		v := values{
			params, "",
		}

		return a.executeTemplateAsResponse(ctx, w, page, v, "../")
	}
}

func (a *WebApp) handleBottle(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	con := middleware.DatabaseFromContext(ctx)
	log := logger.FromContext(ctx)

	type Params struct {
		Digest            digest.Digest `schema:"digest"`
		PartSelectors     []string      `schema:"partSelector"`
		NumGenAncestors   uint          `schema:"numGenAncestors"`
		NumGenDescendants uint          `schema:"numGenDescendants"`
	}

	params := Params{
		NumGenAncestors:   1,
		NumGenDescendants: 1,
	}
	if err := schema.NewDecoder().Decode(&params, r.URL.Query()); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
	}

	log.DebugContext(ctx, "Bottle request", "params", params)

	// bottle ID (digest)
	if err := params.Digest.Validate(); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"digest\" parameter: "+err.Error())
	}

	// we support multiple part selectors
	sel, err := selectors.Parse(params.PartSelectors)
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"partSelector\" parameter: "+err.Error())
	}

	bottle := db.BottleRelative{}
	// Preload Tables in the Bottle struct (argument is a field name in bottle)
	tx := con.Table("bottles").
		Select("bottles.*").
		// Preload(clause.Associations). // preload all fields
		// If we Preload("Blobs") then we would want to exclude some fields with this trick
		// Preload("Blobs", func(tx *gorm.DB) *gorm.DB {
		// 	return tx.Select("id", "digest")
		// }).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("labels.key")
		}).
		Preload("Annotations", func(db *gorm.DB) *gorm.DB {
			return db.Order("annotations.key")
		}).
		Preload("Authors", func(db *gorm.DB) *gorm.DB {
			return db.Order("authors.location")
		}).
		Preload("Metrics", func(db *gorm.DB) *gorm.DB {
			return db.Order("metrics.name")
		}).
		Preload("Data").
		Preload("PublicArtifacts", func(db *gorm.DB) *gorm.DB {
			return db.Order("public_artifacts.location") // We could order by name, path, or index.  Not sure which is best.
		}).
		Preload("Parts", func(db *gorm.DB) *gorm.DB {
			return db.Order("parts.name")
		}).
		Preload("Signatures", func(db *gorm.DB) *gorm.DB {
			return db.Order("signatures.updated_at")
		}).
		Preload("Sources").
		Scopes(db.FilterByDigest(params.Digest, "bottles"), db.IncludeDigests("bottles"))

	if err := tx.First(&bottle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return httputil.NewHTTPError(err, http.StatusNotFound, "Bottle not found")
		}
		return err
	}

	// Get the manifests associated with this bottle
	type Manifestation struct {
		Repository     string
		AuthRequired   bool
		Digest         string
		LastAccessedAt time.Time
	}

	manifestations := []Manifestation{}

	subquery := con.
		Table("events").
		Select(
			"events.repository",
			"events.auth_required",
			"events.manifest_digest AS digest",
			"events.timestamp AS last_accessed_at",
			"RANK() OVER (PARTITION BY events.repository, events.manifest_digest ORDER BY events.timestamp DESC) rank",
		).
		Joins("INNER JOIN bottles ON events.bottle_id = bottles.id").
		Joins("INNER JOIN manifests ON events.bottle_id = manifests.id").
		Scopes(db.FilterByDigest(params.Digest, "bottles"))

	// Does not work in Postgres (but works in SQLite)
	// Model(&db.Event{}).
	// Joins("Bottle").
	// Joins("Manifest").
	// Scopes(db.FilterByDigest(bottleDigest, "Bottle"))

	tx = con.
		Table("(?) as m", subquery).
		Select("m.repository", "m.auth_required", "m.digest", "m.last_accessed_at").
		Where("m.rank = 1").
		Order("m.last_accessed_at DESC")

	if err := tx.Find(&manifestations).Error; err != nil {
		return err
	}

	addPartsToBottleWithSelector(&bottle, sel)
	totalSize := getBottleTotalSize(&bottle)

	// Convert the bottle JSON to pretty format (nice JSON)
	var raw map[string]any
	if err := json.Unmarshal(bottle.Data.RawData, &raw); err != nil {
		return fmt.Errorf("un-marshalling error on RawData: %w", err)
	}
	bottlePrettyJSON, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling error on RawData: %w", err)
	}
	bottlePrettyYAML, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshalling error on RawData: %w", err)
	}

	// Get the viewers for the bottle
	viewerSpecs := a.GetViewerSpecList(bottle.Bottle)
	viewers := a.FindViewers(viewerSpecs.FilterAndSort(mediatype.MediaTypeBottle), params.Digest, params.PartSelectors, nil)
	artifactViewers := a.GetArtifactViewers(viewerSpecs, params.Digest, params.PartSelectors, bottle.PublicArtifacts)

	// Get the bottles deprecated by this bottle
	deprecatesBottleDigests, err := db.FindDeprecatedBy(con, params.Digest)
	if err != nil {
		return err
	}

	// Get the bottle that deprecate this bottle
	deprecatedByBottleDigests, err := db.FindDeprecates(con, params.Digest)
	if err != nil {
		return err
	}

	// Get total bottle pulls
	totalBottlePulls := db.BottlePulls(con, params.Digest)

	// Get pull stats for this bottle
	bottlePulls, err := db.UserPulls(con, params.Digest, 7)
	if err != nil {
		return err
	}

	lineageGraphHTML, err := GetAncestryGraphHTML(ctx, con, &bottle, params.NumGenAncestors, params.NumGenDescendants, a.templates)
	if err != nil {
		return err
	}

	// Add signature annotations
	signatures, err := db.GetSignaturesWithAnnotations(ctx, con, &bottle.Signatures)
	if err != nil {
		return err
	}
	// Reformat signatures to have trust value
	type signatureWithTrust struct {
		Signature db.Signature
		Trusted   bool
	}
	swt := []signatureWithTrust{}
	for _, s := range *signatures {
		swt = append(swt, signatureWithTrust{
			Signature: s,
			// TODO: use appropriate trust anchor
			Trusted: s.Trusted(&db.DefaultTrustAnchor{}),
		})
	}

	values := struct {
		Params
		TotalSize          uint64
		Bottle             *db.BottleRelative
		Manifestations     []Manifestation
		Digests            []digest.Digest // aliases
		PrettyJSON         []byte
		PrettyYAML         []byte
		DeprecatedBy       []digest.Digest
		Deprecates         []digest.Digest
		Viewers            []ViewerLink
		Signatures         []signatureWithTrust
		ArtifactViewers    map[string][]ViewerLink
		TotalBottlePulls   int64
		BottlePullUserNums map[string]int
		LatestAPIVersion   string
		LineageGraphHTML   template.HTML
	}{
		params,
		totalSize, &bottle, manifestations, bottle.Digests, bottlePrettyJSON, bottlePrettyYAML, deprecatedByBottleDigests, deprecatesBottleDigests, viewers, swt, artifactViewers, totalBottlePulls, bottlePulls, latest.GroupVersion.Identifier(), lineageGraphHTML,
	}

	return a.executeTemplateAsResponse(ctx, w, "bottle.html", values, "../")
}

func addPartsToBottleWithSelector(bottle *db.BottleRelative, sel selectors.LabelSelectorSet) {
	// filter parts by selector
	parts := make([]db.Part, 0, len(bottle.Parts))
	artifacts := make([]db.PublicArtifact, 0, len(bottle.PublicArtifacts))
	for _, part := range bottle.Parts {
		if !sel.Matches(labels.Set(part.Labels)) {
			continue
		}
		parts = append(parts, part)

		// prune artifacts
		for _, artifact := range bottle.PublicArtifacts {
			if util.IsPathPrefix(artifact.Path, part.Name) {
				artifacts = append(artifacts, artifact)
			}
		}
	}
	bottle.Parts = parts
	bottle.PublicArtifacts = artifacts
}

func getBottleTotalSize(bottle *db.BottleRelative) uint64 {
	totalSize := uint64(0)
	for _, part := range bottle.Parts {
		totalSize += part.Size
	}
	return totalSize
}

func (a *WebApp) handleSimilarBottles(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	u := *r.URL
	qs := u.Query()
	requirements, exists := qs["requirement"]
	if exists {
		qs.Del("requirement")
		qs.Set("label-selector", strings.Join(requirements, ","))
		u.RawQuery = qs.Encode()
	}

	page := "catalog.html"
	if r.URL.Query().Has("metric") {
		page = "leaderboard.html"
	}
	u.Scheme = ""
	u.User = nil
	u.Path = page

	log.InfoContext(ctx, "Redirecting", "location", u.String())
	w.Header().Set("Location", u.String())
	w.WriteHeader(http.StatusFound)
	return nil
}
