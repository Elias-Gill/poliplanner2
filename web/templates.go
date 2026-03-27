package web

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/config"
)

// NOTE: made like this so the main layouts are parsed only one time on startup
var (
	// Complete layouts to render full pages.
	BaseLayout = parseBaseLayout()

	// Reusable components like buttons, lists, messages, modals, etc., use with
	// HTMX.
	Fragments = parseFragments()
)

// toJS convierte cualquier valor a JSON seguro para insertar directamente en <script>
func toJS(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return template.JS(b)
}

// Parses the layotus and the layout fragments (navbar, sidebar, etc)
func parseBaseLayout() *template.Template {
	tPath := path.Join(config.Get().Paths.BaseDir, "web", "templates", "layouts")
	fragPattern := path.Join(tPath, "fragments", "*.html")
	layout := path.Join(tPath, "base_layout.html")

	tmpl := template.New("base").Funcs(template.FuncMap{"toJS": toJS})
	tmpl = template.Must(tmpl.ParseGlob(fragPattern))
	tmpl = template.Must(tmpl.ParseFiles(layout))

	return tmpl
}

// Parses reusable components like messages, buttons, etc
func parseFragments() *template.Template {
	baseDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "fragments")
	tmpl := template.New("base").Funcs(template.FuncMap{"toJS": toJS})

	err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return tmpl
}
