package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"

	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/db"
)

// API implements the REST API.
type API struct{}

// Initialize setup the API handlers.
func (a *API) Initialize(router chi.Router, scheme *runtime.Scheme) {
	a.addBasicRoutes(router, "blob", "application/octet-stream", &db.BlobProcessor{})
	a.addBasicRoutes(router, "bottle", mediatype.MediaTypeBottleConfig, db.NewBottleProcessor(scheme))
	a.addBasicRoutes(router, "manifest", ocispec.MediaTypeImageManifest, &db.ManifestProcessor{})
	a.addBasicRoutes(router, "event", "application/json", &db.EventProcessor{})
	a.addBasicRoutes(router, "signature", "application/json", &db.SignatureProcessor{})
	// Handler(httputils.SignatureVerifyMiddleware(httputil.RootHandler(handlePutEvent)))

	// Bottle search
	router.Get("/search", httputil.RootHandler(handleBottleSearch).ServeHTTP)

	// Content search
	router.Get("/content", httputil.RootHandler(handleContentSearch).ServeHTTP)

	// Bottle metrics
	router.Get("/metric", httputil.RootHandler(handleGetBottlesFromMetric).ServeHTTP)

	router.Get("/location", httputil.RootHandler(handleGetLocation).ServeHTTP)

	// Bottle Signatures
	router.Get("/signatures", httputil.RootHandler(handleGetSignatures).ServeHTTP)
	router.Get("/signature/validate", httputil.RootHandler(handleGetSigValid).ServeHTTP)
}

func (a *API) addBasicRoutes(router chi.Router, itemType, contentType string, processor db.Processor) {
	// We need the maximum body size limit to be limited (~10MiB).
	// This should be done at the ingress level for all requests.
	// If we decide to do it here we can do that with the following
	// r.Body = http.MaxBytesReader(w, nopCloser{r.Body}, 10*1024*1024)

	path := "/" + itemType

	getData := genericGetData(itemType+"s", contentType).ServeHTTP
	router.Head(path, getData)

	router.Get(path, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("digest") {
			getData(w, r)
		} else {
			genericListData(itemType+"s").ServeHTTP(w, r)
		}
	})

	router.With(middleware.AllowContentType(contentType)).Put(path, genericPutData(processor).ServeHTTP)
}
