package cookie

import (
	"net/http"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/model/auth"
	"github.com/elias-gill/poliplanner2/internal/model/schedule"
)

const (
	SessionIDCookie = "session_id"

	LatestScheduleCookie = "latestSelectedSchedule"
)

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func SetSessionCookie(w http.ResponseWriter, token auth.SessionID) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDCookie,
		Value:    string(token),
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Get().Security.SecureHTTP,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().In(timezone.ParaguayTZ).Add(15 * 24 * time.Hour),
	})
}

func GetSessionCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(SessionIDCookie)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func GetLatestScheduleCookie(r *http.Request) (schedule.ScheduleID, bool) {
	cookie, err := r.Cookie(LatestScheduleCookie)
	if err != nil {
		return -1, false
	}

	id, err := strconv.ParseInt(cookie.Value, 10, 64)
	if err != nil || id < 1 {
		return -1, false
	}

	return schedule.ScheduleID(id), true
}

func SetLatestScheduleCookie(w http.ResponseWriter, scheduleID schedule.ScheduleID) {
	// set this id into a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     LatestScheduleCookie,
		Value:    strconv.FormatInt(int64(scheduleID), 10),
		Path:     "/",
		HttpOnly: true,
		Secure:   config.Get().Security.SecureHTTP,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   30 * 24 * 60 * 60,
	})
}
