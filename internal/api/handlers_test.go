package api_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/runtime"

	bottle "gitlab.com/act3-ai/asce/data/schema/pkg/apis/data.act3-ace.io"
	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
	"gitlab.com/act3-ai/asce/go-common/pkg/redact"
	"gitlab.com/act3-ai/asce/go-common/pkg/test"

	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/api"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/dbtest"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/internal/middleware"
	ttest "gitlab.com/act3-ai/asce/data/telemetry/v2/internal/testing"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	client "gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/client"
	"gitlab.com/act3-ai/asce/data/telemetry/v2/pkg/types"
)

type HandlersTestSuite struct {
	suite.Suite
	server  *httptest.Server
	dataDir string
	log     *slog.Logger
	ctx     context.Context
	token   string
}

// Make sure you run `make template` to ensure that the files are all generated in the testdata directory

func (s *HandlersTestSuite) SetupSuite() {
	s.dataDir = filepath.Join("..", "..", "testdata")
	s.log = test.Logger(s.T(), 0)

	s.ctx = logger.NewContext(context.Background(), s.log)
}

func (s *HandlersTestSuite) SetupTest() {
	scheme := runtime.NewScheme()
	s.NoError(bottle.AddToScheme(scheme))

	// Instead of an evn we can use a the "flags" package to create a flag and default it to the env if set or to file::memory: if not
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "file::memory:"
	}
	u, err := url.Parse(dsn)
	s.NoError(err, "could not parse TEST_DSN dsn to URL")

	if u.Scheme == "postgres" {
		// If using postgres, create a temporary database for each test
		testPgDbDsn, cleanup, err := dbtest.CreateTempPostgresDb(s.T().Name(), u.String())
		s.NoError(err, "could not create test database in postgres with DNS %s", u.String())
		u, err = url.Parse(testPgDbDsn)
		s.NoError(err, "could not URL parse test Postgres dsn %s", testPgDbDsn)
		s.T().Cleanup(cleanup)
	}
	myDB, err := db.Open(s.ctx, v1alpha2.Database{
		DSN: redact.SecretURL(u.String()),
	}, scheme)
	s.NoError(err)

	serveMux := http.NewServeMux()
	wrappedServeMux := httputil.LoggingMiddleware(s.log)(middleware.DatabaseMiddleware(myDB)(serveMux))

	a := &api.API{}
	a.Initialize(serveMux, scheme)

	s.server = httptest.NewServer(wrappedServeMux)
}

func (s *HandlersTestSuite) TearDownTest() {
	// Close the server when test finishes
	s.server.Close()
}

func (s *HandlersTestSuite) makeRequest(method, u string, body io.Reader) *http.Request {
	ctx := logger.NewContext(context.Background(), s.log)
	req, err := http.NewRequestWithContext(ctx, method, s.server.URL+u, body)
	s.NoError(err)
	return req
}

func (s *HandlersTestSuite) performRequest(req *http.Request) (int, http.Header, []byte) {
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

func (s *HandlersTestSuite) TestAPI_handleUpload() {
	t := s.T()

	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("sample.txt", digest.SHA256) })
	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("tabular1.csv", digest.SHA256) })
	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("flame_temperature.ipynb", digest.SHA256) })
	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("doc.md", digest.SHA512) })
	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("parent.html", digest.SHA256) })
	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("child.html", digest.SHA256) })
	t.Run("put-blob", func(t *testing.T) { s.testPutBlob("image1.jpg", digest.SHA256) })
	t.Run("put-bottle", func(t *testing.T) { s.testPutBottle() })
	t.Run("put-manifest", func(t *testing.T) { s.testPutManifest(http.StatusCreated) })
	t.Run("put-manifest-again", func(t *testing.T) { s.testPutManifest(http.StatusNoContent) })
	t.Run("put-event", func(t *testing.T) { s.testPutEvent() })
	t.Run("put-signature", func(t *testing.T) { s.testPutSignature() })
}

func (s *HandlersTestSuite) testPutBlob(file string, digestAlg digest.Algorithm) {
	f, err := os.Open(filepath.Join(s.dataDir, "blob", file))
	s.NoError(err)

	req := s.makeRequest("PUT", "/blob?digest-alg="+digestAlg.String(), f)
	req.Header.Set("Content-Type", "application/octet-stream")

	status, hdrs, _ := s.performRequest(req)

	s.Equal(http.StatusCreated, status)
	s.NotEmpty(hdrs.Get(types.HeaderContentDigest))
}

func (s *HandlersTestSuite) testPutBottle() {
	f, err := os.Open(filepath.Join(s.dataDir, "bottle", "bottle1.json"))
	s.NoError(err)

	req := s.makeRequest("PUT", "/bottle", f)
	req.Header.Set("Content-Type", mediatype.MediaTypeBottleConfig)

	status, hdrs, _ := s.performRequest(req)

	s.Equal(http.StatusCreated, status)
	s.NotEmpty(hdrs.Get(types.HeaderContentDigest))
}

