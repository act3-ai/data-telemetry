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

	"github.com/act3-ai/data-telemetry/v3/internal/db"
	"github.com/act3-ai/data-telemetry/v3/internal/middleware"
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
	Metrics              []string         `schema:"metric"` // Metrickey (string), comparitor (<, >), limitNumber (float)
	SortByMetric         string           `schema:"sort-by-metric"`
	MetricSortAscending  bool             `schema:"metric-sort-ascending"`
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

func getBottlesFromRequestParams(ctx context.Context, params *bottleRequestParams) (*[]bottleResultEntry, *httputil.HTTPError) {
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

	if len(params.Metrics) > 0 {
		tx = tx.Scopes(db.FilterByMetric(params.Metrics))
	}

	if len(params.SortByMetric) > 0 {
		tx = tx.Scopes(db.SortByMetric(params.SortByMetric, params.MetricSortAscending))
	}

	// TODO Distinct is not working because it include Digest
	tx = tx.Table("bottles").
		Preload("Authors", func(db *gorm.DB) *gorm.DB {
			return db.Order("authors.location")
		}).
		Preload("Labels", func(db *gorm.DB) *gorm.DB {
			return db.Order("labels.key")
		}).
		Preload("Metrics", func(db *gorm.DB) *gorm.DB {
			return db.Order("metrics.name")
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

		Select("bottles.id, bottles.description").
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

	var entries []bottleResultEntry
	if err := tx.Find(&entries).Error; err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Issue while retrieving bottle entries")
	}
	return &entries, nil
}

func (a *WebApp) handleBottleSearchIsValid(w http.ResponseWriter, r *http.Request) error {
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

func (a *WebApp) handleBottleSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	type values struct {
		Params  bottleRequestParams
		Entries []bottleResultEntry
		Errors  string
	}
	errorReply := func(err *httputil.HTTPError, params *bottleRequestParams) error {
		v := values{
			*params,
			[]bottleResultEntry{},
			err.Error(),
		}
		w.WriteHeader(err.StatusCode)
		return a.executeTemplateAsResponse(ctx, w, "error-message", v, "../")
	}

	endpoint := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	templateMap := map[string]string{
		"table": "bottle-table",
		"cards": "bottle-cards",
	}
	searchResultTemplate := templateMap[endpoint]
	if len(searchResultTemplate) == 0 {
		return errorReply(httputil.NewHTTPError(fmt.Errorf("endpoint not found in template map"), http.StatusBadRequest, "Invalid endpoint"), &bottleRequestParams{})
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

	w.Header().Add("HX-Trigger-After-Settle", "onNewBottleSearchResults")
	return a.executeTemplateAsResponse(ctx, w, searchResultTemplate, v, "../")
}

func (a *WebApp) handleMetricSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	type values struct {
		Params      bottleRequestParams
		MetricNames []string
		Errors      string
	}

	templateMap := map[string]string{
		"dropdown": "metric-dropdown",
	}
	templateName, requestParams, err := getTemplateNameAndRequestParams(r, templateMap)
	if err != nil {
		return a.basicErrorReply(ctx, w, err)
	}

	entries, err := getMetricNamesFromRequestParams(ctx, requestParams)
	if err != nil {
		return a.basicErrorReply(ctx, w, err)
	}

	v := values{
		*requestParams, *entries, "",
	}

	w.Header().Add("HX-Trigger-After-Settle", "onNewMetricSearchResults")
	return a.executeTemplateAsResponse(ctx, w, templateName, v, "../")
}

// Get all unique metric names from a bottle search.
func getMetricNamesFromRequestParams(ctx context.Context, params *bottleRequestParams) (*[]string, *httputil.HTTPError) {
	con := middleware.DatabaseFromContext(ctx)

	if err := params.validate(); err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle request params")
	}

	tx := getFilteredSearchQuery(con, params)

	// TODO Distinct is not working because it include Digest
	tx = tx.Table("bottles").
		Distinct("bottles.id")

	if !time.Time(params.CreatedBefore).IsZero() {
		tx = tx.Where("bottles.created_at <= ?", time.Time(params.CreatedBefore))
	}

	metricQuery := con.Session(&gorm.Session{NewDB: true}).
		Table("metrics").
		Distinct("metrics.name").
		Where("metrics.bottle_id IN (?)", tx)

		// Joins("INNER JOIN bottles ON metrics.bottle_id = bottles.id").
	// Only need one Digest per bottle here (so we can JOIN), but which digest will it choose (and can it change from query to query)?
	// Should we use the CanonicalDigest to lookup the bottle instead of picking an arbitrary digest?

	var metricNames []string
	if err := metricQuery.Find(&metricNames).Error; err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Issue while retrieving bottle entries")
	}
	return &metricNames, nil
}

