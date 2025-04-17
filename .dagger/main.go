// A generated module for Telemetry functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"dagger/telemetry/internal/dagger"
)

const (
	// repository information
	gitRepo = "act3-ai/data-telemetry"

	// images
	imageGitCliff   = "docker.io/orhunp/git-cliff:2.8.0"
	imageAcedt      = "ghcr.io/act3-ai/data-tool:v1.15.10"
	imageGrype      = "anchore/grype:latest"
	imageSyft       = "anchore/syft:latest"
	imageGo         = "golang:latest" // github.com/sagikazarmark/daggerverse/go convention
	imagePostgres   = "postgres:17-alpine"
	imageGoReleaser = "ghcr.io/goreleaser/goreleaser:v2.8.2"

	// go tools
	goVulnCheck     = "golang.org/x/vuln/cmd/govulncheck@latest"
	goControllerGen = "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.2"
	goCrdRefDocs    = "github.com/elastic/crd-ref-docs@v0.1.0"
)

type Telemetry struct {
	// source code directory
	Source *dagger.Directory

	// +private
	RegistryConfig *dagger.RegistryConfig
	// +private
	Netrc *dagger.Secret
}

func New(
	// top level source code directory
	// +defaultPath="/"
	src *dagger.Directory,
) *Telemetry {
	return &Telemetry{
		Source:         src,
		RegistryConfig: dag.RegistryConfig(),
	}
}

// Add credentials for a registry.
func (t *Telemetry) WithRegistryAuth(
	// registry's hostname
	address string,
	// username in registry
	username string,
	// password or token for registry
	secret *dagger.Secret,
) *Telemetry {
	t.RegistryConfig = t.RegistryConfig.WithRegistryAuth(address, username, secret)
	return t
}

// Removes credentials for a registry.
func (t *Telemetry) WithoutRegistryAuth(
	// registry's hostname
	address string,
) *Telemetry {
	t.RegistryConfig = t.RegistryConfig.WithoutRegistryAuth(address)
	return t
}

// Add netrc credentials for a private git repository.
func (t *Telemetry) WithNetrc(
	// NETRC credentials
	netrc *dagger.Secret,
) *Telemetry {
	t.Netrc = netrc
	return t
}
