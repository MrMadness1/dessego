package middleware

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

const maxIgnoredRequestBodyBytes = 64 << 10

// LimitRequestBody bounds allocations before a game handler decodes its body.
// One MiB leaves ample room for replay uploads and ordinary native requests.
func LimitRequestBody(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		defer body.Close()
		data, err := io.ReadAll(body)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "incomplete request body", http.StatusBadRequest)
			}
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(data))
		h.ServeHTTP(w, r)
	}
}

// DiscardRequestBody consumes unused POST data before sending a response.
// Native PS3 clients send headers and body separately with Connection: close;
// closing the socket with unread body data can reset an otherwise valid reply.
func DiscardRequestBody(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := http.MaxBytesReader(w, r.Body, maxIgnoredRequestBodyBytes)
		defer body.Close()
		if _, err := io.Copy(io.Discard, body); err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "incomplete request body", http.StatusBadRequest)
			}
			return
		}
		h.ServeHTTP(w, r)
	}
}