func (a *WebApp) handleCommonLabelSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	templateMap := map[string]string{
		"list": "label-list",
	}
	templateName, requestParams, err := getTemplateNameAndRequestParams(r, templateMap)
	if err != nil {
		return a.basicErrorReply(ctx, w, err)
	}

	entries, err := getCommonLabelsFromRequestParams(ctx, requestParams)
	if err != nil {
		return a.basicErrorReply(ctx, w, err)
	}

	type values struct {
		Params          bottleRequestParams
		CommonLabelKeys []string
		Errors          string
	}

	v := values{
		*requestParams, *entries, "",
	}

	w.Header().Add("HX-Trigger-After-Settle", "onNewCommonLabelSearchResults")
	return a.executeTemplateAsResponse(ctx, w, templateName, v, "../")
}

// Get all unique label names from a bottle search.
func getCommonLabelsFromRequestParams(ctx context.Context, params *bottleRequestParams) (*[]string, *httputil.HTTPError) {
	con := middleware.DatabaseFromContext(ctx)

	if err := params.validate(); err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle request params")
	}

	tx := getFilteredSearchQuery(con, params)

	txCount := tx.Session(&gorm.Session{}).
		Table("bottles").Select("COUNT(DISTINCT bottles.id) AS c")

	// TODO Distinct is not working because it include Digest
	tx = tx.Table("bottles").
		Distinct("bottles.id")

	if !time.Time(params.CreatedBefore).IsZero() {
		tx = tx.Where("bottles.created_at <= ?", time.Time(params.CreatedBefore))
	}

	// TODO get common labels
	labelQuery := con.Session(&gorm.Session{NewDB: true}).
		Table("labels").
		Select("labels.key").
		Joins("INNER JOIN (?) AS b ON b.id = labels.bottle_id", tx).
		Joins("NATURAL LEFT JOIN (?) AS b_count", txCount).
		Group("labels.key").
		Having("COUNT(DISTINCT b.id) = MAX(b_count.c)")

	var labelKeys []string
	if err := labelQuery.Find(&labelKeys).Error; err != nil {
		return nil, httputil.NewHTTPError(err, http.StatusBadRequest, "Issue while retrieving bottle entries")
	}
	return &labelKeys, nil
}

func getFilteredSearchQuery(con *gorm.DB, params *bottleRequestParams) *gorm.DB {
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

	if len(params.Metrics) > 0 {
		tx = tx.Scopes(db.FilterByMetric(params.Metrics))
	}

	return tx
}

func getTemplateNameAndRequestParams(r *http.Request, templateMap map[string]string) (string, *bottleRequestParams, *httputil.HTTPError) {
	endpoint := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	searchResultTemplate := templateMap[endpoint]
	if len(searchResultTemplate) == 0 {
		return "", &bottleRequestParams{}, httputil.NewHTTPError(fmt.Errorf("endpoint not found in template map"), http.StatusBadRequest, "Invalid endpoint")
	}

	requestParams, err := newBottleRequestParamsFromURLQuery(r.URL.Query())
	if err != nil {
		return "", &bottleRequestParams{}, httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid query parameters")
	}

	// overriding limit if it's 0
	if requestParams.Limit == 0 {
		requestParams.Limit = 9
	}

	return searchResultTemplate, requestParams, nil
}

func (a *WebApp) basicErrorReply(ctx context.Context, w http.ResponseWriter, err *httputil.HTTPError) error {
	v := struct {
		Errors string
	}{
		err.Error(),
	}
	w.WriteHeader(err.StatusCode)
	return a.executeTemplateAsResponse(ctx, w, "error-message", v, "../")
}
