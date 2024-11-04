package webapp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/aohorodnyk/mimeheader"
	"github.com/hetiansu5/urlquery"
	"github.com/opencontainers/go-digest"

	hub "git.act3-ace.com/ace/hub/api/v6/pkg/apis/hub.act3-ace.io/v1beta1"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

// ViewerSpecList is just a list of ViewerSpecs.
type ViewerSpecList []v1alpha2.ViewerSpec

// FilterAndSort returns only viewers supporting mediaType in sorted order (best first).
func (vsl *ViewerSpecList) FilterAndSort(mediaType string) ViewerSpecList {
	type viewerSpecWithQuality struct {
		ViewerSpec v1alpha2.ViewerSpec
		Quality    float32
	}

	// Filter and rank the available viewers for this artifact type.
	matchedSpecs := make([]viewerSpecWithQuality, 0, len(*vsl))
	for _, vs := range *vsl {
		ah := mimeheader.ParseAcceptHeader(vs.Accept) // TODO this does not need to be done every request (can be done on startup)
		accept, _, matched := ah.Negotiate([]string{mediaType}, "")
		if matched {
			matchedSpecs = append(matchedSpecs, viewerSpecWithQuality{vs, accept.Quality})
		}
	}

	sort.Slice(matchedSpecs, func(i, j int) bool {
		return matchedSpecs[i].Quality > matchedSpecs[j].Quality
	})

	// strip out the specs
	specs := make(ViewerSpecList, len(matchedSpecs))
	for i, v := range matchedSpecs {
		specs[i] = v.ViewerSpec
	}

	return specs
}

// ViewerLink represents a viewer that will be presented to the user.
type ViewerLink struct {
	Location string
	Viewer   string
	URL      string
}

const (
	// LabelViewerPrefix is used to denote the annotations that describe viewers.  The part after the "/" in the key is used as the name. The value must be a JSON document representing ViewerSpec.
	LabelViewerPrefix = "viewer.data.act3-ace.io"

	// HubBottleName is the name of the bottle that is used when mounting in ACE Hub
	// This must not include any special characters because it needs to be used as an env var in ACE Hub.
	HubBottleName = "dataset"
)

// GetViewerSpecList will find annotations that denote viewers and parse them.
func (a *WebApp) GetViewerSpecList(bottle db.Bottle) ViewerSpecList {
	annotations := bottle.Annotations
	viewers := make([]v1alpha2.ViewerSpec, 0, len(annotations)+len(a.defaultViewerSpecs))
	for _, annotation := range annotations {
		if path.Dir(annotation.Key) != LabelViewerPrefix {
			continue
		}
		v := v1alpha2.ViewerSpec{}
		if err := json.Unmarshal([]byte(annotation.Value), &v); err != nil {
			// eat errors
			a.log.Error("Unable to decode viewer specification, skipping", "key", annotation.Key, "value", annotation.Value, "error", err)
			continue
		}
		v.Name = path.Base(annotation.Key)

		viewers = append(viewers, v)
	}
	// Add in the common viewer specs
	viewers = append(viewers, a.defaultViewerSpecs...)
	return ViewerSpecList(viewers)
}

// FindViewers computes the viewers for each ACE Hub instance and provided spec.  Each ViewerLink will mount the bottle using the selectors.
func (a *WebApp) FindViewers(specs []v1alpha2.ViewerSpec, bottle digest.Digest, partSelectors []string, artifact *db.PublicArtifact) []ViewerLink {
	viewers := make([]ViewerLink, 0, len(a.hubInstances))
	for _, hubInstance := range a.hubInstances {
		for _, spec := range specs {
			log := a.log.With("hub", hubInstance.Name, "image", spec.ACEHub.Image)

			u, err := getViewerURL(spec, hubInstance.URL, bottle, partSelectors, artifact)
			if err != nil {
				log.Error("could not get viewer URL", "error", err)
				continue
			}

			// Only for logging purposes
			if log.Enabled(context.Background(), slog.LevelError) {
				decoded, err := url.QueryUnescape(u.RawQuery)
				if err != nil {
					a.log.Error("Unable to decode query string", "error", err)
					continue
				}

				log.Info("Viewer Link", "encoded", u.RawQuery, "decoded", decoded)
			}

			viewers = append(viewers, ViewerLink{
				Location: hubInstance.Name,
				Viewer:   spec.Name,
				URL:      u.String(),
			})
		}
	}
	return viewers
}

func addArtifactEnvs(newSpec *hub.HubEnvTemplateSpec, artifact *db.PublicArtifact) {
	if newSpec.Env == nil {
		newSpec.Env = make(map[string]string, 1)
	}

	// NOTE this assumes a path convention from ACE Hub
	newSpec.Env["ACE_OPEN_PATH"] = path.Join("/ace", "bottle", HubBottleName, artifact.Path)

	// These might be handy as well
	newSpec.Env["ACE_OPEN_MEDIATYPE"] = artifact.MediaType
	newSpec.Env["ACE_OPEN_NAME"] = artifact.Name
	newSpec.Env["ACE_OPEN_DIGEST"] = artifact.Digest.String()
}

