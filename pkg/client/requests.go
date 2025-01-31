package client

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.com/act3-ai/asce/data/schema/pkg/mediatype"
	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
	"gitlab.com/act3-ai/asce/data/telemetry/v3/pkg/types"
)

// TODO audit file closing

// NewRequestClientFromConfig creates a http.Client from the given configuration.
func NewRequestClientFromConfig(conf v1alpha2.ClientConfiguration) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	c := &http.Client{
		Jar: jar,
	}

	// Add cookies (for auth)
	for _, location := range conf.Locations {
		cookies := make([]*http.Cookie, 0, len(location.Cookies))
		for k, v := range location.Cookies {
			cookies = append(cookies, &http.Cookie{
				Name:   k,
				Value:  string(v),
				Secure: true,
			})
		}
		u, err := url.Parse(string(location.URL))
		if err != nil {
			return nil, fmt.Errorf("parsing telemetry location URL: %w", err)
		}
		jar.SetCookies(u, cookies)
	}

	return c, nil
}

// getLatestTimestamp reads the index.latest file and parses the timestamp.
func getLatestTimestamp(latestFile string) (*time.Time, error) {
	// use that instead of "since"
	b, err := os.ReadFile(latestFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read latest timestamp: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, string(b))
	if err != nil {
		return nil, fmt.Errorf("unable to parse latest timestamp: %w", err)
	}
	return &t, nil
}

// Download raw objects.
func Download(ctx context.Context, c *http.Client, since time.Time, fromLatest bool, batchSize int, file string, u *url.URL, token string) error {
	dir := filepath.Dir(file)
	objType := filepath.Base(dir)
	log := logger.FromContext(ctx).With("type", objType)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("download operation failed to make directories: %w", err)
	}

	// open for writing and appending (create if not already existing)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("download failed to open file for writing: %w", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)

	latestFile := filepath.Join(dir, "index.latest")
	latest := since
	if fromLatest {
		if t, err := getLatestTimestamp(latestFile); err != nil {
			log.InfoContext(ctx, "Failed to get the latest timestamp", "error", err)
		} else {
			log.InfoContext(ctx, "Using latest", "since", t)
			latest = *t
		}
	}

	for {
		results, err := doListRequest(ctx, c, u, objType, latest, batchSize, WithBearerTokenAuth(token))
		if err != nil {
			return err
		}

		for _, result := range results {
			if err := writeRecordToCsv(log, result, dir, w); err != nil {
				return err
			}

			// update the latest timestamp
			latest = result.CreatedAt
		}
		w.Flush()

		if len(results) < batchSize {
			// we are at the end
			break
		}

		// we write out the latestFile here, as well, to better handle premature network failure
		if err := os.WriteFile(latestFile, []byte(latest.Format(time.RFC3339Nano)), os.ModePerm); err != nil {
			return fmt.Errorf("unable to incrementally write the latest timestamp file: %w", err)
		}
	}

	if err := os.WriteFile(latestFile, []byte(latest.Format(time.RFC3339Nano)), os.ModePerm); err != nil {
		return fmt.Errorf("unable to write the final latest timestamp file: %w", err)
	}

	return nil
}

// DownloadAll raw files given in the file of all types.
func DownloadAll(ctx context.Context, c *http.Client, since time.Time, fromLatest bool, batchSize int, path string, u *url.URL, token string) error {
	// iterate in reverse order
	for i := len(types.TopologicalOrderingOfTypes) - 1; i >= 0; i-- {
		objType := types.TopologicalOrderingOfTypes[i]
		file := filepath.Join(path, objType, "index.csv")
		if err := Download(ctx, c, since, fromLatest, batchSize, file, u, token); err != nil {
			return err
		}
	}
	return nil
}

// Upload all raw files given in the file.
func Upload(ctx context.Context, c *http.Client, file string, u *url.URL, token string, skipInvalid bool) error {
	dir := filepath.Dir(file)
	objType := filepath.Base(dir)
	log := logger.FromContext(ctx).With("type", objType)

	return processIndexFile(file, func(datafile string, dgst digest.Digest, data []byte) error {
		log.InfoContext(ctx, "Uploading", "objType", objType, "file", datafile, "algorithm", dgst.Algorithm())
		err := doPutRequest(ctx, c, u, objType, data, dgst.Algorithm(), WithBearerTokenAuth(token))
		if err != nil {
			target := &types.MissingDigestsError{}
			if errors.As(err, &target) && skipInvalid {
				log.InfoContext(ctx, "failed to upload. skipping", "file", datafile, "error", err)
				return nil
			}
		}
		return err
	})
}

