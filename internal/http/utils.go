package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/elias-gill/poliplanner2/web"
)

// REFACTOR: move to a dedicated cookies section or something
const LatestSelectionCookie = "latestScheduleSelection"

// Makes a correct redirect if the request is from htmx or is a simple http request
func CustomRedirect(w http.ResponseWriter, r *http.Request, target string) {
	if isHtmx(r) {
		w.Header().Add("HX-redirect", target)
	} else {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

func ParseTemplateWithBaseLayout(path string) *template.Template {
	layout := template.Must(web.BaseLayout.Clone())
	return template.Must(layout.ParseFiles(path))
}

func ParseComponentTemplate(path string) *template.Template {
	return template.Must(template.ParseFiles(path))
}

// ---------- validation helpers ----------

func RequiredString(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", fmt.Errorf("required")
	}
	return v, nil
}

func ParseID(idStr string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

func ParseIDList(ids []string) ([]int64, error) {

	out := make([]int64, len(ids))

	for i, idStr := range ids {

		id, err := ParseID(idStr)
		if err != nil {
			return nil, err
		}

		out[i] = id
	}

	return out, nil
}

func isHtmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
