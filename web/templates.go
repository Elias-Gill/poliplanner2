package web

import "html/template"


// NOTE: made like this so the main layouts are parsed only one time on startup
var (
	BaseLayout  = template.Must(template.ParseGlob("web/templates/layout/base_layout.html"))
	CleanLayout = template.Must(template.ParseGlob("web/templates/layout/clean_layout.html"))
)