func getViewerURL(spec v1alpha2.ViewerSpec, hubInstanceURL string, bottle digest.Digest, partSelectors []string, artifact *db.PublicArtifact) (*url.URL, error) {
	// Make a shallow copy of the viewer spec
	newSpec := spec.ACEHub

	// Add the bottle
	if newSpec.Bottles == nil {
		newSpec.Bottles = make([]hub.BottleSpec, 0, 1)
	}
	newSpec.Bottles = append(newSpec.Bottles, hub.BottleSpec{
		Name:      HubBottleName,
		BottleRef: bottle.String(),
		Selector:  partSelectors,
	})

	if artifact != nil {
		// This viewer is for a specific artifact so we include that information
		addArtifactEnvs(&newSpec, artifact)
	}

	// parse resources separately so resource to string conversion is correct
	resourceQueryParams := url.Values{}
	if newSpec.Resources.Requests != nil {
		requests := newSpec.Resources.Requests
		newSpec.Resources.Requests = nil
		for k, v := range requests {
			resourceQueryParams.Add(fmt.Sprintf("resources[requests][%s]", k.String()), (&v).String())
			resourceQueryParams.Add(fmt.Sprintf("resources[limits][%s]", k.String()), (&v).String())
		}
	}

	if newSpec.Resources.Limits != nil {
		limits := newSpec.Resources.Limits
		newSpec.Resources.Limits = nil
		for k, v := range limits {
			resourceQueryParams.Add(fmt.Sprintf("resources[limits][%s]", k.String()), (&v).String())
		}
	} else {
		newSpec.Resources.Limits = nil
		for k, v := range newSpec.Resources.Requests {
			resourceQueryParams.Add(fmt.Sprintf("resources[limits][%s]", k.String()), (&v).String())
		}
	}

	urlqueryEncoder := urlquery.NewEncoder(urlquery.WithNeedEmptyValue(true))
	ds, err := urlqueryEncoder.Marshal(newSpec)
	if err != nil {
		return nil, fmt.Errorf("skipping viewer link: %w", err)
	}
	encoded := string(ds)
	specQueryParams, err := url.ParseQuery(encoded)
	if err != nil {
		return nil, fmt.Errorf("could not parse encoded spec query values: %w", err)
	}

	for k, vals := range resourceQueryParams {
		for _, v := range vals {
			specQueryParams.Add(k, v)
		}
	}

	// make the query parameters lowercased
	lowercaseSpecQueryParams := setQueryParamKeysLowercase(specQueryParams)
	// remove empty query params
	updatedSpecQueryParams := removeEmptyQueryParams(lowercaseSpecQueryParams)

	u, err := url.Parse(hubInstanceURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing ACE Hub URL: %w", err)
	}

	u.RawQuery = updatedSpecQueryParams.Encode()
	u.Path = "/environments/0"

	return u, nil
}

func setQueryParamKeysLowercase(params url.Values) url.Values {
	getFirstCharLowercase := func(s string) string {
		if len(s) == 0 {
			return ""
		}
		firstChar := s[0]
		lowerCaseFirstChar := strings.ToLower(string(firstChar))
		return fmt.Sprintf("%s%s", lowerCaseFirstChar, s[1:])
	}
	lowercaseSpecQueryParams := url.Values{}
	for k, vals := range params {
		lowercaseKey := ""
		paramArrayKeys := strings.SplitAfter(k, "[")
		for _, paramArrayKey := range paramArrayKeys {
			lowercaseKey = fmt.Sprintf("%s%s", lowercaseKey, getFirstCharLowercase(paramArrayKey))
		}
		for _, v := range vals {
			lowercaseSpecQueryParams.Add(lowercaseKey, v)
		}
	}
	return lowercaseSpecQueryParams
}

func removeEmptyQueryParams(params url.Values) url.Values {
	updatedQueryParams := url.Values{}
	for k, vals := range params {
		for _, v := range vals {
			if len(v) > 0 {
				updatedQueryParams.Add(k, v)
			}
		}
	}
	return updatedQueryParams
}

// GetArtifactViewers finds the viewers for each artifact by matching viewerSpecs to artifact's media type.
func (a *WebApp) GetArtifactViewers(specs ViewerSpecList, bottle digest.Digest, partSelectors []string, artifacts []db.PublicArtifact) map[string][]ViewerLink {
	artifactViewers := make(map[string][]ViewerLink, len(artifacts))
	for _, artifact := range artifacts {
		matches := specs.FilterAndSort(artifact.MediaType)
		viewers := a.FindViewers(matches, bottle, partSelectors, &artifact)
		artifactViewers[artifact.Path] = viewers
	}
	return artifactViewers
}
