package bootstrap

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestBootstrapConsumesPS3POSTBody(t *testing.T) {
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("../../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Error(err)
		}
	})
	s, err := NewServer("0", "des.mgn.pub", map[string]string{"US": "18666", "EU": "18667", "JP": "18668"}, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	body := bytes.NewBuffer(make([]byte, 32))
	r := httptest.NewRequest(http.MethodPost, "/demons-souls-us/ss.info", body)
	r.Close = true
	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, r)
	if body.Len() != 0 {
		t.Fatalf("bootstrap left %d POST bytes unread", body.Len())
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	decoded, err := base64.StdEncoding.DecodeString(w.Body.String())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(decoded), "<gameurl1>http://des.mgn.pub:18666/cgi-bin/</gameurl1>") {
		t.Fatal("US game URL missing from bootstrap")
	}
}
