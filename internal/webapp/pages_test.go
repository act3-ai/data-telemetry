package webapp_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/runtime"

	bottle "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/go-common/pkg/test"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/api"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/dbtest"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/middleware"
	ttest "gitlab.com/act3-ai/asce/data/telemetry/internal/testing"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/webapp"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	client "gitlab.com/act3-ai/asce/data/telemetry/pkg/client/v2"
)

// Make sure you run `make template` to ensure that the files are all generated in the testdata directory

type HandlersTestSuite struct {
	suite.Suite
	server   *httptest.Server
	dataDir  string
	assetDir string
	log      *slog.Logger
}

func (s *HandlersTestSuite) SetupSuite() {
	s.dataDir = filepath.Join("..", "..", "testdata")
	s.assetDir = filepath.Join(".", "assets")

	s.log = test.Logger(s.T(), 0)
	ctx := logger.NewContext(context.Background(), s.log)

	scheme := runtime.NewScheme()
	s.NoError(bottle.AddToScheme(scheme))
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "file::memory:"
	}
	u, err := url.Parse(dsn)
	s.NoError(err, "could not parse TEST_DSN dsn to URL %s", dsn)

	if u.Scheme == "postgres" {
		// If using postgres, create a temporary database for each test
		testPgDbDsn, cleanup, err := dbtest.CreateTempPostgresDb(s.T().Name(), u.String())
		s.NoError(err, "could not create test database in postgres with DSN %s", dsn)

		u, err = url.Parse(testPgDbDsn)
		s.NoError(err, "could not URL parse test Postgres dsn %s", testPgDbDsn)
		s.T().Cleanup(cleanup)
	}
	myDB, err := db.Open(ctx, v1alpha2.Database{
		DSN: redact.SecretURL(u.String()),
	}, scheme)
	s.NoError(err)

	router := chi.NewRouter()
	router.Use(
		httputil.LoggingMiddleware(s.log),
		middleware.DatabaseMiddleware(myDB),
	)

	// create a temporary API so we can load data
	router.Route("/_api", func(router chi.Router) {
		a := &api.API{}
		a.Initialize(router, scheme)
	})

	// create the webapp (the unit under test)
	webApp, err := webapp.NewWebApp(v1alpha2.WebApp{
		AssetDir: s.assetDir,
	}, s.log, "test-version")
	s.NoError(err)
	webApp.Initialize(router)

	s.server = httptest.NewServer(router)

	// upload test data
	uploadURL, err := url.Parse(s.server.URL + "/_api")
	s.NoError(err)
	s.NoError(client.UploadAll(ctx, s.server.Client(), s.dataDir, uploadURL, "", false))
}

func (s *HandlersTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *HandlersTestSuite) makeRequest(method, u string, body io.Reader) *http.Request {
	ctx := logger.NewContext(context.Background(), s.log)
	req, err := http.NewRequestWithContext(ctx, method, s.server.URL+u, body)
	s.NoError(err)
	return req
}

func (s *HandlersTestSuite) performRequest(req *http.Request) (int, http.Header, []byte) {
	s.T().Logf("request URL: %s", req.URL.String())
	res, err := s.server.Client().Do(req)
	s.NoError(err)
	defer func() {
		s.NoError(res.Body.Close())
	}()

	body, err := io.ReadAll(res.Body)
	s.NoError(err)
	s.T().Logf("status: %s", res.Status)
	s.T().Logf("body: %s", body)
	s.T().Logf("headers: %s", res.Header)

	return res.StatusCode, res.Header, body
}

func (s *HandlersTestSuite) TestCatalog() {
	u := url.URL{
		Path: "/catalog.html",
		// RawQuery: url.Values{"selector": []string{"mykey=myvalue"}}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, _ := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
}

func (s *HandlersTestSuite) TestLeaderboard() {
	u := url.URL{
		Path: "/leaderboard.html",
		// RawQuery: url.Values{"digest": []string{dgst}}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, _ := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
}

func (s *HandlersTestSuite) TestBottle() {
	dgst, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path:     "/bottle.html",
		RawQuery: url.Values{"digest": []string{dgst.String()}}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, _ := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
}

func (s *HandlersTestSuite) TestArtifactTabular() {
	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path:     path.Join("/artifact", bottleDigest.String(), "foo/bar/data.csv"),
		RawQuery: url.Values{"_type": []string{"raw"}}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, _ := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
}

func (s *HandlersTestSuite) TestArtifactText() {
	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path: path.Join("/artifact", bottleDigest.String(), "foo/bar/sample.txt"),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
	s.Contains(string(body), "The dog jumped over the moon.")
}

func (s *HandlersTestSuite) TestArtifactImage() {
	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle2.json"), "sha512")
	s.NoError(err)

	u := url.URL{
		Path: path.Join("/artifact", bottleDigest.String(), "foo/bar/a.png"),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, _ := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
}

func (s *HandlersTestSuite) TestSimilarBottles() {
	u := url.URL{
		Path: "/similarBottles",
		RawQuery: url.Values{
			"requirement": []string{"foo!=bar"},
			"metric":      []string{"my-metric"},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, _ := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	// TODO check the response
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
