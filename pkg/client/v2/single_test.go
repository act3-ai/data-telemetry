package client

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/runtime"

	bottle "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/go-common/pkg/test"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/api"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/middleware"
	ttest "gitlab.com/act3-ai/asce/data/telemetry/internal/testing"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/types"
)

type SingleTestSuite struct {
	suite.Suite
	server  *httptest.Server
	dataDir string
	blobs   map[digest.Digest][]byte
	log     *slog.Logger
	ctx     context.Context
	client  *Single
	clientB *Single
}

func (s *SingleTestSuite) getBlobByDigest(dgst digest.Digest) ([]byte, error) {
	return s.blobs[dgst], nil
}

func (s *SingleTestSuite) SetupTest() {
	s.dataDir = filepath.Join("..", "..", "..", "testdata")
	s.log = test.Logger(s.T(), 0)
	s.ctx = logger.NewContext(context.Background(), s.log)

	scheme := runtime.NewScheme()
	s.NoError(bottle.AddToScheme(scheme))

	// Instead of an env we can use a the "flags" package to create a flag and default it to the env if set or to file::memory: if not
	dsn := "file::memory:"
	myDB, err := db.Open(s.ctx, v1alpha1.Database{
		DSN: redact.SecretURL(dsn),
	}, scheme)
	s.NoError(err)

	router := chi.NewRouter()
	router.Use(
		httputil.LoggingMiddleware(s.log),
		middleware.DatabaseMiddleware(myDB),
	)
	router.Route("/api", func(router chi.Router) {
		a := &api.API{}
		a.Initialize(router, scheme)
	})

	// process and load the blobs
	s.blobs = make(map[digest.Digest][]byte)
	err = processIndexFile(filepath.Join(s.dataDir, "blob", "index.csv"), func(datafile string, dgst digest.Digest, data []byte) error {
		s.blobs[dgst] = data
		return nil
	})
	s.NoError(err)
	s.server = httptest.NewServer(router)
	// Use Client & URL from our local test server
	sc, err := NewSingleClient(s.server.Client(), s.server.URL, "mycooltoken")
	s.NoError(err)
	s.client = sc

	// mockLocation is a mock example of our config file for testing
	mockLocation := v1alpha1.Location{
		Name:    "MyMockConfig",
		URL:     redact.SecretURL(s.server.URL),
		Cookies: map[string]redact.Secret{"foo": "bar"},
		Token:   "mycooltoken",
	}

	scWithToken, err := NewSingleClient(http.DefaultClient, string(mockLocation.URL), string(mockLocation.Token))
	s.NoError(err)
	s.clientB = scWithToken
}

func (s *SingleTestSuite) TearDownTest() {
	// Close the server when test finishes
	s.server.Close()
}

func (s *SingleTestSuite) TestPutBlob() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "blob", "sample.txt"))
	s.NoError(err)

	err = s.client.PutBlob(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *SingleTestSuite) TestPutBlobSHA512() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "blob", "doc.md"))
	s.NoError(err)

	err = s.client.PutBlob(s.ctx, digest.SHA512, byteValue)
	s.NoError(err)
}

func (s *SingleTestSuite) TestPutBottleMissingDigest() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	err = s.client.PutBottle(s.ctx, digest.SHA256, byteValue)
	missing := &types.MissingDigestsError{}
	s.ErrorAs(err, &missing)
	s.NotEmpty(missing.MissingDigests)
}

func (s *SingleTestSuite) TestPutBottleSuccess() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle4.json"))
	s.NoError(err)

	err = s.client.PutBottle(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *SingleTestSuite) TestPutManifestMissingDigest() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest2.json"))
	s.NoError(err)

	err = s.client.PutManifest(s.ctx, digest.SHA256, byteValue)
	missing := &types.MissingDigestsError{}
	s.ErrorAs(err, &missing)
	s.NotEmpty(missing.MissingDigests)
}

func (s *SingleTestSuite) TestPutManifestSuccess() {
	s.TestPutBottleSuccess()
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest4.json"))
	s.NoError(err)

	err = s.client.PutManifest(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *SingleTestSuite) TestPutEventMissingDigest() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "event", "pull1.json"))
	s.NoError(err)

	err = s.client.PutEvent(s.ctx, digest.SHA256, byteValue)
	missing := &types.MissingDigestsError{}
	s.ErrorAs(err, &missing)
	s.NotEmpty(missing.MissingDigests)
}

