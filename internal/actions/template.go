package actions

import (
	"path/filepath"

	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/testing"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/types"
)

// TemplateRun generates all templates for the given directory.
func TemplateRun(datadir string) error {
	// generate all files from templates
	for _, dir := range types.TopologicalOrderingOfTypes {
		if err := testing.ProcessTemplates(filepath.Join(datadir, dir)); err != nil {
			return err
		}
	}

	return nil
}
