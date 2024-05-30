package webapp

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/microcosm-cc/bluemonday"
	"github.com/opencontainers/go-digest"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"gitlab.com/act3-ai/asce/go-common/pkg/logger"

	"gitlab.com/act3-ai/asce/data/telemetry/internal/db"
	"gitlab.com/act3-ai/asce/data/telemetry/internal/middleware"
)

// TODO consider using GO's filesystem abstraction to build a filesystem for the artifacts based on their path in the bottle
// Once you have the filesystem abstraction then the problem becomes the common problem of simply wanting to
// display content from disk in a browser in an intelligent and safe way with untrusted content.
// There is probably a library we can leverage for this (similar to what github or gitlab does in viewing content)

func (a *WebApp) handleArtifactRaw(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	bottleDigest, err := digest.Parse(chi.URLParam(r, "bottle"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle parameter")
	}

	// we use PublicArtifact.Path to lookup the artifact within a bottle
	artifactPath := chi.URLParam(r, "*")

	artifact, err := db.FindArtifact(con, bottleDigest, artifactPath)
	if err != nil {
		return err
	}

	httputil.AllowCaching(w.Header())
	w.Header().Set("Content-Type", artifact.MediaType)
	_, err = w.Write(artifact.Data.RawData)
	if err != nil {
		return fmt.Errorf("handling raw artifact: %w", err)
	}

	return nil
}

func (a *WebApp) handleArtifact(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	bottleDigest, err := digest.Parse(chi.URLParam(r, "bottle"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle parameter: "+err.Error())
	}

	// we use PublicArtifact.Path to lookup the artifact within a bottle
	artifactPath := chi.URLParam(r, "*")

	artifact, err := db.FindArtifact(con, bottleDigest, artifactPath)
	if err != nil {
		return err
	}

	// parse media type (we do not need the params)
	mediaType, _, err := mime.ParseMediaType(artifact.MediaType)
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusNotAcceptable, "Invalid media type: "+err.Error())
	}
	// mediaType := mime.TypeByExtension(filepath.Ext(artifact.Path))

	// compute the relative location of top
	top := strings.Repeat("../", strings.Count(artifact.Path, "/")+3)

	templateName := "artifact-raw.html"
	var values any = struct {
		db.PublicArtifact
		BottleDigest digest.Digest
	}{*artifact, bottleDigest}

	switch mediaType {
	case "text/csv", "text/tab-separated-values":
		r := csv.NewReader(bytes.NewReader(artifact.Data.RawData))
		if mediaType == "text/tab-separated-values" {
			r.Comma = '\t'
		}

		records, err := r.ReadAll()
		if err != nil {
			return httputil.NewHTTPError(err, http.StatusNotAcceptable, "Unable to read tabular data")
		}

		templateName = "artifact-tabular.html"
		values = struct {
			db.PublicArtifact
			BottleDigest digest.Digest
			Records      [][]string
		}{*artifact, bottleDigest, records}
	case "text/plain":
		templateName = "artifact-text.html"
		// Should we just redirect them to the /api/artifact handler?
	case "text/markdown":
		templateName = "artifact-markdown.html"
		safeHTML := a.convertMarkdown(ctx, artifact, bottleDigest, top)
		values = struct {
			db.PublicArtifact
			BottleDigest digest.Digest
			SafeHTML     template.HTML
		}{*artifact, bottleDigest, template.HTML(safeHTML)}
	case "image/jpeg", "image/png", "image/svg+xml", "image/webp", "image/gif":
		templateName = "artifact-image.html"
		// could just do a redirect to ../api/artifact and not render any template
	case "application/x.jupyter.notebook+json", "text/html":
		templateName = "artifact-iframe.html"
	}

	return a.executeTemplateAsResponse(ctx, w, templateName, values, top)
}

func (a *WebApp) handleArtifactContent(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	// log := logger.FromContext(ctx)
	con := middleware.DatabaseFromContext(ctx)

	bottleDigest, err := digest.Parse(chi.URLParam(r, "bottle"))
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusBadRequest, "Invalid bottle parameter: "+err.Error())
	}

	// we use PublicArtifact.Path to lookup the artifact within a bottle
	// artifactPath := vars["artifact"]
	artifactPath := chi.URLParam(r, "*")

	artifact, err := db.FindArtifact(con, bottleDigest, artifactPath)
	if err != nil {
		return err
	}

	// compute the relative location of top
	top := strings.Repeat("../", strings.Count(artifact.Path, "/")+3)

	// parse media type (we do not need the params)
	mediaType, _, err := mime.ParseMediaType(artifact.MediaType)
	if err != nil {
		return httputil.NewHTTPError(err, http.StatusNotAcceptable, "Invalid media type: "+err.Error())
	}
	// mediaType := mime.TypeByExtension(filepath.Ext(artifact.Path))

	httputil.AllowCaching(w.Header())

	var safeHTML []byte
	switch mediaType {
	// TODO add other native browser types
	case "text/markdown":
		safeHTML = a.convertMarkdown(ctx, artifact, bottleDigest, top)
	case "application/x.jupyter.notebook+json":
		out, err := a.convertNotebook(ctx, artifact, top)
		if err != nil {
			return err
		}
		safeHTML = out
	case "text/html":
		// we might not need to sanitize the HTML since we put it in an iframe with a sandbox, but just to be safe
		// TODO we need to inject <base parent="_parent" /> into the <head> or not use iframes for HTML (display without the navbar instead)
		// We might also need to add ?_type=raw to the img tags for images to show up properly.  Maybe this can be done with <base href="?_type=raw" target="_parent" />
		safeHTML = bluemonday.UGCPolicy().SanitizeBytes(artifact.Data.RawData)
	case "image/jpeg", "image/png", "image/svg+xml", "image/webp", "image/gif":
		w.Header().Set("Content-Type", artifact.MediaType)
		_, err = w.Write(artifact.Data.RawData)
		if err != nil {
			return fmt.Errorf("unable to write image to response: %w", err)
		}
		return nil
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, err = w.Write(safeHTML)
	if err != nil {
		return fmt.Errorf("unable to write safe HTML to response: %w", err)
	}
	return nil
}

