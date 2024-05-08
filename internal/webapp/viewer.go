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

	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
)

// ViewerSpecList is just a list of ViewerSpecs.
type ViewerSpecList []v1alpha1.ViewerSpec

// FilterAndSort returns only viewers supporting mediaType in sorted order (best first).
func (vsl *ViewerSpecList) FilterAndSort(mediaType string) ViewerSpecList {
	type viewerSpecWithQuality struct {
		ViewerSpec v1alpha1.ViewerSpec
		Quality    float32
	}

	// Filter and rank the available viewers for this artifact type
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
	viewers := make([]v1alpha1.ViewerSpec, 0, len(annotations)+len(a.defaultViewerSpecs))
	for _, annotation := range annotations {
		if path.Dir(annotation.Key) != LabelViewerPrefix {
			continue
		}
		v := v1alpha1.ViewerSpec{}
		if err := json.Unmarshal([]byte(annotation.Value), &v); err != nil {
			// eat errors
			a.log.Error("Unable to decode viewer specification, skipping", "key", annotation.Key, "value", annotation.Value, "error", err) //nolint:sloglint
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
func (a *WebApp) FindViewers(specs []v1alpha1.ViewerSpec, bottle digest.Digest, partSelectors []string, artifact *db.PublicArtifact) []ViewerLink {
	viewers := make([]ViewerLink, 0, len(a.hubInstances))
	for _, hub := range a.hubInstances {
		for _, spec := range specs {
			log := a.log.With("hub", hub.Name, "image", spec.ACEHub.Image)

			u, err := getViewerURL(spec, hub, bottle, partSelectors, artifact)
			if err != nil {
				log.Error("could not get viewer URL", "error", err) //nolint:sloglint
				continue
			}

			// Only for logging purposes
			if log.Enabled(context.Background(), slog.LevelError) {
				decoded, err := url.QueryUnescape(u.RawQuery)
				if err != nil {
					a.log.Error("Unable to decode query string", "error", err) //nolint:sloglint
					continue
				}

				log.Info("Viewer Link", "encoded", u.RawQuery, "decoded", decoded) //nolint:sloglint
			}

			viewers = append(viewers, ViewerLink{
				Location: hub.Name,
				Viewer:   spec.Name,
				URL:      u.String(),
			})
		}
	}
	return viewers
}

func addArtifactEnvs(newSpec *v1alpha1.ACEHubLaunchTemplate, artifact *db.PublicArtifact) {
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

func getViewerURL(spec v1alpha1.ViewerSpec, hub v1alpha1.ACEHubInstance, bottle digest.Digest, partSelectors []string, artifact *db.PublicArtifact) (*url.URL, error) {
	// Make a shallow copy
	newSpec := spec.ACEHub

	// Add the bottle
	if newSpec.Bottles == nil {
		newSpec.Bottles = make([]v1alpha1.BottleSpec, 0, 1)
	}
	newSpec.Bottles = append(newSpec.Bottles, v1alpha1.BottleSpec{
		Name:     HubBottleName,
		Bottle:   "bottle:" + bottle.String(),
		Selector: strings.Join(partSelectors, "|"),
	})

	if artifact != nil {
		// This viewer is for a specific artifact so we include that information

		addArtifactEnvs(&newSpec, artifact)
	}

	newSpec.HubName = "" // We want the user to select the name in ACE Hub
	newSpec.Replicas = 1

	u, err := url.Parse(hub.URL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing ACE Hub URL: %w", err)
	}
	// qs := u.Query()

	// https://hub.lion.act3-ace.ai/environments/0?bottles[0][name]=mybottle&bottles[0][bottle]=reg.git.bottle&bottles[0][selector]=foo=bar|dog=cat,x=y&replicas=3&image=sdfsdfsadfsdf&hubName=foo&proxyType=normal&resources[nvidia.com/sharedgpu]=2&shm=64Mi&env[MyENV]=Value&startScript[ACE_START_SCRIPT]=My%20script

	// NOTE that this query string encoding does not allow you to infer the type of the value from the query string.
	// for example foo=1 could have a type of int(1), bool (true), or a string "1".
	// I think a way to encode the type properly is to ensure that the value is a valid JSON value.  So 1 is an int, true is a boolean and "1" is a string

	ds, err := urlquery.Marshal(newSpec)
	if err != nil {
		return nil, fmt.Errorf("Skipping viewer link: %w", err)
	}
	encoded := string(ds)

	u.RawQuery = encoded
	u.Path = "/environments/0"

	return u, nil
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
