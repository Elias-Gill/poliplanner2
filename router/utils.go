package router

import "net/http"

func extractUserID(r *http.Request) int64 {
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
