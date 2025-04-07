// Package testing helps generate test data for use in testing telemetry
package testing

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/opencontainers/go-digest"

	"github.com/act3-ai/bottle-schema/pkg/mediatype"
	"github.com/act3-ai/bottle-schema/pkg/selectors"
)

func fileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return -1, err
	}
	return info.Size(), nil
}

// FileDigest returns the digest for the given file.
func FileDigest(filename string, algorithm digest.Algorithm) (digest.Digest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to digest file: %w", err)
	}
	return digestWithAlgorithm(algorithm, data)
}

func digestWithAlgorithm(algorithm digest.Algorithm, data []byte) (digest.Digest, error) {
	if !algorithm.Available() {
		return "", fmt.Errorf("digest \"%s\" not available", algorithm)
	}
	return algorithm.FromBytes(data), nil
}

// bottleURI takes a hash scheme, hash algorithm, bottle hash, mediatype, and a part selector string to return a URL encoded URI
// hashScheme can be one of "bottle", or "hash"
// dgst is a hash algorithm and a hash separated by a colon like "sha256:xyz"
// partSelectors (optional) are part selector strings like  "partkey!=value1,mykey=value2" and "partkey2=45".
func bottleURI(hashScheme string, dgst digest.Digest, partSelectors ...string) (string, error) {
	// ensure we got a valid digest
	if err := dgst.Validate(); err != nil {
		return "", fmt.Errorf("could not parse digest: digest \"%s\" was not the correct format", dgst)
	}
	alg := dgst.Algorithm()
	if !alg.Available() {
		return "", fmt.Errorf("digest \"%s\" not available", alg)
	}

	u := url.URL{}
	queryParams := url.Values{}
	u.Scheme = hashScheme
	switch u.Scheme {
	case "bottle":
		u.Opaque = dgst.String()
	case "hash":
		u.Host = alg.String()
		u.Path = dgst.Encoded()
		queryParams.Add("type", mediatype.MediaTypeBottleConfig)
	default:
		return "", fmt.Errorf("could not parse BottleURI: hashScheme \"%s\" was not one of \"bottle\" or \"hash\"", hashScheme)
	}

	// Construct the query parameters
	partSelectorLabelSet, err := selectors.Parse(partSelectors)
	if err != nil {
		return "", err
	}

	for _, sel := range partSelectorLabelSet {
		queryParams.Add("selector", sel.String())
	}

	u.RawQuery = queryParams.Encode()
	return u.String(), nil
}

func templateFile(tmpl, out string) error {
	// TODO These functions should really be relative to where the templating is happening
	// So we should probably pass in a path for these functions to use to resolve file names

	// TemplateFuncs are used during templating the .tmpl files
	templateFuncs := template.FuncMap{
		"FileDigest": FileDigest,
		"ReadFile":   os.ReadFile,
		"Digest":     digestWithAlgorithm,
		"FileSize":   fileSize,
		"BottleURI":  bottleURI,
	}

	t, err := template.New("root").
		Funcs(sprig.TxtFuncMap()).
		Funcs(templateFuncs).
		ParseFiles(tmpl)
	if err != nil {
		return fmt.Errorf("unable to parse templates from %s: %w", tmpl, err)
	}

	raw, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("unable to write templated output file: %w", err)
	}
	defer raw.Close()

	// We do not set any values since we use functions to get the digests
	if err := t.ExecuteTemplate(raw, filepath.Base(tmpl), nil); err != nil {
		return fmt.Errorf("unable to execute template: %w", err)
	}

	return raw.Close()
}

// ProcessTemplates converts all .tmpl files into raw files for the given directory.
func ProcessTemplates(base string) error {
	files, err := os.ReadDir(base)
	if err != nil {
		return fmt.Errorf("unable to read template file: %w", err)
	}

	for _, file := range files {
		name := file.Name()
		if file.IsDir() || filepath.Ext(name) != ".tmpl" {
			continue
		}
		sname := name[:len(name)-5]

		if err := templateFile(filepath.Join(base, name), filepath.Join(base, sname)); err != nil {
			return err
		}
	}

	return nil
}
