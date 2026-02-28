package router

import (
	"html/template"
	"net/http"

	"github.com/elias-gill/poliplanner2/internal/logger"
	"github.com/elias-gill/poliplanner2/web"
)

// This function should never fail or panic if the session middleware is functioning correctly.
// If a protected endpoint is reached without a userID set in the request context,
// the application is in an invalid state and something unexpected has occurred.
//
// If this is the case, then probably the endpoint has not been added to the "protected
// endpoints" array list in the middleware configuration.
func extractUserIDFromCtx(r *http.Request) int64 {
	switch id := r.Context().Value("userID").(type) {
	case int64:
		return id
	default:
		panic("The user ID is not set in the request of a protected endpoint, something wrong must have happened with the session manager middleware")
	}
}

func isHtmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Makes a correct redirect if the request is from htmx or is a simple http request
func customRedirect(w http.ResponseWriter, r *http.Request, target string) {
	if isHtmx(r) {
		w.Header().Add("HX-redirect", target)
	} else {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

func parseTemplateWithBaseLayout(path string) *template.Template {
	layout := template.Must(web.BaseLayout.Clone())
	return template.Must(layout.ParseFiles(path))
}

func parseComponentTemplate(path string) *template.Template {
	return template.Must(template.ParseFiles(path))
}

func executeFragment(w http.ResponseWriter, r *http.Request, fragment string, data any) {
	w.Header().Set("Content-Type", "text/html")
	err := web.Fragments.ExecuteTemplate(w, fragment, data)
	if err != nil {
		customRedirect(w, r, "/500")
		logger.Debug("Error executing fragment", "error", err)
	}
}
