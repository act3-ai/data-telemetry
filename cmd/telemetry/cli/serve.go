package cli

import (
	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/actions"
	"gitlab.com/act3-ai/asce/go-common/pkg/config"
)

// NewServeCmd creates a new "serve" subcommand.
func NewServeCmd(telemetryAction *actions.Telemetry) *cobra.Command {
	action := &actions.Serve{
		Telemetry: telemetryAction,
	}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long:  `This connects to the database and runs the REST API and web site`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&action.Listen, "listen", "l", config.EnvOr("ACE_TELEMETRY_LISTEN", "localhost:8100"),
		`Interface and port to listen on.
Use :8100 to listen all on interfaces on the standard port.`)

	return cmd
}
