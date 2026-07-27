package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/elias-gill/poliplanner2/internal/model/user"
)

type contextKey string

const userIDKey contextKey = "userID"

// ==================================
// = 		Middleware utils        =
// ==================================

func InjectUserID(r *http.Request, userID user.UserID) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

func ExtractUserID(r *http.Request) (user.UserID, bool) {
	id, ok := r.Context().Value(userIDKey).(user.UserID)
	return id, ok
}

// This function should never fail or panic if the session middleware is functioning correctly.
// If a protected endpoint is reached without a userID set in the request context,
// the application is in an invalid state and something unexpected has occurred.
//
// If this is the case, then probably the endpoint has not been added to the "protected
// endpoints" array list in the middleware configuration.
func MustExtractUserID(r *http.Request) user.UserID {
	id, ok := ExtractUserID(r)
	if !ok {
		panic("HTTP context error: userID no encontrado en una ruta protegida. Revisa el middleware de sesión.")
	}
	return id
}

// ========================================
// =         Redirection helpers          =
// ========================================

// Makes a correct redirect if the request is from htmx or is a simple http request
func Redirect(w http.ResponseWriter, r *http.Request, target string) {
	if IsHtmx(r) {
		// HTMX needs HX-Redirect header and status 200 or 204
		w.Header().Set("HX-Redirect", target)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Standard HTTP redirection
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func IsHtmx(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// ========================================
// =          Validation helpers          =
// ========================================

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
