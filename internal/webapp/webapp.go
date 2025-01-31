// Package webapp implements the web interface
package webapp

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"code.cloudfoundry.org/bytefmt"
	"github.com/Masterminds/sprig/v3"

	// "github.com/fsnotify/fsnotify".

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"gitlab.com/act3-ai/asce/data/telemetry/v3/pkg/apis/config.telemetry.act3-ace.io/v1alpha2"
)

//go:embed assets
var defaultAssetFS embed.FS

// WebApp implements the HTML web application.
type WebApp struct {
	templates     *template.Template
	templatesLock sync.RWMutex
	staticFS      fs.FS
	templateFS    fs.FS
	ipynbFS       fs.FS
	// watcher            *fsnotify.Watcher
	jupyter            string
	log                *slog.Logger
	hubInstances       []v1alpha2.ACEHubInstance
	defaultViewerSpecs []v1alpha2.ViewerSpec
	globalValues       globalValues
}

// NewWebApp creates the WebApp.
func NewWebApp(conf v1alpha2.WebApp, log *slog.Logger, version string) (*WebApp, error) {
	a := &WebApp{
		log:     log.WithGroup("webapp"),
		jupyter: conf.JupyterExecutable,
	}

	var assetFS fs.FS
	if conf.AssetDir != "" {
		assetFS = os.DirFS(conf.AssetDir)
	} else {
		// get sub filesystem to remove "assets" path prefix
		afs, err := fs.Sub(defaultAssetFS, "assets")
		if err != nil {
			return nil, fmt.Errorf("could not create subFS for \"assets\": %w", err)
		}
		assetFS = afs
	}

	staticFS, err := fs.Sub(assetFS, "static")
	if err != nil {
		return nil, fmt.Errorf("could not find asset subdirectory \"static\": %w", err)
	}
	a.staticFS = staticFS

	templateFS, err := fs.Sub(assetFS, "templates")
	if err != nil {
		return nil, fmt.Errorf("could not find asset subdirectory \"template\": %w", err)
	}
	a.templateFS = templateFS

	ipynbFS, err := fs.Sub(assetFS, "ipynb")
	if err != nil {
		a.log.Warn("could not get ipynbFS from assetDir", "msg", err.Error())
	}
	a.ipynbFS = ipynbFS

	t, err := a.parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	a.templates = t

	// Grab some configuration
	a.hubInstances = conf.ACEHubs

	a.defaultViewerSpecs = conf.Viewers

	a.globalValues = globalValues{
		DefaultBottleSelectors: conf.DefaultBottleSelectors,
		Version:                version,
	}

	// Disabling watcher for now
	// if err := a.startWatching(); err != nil {
	// 	return nil, fmt.Errorf("starting watcher: %w", err)
	// }

	return a, nil
}

