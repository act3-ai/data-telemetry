// Package docs embeds relevant documentation to be surfaced in the ace-dt CLI.
package docs

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/act3-ai/asce/go-common/pkg/embedutil"
)

// apiDocs contains documentation for usage of the Telemetry API
//
//go:embed apis/config.telemetry.act3-ace.io/*.md
var apiDocs embed.FS

// cliDocs contains documentation for usage of the Telemetry CLI
//
//go:embed cli/*.md
var cliDocs embed.FS

// Embedded loads and categorizes the embedded documentation for use in the Telemetry CLI.
func Embedded(root *cobra.Command) (*embedutil.Documentation, error) {
	// walk CLIDocs filesystem to get filepaths to markdown files
	cliDocsPaths := []string{}
	err := fs.WalkDir(cliDocs, ".", func(fsPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking fsys directory: %w", err)
		}
		if !d.IsDir() && strings.HasSuffix(fsPath, ".md") {
			cliDocsPaths = append(cliDocsPaths, fsPath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("problem walking fsys: %w", err)
	}
	cliEmbeddedDocs := []*embedutil.Document{}

	for _, p := range cliDocsPaths {
		name := strings.TrimPrefix(path.Base(p), "telemetry_")
		name = strings.TrimSuffix(name, ".md")
		cliEmbeddedDocs = append(cliEmbeddedDocs, embedutil.LoadMarkdown(
			name,
			name,
			p,
			cliDocs))
	}

	return &embedutil.Documentation{
		Title:   "ACE Telemetry Server",
		Command: root,
		Categories: []*embedutil.Category{
			embedutil.NewCategory(
				"api", "API Documentation", root.Name()+"-api", 1,
				embedutil.LoadMarkdown("api", "API", "apis/config.telemetry.act3-ace.io/v1alpha1.md", apiDocs),
			),
			embedutil.NewCategory(
				"cli", "CLI Documentation", root.Name()+"-cli", 1,
				cliEmbeddedDocs...,
			),
		},
	}, nil
}