func (s *SingleTestSuite) TestPutEventSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle4.json"))
	s.NoError(err)
	err = s.client.PutBottle(s.ctx, digest.SHA256, btl)
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest4.json"))
	s.NoError(err)
	err = s.client.PutManifest(s.ctx, digest.SHA256, manifest)
	s.NoError(err)

	event, err := os.ReadFile(filepath.Join(s.dataDir, "event", "pull4.json"))
	s.NoError(err)
	err = s.client.PutEvent(s.ctx, digest.SHA256, event)
	s.NoError(err)
}

func (s *SingleTestSuite) TestSendBottleSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	err = s.client.SendBottle(s.ctx, digest.SHA256, btl, s.getBlobByDigest)
	s.NoError(err)
}

func (s *SingleTestSuite) TestSendManifestSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest1.json"))
	s.NoError(err)

	err = s.client.SendManifest(s.ctx, digest.SHA256, manifest, btl, s.getBlobByDigest)
	s.NoError(err)
}

func (s *SingleTestSuite) TestSendEventSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest1.json"))
	s.NoError(err)

	event, err := os.ReadFile(filepath.Join(s.dataDir, "event", "pull1.json"))
	s.NoError(err)

	err = s.client.SendEvent(s.ctx, digest.SHA256, event, manifest, btl, s.getBlobByDigest)
	s.NoError(err)
}

func (s *SingleTestSuite) TestListBlobs() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	askTime, err := time.Parse(time.RFC3339, "2021-11-15T11:06:36.762880891-05:00")
	s.NoError(err)

	getblob, err := s.client.ListBlobs(s.ctx, askTime, 10)
	s.NoError(err)

	s.NotEmpty(getblob)
	for _, b := range getblob {
		s.NotEmpty(b.Digests)
		s.NotEmpty(b.Data)
		s.Condition(func() bool { return b.CreatedAt.After(askTime) })
		s.Condition(func() bool { return b.CreatedAt.Before(time.Now()) })

	}
}

func (s *SingleTestSuite) TestListBottles() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	askTime, err := time.Parse(time.RFC3339, "2021-11-15T11:06:36.762880891-05:00")
	s.NoError(err)

	getbottles, err := s.client.ListBottles(s.ctx, askTime, 10)
	s.NoError(err)

	s.NotEmpty(getbottles)
	for _, b := range getbottles {
		s.NotEmpty(b.Digests)
		s.NotEmpty(b.Data)
		s.Condition(func() bool { return b.CreatedAt.After(askTime) })
		s.Condition(func() bool { return b.CreatedAt.Before(time.Now()) })

	}
}

func (s *SingleTestSuite) TestListManifests() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))
	askTime, err := time.Parse(time.RFC3339, "2021-11-15T11:06:36.762880891-05:00")
	s.NoError(err)

	getManifests, err := s.client.ListManifests(s.ctx, askTime, 10)
	s.NoError(err)

	s.NotEmpty(getManifests)
	for _, b := range getManifests {
		s.NotEmpty(b.Digests)
		s.NotEmpty(b.Data)
		s.Condition(func() bool { return b.CreatedAt.After(askTime) })
		s.Condition(func() bool { return b.CreatedAt.Before(time.Now()) })

	}
}

func (s *SingleTestSuite) TestListEvents() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))
	askTime, err := time.Parse(time.RFC3339, "2021-11-15T11:06:36.762880891-05:00")
	s.NoError(err)

	getEvents, err := s.client.ListEvents(s.ctx, askTime, 10)
	s.NoError(err)

	s.NotEmpty(getEvents)
	for _, b := range getEvents {
		s.NotEmpty(b.Digests)
		s.NotEmpty(b.Data)
		s.Condition(func() bool { return b.CreatedAt.After(askTime) })
		s.Condition(func() bool { return b.CreatedAt.Before(time.Now()) })

	}
}

