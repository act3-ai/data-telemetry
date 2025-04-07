package client

import (
	"context"
	"time"

	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"github.com/act3-ai/data-telemetry/v3/pkg/types"
)

// Dummy will be used for logging.
type Dummy struct{}

// Initialize with our Client Interface.
var _ Client = &Dummy{}

// SendEvent will make a Dummy sendevent call.
func (dc *Dummy) SendEvent(ctx context.Context, alg digest.Algorithm, eventJSON, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// SendManifest will make a Dummy SendManifest call.
func (dc *Dummy) SendManifest(ctx context.Context, alg digest.Algorithm, manifestJSON, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// SendBottle will make a Dummy SendBottle call.
func (dc *Dummy) SendBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// SendSignature will make a Dummy SendBottle call.
func (dc *Dummy) SendSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte, getArtifactData GetArtifactDataFunc) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// PutEvent will make a Dummy PutEvent call.
func (dc *Dummy) PutEvent(ctx context.Context, alg digest.Algorithm, eventJSON []byte) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// PutManifest will make a Dummy PutManifest call.
func (dc *Dummy) PutManifest(ctx context.Context, alg digest.Algorithm, manifestJSON []byte) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// PutBottle will make a Dummy PutBottle call.
func (dc *Dummy) PutBottle(ctx context.Context, alg digest.Algorithm, bottleConfigJSON []byte) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// PutSignature will make a Dummy PutBottle call.
func (dc *Dummy) PutSignature(ctx context.Context, alg digest.Algorithm, signatureJSON []byte) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// PutBlob will make a Dummy PutBlob call.
func (dc *Dummy) PutBlob(ctx context.Context, alg digest.Algorithm, blob []byte) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// ListBlobs will make a Dummy ListBlobs call.
func (dc *Dummy) ListBlobs(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// ListBottles will make a Dummy ListBottles call.
func (dc *Dummy) ListBottles(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// ListManifests will make a Dummy ListManifests call.
func (dc *Dummy) ListManifests(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// ListEvents will make a Dummy ListEvents call.
func (dc *Dummy) ListEvents(ctx context.Context, since time.Time, limit int) ([]types.ListResultEntry, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// GetBlob will make a Dummy GetBlob call.
func (dc *Dummy) GetBlob(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// GetBottle will make a Dummy GetBottle call.
func (dc *Dummy) GetBottle(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// GetManifest will make a Dummy GetManifest call.
func (dc *Dummy) GetManifest(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// GetEvent will make a Dummy GetEvent call.
func (dc *Dummy) GetEvent(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// Upload will make a Dummy Upload call.
func (dc *Dummy) Upload(ctx context.Context, file string, skipInvalid bool) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// UploadAll will make a Dummy UploadAll call.
func (dc *Dummy) UploadAll(ctx context.Context, path string, skipInvalid bool) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// Download will make a Dummy Download call.
func (dc *Dummy) Download(ctx context.Context, since time.Time, fromlatest bool, batchSize int, file string) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// DownloadAll will make a Dummy DownloadAll call.
func (dc *Dummy) DownloadAll(ctx context.Context, since time.Time, fromlatest bool, batchSize int, path string) error {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil
}

// GetLocations will make a Dummy GetLocations call.
func (dc *Dummy) GetLocations(ctx context.Context, bottledigest digest.Digest) ([]types.LocationResponse, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// BottleSearch will make a Dummy BottleSearch call.
func (dc *Dummy) BottleSearch(ctx context.Context, selectors []string, description string, limit int, digestOnly bool) ([]types.SearchResult, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}

// GetBottlesFromMetric will make a Dummy GetBottlesFromMetric call.
func (dc *Dummy) GetBottlesFromMetric(ctx context.Context, selectors []string, metric string, limit int, desc bool) ([]byte, error) {
	log := logger.FromContext(ctx)
	log.InfoContext(ctx, "Dummy client called")
	return nil, ErrNotFound
}
