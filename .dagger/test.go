package main

import (
	"context"
	"dagger/telemetry/internal/dagger"
	"errors"
)

// Run tests.
func (t *Telemetry) Test() *Test {
	return &Test{
		Source:         t.Source,
		Netrc:          t.Netrc,
		RegistryConfig: t.RegistryConfig,
	}
}

// Test organizes test functions.
type Test struct {
	// source code directory
	// +defaultPath="/"
	Source *dagger.Directory

	// NETRC credentials
	// +private
	Netrc *dagger.Secret
	// +private
	RegistryConfig *dagger.RegistryConfig
}

// Run all tests.
func (tt *Test) All(ctx context.Context) (string, error) {
	unitResults, unitErr := tt.Unit(ctx)

	funcResults, funcErr := tt.Functional(ctx)

	out := "Unit Test Results:\n" + unitResults + "\n=====\n\nFunctional Test Results:\n" + funcResults
	return out, errors.Join(unitErr, funcErr)
}

// Run unit tests.
func (tt *Test) Unit(ctx context.Context) (string, error) {
	return tt.TemplateTestData(ctx).
		WithExec([]string{"go", "test", "./...", "-tags", "sqlite_fts5"}).
		Stdout(ctx)
}

// Run functional tests.
func (tt *Test) Functional(ctx context.Context) (string, error) {
	ctx, span := Tracer().Start(ctx, "Functional Tests")
	defer span.End()

	postgres := tt.Postgres()
	postgresService, err := postgres.Start(ctx)
	if err != nil {
		return "", err
	}
	defer postgres.Stop(ctx)

	telemServer := tt.Server(ctx, postgresService)
	telemServer, err = telemServer.Start(ctx)
	if err != nil {
		return "", err
	}
	defer telemServer.Stop(ctx)

	telemEndpoint, err := telemServer.Endpoint(ctx, dagger.ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return "", err
	}

	uploadResult, err := tt.TemplateTestData(ctx).
		WithServiceBinding("server", telemServer).
		WithExec([]string{"telemetry", "client", "upload", "./testdata", telemEndpoint, "--all"}).
		Stdout(ctx)
	if err != nil {
		return uploadResult, err
	}

	downloadResult, err := tt.TemplateTestData(ctx).
		WithServiceBinding("server", telemServer).
		WithDirectory("testdata-download", dag.Directory()).
		WithExec([]string{"telemetry", "client", "download", "testdata-download", telemEndpoint, "--all", "--from-latest"}).
		Stdout(ctx)
	if err != nil {
		return downloadResult, err
	}

	webappResult, err := tt.TemplateTestData(ctx).
		WithServiceBinding("server", telemServer).
		WithExec([]string{"hack/test-webapp.sh", telemEndpoint}).
		Stdout(ctx)
	if err != nil {
		return webappResult, err
	}

	_, err = telemServer.Stop(ctx)
	if err != nil {
		return "", err
	}

	_, err = postgres.Stop(ctx)
	if err != nil {
		return "", err
	}

	return uploadResult + "\n" + downloadResult + "\n" + webappResult, nil
}

// Start the telemetry server as a service.
func (tt *Test) Server(ctx context.Context,
	// postgres service
	postgres *dagger.Service,
) *dagger.Service {
	return tt.TemplateTestData(ctx).
		WithEnvVariable("ACE_TELEMETRY_DSN", "postgres://testUser:testPassword@postgres/testdb").
		WithServiceBinding("postgres", postgres).
		WithExposedPort(8100).
		AsService(dagger.ContainerAsServiceOpts{
			Args: []string{"telemetry", "serve", "--listen", ":8100"},
		})
}

// Start postgres as a service.
func (tt *Test) Postgres() *dagger.Service {
	return dag.Container().
		From(imagePostgres).
		// Notice: changes to these env vars must be reflected in uses of
		// ACE_TELEMETRY_DSN
		WithEnvVariable("POSTGRES_DB", "testdb").
		WithEnvVariable("POSTGRES_USER", "testUser").
		WithEnvVariable("POSTGRES_PASSWORD", "testPassword").
		WithEnvVariable("POSTGRES_HOST_AUTH_METHOD", "trust").
		WithExposedPort(5432).
		AsService(dagger.ContainerAsServiceOpts{UseEntrypoint: true})
}

// ServerWithPostgres starts a telemetry service with postgres.
//
// This function makes it easier to expose a telemetry server to the host
// with less hassle connecting it to postgres.
func (tt *Test) ServerWithPostgres(ctx context.Context) (*dagger.Service, error) {
	telemServer := tt.Server(ctx, tt.Postgres())
	telemServer, err := telemServer.Start(ctx)
	if err != nil {
		return nil, err
	}

	return telemServer, nil
}

// TemplateTestData returns a telemetry container after executing
// `telemetry template` with the testdata directory.
func (tt *Test) TemplateTestData(ctx context.Context) *dagger.Container {
	bin := build(ctx, tt.Source, "linux/amd64", true)

	return dag.Go().
		WithSource(tt.Source).
		Container().
		WithUser("0").
		WithMountedFile("/usr/local/bin/telemetry", bin).
		WithExec([]string{"telemetry", "template", "./testdata"})
}
