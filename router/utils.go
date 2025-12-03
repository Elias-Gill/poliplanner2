package router

import "net/http"

// This function should never fail or panic if the session middleware is functioning correctly.
// If a protected endpoint is reached without a userID set in the request context,
// the application is in an invalid state and something unexpected has occurred.
//
// If this is the case, then probably the endpoint has not been added to the "protected
// endpoints" array list in the middleware configuration.
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
