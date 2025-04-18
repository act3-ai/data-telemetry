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

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/runtime"

	bottle "github.com/act3-ai/bottle-schema/pkg/apis/data.act3-ace.io"
	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/act3-ai/go-common/pkg/logger"
	"github.com/act3-ai/go-common/pkg/redact"
	"github.com/act3-ai/go-common/pkg/test"

	"github.com/act3-ai/data-telemetry/v3/internal/api"
	"github.com/act3-ai/data-telemetry/v3/internal/db"
	"github.com/act3-ai/data-telemetry/v3/internal/middleware"
	ttest "github.com/act3-ai/data-telemetry/v3/internal/testing"
	"github.com/act3-ai/data-telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	"github.com/act3-ai/data-telemetry/v3/pkg/types"
)

type MultiTestSuite struct {
	suite.Suite
	serverA *httptest.Server
	serverB *httptest.Server
	dataDir string
	blobs   map[digest.Digest][]byte
	log     *slog.Logger
	ctx     context.Context
	client  *MultiClient
}

func (s *MultiTestSuite) getBlobByDigest(dgst digest.Digest) ([]byte, error) {
	return s.blobs[dgst], nil
}

func (s *MultiTestSuite) SetupTest() {
	scheme := runtime.NewScheme()
	s.NoError(bottle.AddToScheme(scheme))

	s.dataDir = filepath.Join("..", "..", "testdata")
	s.log = test.Logger(s.T(), -8)
	s.ctx = logger.NewContext(context.Background(), s.log)

	dsn := "file::memory:"

	myDB, err := db.Open(s.ctx, v1alpha2.Database{
		DSN: redact.SecretURL(dsn),
	}, scheme)
	s.NoError(err)

	myDB2, err := db.Open(s.ctx, v1alpha2.Database{
		DSN: redact.SecretURL(dsn),
	}, scheme)
	s.NoError(err)

	// initializing 2 apis for different clients
	apiA := api.API{}
	apiMuxA := http.NewServeMux()
	apiA.Initialize(apiMuxA, scheme)
	routerA := http.NewServeMux()
	routerA.Handle("/api/", http.StripPrefix("/api", apiMuxA))
	wrappedRouterA := httputil.LoggingMiddleware(s.log)(middleware.DatabaseMiddleware(myDB)(routerA))

	apiB := api.API{}
	apiMuxB := http.NewServeMux()
	apiB.Initialize(apiMuxB, scheme)
	routerB := http.NewServeMux()
	routerB.Handle("/api/", http.StripPrefix("/api", apiMuxB))
	wrappedRouterB := httputil.LoggingMiddleware(s.log)(middleware.DatabaseMiddleware(myDB2)(routerB))

	// process and load the blobs
	s.blobs = make(map[digest.Digest][]byte)
	err = processIndexFile(filepath.Join(s.dataDir, "blob", "index.csv"), func(datafile string, dgst digest.Digest, data []byte) error {
		s.blobs[dgst] = data
		return nil
	})
	s.NoError(err)

	// different clients will talk to different servers
	s.serverA = httptest.NewServer(wrappedRouterA)
	s.serverB = httptest.NewServer(wrappedRouterB)

	client1, err := NewSingleClient(s.serverA.Client(), s.serverA.URL, "mycooltoken")
	s.NoError(err)

	client2 := &Dummy{}

	client3, err := NewSingleClient(s.serverB.Client(), s.serverB.URL, "mycooltoken")
	s.NoError(err)

	// initiate a new multiclient
	s.client = NewMultiClient([]Client{client1, client2, client3})
}

func (s *MultiTestSuite) TearDownTest() {
	// Close the server when test finishes
	s.serverA.Close()
	s.serverB.Close()
}

func (s *MultiTestSuite) TestUploadAll() {
	err := s.client.UploadAll(s.ctx, s.dataDir, false)
	s.NoError(err)
}

