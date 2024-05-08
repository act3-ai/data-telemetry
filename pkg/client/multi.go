package client

import (
	"context"
	"errors"
	"time"

	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
)

// Implement the Client interface.
var _ Client = &MultiClient{}

// MultiClient has an array of Client Interfaces.
type MultiClient struct {
	clients []Client
}

// NewMultiClient creates a MultiClient from an array of Clients.
func NewMultiClient(clients []Client) *MultiClient {
	return &MultiClient{clients}
}

// NewMultiClientConfig will create a MultiClient using slices of config file.
func NewMultiClientConfig(locations []v1alpha1.Location) *MultiClient {
	singleClients := make([]Client, 0, len(locations))
	for _, location := range locations {
		sc, err := NewSingleClientFromConfig(location)
		if err != nil {
			continue
		}

		// append each return result to the collecting bucket
		singleClients = append(singleClients, sc)
	}
	return NewMultiClient(singleClients)
}

// SendEvent will send an event to the api using the MultiClient.
func (mc *MultiClient) SendEvent(ctx context.Context, alg digest.Algorithm, eventJSON, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.SendEvent(ctx, alg, eventJSON, manifestJSON, bottleConfigJSON, getArtifactData)
	})
	return errors.Join(errs...)
}

// SendManifest will send a manifest JSON to the api.
func (mc *MultiClient) SendManifest(ctx context.Context, alg digest.Algorithm, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.SendManifest(ctx, alg, manifestJSON, bottleConfigJSON, getArtifactData)
	})
	return errors.Join(errs...)
}

// SendBottle will send a bottle JSON to the api.
func (mc *MultiClient) SendBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.SendBottle(ctx, alg, bottleConfigJSON, getArtifactData)
	})
	return errors.Join(errs...)
}

// SendSignature will send an event to the api using the MultiClient.
func (mc *MultiClient) SendSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte, getArtifactData GetArtifactDataFunc) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.SendSignature(ctx, alg, signatureJSON, getArtifactData)
	})
	return errors.Join(errs...)
}

// PutEvent will make a put event request to the api.
func (mc *MultiClient) PutEvent(ctx context.Context, alg digest.Algorithm, eventJSON []byte) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.PutEvent(ctx, alg, eventJSON)
	})
	return errors.Join(errs...)
}

// PutManifest will make a put manifest request to the api.
func (mc *MultiClient) PutManifest(ctx context.Context, alg digest.Algorithm, manifestJSON []byte) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.PutManifest(ctx, alg, manifestJSON)
	})
	return errors.Join(errs...)
}

// PutBottle will make a put bottle request to the api.
func (mc *MultiClient) PutBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.PutBottle(ctx, alg, bottleConfigJSON)
	})
	return errors.Join(errs...)
}

// PutSignature will make a put signature request to the api.
func (mc *MultiClient) PutSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.PutSignature(ctx, alg, signatureJSON)
	})
	return errors.Join(errs...)
}

// PutBlob will make a put blob request to the api.
func (mc *MultiClient) PutBlob(ctx context.Context, alg digest.Algorithm, blob []byte) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.PutBlob(ctx, alg, blob)
	})
	return errors.Join(errs...)
}

// Upload will upload files to api.
func (mc *MultiClient) Upload(ctx context.Context, file string, skipInvalid bool) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.Upload(ctx, file, skipInvalid)
	})
	return errors.Join(errs...)
}

// UploadAll will call UploadAll function from requests.
func (mc *MultiClient) UploadAll(ctx context.Context, path string, skipInvalid bool) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.UploadAll(ctx, path, skipInvalid)
	})
	return errors.Join(errs...)
}

// Download will use our client to call Download function.
func (mc *MultiClient) Download(ctx context.Context, since time.Time, fromlatest bool, batchSize int, file string) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.Download(ctx, since, fromlatest, batchSize, file)
	})
	return errors.Join(errs...)
}

// DownloadAll will use our client to call DownloadAll request function.
func (mc *MultiClient) DownloadAll(ctx context.Context, since time.Time, fromlatest bool, batchSize int, path string) error {
	errs := parallelMap(mc.clients, func(client Client, _ int) error {
		return client.DownloadAll(ctx, since, fromlatest, batchSize, path)
	})
	return errors.Join(errs...)
}

// ListBlobs will make a get a list of blobs from a specific time.
func (mc *MultiClient) ListBlobs(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return genericGet(mc, func(client Client) ([]types.ListResultEntry, error) {
		return client.ListBlobs(ctx, since, limit)
	})
}

// ListBottles will make a get a bottles list from a specific time.
func (mc *MultiClient) ListBottles(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return genericGet(mc, func(client Client) ([]types.ListResultEntry, error) {
		return client.ListBottles(ctx, since, limit)
	})
}

// ListManifests will make a get a manifests list from a specific time.
func (mc *MultiClient) ListManifests(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return genericGet(mc, func(client Client) ([]types.ListResultEntry, error) {
		return client.ListManifests(ctx, since, limit)
	})
}

// ListEvents will make a get an events list from a specific time.
func (mc *MultiClient) ListEvents(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	return genericGet(mc, func(client Client) ([]types.ListResultEntry, error) {
		return client.ListEvents(ctx, since, limit)
	})
}

// GetBlob will get blob against each client.
func (mc *MultiClient) GetBlob(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return genericGet(mc, func(client Client) ([]byte, error) {
		return client.GetBlob(ctx, dgst)
	})
}

// GetBottle will make a get bottle request to the api with the digest.
func (mc *MultiClient) GetBottle(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return genericGet(mc, func(client Client) ([]byte, error) {
		return client.GetBottle(ctx, dgst)
	})
}

// GetManifest will make a get manifest request to the api with the digest.
func (mc *MultiClient) GetManifest(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return genericGet(mc, func(client Client) ([]byte, error) {
		return client.GetManifest(ctx, dgst)
	})
}

// GetEvent will make a get event request to the api with the digest.
func (mc *MultiClient) GetEvent(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	return genericGet(mc, func(client Client) ([]byte, error) {
		return client.GetEvent(ctx, dgst)
	})
}

// GetLocations will return the Location Response.
func (mc *MultiClient) GetLocations(ctx context.Context, bottledigest digest.Digest) ([]types.LocationResponse, error) {
	return genericGet(mc, func(client Client) ([]types.LocationResponse, error) {
		return client.GetLocations(ctx, bottledigest)
	})
}

// BottleSearch will return the Location Response.
func (mc *MultiClient) BottleSearch(ctx context.Context, selectors []string, description string, limit int, digestOnly bool) ([]types.SearchResult, error) {
	return genericGet(mc, func(client Client) ([]types.SearchResult, error) {
		return client.BottleSearch(ctx, selectors, description, limit, digestOnly)
	})
}

// GetBottlesFromMetric will return the bottles using metric.
func (mc *MultiClient) GetBottlesFromMetric(ctx context.Context, selectors []string, metric string, limit int, desc bool) ([]byte, error) {
	return genericGet(mc, func(client Client) ([]byte, error) {
		return client.GetBottlesFromMetric(ctx, selectors, metric, limit, desc)
	})
}
