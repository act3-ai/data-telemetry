// Package main is the main package for the telemetry program
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	commands "gitlab.com/act3-ai/asce/go-common/pkg/cmd"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/runner"
	vv "gitlab.com/act3-ai/asce/go-common/pkg/version"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/cmd/telemetry/cli"
	"gitlab.com/act3-ai/asce/data/telemetry/v3/docs"
)

// getVersionInfo retreives the proper version information for this executable.
func getVersionInfo() vv.Info {
	info := vv.Get()
	if version != "" {
		info.Version = version
	}
	return info
}

func main() {
	info := getVersionInfo()
	root := cli.NewTelemetryCmd(info)
	ctx := context.Background()

	// add embedded documentation command
	embeddedDocs, err := docs.Embedded(root)
	if err != nil {
		panic(fmt.Errorf("could not embed docs: %w", err))
	}

	root.AddCommand(
		commands.NewVersionCmd(info),
		commands.NewGendocsCmd(embeddedDocs),
	)

	root.SilenceUsage = true
	// root.SilenceErrors = true

	root.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		log := logger.FromContext(ctx)
		log.InfoContext(ctx, "Software", "version", info.Version)
		log.DebugContext(ctx, "Software details", "info", info)
	}

	if err := runner.Run(ctx, root, "ACE_TELEMETRY_VERBOSITY"); err != nil {
		// fmt.Fprintln(os.Stderr, "Error occurred", err)
		os.Exit(1)
	}
}
