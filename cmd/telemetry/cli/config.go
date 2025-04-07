package cli

import (
	"github.com/spf13/cobra"

	"github.com/act3-ai/data-telemetry/v3/internal/actions"
)

// NewConfigCmd creates a new "config" subcommand.
func NewConfigCmd(telemetryAction *actions.Telemetry) *cobra.Command {
	action := &actions.Config{
		Telemetry: telemetryAction,
	}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show the current configuration",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return action.Run(cmd.Context(), cmd.OutOrStdout())
		},
	}

	cmd.Example = `Configuration can be modified with the following environment variables:
ACE_TELEMETRY_DSN: Database connection URL

ACE_TELEMETRY_JUPYTER: Path to the jupyter executable to support ipynb conversion.
ACE_TELEMETRY_SELECTORS: A "|" separated list of label selectors to use as the default bottle selectors.
ACE_TELEMETRY_ASSETS: Assets directory
`

	cmd.Flags().BoolVarP(&action.Sample, "sample", "s", false,
		"Output a sample configuration that can be used in a configuration file.")

	return cmd
}
