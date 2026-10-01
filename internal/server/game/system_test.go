package game

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danmrichards/dessego/internal/service/gamestate"
	"github.com/rs/zerolog"
)

func TestSystemRoutesConsumePOSTBody(t *testing.T) {
	for _, tc := range []struct {
		path    string
		command byte
	}{
		{"/cgi-bin/login.spd", 0x02},
		{"/cgi-bin/getTimeMessage.spd", 0x22},
	} {
		t.Run(tc.path, func(t *testing.T) {
			s := &Server{r: http.NewServeMux(), gs: gamestate.NewMemory(), l: zerolog.Nop()}
			s.routes()
			body := bytes.NewBuffer(make([]byte, 32))
			r := httptest.NewRequest(http.MethodPost, tc.path, body)
			r.Close = true
			w := httptest.NewRecorder()
			s.r.ServeHTTP(w, r)
			if body.Len() != 0 {
				t.Fatalf("system route left %d POST bytes unread", body.Len())
			}
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			if !strings.HasSuffix(w.Body.String(), "\n") {
				t.Fatal("response missing final newline")
			}
			decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Body.String()))
			if err != nil {
				t.Fatal(err)
			}
			if len(decoded) < 5 || decoded[0] != tc.command {
				t.Fatal("incorrect response command")
			}
			if binary.LittleEndian.Uint32(decoded[1:5]) != uint32(len(decoded)) {
				t.Fatal("incorrect response length")
			}
		})
	}
}