func (s *HandlersTestSuite) testPutManifest(expectedStatus int) {
	f, err := os.Open(filepath.Join(s.dataDir, "manifest", "manifest1.json"))
	s.NoError(err)

	req := s.makeRequest("PUT", "/manifest", f)
	req.Header.Set("Content-Type", ocispec.MediaTypeImageManifest)

	status, hdrs, _ := s.performRequest(req)

	s.Equal(expectedStatus, status)
	s.NotEmpty(hdrs.Get(types.HeaderContentDigest))
}

func (s *HandlersTestSuite) testPutEvent() {
	f, err := os.Open(filepath.Join(s.dataDir, "event", "push1.json"))
	s.NoError(err)

	req := s.makeRequest("PUT", "/event", f)
	req.Header.Set("Content-Type", "application/json")

	status, hdrs, _ := s.performRequest(req)

	s.Equal(http.StatusCreated, status)
	s.NotEmpty(hdrs.Get(types.HeaderContentDigest))
}

func (s *HandlersTestSuite) testPutSignature() {
	// PUT signature 1
	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	manifestDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "manifest", "manifest1.json"), "sha256")
	s.NoError(err)

	// TODO use oci image created from other test data.
	signature, err := getTestSignature(manifestDigest, bottleDigest, []byte("test"))
	s.NoError(err)
	sigJSON, err := json.Marshal(*signature)
	s.NoError(err)

	req := s.makeRequest("PUT", "/signature", bytes.NewReader(sigJSON))
	req.Header.Set("Content-Type", "application/json")

	status, hdrs, _ := s.performRequest(req)

	s.Equal(http.StatusCreated, status)
	s.NotEmpty(hdrs.Get(types.HeaderContentDigest))

	// PUT signature 2
	f, err := os.Open(filepath.Join(s.dataDir, "signature", "signature1.json"))
	s.NoError(err)

	req = s.makeRequest("PUT", "/signature", f)
	req.Header.Set("Content-Type", "application/json")

	status, hdrs, _ = s.performRequest(req)

	s.Equal(http.StatusCreated, status)
	s.NotEmpty(hdrs.Get(types.HeaderContentDigest))
}

func (s *HandlersTestSuite) TestAPI_handleGetBottlesFromMetric() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)

	s.NoError(client.Upload(s.ctx, s.server.Client(), filepath.Join(s.dataDir, "blob", "index.csv"), uploadURL, s.token, false))
	s.NoError(client.Upload(s.ctx, s.server.Client(), filepath.Join(s.dataDir, "bottle", "index.csv"), uploadURL, s.token, false))

	u := url.URL{
		Path: "/metric",
		RawQuery: url.Values{
			"metric":   []string{"training loss"},
			"selector": []string{"type=testing,myotherkey=myothervalue2", "myotherkey=doesnotexist"},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)
	s.Equal(http.StatusOK, status)

	// Only expect bottle2
	// TODO actually unmarshal the JSON body and check the bottles.  assert.JSONEq()
	s.NotContains(body, "3.141592654")
	s.Contains(string(body), "52")
	s.Contains(string(body), "sha512")
}

func (s *HandlersTestSuite) TestAPI_handleGetLocation() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path: "/location",
		RawQuery: url.Values{
			"bottle_digest": []string{bottleDigest.String()},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	s.Contains(string(body), "reg2.example.com/bar/somewhere/else")
	s.Contains(string(body), "reg.example.com/foo")
}

