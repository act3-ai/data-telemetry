package webapp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/middleware"
)

type bottleRequestParams struct {
	Bottle               digest.Digest    `schema:"bottle"`
	LabelSelectors       []string         `schema:"label-selector"`
	Author               string           `schema:"author"`
	Description          string           `schema:"description"`
	SignatureFingerprint digest.Digest    `schema:"signature-fingerprint"`
	SignatureAnnotations []string         `schema:"signature-annotation"`
	ParentsOf            digest.Digest    `schema:"parents-of"`
	ChildrenOf           digest.Digest    `schema:"children-of"`
	DeprecatedBy         digest.Digest    `schema:"deprecated-by"`
	Deprecates           digest.Digest    `schema:"deprecates"`
	ShowDeprecated       bool             `schema:"show-deprecated"`
	PartDigests          []digest.Digest  `schema:"part-digest"`
	Limit                int              `schema:"limit"`
	Page                 int              `schema:"page"`
	CreatedBefore        requestTimestamp `schema:"created-before"`
	BottleRepo           string           `schema:"bottle-repository"`
}

type requestTimestamp time.Time

func (rt *requestTimestamp) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse int from timestamp text (%s): %w", s, err)
	}
	*rt = requestTimestamp(time.UnixMilli(i))
	return nil
}

func (rt *requestTimestamp) MarshalText() (text []byte, err error) {
	return []byte(strconv.FormatInt(time.Time(*rt).UnixMilli(), 10)), nil
}

func (rt requestTimestamp) String() string {
	return strconv.FormatInt(time.Time(rt).UnixMilli(), 10)
}

func newBottleRequestParamsFromURLQuery(values url.Values) (*bottleRequestParams, error) {
	brp := bottleRequestParams{}
	err := brp.populateFromURLQuery(values)
	return &brp, err
}

func (p *bottleRequestParams) populateFromURLQuery(values url.Values) error {
	// trim whitespace from query parameters
	for i, k := range values {
		for j, v := range k {
			values[i][j] = strings.TrimSpace(v)
		}
	}

	if err := schema.NewDecoder().Decode(p, values); err != nil {
		return fmt.Errorf("could not decode values into params: %w", err)
	}

	return p.validate()
}

// returns an error if any fields are invalid.
func (p *bottleRequestParams) validate() error {
	var multiError error

	validateDigest := func(dgst *digest.Digest, fieldName string) {
		if len(dgst.String()) > 0 {
			err := dgst.Validate()
			if err != nil {
				multiError = errors.Join(multiError, fmt.Errorf("invalid search param \"%s\" (%s): %w", fieldName, dgst.String(), err))
			}
		}
	}

	validateDigest(&p.Bottle, "bottle")
	validateDigest(&p.SignatureFingerprint, "signature-fingerprint")
	validateDigest(&p.ParentsOf, "parents-of")
	validateDigest(&p.ChildrenOf, "children-of")
	validateDigest(&p.DeprecatedBy, "deprecated-by")
	validateDigest(&p.Deprecates, "deprecates")
	for _, part := range p.PartDigests {
		validateDigest(&part, "part-digest")
	}

	for _, labelSelector := range p.LabelSelectors {
		_, err := labels.Parse(labelSelector)
		if err != nil {
			multiError = errors.Join(multiError, fmt.Errorf("invalid search param \"label-selector\" (%s): %w", labelSelector, err))
		}
	}

	for _, signatureAnnotation := range p.SignatureAnnotations {
		sigAnnParts := strings.Split(signatureAnnotation, "=")
		if len(sigAnnParts) != 2 {
			multiError = errors.Join(multiError, fmt.Errorf("invalid search param \"signature-annotation\" (%s)", signatureAnnotation))
		}
	}

	return multiError
}

