package api

import (
	"fmt"
	"net/http"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/act3-ai/bottle-schema/pkg/mediatype"
	"github.com/act3-ai/go-common/pkg/httputil"

	"github.com/act3-ai/data-telemetry/v3/internal/db"
)

// API implements the REST API.
type API struct{}

// Initialize setup the API handlers.
func (a *API) Initialize(serveMux *http.ServeMux, scheme *runtime.Scheme) {
	a.addBasicRoutes(serveMux, "blob", "application/octet-stream", &db.BlobProcessor{})
	a.addBasicRoutes(serveMux, "bottle", mediatype.MediaTypeBottleConfig, db.NewBottleProcessor(scheme))
	a.addBasicRoutes(serveMux, "manifest", ocispec.MediaTypeImageManifest, &db.ManifestProcessor{})
	a.addBasicRoutes(serveMux, "event", "application/json", &db.EventProcessor{})
	a.addBasicRoutes(serveMux, "signature", "application/json", &db.SignatureProcessor{})
	// Handler(httputils.SignatureVerifyMiddleware(httputil.RootHandler(handlePutEvent)))

	// Bottle search
	serveMux.Handle("GET /search", httputil.RootHandler(handleBottleSearch))

	// Content search
	serveMux.Handle("GET /content", httputil.RootHandler(handleContentSearch))

	// Bottle metrics
	serveMux.Handle("GET /metric", httputil.RootHandler(handleGetBottlesFromMetric))

	serveMux.Handle("GET /location", httputil.RootHandler(handleGetLocation))

	// Bottle Signatures
	serveMux.Handle("GET /signatures", httputil.RootHandler(handleGetSignatures))
	serveMux.Handle("GET /signature/validate", httputil.RootHandler(handleGetSigValid))
}

func (a *API) addBasicRoutes(serveMux *http.ServeMux, itemType, contentType string, processor db.Processor) {
	// We need the maximum body size limit to be limited (~10MiB).
	// This should be done at the ingress level for all requests.
	// If we decide to do it here we can do that with the following
	// r.Body = http.MaxBytesReader(w, nopCloser{r.Body}, 10*1024*1024)

	path := "/" + itemType

	getData := genericGetData(itemType+"s", contentType).ServeHTTP
	serveMux.HandleFunc(fmt.Sprintf("HEAD %s", path), getData)

	serveMux.HandleFunc(fmt.Sprintf("GET %s", path), func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("digest") {
			getData(w, r)
		} else {
			genericListData(itemType+"s").ServeHTTP(w, r)
		}
	})

	serveMux.Handle(fmt.Sprintf("PUT %s", path), httputil.AllowContentTypeMiddleware(genericPutData(processor), contentType))
}
