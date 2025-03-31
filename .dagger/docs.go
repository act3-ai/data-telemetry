package main

import (
	"context"
	"dagger/telemetry/internal/dagger"
	"fmt"
	"path/filepath"
)

const (
	pkgPath  = "pkg/apis/config.telemetry.act3-ace.io"
	docsPath = "docs/apis/config.telemetry.act3-ace.io"
)

// Generate CLI documentation.
func (t *Telemetry) CLIDocs(ctx context.Context) *dagger.Directory {
	telem := t.Build(ctx, "linux/amd64", "", "")

	cliDocsPath := "docs/cli"
	return dag.Go().
		WithSource(t.Source).
		Container().
		WithFile("/usr/local/bin/telemetry", telem).
		WithExec([]string{"telemetry", "gendocs", "md", "--only-commands", cliDocsPath}).
		Directory(cliDocsPath)
}

// Generate API documentation.
func (t *Telemetry) APIDocs() (*dagger.Directory, error) {

	ctr := dag.Go().
		WithSource(t.Source).
		Exec([]string{"go", "install", goCrdRefDocs})

	ctr = ctr.WithExec([]string{"crd-ref-docs", "--config=apidocs.yaml", "--renderer=markdown",
		fmt.Sprintf("--source-path=%s/", pkgPath),
		fmt.Sprintf("--output-path=%s/", docsPath),
	})

	return ctr.WithoutFile(filepath.Join(docsPath, "out.md")).
		Directory(docsPath), nil
}

// Generate pkg/apis with controller-gen.
func (t *Telemetry) Generate() *dagger.Directory {
	ctr := dag.Go().
		WithSource(t.Source).
		WithEnvVariable("GOBIN", "/work/src/tool").
		Exec([]string{"go", "install", goControllerGen}).
		WithExec([]string{"go", "generate", "./..."})

	return ctr.Directory(pkgPath)
}

// Template swagger.yml.
func (t *Telemetry) Swagger() *dagger.File {
	return dag.Wolfi().
		Container(dagger.WolfiContainerOpts{
			Packages: []string{"yq"},
		}).
		WithDirectory("/work/src", t.Source).
		WithWorkdir("/work/src").
		WithExec([]string{"yq", "eval", "--inplace", `.paths."/api/bottle".put.requestBody.content."application/json".examples.bottleJSON.value=load_str("testdata/bottle/bottle1.json")`, "swagger.yml"}).
		WithExec([]string{"yq", "eval", "--inplace", `.paths."/api/manifest".put.requestBody.content."application/json".examples.manifestJSON.value=load_str("testdata/manifest/manifest1.json")`, "swagger.yml"}).
		WithExec([]string{"yq", "eval", "--inplace", `.paths."/api/event".put.requestBody.content."application/json".examples.eventJSON.value=load_str("testdata/event/push1.json")`, "swagger.yml"}).
		File("/work/src/swagger.yml")
}
