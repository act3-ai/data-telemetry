package main

import (
	"context"
	"dagger/telemetry/internal/dagger"
	"fmt"
	"path"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"oras.land/oras-go/pkg/registry"
)

// Generate a directory of telemetry executables built for all supported platforms, concurrently.
func (t *Telemetry) BuildPlatforms(ctx context.Context,
	// snapshot build, skip goreleaser validations
	// +optional
	snapshot bool,
) *dagger.Directory {
	return GoReleaser(t.Source).
		WithExec([]string{"goreleaser", "build", "--clean", "--auto-snapshot", "--timeout=10m", fmt.Sprintf("--snapshot=%v", snapshot)}).
		Directory("dist")
}

// Build an executable for the specified platform, named "telemetry-{GOOS}-{GOARCH}".
//
// Supported Platform Matrix:
//
//	GOOS: linux, windows, darwin
//	GOARCH: amd64, arm64
func (t *Telemetry) Build(ctx context.Context,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
	// snapshot build, skip goreleaser validations
	// +optional
	snapshot bool,
) *dagger.File {
	return build(ctx, t.Source, platform, snapshot)
}

func build(ctx context.Context,
	src *dagger.Directory,
	platform dagger.Platform,
	// snapshot build, skip goreleaser validations
	snapshot bool,
) *dagger.File {
	name := binaryName(string(platform))

	_, span := Tracer().Start(ctx, fmt.Sprintf("Build %s", name))
	defer span.End()

	os, arch, _ := strings.Cut(string(platform), "/")
	return GoReleaser(src).
		WithEnvVariable("GOOS", os).
		WithEnvVariable("GOARCH", arch).
		WithExec([]string{"goreleaser", "build", "--auto-snapshot", "--timeout=10m", "--single-target", "--output", name, fmt.Sprintf("--snapshot=%v", snapshot)}).
		File(name)
}

// binaryName constructs the name of a telemetry executable based on build params.
// All arguments are optional, building up to "telemetry-v{VERSION}-fips-{GOOS}-{GOARCH}".
func binaryName(platform string) string {
	str := strings.Builder{}
	str.WriteString("telemetry")

	if platform != "" {
		platform = strings.ReplaceAll(string(platform), "/", "-")
		str.WriteString("-")
		str.WriteString(platform)
	}

	return str.String()
}

// Create an image with the telemetry executable.
func (t *Telemetry) Image(ctx context.Context,
	// image version
	version string,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
) *dagger.Container {
	// ensure to copy files, not mount them; else they won't be in the final image
	ctr := dag.Container(dagger.ContainerOpts{Platform: platform}).
		From("cgr.dev/chainguard/static").
		WithFile("/usr/local/bin/telemetry", t.Build(ctx, platform, false)).
		WithEntrypoint([]string{"telemetry"}).
		WithExposedPort(8100).
		WithWorkdir("/").
		WithLabel("description", "ACE Data Telemetry -- bottle and experiment tracking")

	return withCommonLabels(ctr, version)
}

// Create and publish a multi-platform image index.
// Uses docker mediatypes.
func (t *Telemetry) ImageIndex(ctx context.Context,
	// image version
	version string,
	// OCI Reference
	address string,
	// build platforms
	platforms []dagger.Platform,
) (string, error) {
	ref, err := registry.ParseReference(address)
	if err != nil {
		return "", fmt.Errorf("parsing address: %w", err)
	}
	imgURL := "https://" + path.Join(ref.Registry, ref.Repository)
	// i := imageURL.
	p := pool.NewWithResults[*dagger.Container]().WithContext(ctx)
	for _, platform := range platforms {
		p.Go(func(ctx context.Context) (*dagger.Container, error) {
			img := t.Image(ctx, version, platform).
				WithLabel("org.opencontainers.image.url", imgURL)
			return img, nil
		})
	}

	platformVariants, err := p.Wait()
	if err != nil {
		return "", fmt.Errorf("building images: %w", err)
	}

	return dag.Container().
		Publish(ctx, address, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
			MediaTypes:       dagger.ImageMediaTypesDockerMediaTypes,
		})
}