func (s *MultiTestSuite) TestPutBlob() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "blob", "sample.txt"))
	s.NoError(err)

	err = s.client.PutBlob(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *MultiTestSuite) TestPutBlobWithToken() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "blob", "sample.txt"))
	s.NoError(err)

	err = s.client.PutBlob(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *MultiTestSuite) TestPutBlobSHA512() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "blob", "doc.md"))
	s.NoError(err)

	err = s.client.PutBlob(s.ctx, digest.SHA512, byteValue)
	s.NoError(err)
}

func (s *MultiTestSuite) TestUploadBlob() {
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "blob", "index.csv"), false)
	s.NoError(err)
}

func (s *MultiTestSuite) TestGetBlob() {
	s.TestUploadBlob()
	blobDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "blob", "doc.md"), "sha512")
	s.NoError(err)

	blob, err := s.client.GetBlob(s.ctx, digest.Digest(blobDigest))
	s.NoError(err)
	s.NotEmpty(blob)

	alg := digest.SHA512
	hash := alg.FromBytes(blob)

	// check if both digests are equal
	s.Equal(blobDigest, hash)
}

// -----------------------------------------------------------------------------------------------

func (s *MultiTestSuite) TestPutBottleMissingDigest() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	err = s.client.PutBottle(s.ctx, digest.SHA256, byteValue)

	missing := &types.MissingDigestsError{}
	s.ErrorAs(err, &missing)
}

func (s *MultiTestSuite) TestPutBottleSuccess() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle4.json"))
	s.NoError(err)

	err = s.client.PutBottle(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *MultiTestSuite) TestPutManifestMissingDigest() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest2.json"))
	s.NoError(err)

	err = s.client.PutManifest(s.ctx, digest.SHA256, byteValue)

	missing := &types.MissingDigestsError{}
	s.ErrorAs(err, &missing)
}

func (s *MultiTestSuite) TestPutManifestSuccess() {
	s.TestPutBottleSuccess()
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest4.json"))
	s.NoError(err)

	err = s.client.PutManifest(s.ctx, digest.SHA256, byteValue)
	s.NoError(err)
}

func (s *MultiTestSuite) TestPutEventMissingDigest() {
	byteValue, err := os.ReadFile(filepath.Join(s.dataDir, "event", "pull1.json"))
	s.NoError(err)

	err = s.client.PutEvent(s.ctx, digest.SHA256, byteValue)

	missing := &types.MissingDigestsError{}
	s.ErrorAs(err, &missing)
}

func (s *MultiTestSuite) TestPutEventSuccess() {
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

func (s *MultiTestSuite) TestPutSignatureSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle4.json"))
	s.NoError(err)
	err = s.client.PutBottle(s.ctx, digest.SHA256, btl)
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest4.json"))
	s.NoError(err)
	err = s.client.PutManifest(s.ctx, digest.SHA256, manifest)
	s.NoError(err)

	// signature, err := os.ReadFile(filepath.Join(s.dataDir, "signature", "signature2.json"))
	// s.NoError(err)
	//
	// err = s.client.PutSignature(s.ctx, digest.SHA256, signature)
	// s.NoError(err)
}

func (s *MultiTestSuite) TestSendBottleSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	err = s.client.SendBottle(s.ctx, digest.SHA256, btl, s.getBlobByDigest)
	s.NoError(err)
}

func (s *MultiTestSuite) TestSendManifestSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest1.json"))
	s.NoError(err)

	err = s.client.SendManifest(s.ctx, digest.SHA256, manifest, btl, s.getBlobByDigest)
	s.NoError(err)
}

func (s *MultiTestSuite) TestSendEventSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest1.json"))
	s.NoError(err)

	event, err := os.ReadFile(filepath.Join(s.dataDir, "event", "pull1.json"))
	s.NoError(err)

	err = s.client.SendEvent(s.ctx, digest.SHA256, event, manifest, btl, s.getBlobByDigest)
	s.NoError(err)
}

