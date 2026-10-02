package middleware

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

type responseLog struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseLog) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func (w *responseLog) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	if status >= 200 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseLog) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.bytes += n
	return n, err
}

// LogRequest records completed responses without exposing request payloads.
func LogRequest(l zerolog.Logger, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		response := &responseLog{ResponseWriter: w}
		completed := false
		defer func() {
			status := response.status
			if status == 0 && completed {
				status = http.StatusOK
			}
			l.Info().Str("method", r.Method).Str("path", r.URL.Path).
				Str("client", r.RemoteAddr).Int("status", status).
				Int("bytes", response.bytes).Dur("duration_ms", time.Since(start)).
				Bool("completed", completed).Msg("request")
		}()
		h.ServeHTTP(response, r)
		completed = true
	}
}