func getBottlesFromRequestParams(ctx context.Context, params *bottleRequestParams) (*[]resultEntry, *httputil.HTTPError) {
	con := middleware.DatabaseFromContext(ctx)

	if err := params.validate(); err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle request params")
	}

	tx := con

	// add selectors
	tx = tx.Scopes(db.FilterBySelectors(params.LabelSelectors))

	// description matching
	tx = tx.Scopes(db.RankByDescription(params.Description))

	// searching Author name or Author email
	tx = tx.Scopes(db.SearchByAuthor(params.Author))

	tx = tx.Scopes(db.SearchByRepository(params.BottleRepo))

	if len(params.SignatureFingerprint) > 0 {
		tx = tx.Scopes(db.WithSignature([]digest.Digest{params.SignatureFingerprint}))
	}

	// signature annotation matching
	tx = tx.Scopes(db.WithSignatureAnnotations(params.SignatureAnnotations))

	if len(params.ParentsOf) > 0 {
		tx = tx.Scopes(db.ParentsOf([]digest.Digest{params.ParentsOf}))
	}

	if len(params.ChildrenOf) > 0 {
		tx = tx.Scopes(db.ChildrenOf([]digest.Digest{params.ChildrenOf}))
	}

	if len(params.DeprecatedBy) > 0 {
		tx = tx.Scopes(db.DeprecatedBy(params.DeprecatedBy))

		// If we are using "deprecated by" we can assume they want to see deprecated bottles
		params.ShowDeprecated = true
	}

	if len(params.Deprecates) > 0 {
		tx = tx.Scopes(db.DeprecatesThis(params.Deprecates))
	}

	if !params.ShowDeprecated {
		tx = tx.Scopes(db.ExcludeDeprecated())
	}

	if len(params.PartDigests) > 0 {
		tx = tx.Scopes(db.FilterByParts(params.PartDigests))
	}

	// TODO Distinct is not working because it include Digest
	tx = tx.Table("bottles").
		Preload("Authors", func(db *gorm.DB) *gorm.DB {
			return db.Order("authors.location")
		}).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("labels.key")
		}).
		Preload("Signatures", func(db *gorm.DB) *gorm.DB {
			return db.Order("signatures.updated_at")
		}).
		Preload("Metrics", func(db *gorm.DB) *gorm.DB {
			return db.Order("metrics.name")
		}).

		// Preload("Blobs", func(tx *gorm.DB) *gorm.DB {
		// 	return tx.Select("id", "digest")
		// }).

		Distinct("bottles.id", "bottles.description").
		Scopes(db.IncludeDigests("bottles"), db.IncludeIsDeprecated(), db.IncludeNumPulls()).
		Limit(params.Limit).
		Offset(params.Page * params.Limit)

	if !time.Time(params.CreatedBefore).IsZero() {
		tx = tx.Where("bottles.created_at <= ?", time.Time(params.CreatedBefore))
	}

	tx.Statement.Selects = append(
		tx.Statement.Selects,
		"COUNT(*) OVER () AS total_count",
	)

	// Only need one Digest per bottle here (so we can JOIN), but which digest will it choose (and can it change from query to query)?
	// Should we use the CanonicalDigest to lookup the bottle instead of picking an arbitrary digest?

	var entries []resultEntry
	if err := tx.Find(&entries).Error; err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Issue while retrieving bottle entries")
	}
	return &entries, nil
}

func (a *WebApp) handleBottleSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	type values struct {
		Params bottleRequestParams
		Errors string
	}

	errorReply := func(err *httputil.HTTPError, params *bottleRequestParams) error {
		v := values{
			*params,
			err.Error(),
		}
		w.WriteHeader(err.StatusCode)
		return a.executeTemplateAsResponse(ctx, w, "bottle-search-bar", v, "../")
	}

	requestParams, err := newBottleRequestParamsFromURLQuery(r.URL.Query())
	if err != nil {
		return errorReply(httputil.NewHTTPError(err, http.StatusUnprocessableEntity, "Invalid query parameters"), requestParams)
	}

	// if the bottle param is included, redirect to the bottle detail page for that bottle
	if len(requestParams.Bottle.String()) >= 1 {
		w.Header().Add("HX-Redirect", "/www/bottle.html?digest="+requestParams.Bottle.String())
		w.WriteHeader(http.StatusOK)
		return nil
	}

	v := values{
		Params: *requestParams,
		Errors: "",
	}

	if httpErr := setCurrentURLParams(w, r, &v.Params); httpErr != nil {
		return errorReply(httpErr, requestParams)
	}

	w.Header().Add("HX-Trigger-After-Settle", "onValidSearch")
	return a.executeTemplateAsResponse(ctx, w, "bottle-search-bar", v, "../")
}

func (a *WebApp) handleBottleCards(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	type values struct {
		Params  bottleRequestParams
		Entries []resultEntry
		Errors  string
	}
	errorReply := func(err *httputil.HTTPError, params *bottleRequestParams) error {
		v := values{
			*params,
			[]resultEntry{},
			err.Error(),
		}
		w.WriteHeader(err.StatusCode)
		return a.executeTemplateAsResponse(ctx, w, "bottle-cards", v, "../")
	}

	requestParams, err := newBottleRequestParamsFromURLQuery(r.URL.Query())
	if err != nil {
		return errorReply(httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid query parameters"), requestParams)
	}

	// overriding limit if it's 0
	if requestParams.Limit == 0 {
		requestParams.Limit = 9
	}

	entries, httpErr := getBottlesFromRequestParams(ctx, requestParams)
	if httpErr != nil {
		return errorReply(httpErr, requestParams)
	}

	v := values{
		*requestParams, *entries, "",
	}

	return a.executeTemplateAsResponse(ctx, w, "bottle-cards", v, "../")
}

func removeUnsetURLParams(values *url.Values) url.Values {
	newValues := url.Values{}
	for k, vSlice := range *values {
		if values.Get(k) != "" {
			for _, v := range vSlice {
				newValues.Add(k, v)
			}
		}
	}
	return newValues
}
