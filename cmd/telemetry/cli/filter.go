package cli

import (
	_ "embed"
	"os"

	"github.com/spf13/cobra"
)

// NewFilterCmd creates a new logging filter subcommand.
func NewFilterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filters to use when pretty printing logs",
	}
	cmd.SetOut(os.Stdout)

	cmd.AddCommand(
		newJqCmd(),
		newYqCmd(),
	)

	return cmd
}

//go:embed log.jq
var jqFilter string

func newJqCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jq",
		Short: "Print a jq filter to use when pretty printing logs",
		Long: `To use this filter you must have "jq" installed.

Save the filter to log.jq:
	telemetry filter jq > log.jq

Pretty logs:
	jq -j -f log.jq telemetry.log

Pipe from telemetry directly (requires bash):
	telemetry serve -v=1 2> >(jq -j -f log.jq)`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(jqFilter)
			return nil
		},
	}

	return cmd
}

//go:embed log.yq
var yqFilter string

func newYqCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yq",
		Short: "Print a yq filter to use when pretty printing logs",
		Long: `To use this filter you must have "yq" installed.

Save the filter to log.jq:
	telemetry filter yq > log.yq

Pretty logs
	yq -p=json --from-file=log.yq telemetry.logs

Pipe from telemetry directly (requires bash):
	telemetry serve -v=1 2> >(yq -p=json --from-file=log.yq)`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(yqFilter)
			return nil
		},
	}

	return cmd
}