// Create an image with the telemetry executable, and ipynb support.
func (t *Telemetry) ImageIpynb(ctx context.Context,
	// image version
	version string,
	// Build target platform
	// +optional
	// +default="linux/amd64"
	platform dagger.Platform,
) *dagger.Container {
	venv := dag.Container(dagger.ContainerOpts{Platform: platform}).
		From("cgr.dev/chainguard/python:latest-dev").
		WithEnvVariable("LANG", "C.UTF-8").
		WithEnvVariable("PYTHONDONTWRITEBYTECODE", "1").
		WithEnvVariable("PYTHONUNBUFFERED", "1").
		WithEnvVariable("PATH", "/opt/venv/bin:$PATH", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithEnvVariable("ACE_TELEMETRY_JUPYTER", "/opt/venv/bin/jupyter").
		WithWorkdir("/opt/venv").
		WithExec([]string{"python", "-m", "venv", "/opt/venv"}).
		WithFile("/opt/requirements.txt", t.Source.File("requirements.txt")). // must copy // TODO: but why?
		WithExec([]string{"pip", "install", "--disable-pip-version-check", "--no-cache-dir", "--only-binary=:all:", "-r", "/opt/requirements.txt"}).
		Directory("/opt/venv")

		// ensure to copy files, not mount them; else they won't be in the final image
	ctr := dag.Container(dagger.ContainerOpts{Platform: platform}).
		From("cgr.dev/chainguard/python").
		WithEnvVariable("PYTHONUNBUFFERED", "1").
		WithEnvVariable("PATH", "/opt/venv/bin:$PATH", dagger.ContainerWithEnvVariableOpts{Expand: true}).
		WithEnvVariable("ACE_TELEMETRY_JUPYTER", "/opt/venv/bin/jupyter").
		WithFile("/usr/local/bin/telemetry", t.Build(ctx, platform, false)).
		WithDirectory("/opt/venv", venv).
		WithEntrypoint([]string{"telemetry"}).
		WithExposedPort(8100).
		WithWorkdir("/home/nonroot").
		WithLabel("description", "ACE Data Telemetry -- bottle and experiment tracking, with ipynb support")

	return withCommonLabels(ctr, version)
}

// Create and publish a multi-platform image index, with ipynb support.
// Uses docker mediatypes.
func (t *Telemetry) ImageIpynbIndex(ctx context.Context,
	// image version
	version string,
	// OCI Reference
	address string,
	// build platforms
	platforms []dagger.Platform,
) (string, error) {
	ref, err := registry.ParseReference(address)
	if err != nil {
		return "", fmt.Errorf("parsing address: %w", err)
	}
	imgURL := "https://" + path.Join(ref.Registry, ref.Repository)
	p := pool.NewWithResults[*dagger.Container]().WithContext(ctx)
	for _, platform := range platforms {
		p.Go(func(ctx context.Context) (*dagger.Container, error) {
			img := t.ImageIpynb(ctx, version, platform).
				WithLabel("org.opencontainers.image.url", imgURL)
			return img, nil
		})
	}

	platformVariants, err := p.Wait()
	if err != nil {
		return "", fmt.Errorf("building images: %w", err)
	}

	return dag.Container().
		Publish(ctx, address, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
			MediaTypes:       dagger.ImageMediaTypesDockerMediaTypes,
		})
}

// withCommonLabels applies common labels to a container, e.g. maintainers, vendor, etc.
func withCommonLabels(ctr *dagger.Container, version string) *dagger.Container {
	return ctr.
		WithLabel("maintainers", "Nathan D. Joslin <nathan.joslin@udri.udayton.edu>").
		WithLabel("org.opencontainers.image.vendor", "AFRL ACT3").
		WithLabel("org.opencontainers.image.version", version).
		WithLabel("org.opencontainers.image.title", "Telemetry").
		WithLabel("org.opencontainers.image.url", "ghcr.io/act3-ai/data-telemetry").
		WithLabel("org.opencontainers.image.description", "ACE Data Tool Telemetry Server").
		WithLabel("org.opencontainers.image.source", "https://github.com/act3-ai/data-telemetry")
}
