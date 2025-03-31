package client

import (
	"context"
	"net/url"
	"path"
	"time"

	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/pkg/types"
)

// GetArtifactDataFunc returns the data for an artifact with the given digest.
type GetArtifactDataFunc func(dgst digest.Digest) ([]byte, error)

// Client Interface defines an interface for interracting with a telemetry server.
type Client interface {
	// Params :
	// eventJSON - contains details for an Event in JSON
	// manifestJSON - contains details of a Manifest in JSON
	// bottleConfigJSON - contains details of a bottle in JSON
	// getArtifactData - function to get artifact/blobs

	// TODO we artificially restrict the event, manifest, and bottle, and artifacts to be the same digest algorithm.  They should be allowed to be different. In fact, we should only use the provided alg for the top level item. There after the digest is dictated by the JSON itself (e.g., the manifest references the bottle config with a specific algorithm, and the layers/parts by a specific algorithm and it can be different for each, just for fun).

	// SendEvent sends an event to the Telemetry server using context, algorithm, eventJSON, manifestJSON, bottleConfigJSON, getArtifactData function
	SendEvent(ctx context.Context, alg digest.Algorithm, eventJSON, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error
	// SendManifest sends a manifest to the Telemetry server using context, algorithm, manifestJSON, bottleConfigJSON, getArtifactData function
	SendManifest(ctx context.Context, alg digest.Algorithm, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error
	// SendBottle sends a bottle JSON to the Telemetry Server using context, algorithm, bottleConfigJSON, getArtifactData function
	SendBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error
	// SendSignature sends a signature JSON to the Telemetry Server using context, algorithm, signatureJSON, getArtifactData function
	SendSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte, getArtifactData GetArtifactDataFunc) error
	// PutEvent makes a PUT request to the Telemetry Server on an Event using algorithm and eventJSON
	PutEvent(ctx context.Context, alg digest.Algorithm, eventJSON []byte) error
	// PutManifest makes a PUT request to the Telemetry Server on a Manifest using algorithm and manifestJSON
	PutManifest(ctx context.Context, alg digest.Algorithm, manifestJSON []byte) error
	// PutBottle makes a PUT request to the Telemetry Server on a Bottle using algorithm and bottleConfigJSON
	PutBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte) error
	// PutSignature makes a PUT request to the Telemetry Server on a Signature using algorithm and signatureJSON
	PutSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte) error
	// PutBlob makes a PUT request to the Telemetry Server on a Blob using algorithm and blob of JSON type
	PutBlob(ctx context.Context, alg digest.Algorithm, blob []byte) error

	// TODO Return the latest time as well????

	// ListBlobs returns Blobs Lists details of type ListResultEntry and an error based of specified time and limit, on successful GET request to the telemetry Server
	ListBlobs(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error)
	// ListBottles returns Bottles Lists details of type ListResultEntry and an error based of specified time and limit, on successful GET request to the telemetry Server
	ListBottles(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error)
	// ListManifests returns Manifests Lists details of type ListResultEntry and an error based of specified time and limit, on successful GET request to the telemetry Server
	ListManifests(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error)
	// ListEvents returns Events Lists details of type ListResultEntry and an error based of specified time and limit, on successful GET request to the telemetry Server
	ListEvents(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error)

	// GetBlob retrieves and returns a blob with its digest
	GetBlob(ctx context.Context, dgst digest.Digest) ([]byte, error)
	// GetBottle retrieves and returns a bottle with its digest
	GetBottle(ctx context.Context, dgst digest.Digest) ([]byte, error)
	// GetManifest retrieves and returns a Manifest with its digest
	GetManifest(ctx context.Context, dgst digest.Digest) ([]byte, error)
	// GetEvent retrieves and returns an Event with its digest
	GetEvent(ctx context.Context, dgst digest.Digest) ([]byte, error)

	// Upload will upload a file to server
	Upload(ctx context.Context, file string, skipInvalid bool) error
	// UploadAll will upload files in the directory given
	UploadAll(ctx context.Context, directory string, skipInvalid bool) error

	// TODO Return the latest time as well????

	// Download will download a specified file from specified time parameter
	Download(ctx context.Context, since time.Time, fromlatest bool, batchSize int, file string) error
	// DownloadAll will download all files in specified path from specified time parameter
	DownloadAll(ctx context.Context, since time.Time, fromlatest bool, batchSize int, directory string) error

	// GetLocations will use a bottle digest to return and retrieve the location of type LocationResponse for a bottle
	GetLocations(ctx context.Context, bottledigest digest.Digest) ([]types.LocationResponse, error)
	// BottleSearch will search for a Bottle with selectors, description and return result in SearchResult type format
	BottleSearch(ctx context.Context, selectors []string, description string, limit int, digestOnly bool) ([]types.SearchResult, error)
	// GetBottlesFromMetric will retrieve and return bottles with selectors and metric metric in a slice
	GetBottlesFromMetric(ctx context.Context, selectors []string, metric string, limit int, desc bool) ([]byte, error)
}

// BottleDetailURL returns the URL to use to view the bottle with the given bottleDigest in a browser.  u is the telemetry server base URL and dgst is the bottle config digest.
func BottleDetailURL(u url.URL, dgst digest.Digest) string {
	// modify u (we have a copy of it so this is OK)
	u.Path = path.Join(u.Path, "www", "bottle.html")
	qs := u.Query()
	qs["digest"] = []string{dgst.String()}
	u.RawQuery = qs.Encode()
	return u.String()
}
