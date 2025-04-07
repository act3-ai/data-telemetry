package actions

import (
	"context"
	"fmt"
	"time"

	client "github.com/act3-ai/data-telemetry/v3/pkg/client"
)

// Download is the action for the download operations.
type Download struct {
	*Client

	Since      string
	BatchSize  int
	All        bool
	FromLatest bool
}

// Run is the action method.
func (action *Download) Run(ctx context.Context, path, telemetryServerURL string) error {
	clientConfig, err := action.GetClientConfig(ctx)
	if err != nil {
		return err
	}

	var since time.Time
	if action.Since != "" {
		s, err := time.Parse(time.RFC3339Nano, action.Since)
		if err != nil {
			return fmt.Errorf("parsing \"since\" date: %w", err)
		}
		since = s
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
		return c.DownloadAll(ctx, since, action.FromLatest, action.BatchSize, path)
	}
	return c.Download(ctx, since, action.FromLatest, action.BatchSize, path)
}