func (s *MultiTestSuite) TestSendSignatureSuccess() {
	btl, err := os.ReadFile(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	err = s.client.SendBottle(s.ctx, digest.SHA256, btl, s.getBlobByDigest)
	s.NoError(err)

	manifest, err := os.ReadFile(filepath.Join(s.dataDir, "manifest", "manifest1.json"))
	s.NoError(err)

	err = s.client.SendManifest(s.ctx, digest.SHA256, manifest, btl, s.getBlobByDigest)
	s.NoError(err)

	// signature, err := os.ReadFile(filepath.Join(s.dataDir, "signature", "signature1.json"))
	// s.NoError(err)
	//
	// err = s.client.SendSignature(s.ctx, digest.SHA256, signature, s.getBlobByDigest)
	// s.NoError(err)
}

func (s *MultiTestSuite) TestListBlobs() {
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

func (s *MultiTestSuite) TestListBottles() {
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

func (s *MultiTestSuite) TestListManifests() {
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

func (s *MultiTestSuite) TestListEvents() {
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

func (s *MultiTestSuite) TestGetBottle() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	getBottleByDigest, err := s.client.GetBottle(s.ctx, digest.Digest(bottleDigest))
	s.NoError(err)
	s.NotEmpty(getBottleByDigest)

	alg := digest.SHA256
	hash := alg.FromBytes(getBottleByDigest)

	// check if both digests are equal
	s.Equal(bottleDigest, hash)
}

func (s *MultiTestSuite) TestGetManifest() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	manifestDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "manifest", "manifest1.json"), "sha256")
	s.NoError(err)

	getManifestByDigest, err := s.client.GetManifest(s.ctx, digest.Digest(manifestDigest))
	s.NoError(err)
	s.NotEmpty(getManifestByDigest)

	alg := digest.SHA256
	hash := alg.FromBytes(getManifestByDigest)

	// check if both digests are equal
	s.Equal(manifestDigest, hash)
}

func (s *MultiTestSuite) TestGetEvent() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	alg := digest.SHA256
	eventDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "event", "pull1.json"), alg)
	s.NoError(err)

	getEventByDigest, err := s.client.GetEvent(s.ctx, eventDigest)
	s.NoError(err)
	s.NotEmpty(getEventByDigest)

	hash := alg.FromBytes(getEventByDigest)

	// check if both digests are equal
	s.Equal(eventDigest, hash)
}

func (s *MultiTestSuite) TestUploadBottle() {
	s.TestUploadBlob()
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "bottle", "index.csv"), false)
	s.NoError(err)
}

func (s *MultiTestSuite) TestUploadManifest() {
	s.TestUploadBottle()
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "manifest", "index.csv"), false)
	s.NoError(err)
}

func (s *MultiTestSuite) TestUploadEvent() {
	s.TestUploadManifest()
	err := s.client.Upload(s.ctx, filepath.Join(s.dataDir, "event", "index.csv"), false)
	s.NoError(err)
}

func (s *MultiTestSuite) TestGetLocations() {
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

func (s *MultiTestSuite) TestBottleSearch() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	selectorval := []string{"type=testing,group=testset"}

	bottleSearch, err := s.client.BottleSearch(s.ctx, selectorval, "image", 7, true)
	s.NoError(err)

	s.NotEmpty(bottleSearch)
}

func (s *MultiTestSuite) TestGetBottlesFromMetric() {
	s.NoError(s.client.UploadAll(s.ctx, s.dataDir, false))

	metric := "training loss"
	selector := []string{"type=testing,myotherkey=myothervalue2", "myotherkey=doesnotexist"}

	metricSearch, err := s.client.GetBottlesFromMetric(s.ctx, selector, metric, 7, true)
	s.NoError(err)

	s.NotEmpty(metricSearch)
}

func TestMultiTestSuite(t *testing.T) {
	suite.Run(t, new(MultiTestSuite))
}
