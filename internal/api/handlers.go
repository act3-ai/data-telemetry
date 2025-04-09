package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/schema"
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-telemetry/v3/internal/db"
	"github.com/act3-ai/data-telemetry/v3/internal/middleware"
	"github.com/act3-ai/data-telemetry/v3/pkg/types"
)

func handleGetBottlesFromMetric(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	// TODO make this DRY w.r.t. pkg/api/handlers.go

	type Params struct {
		Selectors  []string `schema:"selector"`
		Metric     string   `schema:"metric"`
		Limit      int      `schema:"limit"`
		Descending bool     `schema:"descending"`
	}

	params := Params{
		Limit: 7, // default value
	}
	if err := schema.NewDecoder().Decode(&params, r.URL.Query()); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid query parameters")
	}
	tx := con.Table("bottles").
		Joins("INNER JOIN metrics ON metrics.bottle_id = bottles.id").
		Where("metrics.name = ?", params.Metric).
		Scopes(db.IncludeDigests("bottles")).
		Group("metrics.description, metrics.value").
		Distinct("metrics.description", "metrics.value")

	// add selectors
	tx = tx.Scopes(db.FilterBySelectors(params.Selectors))

	// parse order, optional, valid values are ascending or descending
	order := "ASC"
	if params.Descending {
		order = "DESC"
	}
	tx = tx.Order("metrics.value " + order).Limit(params.Limit)

	type ResultEntry struct {
		db.Digested
		Description string
		Value       float64
	}

	var entries []ResultEntry
	if err := tx.Find(&entries).Error; err != nil {
		return err
	}

	if err := httputil.WriteJSON(w, map[string]any{"Results": entries}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

func handleGetLocation(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	bottleDigest, err := digest.Parse(r.URL.Query().Get("bottle_digest"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"bottle_digest\" parameter")
	}

	tx := con.
		Table("events").
		// Preload("Manifest").

		// Does not work in Postgres (but works in SQLite)
		// Select("events.repository", "events.auth_required", "Manifest.digest").
		// Joins("Bottle").
		// Joins("Manifest").
		// Where("Bottle.digest = ?", bottleDigest.String()).

		Distinct("events.repository", "events.auth_required", "events.manifest_digest AS digest").
		Joins("INNER JOIN bottles ON events.bottle_id = bottles.id").
		Joins("INNER JOIN manifests ON events.bottle_id = manifests.bottle_id").
		Scopes(db.FilterByDigest(bottleDigest, "bottles"))

	entries := []types.LocationResponse{}
	if err := tx.Find(&entries).Error; err != nil {
		return err
	}

	// TODO We could consider returning the bottle config (then the requester would know the content digests of each part)
	// TODO we could also consider returning all the known manifests so they know the layer digests of parts (there can be many layer digests for the same content digest)
	if err := httputil.WriteJSON(w, map[string]any{"Results": entries}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

// handleGetSignatures is an HTTP handler function that responds with an array of signature validation summaries in JSON.
// The signatures that are returned are selected with URL parameters:
//   - "bottle_digest" -> get data for signatures associated with the given bottle.
func handleGetSignatures(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	con := middleware.DatabaseFromContext(ctx)

	bottleDigest, err := digest.Parse(r.URL.Query().Get("bottle_digest"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"bottle_digest\" parameter")
	}

	tx := con.
		Table("signatures").
		Joins("INNER JOIN bottles ON signatures.bottle_id = bottles.id").
		Scopes(db.FilterByDigest(bottleDigest, "bottles"))

	entries := []db.Signature{}
	if err := tx.Find(&entries).Error; err != nil {
		return err
	}

	dtoEntries := []types.SignatureValidationSummary{}
	for _, e := range entries {
		annos := map[string]string{}
		for _, a := range e.Annotations {
			annos[a.Key] = a.Value
		}
		dtoEntries = append(dtoEntries, types.SignatureValidationSummary{
			SubjectManifest: e.ManifestDigest,
			SubjectBottleid: e.BottleDigest,
			Validated:       true, // if it exists in the database, it has been validated
			// TODO: use appropriate trust anchor
			Trusted:     e.Trusted(&db.DefaultTrustAnchor{}),
			Fingerprint: e.PublicKeyFingerPrint.String(),
			Annotations: annos,
		})
	}

	if err := httputil.WriteJSON(w, map[string]any{"Results": dtoEntries}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

// handleGetSigValid is an HTTP handler function that responds with an array of signature validation status data in JSON.
// The signatures that are returned are selected with URL parameters:
//   - "bottle_digest" -> get data for signatures associated with the given bottle.
//   - "key_fingerprint" -> get data for signatures made with the given key fingerprint.
//   - "trust_level" -> get data for signatures that match the given trust level ("trusted", or "validated" (default)).
func handleGetSigValid(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	con := middleware.DatabaseFromContext(ctx)

	bottleDigest, err := digest.Parse(r.URL.Query().Get("bottle_digest"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"bottle_digest\" parameter")
	}
	keyFP, err := digest.Parse(r.URL.Query().Get("key_fingerprint"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"key_fingerprint\" parameter")
	}
	trustlevel := r.URL.Query().Get("trust_level")
	if trustlevel == "" {
		trustlevel = "validated"
	}

	tx := con.
		Table("signatures").
		Joins("INNER JOIN bottles ON signatures.bottle_id = bottles.id").
		Scopes(db.FilterByDigest(bottleDigest, "bottles")).
		Scopes(db.FilterByPublicKeyFP(keyFP)).
		Scopes(db.FilterByTrustLevel(trustlevel))

	entries := []db.Signature{}
	if err := tx.Find(&entries).Error; err != nil {
		return err
	}

	dtoEntries := []types.SignatureValid{}
	for _, e := range entries {
		dtoEntries = append(dtoEntries, types.SignatureValid{
			BottleID:  e.BottleDigest,
			KeyFp:     e.PublicKeyFingerPrint.String(),
			Validated: true, // if it exists in the database, it has been validated
		})
	}
	if err := httputil.WriteJSON(w, map[string]any{"Results": dtoEntries}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

type listResultEntry struct {
	db.Base
	db.Digested
}

// We only want some fields in the result so we select them here.
func (e *listResultEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(&types.ListResultEntry{
		Digests:   e.Digests,
		CreatedAt: e.CreatedAt,
		Data:      e.Data.RawData,
	})
}

func genericListData(table string) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		// log := logger.FromContext(ctx)
		con := middleware.DatabaseFromContext(ctx)

		since, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("since"))
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"since\" parameter")
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"limit\" parameter")
		}

		tx := con.Table(table).
			Preload("Data").
			Select(table+".created_at", table+".data_id").
			Where(table+".created_at > ?", since).
			Order(table + ".created_at ASC").Limit(limit).
			Scopes(db.IncludeDigests(table))

		// What digests reference a object?  We can pull all of them that point to the same piece of data
		// but (since an artifact can be referenced as sha256 in one bottle and sha512 in another bottle,
		// we need both references/aliases).  The safe thing it to return all digests.
		// But should we track the aliases exactly?  Probably not worth it.  We just return all references since that is what would create a perfect mirror.  It is settled then.  Mirror everything.

		var entries []listResultEntry
		if err := tx.Find(&entries).Error; err != nil {
			return err
		}

		return httputil.WriteJSON(w, map[string]any{"Results": entries})
	})
}