// Initialize the routes.
func (a *WebApp) Initialize(serveMux *http.ServeMux) {
	serveMux.Handle("GET /static/", http.StripPrefix("/static", http.FileServerFS(a.staticFS)))

	// redirect / to the about page
	serveMux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		// We do not use this one because it converts it to an absolute path (preventing relocation behind a reverse proxy)
		// http.Redirect(w, r, "about.html", http.StatusFound)
		if r.Method == http.MethodGet {
			w.Header().Set("Location", "catalog.html")
			w.WriteHeader(http.StatusFound)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// pages
	serveMux.Handle("GET /documentation.html", httputil.RootHandler(a.handleAbout))
	serveMux.Handle("GET /catalog.html", httputil.RootHandler(a.getPageHandler("catalog.html")))
	serveMux.Handle("GET /leaderboard.html", httputil.RootHandler(a.getPageHandler("leaderboard.html")))
	serveMux.Handle("GET /bottle.html", httputil.RootHandler(a.handleBottle))
	serveMux.Handle("GET /similarBottles", httputil.RootHandler(a.handleSimilarBottles))

	// search components
	searchMux := http.NewServeMux()
	serveMux.Handle("/search/", http.StripPrefix("/search", searchMux))
	searchMux.Handle("GET /", httputil.RootHandler(a.handleBottleSearchIsValid))

	bottleComponentMux := http.NewServeMux()
	searchMux.Handle("GET /bottle/", http.StripPrefix("/bottle", bottleComponentMux))
	bottleComponentMux.Handle("GET /cards", httputil.RootHandler(a.handleBottleSearch))
	bottleComponentMux.Handle("GET /table", httputil.RootHandler(a.handleBottleSearch))

	metricComponentMux := http.NewServeMux()
	searchMux.Handle("GET /metric/", http.StripPrefix("/metric", metricComponentMux))
	metricComponentMux.Handle("GET /dropdown", httputil.RootHandler(a.handleMetricSearch))

	labelComponentMux := http.NewServeMux()
	searchMux.Handle("GET /label/", http.StripPrefix("/label", labelComponentMux))
	labelComponentMux.Handle("GET /list", httputil.RootHandler(a.handleCommonLabelSearch))
	// Note that we want to serve <img> requests (Sec-Fetch-Dest=image) with actual images and not html.

	serveMux.HandleFunc("GET /artifact/{bottle}/{path...}", func(w http.ResponseWriter, r *http.Request) {
		qs := r.URL.Query()
		h := a.handleArtifact
		switch qs.Get("_type") {
		case "content":
			h = a.handleArtifactContent
		case "raw":
			h = a.handleArtifactRaw
		default:
			if r.Header.Get("Sec-Fetch-Dest") == "image" {
				h = a.handleArtifactRaw
			}
		}
		httputil.RootHandler(h).ServeHTTP(w, r)
	})
}

/*
func (a *WebApp) startWatching() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("starting to watch templates: %w", err)
	}
	a.watcher = watcher

	// defer a.watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				a.log.V(2).Info("event", "item", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					a.reloadTemplates()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				a.log.Error(err, "fsnotify error")
			}
		}
	}()

	err = watcher.Add(filepath.Join(a.assetDir, "templates"))
	if err != nil {
		return fmt.Errorf("unable to add templates to watcher: %w", err)
	}

	return nil
}
*/

func (a *WebApp) parseTemplates() (*template.Template, error) {
	// TemplateFuncs are functions usable in the HTML templates
	templateFuncs := template.FuncMap{
		"ByteSize":        bytefmt.ByteSize,
		"ToAge":           toAge,
		"GetCommonLabels": getCommonLabelsFromBotleEntries,
		"RemoveLabels":    removeLabels,
	}

	tempateGlobPatterns := []string{}
	err := fs.WalkDir(a.templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk templateFS directories: %w", err)
		}
		if !d.IsDir() {
			tempateGlobPatterns = append(tempateGlobPatterns, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	t, err := template.New("base").
		Funcs(sprig.FuncMap()).
		Funcs(templateFuncs).
		ParseFS(a.templateFS, tempateGlobPatterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return t, nil
}

/*
func (a *WebApp) reloadTemplates() {
	a.log.Info("Reloading HTML templates")
	t, err := a.parseTemplates()
	if err != nil {
		a.log.Error(err, "Failed to parse templates, keeping old templates")
		return
	}
	a.templatesLock.Lock()
	defer a.templatesLock.Unlock()
	a.templates = t
}
*/

type globalValues struct {
	DefaultBottleSelectors []string
	Version                string

	// Top is the relative path to the top of the app (right above above www)
	Top string
}

func (a *WebApp) executeTemplateAsResponse(ctx context.Context, w http.ResponseWriter, templateName string, values any, top string) error {
	a.templatesLock.RLock()
	defer a.templatesLock.RUnlock()

	allValues := struct {
		Values       any
		Globals      globalValues
		RootTemplate string
	}{values, a.globalValues, templateName}
	allValues.Globals.Top = top

	log := logger.FromContext(ctx).With("values", allValues)
	log.DebugContext(ctx, "Rendering template", "name", templateName)

	// See https://medium.com/@leeprovoost/dealing-with-go-template-errors-at-runtime-1b429e8b854a
	// We do not write directly to the HTML response because if a template error occurs we send a partial response
	// TODO cache these buffers in a buffer pool for performance
	var buf bytes.Buffer
	if err := a.templates.ExecuteTemplate(&buf, templateName, allValues); err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}
	// if all is good then write buffer to the response writer
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	_, err := buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("writing template to response: %w", err)
	}

	return nil
}
