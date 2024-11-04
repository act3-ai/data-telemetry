// Package cli implements the command line interface for the telemetry command
package cli

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/go-common/pkg/config"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/go-common/pkg/version"

	"gitlab.com/act3-ai/asce/data/telemetry/cmd/telemetry/cli/client"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/actions"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// NewTelemetryCmd create a new root command.
func NewTelemetryCmd(info version.Info) *cobra.Command {
	action := actions.NewTelemetry(info)

	// rootCmd represents the base command when called without any subcommands
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "ACE Data Bottle Telemetry Server",
		Long:  `Manages the discovery and advanced queries on metadata on bottles`,
	}

	defaultConfigLocations := config.DefaultConfigSearchPath("ace", "telemetry", "config.yaml")

	cmd.PersistentFlags().StringArrayVar(&action.ConfigFiles, "config",
		config.EnvPathOr("ACE_TELEMETRY_CONFIG", defaultConfigLocations),
		`server configuration file location (setable with env "ACE_TELEMETRY_CONFIG"). 
The first configuration file present is used.  Others are ignored.
`)
	// Add environment variable configuration overrides
	action.AddServerConfigOverride(ServerConfigurationOverrides)

	// add subcommands
	cmd.AddCommand(
		NewServeCmd(action),
		NewTemplateCmd(),
		client.NewClientCmd(action),
		NewConfigCmd(action),
		NewFilterCmd(),
	)

	return cmd
}

// ServerConfigurationOverrides applies environment variables to the configuration.
func ServerConfigurationOverrides(ctx context.Context, c *v1alpha2.ServerConfiguration) error {
	log := logger.FromContext(ctx)
	// Database overrides
	name := "ACE_TELEMETRY_DSN"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.DB.DSN = redact.SecretURL(value)
	}

	name = "ACE_TELEMETRY_DB_PASS"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.DB.Password = redact.Secret(value)
	}

	// WebApp overrides
	name = "ACE_TELEMETRY_JUPYTER"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.WebApp.JupyterExecutable = value
	}

	name = "ACE_TELEMETRY_SELECTORS"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.WebApp.DefaultBottleSelectors = strings.Split(value, "|")
	}

	name = "ACE_TELEMETRY_ASSETS"
	if value, exists := os.LookupEnv(name); exists {
		log.InfoContext(ctx, "Using environment variable", "name", name)
		c.WebApp.AssetDir = value
	}

	// Maybe issue a warning if any environment variable starting with ACE_TELEMETRY is not used.
	// os.Environ()

	return nil
}