// UploadAll raw files given in the file of all types.
func UploadAll(ctx context.Context, c *http.Client, path string, u *url.URL, token string, skipInvalid bool) error {
	for _, objType := range types.TopologicalOrderingOfTypes {
		file := filepath.Join(path, objType, "index.csv")
		if err := Upload(ctx, c, file, u, token, skipInvalid); err != nil {
			return err
		}
	}
	return nil
}

type apiEntry struct {
	Path        string
	ContentType string
}

var apiMapper = map[string]apiEntry{
	"blob":      {"/blob", "application/octet-stream"},
	"bottle":    {"/bottle", mediatype.MediaTypeBottleConfig},
	"manifest":  {"/manifest", ocispec.MediaTypeImageManifest},
	"event":     {"/event", "application/json"},
	"signature": {"/signature", "application/json"},
}

// doPutRequest actually makes the request to the handler if given, otherwise to the url in the request.
func doPutRequest(ctx context.Context, c *http.Client,
	u *url.URL,
	objType string,
	data []byte, alg digest.Algorithm, options ...AuthRequestOptsFunc,
) error {
	log := logger.FromContext(ctx).WithGroup("put-request")

	entry, exists := apiMapper[objType]
	if !exists {
		return fmt.Errorf("unknown api type \"%s\"", objType)
	}

	// compute the digest
	dgst := alg.FromBytes(data)

	uu := *u
	uu.Path += entry.Path
	uu.RawQuery = url.Values{"digest": []string{dgst.String()}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uu.String(), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unable create put request: %w", err)
	}
	req.Header.Set("Content-Type", entry.ContentType)

	for _, fn := range options {
		if err := fn(req); err != nil {
			return err
		}
	}

	log.DebugContext(ctx, "Request", "url", req.URL)

	// execute the remote request (use the HTTP protocol)
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP put request: %w", err)
	}
	defer res.Body.Close()
	log.DebugContext(ctx, "put request http handler", "instance", res.Header.Get(httputil.HeaderInstance))

	// > 300 status code
	if res.StatusCode >= http.StatusMultipleChoices {
		return processErrorResponse(res)
	}
	if res.Header.Get(types.HeaderContentDigest) == "" {
		return fmt.Errorf("expected header %s is missing", types.HeaderContentDigest)
	}

	return res.Body.Close()
}

// processResponse will process the responseData byte and unmarshall to give.
func processErrorResponse(response *http.Response) error {
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read the rest of the HTTP response body: %w", err)
	}
	if response.StatusCode == http.StatusPreconditionFailed && response.Header.Get("Content-Type") == httputil.MediaTypeProblem {
		var result types.MissingDigestsError
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
		return &result
	}
	return fmt.Errorf("unknown HTTP error: %s, %s", response.Status, string(data))
}

