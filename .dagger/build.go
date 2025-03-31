package main

import (
	"context"
	"dagger/telemetry/internal/dagger"
	"fmt"
	"path"
	"strings"

	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"oras.land/oras-go/pkg/registry"
)

// Generate a directory of telemetry executables built for all supported platforms, concurrently.
func (t *Telemetry) BuildPlatforms(ctx context.Context,
	// release version
	// +optional
	version string,
) (*dagger.Directory, error) {
	// build matrix
	gooses := []string{"linux", "windows", "darwin"}
	goarches := []string{"amd64", "arm64"}

	ctx, span := Tracer().Start(ctx, "Build Platforms", trace.WithAttributes(attribute.StringSlice("GOOS", gooses), attribute.StringSlice("GOARCH", goarches)))
	defer span.End()

	buildsDir := dag.Directory()
	p := pool.NewWithResults[*dagger.File]().WithContext(ctx)

	for _, goos := range gooses {
		for _, goarch := range goarches {
			p.Go(func(ctx context.Context) (*dagger.File, error) {
				platform := fmt.Sprintf("%s/%s", goos, goarch)
				bin := t.Build(ctx, dagger.Platform(platform), version, "latest")
				return bin, nil
			})
		}
	}

	bins, err := p.Wait()
	if err != nil {
		return nil, err
	}
	return buildsDir.WithFiles(".", bins), nil
}

// Build an executable for the specified platform, named "telemetry-v{VERSION}-{GOOS}-{GOARCH}".
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
	// Release version, included in file name
	// +optional
	version string,
	// value of GOFIPS140, accepts modes "off", "latest", and "v1.0.0"
	// +optional
	// +default="latest"
	fipsMode string,
) *dagger.File {
	return build(ctx, t.Source, t.Netrc, platform, version, fipsMode)
}

func build(ctx context.Context,
	src *dagger.Directory,
	netrc *dagger.Secret,
	platform dagger.Platform,
	version string,
	fipsMode string,
) *dagger.File {
	// only name the result "fips" if it
	name := binaryName(string(platform), version)

	_, span := Tracer().Start(ctx, fmt.Sprintf("Build %s", name))
	defer span.End()

	return dag.Go(
		dagger.GoOpts{
			Container: dag.Container().
				From(imageGo).                            // same as dag.Go, but...
				WithMountedSecret("/root/.netrc", netrc), // allows us to mount this secret
		}).
		WithSource(src).
		WithCgoDisabled().
		WithEnvVariable("GO_PRIVATE", gitlabHost).
		WithEnvVariable("GOFIPS140", fipsMode).
		Build(dagger.GoWithSourceBuildOpts{
			Pkg:      "./cmd/telemetry",
			Platform: platform,
			Ldflags:  []string{"-s", "-w", fmt.Sprintf("-X 'main.version=%s'", version)},
			Trimpath: true,
		}).
		WithName(name)
}

// binaryName constructs the name of a telemetry executable based on build params.
// All arguments are optional, building up to "telemetry-v{VERSION}-fips-{GOOS}-{GOARCH}".
func binaryName(platform string, version string) string {
	str := strings.Builder{}
	str.Grow(35) // est. max = len("telemetry-v1.11.11-fips-linux-amd64")
	str.WriteString("telemetry")

	if version != "" {
		str.WriteString("-v")
		str.WriteString(version)
	}

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
		WithFile("/usr/local/bin/telemetry", t.Build(ctx, platform, "", "latest")).
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
		WithFile("/usr/local/bin/telemetry", t.Build(ctx, platform, "", "latest")).
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

// Create an image tailored for ace hub.
func (t *Telemetry) ImageHub(ctx context.Context,
	// image version
	version string,
	// gitlab api access token name
	secretName string,
	// gitlab api access token value
	secretValue *dagger.Secret,
) (*dagger.Container, error) {
	tk, err := secretValue.Plaintext(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting plaintext secret value: %w", err)
	}

	ctr := t.Source.Directory(".acehub").
		// At this point in time we only publish a linux/amd64 hub image
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Platform: "linux/amd64",
			Secrets:  []*dagger.Secret{dag.SetSecret(secretName, tk)},
		})

	return withCommonLabels(ctr, version), nil
}

// withCommonLabels applies common labels to a container, e.g. maintainers, vendor, etc.
func withCommonLabels(ctr *dagger.Container, version string) *dagger.Container {
	return ctr.
		WithLabel("maintainers", "Nathan D. Joslin <nathan.joslin@udri.udayton.edu>").
		WithLabel("org.opencontainers.image.vendor", "AFRL ACT3").
		WithLabel("org.opencontainers.image.version", version).
		WithLabel("org.opencontainers.image.title", "Telemetry").
		WithLabel("org.opencontainers.image.url", path.Join(gitlabHost, gitlabProject)).
		WithLabel("org.opencontainers.image.description", "ACE Data Tool Telemetry Server")
}