func (s *HandlersTestSuite) TestAPI_handleBottleSearch() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	u := url.URL{
		Path: "/search",
		RawQuery: url.Values{
			"selector":    []string{"type=testing,group=testset"},
			"description": []string{"image"},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	s.Contains(string(body), "sha256")
	s.Contains(string(body), "sha512")
}

func (s *HandlersTestSuite) TestAPI_handleContentSearch() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	u := url.URL{
		Path: "/content",
		RawQuery: url.Values{
			"contentDigest": []string{"sha256:0b1de4364cfd94d75e7bda5d0583bcb136d6437c88a36dc06bcd64566a3530ae"},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	s.Contains(string(body), "sha256:625b0528ec90bd34498563b8380db33f2f374256181a62a23a6cdcaf41b19304")
	s.Contains(string(body), "sha256:725b0528ec90bd34498563b8380db33f2f374256181a62a23a6cdcaf41b19304")
}

func (s *HandlersTestSuite) TestAPI_handleGetData() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	blobDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "blob", "image1.png"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path: "/blob",
		RawQuery: url.Values{
			"digest": []string{blobDigest.String()},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)

	s.Contains(string(body), "5")

	{
		// does not exist
		u := url.URL{
			Path: "/blob",
			RawQuery: url.Values{
				"digest": []string{"sha256:deadbeef4cfd94d75e7bda5d0583bcb136d6437c88a36dc06bcd64566a3aaaaa"},
			}.Encode(),
		}
		req := s.makeRequest("GET", u.String(), nil)

		status, _, body := s.performRequest(req)

		s.Equal(http.StatusNotFound, status)
		s.Contains(string(body), "not found")
	}
}

func (s *HandlersTestSuite) TestAPI_handleListData() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	var since time.Time

	u := url.URL{
		Path: "/blob",
		RawQuery: url.Values{
			"since": []string{since.Format(time.RFC3339Nano)},
			"limit": []string{"2"},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	s.Contains(string(body), "sha256")
	s.NotContains(string(body), "null")
}

func (s *HandlersTestSuite) TestAPI_handleGetSignatures() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	manifestDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "manifest", "manifest1.json"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path: "/signatures",
		RawQuery: url.Values{
			"bottle_digest": []string{bottleDigest.String()},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	s.Contains(string(body), manifestDigest.String())
	s.NotContains(string(body), "null")
}

func (s *HandlersTestSuite) TestAPI_handleGetSigValid() {
	uploadURL, err := url.Parse(s.server.URL)
	s.NoError(err)
	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))

	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
	s.NoError(err)

	keyFP, err := ttest.FileDigest(filepath.Join(s.dataDir, "signature", "pub.pem"), "sha256")
	s.NoError(err)

	u := url.URL{
		Path: "/signature/validate",
		RawQuery: url.Values{
			"bottle_digest":   []string{bottleDigest.String()},
			"key_fingerprint": []string{keyFP.String()},
			"trust_level":     []string{"validated"},
		}.Encode(),
	}
	req := s.makeRequest("GET", u.String(), nil)

	status, _, body := s.performRequest(req)

	s.Equal(http.StatusOK, status)
	s.Contains(string(body), keyFP.String())
	s.Contains(string(body), bottleDigest.String())
	s.NotContains(string(body), "null")
}

// func (s *HandlersTestSuite) TestAPI_handleGetSigIdent() {
// 	uploadURL, err := url.Parse(s.server.URL)
// 	s.NoError(err)
// 	s.NoError(client.UploadAll(s.ctx, s.server.Client(), s.dataDir, uploadURL, s.token, false))
//
// 	bottleDigest, err := ttest.FileDigest(filepath.Join(s.dataDir, "bottle", "bottle1.json"), "sha256")
// 	s.NoError(err)
//
// 	keyFP, err := ttest.FileDigest(filepath.Join(s.dataDir, "signature", "test-public-key-cosign.pub"), "sha256")
// 	s.NoError(err)
//
// 	u := url.URL{
// 		Path: "/signature/validate",
// 		RawQuery: url.Values{
// 			"bottle_digest": []string{bottleDigest.String()},
// 			"key_fingerprint": []string{keyFP.String()},
// 		}.Encode(),
// 	}
// 	req := s.makeRequest("GET", u.String(), nil)
//
// 	status, _, body := s.performRequest(req)
//
// 	s.Equal(http.StatusOK, status)
// 	s.Contains(string(body), "sha256")
// 	s.NotContains(string(body), "null")
// }

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

// getTestSignature returns a SignatureSummary with valid signature.
func getTestSignature(manifestDigest, bottleDigest digest.Digest, signaturePayload []byte) (*types.SignaturesSummary, error) {
	// Create an ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("could not generate private key: %w", err)
	}

	// Hash our payload
	hash := sha256.Sum256(signaturePayload)

	signatureRaw, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("could not sign data: %w", err)
	}

	// Make sure we have a valid signature
	valid := ecdsa.VerifyASN1(&privateKey.PublicKey, hash[:], signatureRaw)
	if !valid {
		return nil, fmt.Errorf("could not verify test signature: %w", err)
	}

	signature := base64.StdEncoding.EncodeToString(signatureRaw)

	x509EncodedPublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("could not encode public key to x509: %w", err)
	}
	pemEncodedPublicKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509EncodedPublicKey,
	})

	sd := &types.SignaturesSummary{
		SubjectManifest: manifestDigest,
		SubjectBottleID: bottleDigest,
		Signatures: []types.SignatureDetail{
			{
				SignatureType: "dev.cosignproject.cosign/signature",
				Signature:     signature,
				Descriptor: ocispec.Descriptor{
					MediaType: "application/vnd.dev.cosign.simplesigning.v1+json",
					Digest:    digest.NewDigestFromBytes(digest.SHA256, hash[:]),
					Size:      int64(len(signature)),
				},
				PublicKey: string(pemEncodedPublicKey),
				Annotations: map[string]string{
					"test": "true",
				},
			},
		},
	}

	return sd, nil
}
