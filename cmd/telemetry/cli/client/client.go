// Package client contains the CLI for clients interacting with the Telemetry server
package client

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-telemetry/v3/internal/actions"
	"github.com/act3-ai/go-common/pkg/config"
)

// NewClientCmd creates a new "upload" command.
func NewClientCmd(telemetryAction *actions.Telemetry) *cobra.Command {
	action := &actions.Client{
		Telemetry: telemetryAction,
	}

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Client commands for interacting with a telemetry server at a low level.",
		Long:  `Typical users will use ace-dt to interact with the REST API of the telemetry server.  The subcommands here offer a lower-level API for administrators and developers.`,
	}

	defaultConfigLocations := config.DefaultConfigSearchPath("ace", "telemetry", "client-config.yaml")

	cmd.PersistentFlags().StringArrayVar(&action.ConfigFiles, "client-config",
		config.EnvPathOr("ACE_TELEMETRY_CLIENT_CONFIG", defaultConfigLocations),
		`client configuration file location (setable with env "ACE_TELEMETRY_CLIENT_CONFIG")
May specify multiple files separated by ":".  
The first configuration file present is used.  Others are ignored.
`)

	cmd.AddCommand(
		NewUploadCmd(action),
		NewDownloadCmd(action),
		NewClientConfigCmd(action),
	)
	return cmd
}