func genericGetData(table, mediaType string) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		// log := logger.FromContext(ctx)
		con := middleware.DatabaseFromContext(ctx)

		dgst, err := digest.Parse(r.URL.Query().Get("digest"))
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid \"digest-alg\" parameter")
		}

		base := db.Base{}
		tx := con.Table(table).
			Preload("Data").
			// Select("created_at", "data").
			Scopes(db.FilterByDigest(dgst, table))
		if err := tx.First(&base).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return httputil.NewHTTPError(err, http.StatusNotFound, "Data in "+table+" not found")
			}
			return err
		}

		w.Header().Add(types.HeaderContentDigest, dgst.String())
		w.Header().Set(httputil.HeaderCreationDate, base.CreatedAt.Format(time.RFC3339Nano))
		httputil.AllowCaching(w.Header())
		w.Header().Set("Content-Type", mediaType)
		switch r.Method {
		case http.MethodGet:
			_, err := w.Write(base.Data.RawData)
			if err != nil {
				return fmt.Errorf("getting data: %w", err)
			}
			return nil
		case http.MethodHead:
			return nil
		default:
			return httputil.NewHTTPError(fmt.Errorf("method %s is not allowed", r.Method), http.StatusMethodNotAllowed, "Only GET and HEAD are allowed on this endpoint.")
		}
	})
}

func parseDataPutParams(data []byte, r *http.Request) (*digest.Digest, error) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var serverDigest digest.Digest
	if clientDigestStr := r.URL.Query().Get("digest"); clientDigestStr != "" {
		clientDigest, err := digest.Parse(clientDigestStr)
		if err != nil {
			return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid digest parameter")
		}

		// digest the data with the same crypto hash algorithm
		alg := clientDigest.Algorithm()
		if !alg.Available() {
			return nil, httputil.NewHTTPError(nil, http.StatusBadRequest, "Unknown digest algorithm")
		}
		serverDigest = alg.FromBytes(data)
		if serverDigest != clientDigest {
			log.InfoContext(ctx, "Digests for content and specified in query param do not match",
				"providedDigest", clientDigest, "calculatedDigest", serverDigest)
			return nil, httputil.NewHTTPError(nil, http.StatusConflict, "Digests for content and specified in query param do not match")
		}
		return &serverDigest, nil
	}

	if algStr := r.URL.Query().Get("digest-alg"); algStr != "" {
		alg := digest.Algorithm(algStr)
		if !alg.Available() {
			return nil, httputil.NewHTTPError(nil, http.StatusBadRequest, "Unknown digest algorithm")
		}

		serverDigest = alg.FromBytes(data)
		return &serverDigest, nil
	}

	// otherwise use the default hash function
	serverDigest = digest.FromBytes(data)
	return &serverDigest, nil
}

