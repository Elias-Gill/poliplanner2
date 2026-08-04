package middleware

import (
	"net/http"
	"sync"
	"time"
)

type GlobalPDFLimiter struct {
	mu        sync.Mutex
	count     int
	max       int
	window    time.Duration
	lastReset time.Time
}

func NewGlobalPDFLimiter(max int, window time.Duration) *GlobalPDFLimiter {
	return &GlobalPDFLimiter{
		max:       max,
		window:    window,
		lastReset: time.Now(),
	}
}

func (l *GlobalPDFLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.mu.Lock()
		now := time.Now()

		// Si ya pasó el minuto, reseteamos el contador
		if now.Sub(l.lastReset) >= l.window {
			l.count = 0
			l.lastReset = now
		}

		// Validar si superó el límite
		if l.count >= l.max {
			l.mu.Unlock()
			http.Error(w, "Límite global de generación de PDFs alcanzado por este minuto. Intenta de nuevo en breve.", http.StatusTooManyRequests)
			return
		}

		l.count++
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