// doGetRequest will make a get request with its digest.
func doGetRequest(ctx context.Context, c *http.Client,
	u *url.URL,
	objType string,
	dgst digest.Digest, options ...AuthRequestOptsFunc,
) ([]byte, error) {
	log := logger.FromContext(ctx).WithGroup("get-request")
	ctx = logger.NewContext(ctx, log)

	entry, exists := apiMapper[objType]
	if !exists {
		return nil, fmt.Errorf("unknown api type \"%s\"", objType)
	}

	uu := *u
	uu.Path += entry.Path
	uu.RawQuery = url.Values{"digest": []string{dgst.String()}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uu.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create GET request: %w", err)
	}

	for _, fn := range options {
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	body, err := doRequest(req, c)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// doListRequest actually makes the request to the handler if given, otherwise to the url in the request.
func doListRequest(ctx context.Context, c *http.Client,
	u *url.URL,
	objType string,
	since time.Time, limit int, options ...AuthRequestOptsFunc,
) ([]types.ListResultEntry, error) {
	log := logger.FromContext(ctx).WithGroup("list-request")
	ctx = logger.NewContext(ctx, log)

	entry, exists := apiMapper[objType]
	if !exists {
		return nil, fmt.Errorf("unknown api type \"%s\"", objType)
	}

	uu := *u
	uu.Path += entry.Path
	uu.RawQuery = url.Values{
		"since": []string{since.Format(time.RFC3339Nano)},
		"limit": []string{strconv.Itoa(limit)},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uu.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create list request: %w", err)
	}

	for _, fn := range options {
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	body, err := doRequest(req, c)
	if err != nil {
		return nil, err
	}

	results := struct {
		Results []types.ListResultEntry
	}{}
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	return results.Results, nil
}

func processIndexFile(file string, f func(datafile string, dgst digest.Digest, data []byte) error) error {
	dir := filepath.Dir(file)
	csvfile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("unable to open index file: %w", err)
	}
	defer csvfile.Close()

	// parse the file
	r := csv.NewReader(csvfile)
	// allow variable length rows
	r.FieldsPerRecord = -1

	for {
		// Read each record from csv
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if len(record) < 2 {
			return errors.New("invalid test file format")
		}
		// record[0] is the file and record[1] and after is the algorithm

		for _, algStr := range record[1:] {
			alg := digest.Algorithm(algStr)
			if !alg.Available() {
				return fmt.Errorf("unknown digest algorithm %s", algStr)
			}

			// read the data from file
			datafile := filepath.Join(dir, record[0])
			data, err := os.ReadFile(datafile)
			if err != nil {
				return fmt.Errorf("unable to read file references in index at %s: %w", file, err)
			}

			// compute the digest
			dgst := alg.FromBytes(data)

			// process this file
			if err := f(datafile, dgst, data); err != nil {
				return err
			}
		}
	}

	return csvfile.Close()
}

// GetLocations returns the location response for the bottle digest passed in.
func GetLocations(ctx context.Context, c *http.Client, handler http.Handler,
	u *url.URL,
	dgst digest.Digest, options ...AuthRequestOptsFunc,
) ([]types.LocationResponse, error) {
	log := logger.FromContext(ctx).WithGroup("get-locations")
	ctx = logger.NewContext(ctx, log)

	uu := *u
	uu.Path += "/location"
	uu.RawQuery = url.Values{
		"bottle_digest": []string{dgst.String()},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uu.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create locations request: %w", err)
	}

	for _, fn := range options {
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	body, err := doRequest(req, c)
	if err != nil {
		return nil, err
	}

	results := struct {
		Results []types.LocationResponse
	}{}
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	return results.Results, nil
}

// GetBottlesFromMetric will call the handleGetBottlesFromMetric from api.
func GetBottlesFromMetric(ctx context.Context, c *http.Client, handler http.Handler,
	u *url.URL, selector []string, metric string, limit int, desc bool, options ...AuthRequestOptsFunc,
) ([]byte, error) {
	log := logger.FromContext(ctx).WithGroup("get-bottle-from-metric")
	ctx = logger.NewContext(ctx, log)

	uu := *u
	uu.Path += "/metric"
	uu.RawQuery = url.Values{
		"selector":   selector,
		"metric":     []string{metric},
		"limit":      []string{strconv.Itoa(limit)},
		"descending": []string{strconv.FormatBool(desc)},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uu.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create metrics request: %w", err)
	}

	for _, fn := range options {
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	body, err := doRequest(req, c)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// BottleSearch will make a call to the BottleSearch Handler.
func BottleSearch(ctx context.Context, c *http.Client, handler http.Handler,
	u *url.URL, selectors []string, description string, limit int, digestOnly bool, options ...AuthRequestOptsFunc,
) ([]types.SearchResult, error) {
	log := logger.FromContext(ctx).WithGroup("bottle-search")
	ctx = logger.NewContext(ctx, log)

	uu := *u
	uu.Path += "/search"
	uu.RawQuery = url.Values{
		"description": []string{description},
		"limit":       []string{strconv.Itoa(limit)},
		"selector":    selectors,
		"digestOnly":  []string{strconv.FormatBool(digestOnly)},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uu.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create search request: %w", err)
	}

	for _, fn := range options {
		if err := fn(req); err != nil {
			return nil, err
		}
	}

	body, err := doRequest(req, c)
	if err != nil {
		return nil, err
	}

	results := struct {
		Results []types.SearchResult
	}{}
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	return results.Results, nil
}

func writeRecordToCsv(log *slog.Logger, result types.ListResultEntry, dir string, w *csv.Writer) error {
	// We just pick the first one (these are sorted by the server)
	primaryDigest := result.Digests[0]

	log.Info("Saving", "digest", primaryDigest)

	// by default there is a ":" in the name but that is not a valid filename on some platforms (Windows)
	filename := primaryDigest.Algorithm().String() + "-" + primaryDigest.Encoded()
	if err := os.WriteFile(filepath.Join(dir, filename), result.Data, 0o666); err != nil {
		return fmt.Errorf("unable to open index file for writing: %w", err)
	}

	// construct the record (filename, digestAlg1, digestAlg2, ...)
	record := []string{filename}
	for _, dgst := range result.Digests {
		record = append(record, dgst.Algorithm().String())
	}

	return w.Write(record)
}

func doRequest(req *http.Request, c *http.Client) ([]byte, error) {
	ctx := req.Context()
	log := logger.FromContext(ctx)

	// execute the remote request (use the HTTP protocol)
	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform HTTP request: %w", err)
	}
	defer res.Body.Close()

	log.DebugContext(ctx, "handling request with http handler", "instance", res.Header.Get(httputil.HeaderInstance))

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	// >= 300
	if res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("failed loading: %d, %s", res.StatusCode, body)
	}

	return body, nil
}