func genericPutData(processor db.Processor) http.Handler {
	return httputil.RootHandler(func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		// log := logger.FromContext(ctx)
		con := middleware.DatabaseFromContext(ctx)

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusBadRequest, "Unable to read the body")
		}

		// TODO support empty (Content-Length: 0) bodies.  Just pull the data with the "digest" from the params.
		// We would need a HEAD (and GET for completeness) for data (not just blobs)
		dgst, err := parseDataPutParams(data, r)
		if err != nil {
			return err
		}
		w.Header().Add(types.HeaderContentDigest, dgst.String())

		// compute the CanonicalDigest if we do not already have it
		canonicalDigest := *dgst
		if dgst.Algorithm() != db.CanonicalDigestAlgorithm {
			canonicalDigest = db.CanonicalDigestAlgorithm.FromBytes(data)
		}

		// Step 1: Make sure the Data record exists
		// Step 2: Make sure the Digest record exists
		// Step 3: Make sure the Object (bottle, event, ..) record exists

		// Step 1
		tx := con.Where(db.Data{
			CanonicalDigest: canonicalDigest,
		}).Attrs(db.Data{
			RawData: data,
		})
		dataRecord := db.Data{}
		if err := tx.FirstOrCreate(&dataRecord).Error; err != nil {
			return err
		}

		// Step 2
		tx = con.Where(db.Digest{
			DataID: dataRecord.ID, // This is slightly redundant.  If a record exists with the digest then the data better be the same or we found a collision.
			Digest: *dgst,
		})
		digestRecord := db.Digest{}
		if err := tx.FirstOrCreate(&digestRecord).Error; err != nil {
			return err
		}

		// Step 3
		tx = con.Table(processor.PrimaryTable()).
			Where(db.Base{DataID: dataRecord.ID})
		base := db.Base{}
		if err := tx.First(&base).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				// a real error occurred
				return err
			}
		} else {
			// we found one, it already exists
			w.WriteHeader(http.StatusNoContent)

			if base.ProcessorVersion == processor.Version() {
				// short circuit
				return nil
			}
			// else we "reprocess" the object to make its processor version up to date (fallthrough)
		}

		base.ProcessorVersion = processor.Version()
		base.Data = dataRecord
		base.DataID = dataRecord.ID

		if err := processor.Process(con, base); err != nil {
			return err
		}

		/*
			username := r.Header.Get(middleware.HeaderUsername)
			// FIXME
			if evt := event.(*db.Event); username != evt.Username {
				log.Info("username mismatch", "header", username, "event", evt.Username)
				// return httputil.NewHTTPError(nil, htt.StatusUnauthorized, "Username in the event does not match the header")
			}
		*/

		w.WriteHeader(http.StatusCreated)
		return nil
	})
}

type searchResultEntry struct {
	db.Base
	db.Digested
}

// We only want some fields in the result so we select them here.
func (e *searchResultEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Digests []digest.Digest
		Bottle  json.RawMessage `json:",omitempty"`
	}{
		Digests: e.Digests,
		Bottle:  e.Data.RawData,
	})
}

func handleBottleSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	type Params struct {
		Selectors   []string        `schema:"selector"`
		Description string          `schema:"description"`
		Limit       int             `schema:"limit"`
		DigestOnly  bool            `schema:"digestOnly"`
		PartDigests []digest.Digest `schema:"partDigest"`
	}

	params := Params{
		DigestOnly: true,
		Limit:      7,
	}

	if err := schema.NewDecoder().Decode(&params, r.URL.Query()); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid query parameters")
	}
	log.InfoContext(ctx, "Parameters", "params", params)

	tx := con.Table("bottles").
		Scopes(
			db.RankByDescription(params.Description),
			db.FilterBySelectors(params.Selectors),
			db.IncludeDigests("bottles"),
			db.FilterByParts(params.PartDigests),
		).
		Distinct("bottles.data_id").
		Limit(params.Limit)

	if !params.DigestOnly {
		tx = tx.Preload("Data")
	}

	var entries []searchResultEntry
	if err := tx.Find(&entries).Error; err != nil {
		return err
	}

	if err := httputil.WriteJSON(w, map[string]any{"Results": entries}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}

func handleContentSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	type Params struct {
		ContentDigest digest.Digest `schema:"contentDigest"`
		Limit         int           `schema:"limit"`
	}

	params := Params{
		Limit: 7,
	}

	if err := schema.NewDecoder().Decode(&params, r.URL.Query()); err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid query parameters")
	}
	log.InfoContext(ctx, "Parameters", "params", params)

	tx := con.Table("layers").
		Joins("INNER JOIN manifests ON layers.manifest_id = manifests.id").
		Joins("INNER JOIN bottles ON manifests.bottle_id = bottles.id").
		Joins("INNER JOIN parts ON bottles.id = parts.bottle_id").
		Where("parts.digest = ?", params.ContentDigest).
		Distinct("layers.digest").
		Limit(params.Limit)

	var entries []digest.Digest
	if err := tx.Find(&entries).Error; err != nil {
		return err
	}

	if err := httputil.WriteJSON(w, map[string]any{"Results": entries}); err != nil {
		return fmt.Errorf("could not write JSON results: %w", err)
	}
	return nil
}
