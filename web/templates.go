package web

import (
	"html/template"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
)

// NOTE: made like this so the main layouts are parsed only one time on startup
var (
	BaseLayout = parseTemplates()
)

func parseTemplates() *template.Template {
	tPath := path.Join(config.Get().Paths.BaseDir, "web", "templates")
	fragPattern := path.Join(tPath, "fragments", "*.html")
	layout := path.Join(tPath, "layouts", "base_layout.html")

	tmpl := template.New("base").Funcs(template.FuncMap{ /* custom functions here */ })
	tmpl = template.Must(tmpl.ParseGlob(fragPattern))
	tmpl = template.Must(tmpl.ParseFiles(layout))

	return tmpl
}