func (s *SingleTestSuite) TestGetBlob() {
	s.TestUploadBlob()
	alg := digest.SHA512
	blobDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "blob", "doc.md"), alg)
	s.NoError(err)

	getblobsbyDigest, err := s.client.GetBlob(s.ctx, digest.Digest(blobDigest))
	s.NoError(err)
	s.NotEmpty(getblobsbyDigest)

	hash := alg.FromBytes(getblobsbyDigest)

	// check if both digests are equal
	s.Equal(blobDigest, hash)
}

func (s *SingleTestSuite) TestGetBottle() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	alg := digest.SHA256
	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), alg)
	s.NoError(err)

	getBottleByDigest, err := s.client.GetBottle(s.ctx, digest.Digest(bottleDigest))
	s.NoError(err)
	s.NotEmpty(getBottleByDigest)

	hash := alg.FromBytes(getBottleByDigest)

	// check if both digests are equal
	s.Equal(bottleDigest, hash)
}

func (s *SingleTestSuite) TestGetManifest() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	alg := digest.SHA256
	manifestDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "manifest", "manifest1.json"), alg)
	s.NoError(err)

	getManifestByDigest, err := s.client.GetManifest(s.ctx, digest.Digest(manifestDigest))
	s.NoError(err)
	s.NotEmpty(getManifestByDigest)

	hash := alg.FromBytes(getManifestByDigest)

	// check if both digests are equal
	s.Equal(manifestDigest, hash)
}

func (s *SingleTestSuite) TestGetEvent() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	alg := digest.SHA256
	eventDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "event", "pull1.json"), alg)
	s.NoError(err)

	getEventByDigest, err := s.client.GetEvent(s.ctx, digest.Digest(eventDigest))
	s.NoError(err)
	s.NotEmpty(getEventByDigest)

	hash := alg.FromBytes(getEventByDigest)

	// check if both digests are equal
	s.Equal(eventDigest, hash)
}

func (s *SingleTestSuite) TestUploadBlob() {
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "blob", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadBlobFromConfig() {
	err := s.clientB.Upload(s.ctx, filepath.Join(s.dataDir, "blob", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadBottle() {
	s.TestUploadBlob()
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "bottle", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadBottleFromConfig() {
	s.TestUploadBlobFromConfig()
	err := s.clientB.Upload(s.ctx, filepath.Join(s.dataDir, "bottle", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadManifest() {
	s.TestUploadBottle()
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "manifest", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadManifestFromConfig() {
	s.TestUploadBottleFromConfig()
	err := s.clientB.Upload(s.ctx, filepath.Join(s.dataDir, "manifest", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadEvent() {
	s.TestUploadManifest()
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "event", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadEventFromConfig() {
	s.TestUploadManifestFromConfig()
	err := s.clientB.Upload(s.ctx, filepath.Join(s.dataDir, "event", "index.csv"), false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadAll() {
	err := s.client.UploadAll(s.ctx, s.dataDir, false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestUploadAllFromConfig() {
	err := s.clientB.UploadAll(s.ctx, s.dataDir, false)
	s.NoError(err)
}

func (s *SingleTestSuite) TestGetLocations() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	bottleLocation, err := s.client.GetLocations(s.ctx, digest.Digest(bottleDigest))
	s.NoError(err)

	s.NotEmpty(bottleLocation)
	for _, b := range bottleLocation {
		s.NotEmpty(b.Digest)
		s.NotEmpty(b.Repository)
	}
}

func (s *SingleTestSuite) TestBottleSearch() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	selectorval := []string{"type=testing,group=testset"}

	bottleSearch, err := s.client.BottleSearch(s.ctx, selectorval, "image", 7, true)
	s.NoError(err)

	s.NotEmpty(bottleSearch)
}

func (s *SingleTestSuite) TestGetBottlesFromMetric() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	metric := "training loss"
	selector := []string{"type=testing,myotherkey=myothervalue2", "myotherkey=doesnotexist"}

	metricSearch, err := s.client.GetBottlesFromMetric(s.ctx, selector, metric, 7, true)
	s.NoError(err)

	s.NotEmpty(metricSearch)
}

func TestSingleTestSuite(t *testing.T) {
	suite.Run(t, new(SingleTestSuite))
}
