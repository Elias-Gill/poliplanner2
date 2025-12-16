package web

import (
	"html/template"
	"path"

	"github.com/elias-gill/poliplanner2/internal/config"
)

// NOTE: made like this so the main layouts are parsed only one time on startup
var (
	BaseLayout  = template.Must(template.ParseGlob(path.Join(config.Get().Paths.BaseDir, "web/templates/layout/base_layout.html")))
	CleanLayout = template.Must(template.ParseGlob(path.Join(config.Get().Paths.BaseDir, "web/templates/layout/clean_layout.html")))
)
