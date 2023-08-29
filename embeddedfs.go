package yahs

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	//go:embed wwwroot/assets
	fsAssets embed.FS

	//go:embed wwwroot/layouts wwwroot/pages
	fsTemplates embed.FS
)

type WWWRoot struct {
	AssetsDir  string
	LayoutsDir string
	PagesDir   string
	Assets     embed.FS
	Templates  embed.FS
}

func NewWWWRoot() WWWRoot {
	wr := WWWRoot{
		AssetsDir:  "wwwroot/assets",
		LayoutsDir: "wwwroot/layouts",
		PagesDir:   "wwwroot/pages",
		Assets:     fsAssets,
		Templates:  fsTemplates,
	}

	return wr
}

// WithEmbeddedFS() takes a WWWRoot that includes embedded file systems and
// path metadata. Templates are parsed once at startup and stored in a hash map
// to reduce compute and serve latency at runtime when handling inbound
// requests.
// Consider alternatives like build flag based determining if live or embed.
// See after enable option to serve live files too.
func WithEmbeddedFS(wwwroot WWWRoot) Option {
	return func(hs *Server) error {
		err := hs.loadTemplates(wwwroot.Templates, wwwroot.PagesDir, wwwroot.LayoutsDir)
		if err != nil {
			return fmt.Errorf("failed loading templates: %w", err)
		}

		fsys, err := fs.Sub(wwwroot.Assets, wwwroot.AssetsDir)
		if err != nil {
			return fmt.Errorf("failed loading assets filesystem: %w", err)
		}

		hs.assets = http.FS(fsys)

		hs.AddRoute(
			"GET",
			"/static/.*",
			http.StripPrefix("/static/", hs.handleStaticFiles()).ServeHTTP,
		)

		hs.AddRoute(
			"GET",
			"/.*",
			hs.handleTemplates(),
		)

		return nil
	}
}

func (hs *Server) loadTemplates(fsTemplates fs.FS, pagesDir, layoutsDir string) error {
	if hs.templates == nil {
		hs.templates = make(map[string]*template.Template)
	}

	if fsTemplates == nil {
		return errors.New("fsTemplates === nil")
	}

	loadTemplate := func(fsPath string, d fs.DirEntry, err error) error {
		if fsPath == pagesDir && d == nil && err != fs.ErrNotExist {
			return fmt.Errorf("root directory does not exist: %s", fsPath)
		}

		if err != nil {
			return fmt.Errorf("failed to read fsPath: %s error: %w", fsPath, err)
		}

		if d.IsDir() {
			return nil
		}

		pt, err2 := template.ParseFS(fsTemplates, fsPath, layoutsDir+"/*.gohtml")
		if err2 != nil {
			return fmt.Errorf("page not found for path: %s error: %w", fsPath, err2)
		}

		// Trim embedded filesystem pages prefix so map key becomes URL path.
		webPath, _ := strings.CutPrefix(fsPath, pagesDir)

		hs.templates[webPath] = pt

		return nil
	}

	err := fs.WalkDir(fsTemplates, pagesDir, loadTemplate)
	if err != nil {
		return fmt.Errorf("failed walking directory: %w", err)
	}

	return nil
}

func (hs *Server) handleTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fp := filepath.Join("", filepath.Clean(r.URL.Path))

		if strings.HasSuffix(r.URL.Path, "/") {
			fp = filepath.Join(fp, "index.html")
		}

		tmpl, ok := hs.templates[fp]
		if !ok {
			http.NotFoundHandler().ServeHTTP(w, r)

			return
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "layout", nil); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		_, err := buf.WriteTo(w)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	}
}

func (hs *Server) handleStaticFiles() http.HandlerFunc {
	fs := http.FileServer(hs.assets)

	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Redacted())

		f, err := hs.assets.Open(path)
		if err != nil {
			http.NotFoundHandler().ServeHTTP(w, r)

			return
		}

		stat, err := f.Stat()
		if err != nil {
			http.NotFoundHandler().ServeHTTP(w, r)

			return
		}

		if stat.IsDir() {
			http.NotFoundHandler().ServeHTTP(w, r)

			return
		}

		closeErr := f.Close()
		if closeErr != nil {
			http.NotFoundHandler().ServeHTTP(w, r)

			return
		}

		fs.ServeHTTP(w, r)
	}
}
