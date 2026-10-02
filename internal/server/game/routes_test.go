package game

import (
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnknownRoutesPreserveNativeFallback(t *testing.T) {
	s := &Server{r: http.NewServeMux(), l: zerolog.Nop()}
	s.routes()
	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, httptest.NewRequest("GET", "/unknown", nil))
	if w.Code != 404 {
		t.Fatalf("GET status %d", w.Code)
	}
	func() {
		defer func() {
			if recover() != http.ErrAbortHandler {
				t.Fatal("unknown native POST must abort")
			}
		}()
		s.r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/unknown", strings.NewReader("native body")))
	}()
	// Known ignored-body routes retain their existing wire response.
	w = httptest.NewRecorder()
	s.r.ServeHTTP(w, httptest.NewRequest("POST", "/cgi-bin/getTimeMessage.spd", strings.NewReader("native body")))
	if w.Code != 200 || !strings.HasSuffix(w.Body.String(), "\n") {
		t.Fatalf("known response: %v", w)
	}
}
