package middleware

import "net/http"

// HeadAsGet rewrites HEAD requests as GET before the router so routes registered
// with Get also answer HEAD. net/http discards the response body for HEAD requests.
func HeadAsGet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			r = r.Clone(r.Context())
			r.Method = http.MethodGet
		}

		next.ServeHTTP(w, r)
	})
}