func (a *WebApp) convertMarkdown(ctx context.Context, artifact *db.PublicArtifact, bottleDigest digest.Digest, top string) []byte {
	log := logger.FromContext(ctx)
	// We use www/artifact because it includes the header
	// TODO "www" here breaks the webapp's modularity w.r.t. the App.
	prefix := top + "www/artifact/" + bottleDigest.String() + "/" + path.Dir(artifact.Path)
	log.DebugContext(ctx, "Rendering markdown", "absolutePrefix", prefix)
	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		AbsolutePrefix: prefix,
	}
	renderer := html.NewRenderer(opts)
	mdHTML := markdown.ToHTML(artifact.Data.RawData, nil, renderer)
	return bluemonday.UGCPolicy().SanitizeBytes(mdHTML)
}

func (a *WebApp) convertNotebook(ctx context.Context, artifact *db.PublicArtifact, top string) ([]byte, error) {
	log := logger.FromContext(ctx)

	// https://nbconvert.readthedocs.io/en/latest/customizing.html

	// TODO this command can be slow (take a few seconds) so caching the result would be ideal.

	// https://github.com/SylvainCorlay/nbconvert-acme/tree/master/share/jupyter/nbconvert/templates/acme

	// we copy assets to the local filesystem so the external command can access them
	tmpDirPath, err := os.MkdirTemp("", "tmp-ipynb-templates-*")
	if err != nil {
		return nil, fmt.Errorf("creating ipynb templates temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDirPath)
	err = copyFSToDir(a.ipynbFS, tmpDirPath)
	if err != nil {
		return nil, fmt.Errorf("populating ipynb templates temp directory: %w", err)
	}

	// TODO this wont work with assetFS
	// probably need to mount the assetFS first
	cmd := exec.CommandContext(ctx, a.jupyter, "nbconvert", "--to=html", "--stdin", "--stdout",
		"--HTMLExporter.require_js_url=internal/webapp/assets/static/libs/requirejs/2.3.6/require.min.js",
		"--HTMLExporter.mathjax_url=internal/webapp/assets/static/libs/mathjax/3.2.2/es5/tex-mml-chtml.js",
		"--TemplateExporter.extra_template_basedirs="+filepath.Join(tmpDirPath, "templates"),
		"--template=telemetry",
	)

	log.DebugContext(ctx, "Converting ipynb", "command", cmd.String())

	cmd.Stdin = bytes.NewReader(artifact.Data.RawData)
	out, err := cmd.Output()
	exitError := &exec.ExitError{}
	if errors.As(err, &exitError) {
		// special logging for exitErrors to log the extra information useful in troubleshooting
		log.ErrorContext(ctx, "jupyter nbconvert", "stdout", out, "stderr", string(exitError.Stderr), "error", err)
	}
	if err != nil {
		return nil, fmt.Errorf("unexpected error from Jupyter: %w", err)
	}

	// Sanitizing kills the notebook, but it should not be necessary since it is in a sandbox.
	// safeHTML = bluemonday.UGCPolicy().SanitizeBytes(out)
	return out, nil
}

// copyFSToDir walks the file tree in fsys and copies the files to dirPath.
func copyFSToDir(fsys fs.FS, dirPath string) error {
	err := fs.WalkDir(fsys, ".", func(fsPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking fsys directory: %w", err)
		}
		if d.IsDir() {
			err = os.MkdirAll(filepath.Join(dirPath, fsPath), os.ModePerm)
			if err != nil {
				return fmt.Errorf("could not create directory (%s): %w", fsPath, err)
			}
		} else {
			fileContents, err := fs.ReadFile(fsys, fsPath)
			if err != nil {
				return fmt.Errorf("getting file contents from fsys (%s): %w", fsPath, err)
			}
			dest := filepath.Join(dirPath, fsPath)
			err = os.WriteFile(dest, fileContents, os.FileMode(0o644))
			if err != nil {
				return fmt.Errorf("writing file to directory (%s): %w", dest, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("problem copying fsys to directory: %w", err)
	}
	return nil
}
