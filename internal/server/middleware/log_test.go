package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogRequestPreservesResponseAndRecordsOutcome(t *testing.T) {
	for _, status := range []int{200, 404, 413} {
		var logs bytes.Buffer
		h := LogRequest(zerolog.New(&logs).With().Str("region", "US").Logger(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "kept")
			if status != 200 {
				w.WriteHeader(status)
			}
			w.Write([]byte("payload\n"))
			w.WriteHeader(500) // Must not replace the first final status.
		}))
		w := httptest.NewRecorder()
		h(w, httptest.NewRequest("POST", "/cgi-bin/test.spd", nil))
		if w.Code != status || w.Body.String() != "payload\n" || w.Header().Get("X-Test") != "kept" {
			t.Fatalf("response altered: %v", w)
		}
		var event map[string]interface{}
		if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
			t.Fatal(err)
		}
		if event["status"] != float64(status) || event["bytes"] != float64(8) || event["region"] != "US" || event["completed"] != true {
			t.Fatalf("event: %v", event)
		}
		if _, ok := event["duration_ms"]; !ok {
			t.Fatal("missing duration")
		}
	}
}

func TestLogRequestPreservesAbort(t *testing.T) {
	var logs bytes.Buffer
	h := LogRequest(zerolog.New(&logs), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic(http.ErrAbortHandler) }))
	func() {
		defer func() {
			if recover() != http.ErrAbortHandler {
				t.Fatal("abort not preserved")
			}
		}()
		h(httptest.NewRecorder(), httptest.NewRequest("POST", "/unknown", nil))
	}()
	var event map[string]interface{}
	if err := json.Unmarshal(logs.Bytes(), &event); err != nil {
		t.Fatal(err)
	}
	if event["completed"] != false || event["status"] != float64(0) {
		t.Fatalf("abort reported as success: %v", event)
	}
}
