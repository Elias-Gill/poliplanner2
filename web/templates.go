package web

import (
	"html/template"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
)

// NOTE: made like this so the main layouts are parsed only one time on startup
var (
	// Complete layouts to render full pages.
	BaseLayout = parseBaseLayout()

	// Reusable components like buttons, lists, messages, modals, etc., use with
	// HTMX.
	Fragments  = parseFragments()
)

// Parses the layotus and the layout fragments (navbar, sidebar, etc)
func parseBaseLayout() *template.Template {
	tPath := path.Join(config.Get().Paths.BaseDir, "web", "templates", "layouts")
	fragPattern := path.Join(tPath, "fragments", "*.html")
	layout := path.Join(tPath, "base_layout.html")

	tmpl := template.New("base").Funcs(template.FuncMap{ /* custom functions here */ })
	tmpl = template.Must(tmpl.ParseGlob(fragPattern))
	tmpl = template.Must(tmpl.ParseFiles(layout))

	return tmpl
}

// Parses reusable components like messages, buttons, etc
func parseFragments() *template.Template {
	fragPattern := path.Join(config.Get().Paths.BaseDir, "web", "templates", "fragments", "**", "*.html")

	tmpl := template.New("base").Funcs(template.FuncMap{ /* custom functions here */ })
	tmpl = template.Must(tmpl.ParseGlob(fragPattern))

	return tmpl
}
