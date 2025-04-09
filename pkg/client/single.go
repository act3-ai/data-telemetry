package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/opencontainers/go-digest"

	"github.com/act3-ai/go-common/pkg/logger"

	"github.com/act3-ai/data-telemetry/v3/pkg/types"
)

// ensure Client interface implementation.
var _ Client = &Single{}

// Single will represent our client to make http requests.
type Single struct {
	// consider using https://pkg.go.dev/github.com/hashicorp/go-retryablehttp#section-readme
	client *http.Client

	// apiURL is the URL object that represents the URL to the without the "/api" path on it
	// TODO consider switching this to a string
	apiURL *url.URL

	// Token used to make authenticated http calls
	token string
}

// NewSingleClient creates a new client to connect to the given telemetry server.
func NewSingleClient(httpClient *http.Client, serverURL string, token string) (*Single, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL for single client: %w", err)
	}

	// TODO we could test the URL to determine that the server is responsive at this point

	return &Single{
		client: httpClient,
		apiURL: u.JoinPath("api"),
		token:  token,
	}, nil
}

// SendEvent will send an event JSON to the api.
func (sc *Single) SendEvent(ctx context.Context, alg digest.Algorithm, eventJSON, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	missing := &types.MissingDigestsError{}
	if err := sc.PutEvent(ctx, alg, eventJSON); errors.As(err, &missing) && len(missing.MissingDigests) == 1 {
		if err := sc.SendManifest(ctx, missing.MissingDigests[0].Algorithm(), manifestJSON, bottleConfigJSON, getArtifactData); err != nil {
			return fmt.Errorf("failed to send manifest to server %s: %w", sc.apiURL.String(), err)
		}
		log.InfoContext(ctx, "trying to push the event again")
		if err := sc.PutEvent(ctx, alg, eventJSON); err != nil {
			return fmt.Errorf("failed to re-push event to server %s: %w", sc.apiURL.String(), err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to push event to server %s: %w", sc.apiURL.String(), err)
	}
	log.InfoContext(ctx, "event push successful")
	return nil
}

// SendManifest will send a manifest JSON to the api.
func (sc *Single) SendManifest(ctx context.Context, alg digest.Algorithm, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	missing := &types.MissingDigestsError{}
	if err := sc.PutManifest(ctx, alg, manifestJSON); errors.As(err, &missing) && len(missing.MissingDigests) == 1 {
		if err := sc.SendBottle(ctx, missing.MissingDigests[0].Algorithm(), bottleConfigJSON, getArtifactData); err != nil {
			return err
		}
		log.InfoContext(ctx, "trying to push the manifest again")
		if err := sc.PutManifest(ctx, alg, manifestJSON); err != nil {
			return fmt.Errorf("failed to re-push manifest: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to push manifest: %w", err)
	}
	log.InfoContext(ctx, "manifest push successful")
	return nil
}

// SendBottle will send a bottle JSON to the api.
func (sc *Single) SendBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	missing := &types.MissingDigestsError{}
	if err := sc.PutBottle(ctx, alg, bottleConfigJSON); errors.As(err, &missing) {
		for _, dgst := range missing.MissingDigests {
			data, err := getArtifactData(dgst)
			if err != nil {
				return fmt.Errorf("failed to get artifact data with digest %s: %w", dgst.String(), err)
			}
			if err := sc.PutBlob(ctx, dgst.Algorithm(), data); err != nil {
				return fmt.Errorf("failed to push blob: %w", err)
			}
		}
		log.InfoContext(ctx, "trying to push the bottle again")
		if err := sc.PutBottle(ctx, alg, bottleConfigJSON); err != nil {
			return fmt.Errorf("failed to re-push bottle: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to push bottle: %w", err)
	}
	log.InfoContext(ctx, "bottle push successful")
	return nil
}

// SendSignature will send an event JSON to the api.
func (sc *Single) SendSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	missing := &types.MissingDigestsError{}
	if err := sc.PutSignature(ctx, alg, signatureJSON); errors.As(err, &missing) && len(missing.MissingDigests) == 1 {
		log.InfoContext(ctx, "trying to push the signature again")
		if err := sc.PutSignature(ctx, alg, signatureJSON); err != nil {
			return fmt.Errorf("failed to re-push signature to server %s: %w", sc.apiURL.String(), err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to push signature to server %s: %w", sc.apiURL.String(), err)
	}
	log.InfoContext(ctx, "signature push successful")
	return nil
}

// PutEvent will make a put event request to the api.
func (sc *Single) PutEvent(ctx context.Context, alg digest.Algorithm, eventJSON []byte) error {
	return doPutRequest(ctx, sc.client, sc.apiURL, "event", eventJSON, alg, WithBearerTokenAuth(sc.token))
}

// PutManifest will make a put manifest request to the api.
func (sc *Single) PutManifest(ctx context.Context, alg digest.Algorithm, manifestJSON []byte) error {
	return doPutRequest(ctx, sc.client, sc.apiURL, "manifest", manifestJSON, alg, WithBearerTokenAuth(sc.token))
}

// PutBottle will make a put bottle request to the api.
func (sc *Single) PutBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte) error {
	return doPutRequest(ctx, sc.client, sc.apiURL, "bottle", bottleConfigJSON, alg, WithBearerTokenAuth(sc.token))
}

// PutSignature will make a put event request to the api.
func (sc *Single) PutSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte) error {
	return doPutRequest(ctx, sc.client, sc.apiURL, "signature", signatureJSON, alg, WithBearerTokenAuth(sc.token))
}

// PutBlob will make a put blob request to the api.
func (sc *Single) PutBlob(ctx context.Context, alg digest.Algorithm, blob []byte) error {
	return doPutRequest(ctx, sc.client, sc.apiURL, "blob", blob, alg, WithBearerTokenAuth(sc.token))
}

// ListBlobs will make a get a blobs list using since.
func (sc *Single) ListBlobs(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return doListRequest(ctx, sc.client, sc.apiURL, "blob", since, limit, WithBearerTokenAuth(sc.token))
}

// ListBottles will make a get a bottles list using since.
func (sc *Single) ListBottles(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return doListRequest(ctx, sc.client, sc.apiURL, "bottle", since, limit, WithBearerTokenAuth(sc.token))
}

// ListManifests will make a get a manifests list using since.
func (sc *Single) ListManifests(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return doListRequest(ctx, sc.client, sc.apiURL, "manifest", since, limit, WithBearerTokenAuth(sc.token))
}

// ListEvents will make a get an events list using since.
func (sc *Single) ListEvents(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return doListRequest(ctx, sc.client, sc.apiURL, "event", since, limit, WithBearerTokenAuth(sc.token))
}

// GetBlob will make a get blob request to the api with the digest.
func (sc *Single) GetBlob(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return doGetRequest(ctx, sc.client, sc.apiURL, "blob", dgst, WithBearerTokenAuth(sc.token))
}

// GetBottle will make a get bottle request to the api with the digest.
func (sc *Single) GetBottle(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return doGetRequest(ctx, sc.client, sc.apiURL, "bottle", dgst, WithBearerTokenAuth(sc.token))
}

// GetManifest will make a get manifest request to the api with the digest.
func (sc *Single) GetManifest(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return doGetRequest(ctx, sc.client, sc.apiURL, "manifest", dgst, WithBearerTokenAuth(sc.token))
}

// GetEvent will make a get event request to the api with the digest.
func (sc *Single) GetEvent(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return doGetRequest(ctx, sc.client, sc.apiURL, "event", dgst, WithBearerTokenAuth(sc.token))
}

// Upload will upload files to api.
func (sc *Single) Upload(ctx context.Context, file string, skipInvalid bool) error {
	return Upload(ctx, sc.client, file, sc.apiURL, sc.token, skipInvalid)
}

// UploadAll will call UploadAll function from requests.
func (sc *Single) UploadAll(ctx context.Context, path string, skipInvalid bool) error {
	return UploadAll(ctx, sc.client, path, sc.apiURL, sc.token, skipInvalid)
}

// Download will use our client to call Download function.
func (sc *Single) Download(ctx context.Context, since time.Time, fromlatest bool, batchSize int, file string) error {
	return Download(ctx, sc.client, since, fromlatest, batchSize, file, sc.apiURL, sc.token)
}

// DownloadAll will use our client to call DownloadAll request function.
func (sc *Single) DownloadAll(ctx context.Context, since time.Time, fromlatest bool, batchSize int, path string) error {
	return DownloadAll(ctx, sc.client, since, fromlatest, batchSize, path, sc.apiURL, sc.token)
}

// GetLocations will return the Location Response.
func (sc *Single) GetLocations(ctx context.Context, bottledigest digest.Digest) ([]types.LocationResponse, error) {
	return GetLocations(ctx, sc.client, nil, sc.apiURL, bottledigest)
}

// BottleSearch will return the Location Response.
func (sc *Single) BottleSearch(ctx context.Context, selectors []string, description string, limit int, digestOnly bool) ([]types.SearchResult, error) {
	return BottleSearch(ctx, sc.client, nil, sc.apiURL, selectors, description, limit, digestOnly)
}

// GetBottlesFromMetric will return the bottles using metric.
func (sc *Single) GetBottlesFromMetric(ctx context.Context, selectors []string, metric string, limit int, desc bool) ([]byte, error) {
	return GetBottlesFromMetric(ctx, sc.client, nil, sc.apiURL, selectors, metric, limit, desc)
}
