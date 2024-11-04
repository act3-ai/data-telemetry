package actions

import (
	"context"

	client "gitlab.com/act3-ai/asce/data/telemetry/pkg/client/v2"
)

// Upload is the upload action.
type Upload struct {
	*Client

	All         bool
	SkipInvalid bool
}

// Run is the action method.
func (action *Upload) Run(ctx context.Context, path, telemetryServerURL string) error {
	clientConfig, err := action.GetClientConfig(ctx)
	if err != nil {
		return err
	}

	newconfig, err := matchURLConfig(telemetryServerURL, clientConfig)
	if err != nil {
		return err
	}

	c, err := client.NewSingleClient(authClientOrDefault(ctx, newconfig), telemetryServerURL, "")
	if err != nil {
		return err
	}

	if action.All {
		// return client.UploadAll(ctx, c, path, u, handler)
		return c.UploadAll(ctx, path, action.SkipInvalid)
	}
	// return c.Upload(ctx, c, path, u, handler)
	return c.Upload(ctx, path, action.SkipInvalid)
}
