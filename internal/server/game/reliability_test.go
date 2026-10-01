package game

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/danmrichards/dessego/internal/service/msg"
	"github.com/danmrichards/dessego/internal/service/replay"
	"github.com/rs/zerolog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type missingReplayService struct {
	Replays
	err error
}

func (s missingReplayService) Get(uint32) (*replay.Replay, error) { return nil, s.err }
func TestMissingReplayReturnsErrorWithoutPanic(t *testing.T) {
	for _, tc := range []struct {
		err    error
		status int
	}{{sql.ErrNoRows, 404}, {errors.New("database failure"), 500}, {nil, 500}} {
		s := &Server{rd: sosPlaintextDecrypter{}, rs: missingReplayService{err: tc.err}, l: zerolog.Nop()}
		w := httptest.NewRecorder()
		s.getReplayDataHandler()(w, httptest.NewRequest("POST", "/", strings.NewReader("ghostID=999")))
		if w.Code != tc.status {
			t.Fatalf("status %d, want %d", w.Code, tc.status)
		}
	}
}

type recordingMessages struct {
	Messages
	added msg.BloodMsg
}

func (m *recordingMessages) Add(b msg.BloodMsg) error { m.added = b; return nil }
func TestBloodMessageAnglesDecodeIndependently(t *testing.T) {
	m := &recordingMessages{}
	s := &Server{rd: sosPlaintextDecrypter{}, ms: m, l: zerolog.Nop()}
	w := httptest.NewRecorder()
	s.addBloodMsgHandler()(w, httptest.NewRequest("POST", "/", strings.NewReader("angx=1&angy=2&angz=3")))
	if w.Code != 200 || m.added.AngX != 1 || m.added.AngY != 2 || m.added.AngZ != 3 {
		t.Fatalf("status %d angles %+v", w.Code, m.added)
	}
}
func TestListHandlersRejectInvalidCountsBeforeServices(t *testing.T) {
	s := &Server{rd: sosPlaintextDecrypter{}, l: zerolog.Nop()}
	for _, tc := range []struct {
		h     http.HandlerFunc
		field string
	}{{s.getSosDataHandler(), "maxSosNum"}, {s.getSosDataHandler(), "sosNum"}, {s.getGhostHandler(), "maxGhostNum"}, {s.getBloodMsgHandler(), "replayNum"}, {s.replayListHandler(), "replayNum"}, {s.worldTendencyHandler(), "maxNum"}} {
		for _, value := range []string{"-1", "1025", "2147483647"} {
			w := httptest.NewRecorder()
			tc.h(w, httptest.NewRequest("POST", "/", strings.NewReader(tc.field+"="+value)))
			if w.Code != 400 {
				t.Fatalf("%s=%s status %d", tc.field, value, w.Code)
			}
		}
	}
}
func TestReplayParserBounds(t *testing.T) {
	r := replay.Replay{CharacterID: "test", BlockID: 60070, AngX: 1, AngZ: 3, Data: []byte("abc")}
	data := append(r.Header(), r.Data...)
	data = append(data, 0)
	for n := 0; n < len(data); n++ {
		if _, err := replay.NewReplayFromBytes(data[:n]); err == nil {
			t.Fatalf("accepted truncation %d", n)
		}
	}
	parsed, err := replay.NewReplayFromBytes(data)
	if err != nil || !bytes.Equal(parsed.Data, r.Data) || parsed.AngX != r.AngX {
		t.Fatalf("valid replay: %+v %v", parsed, err)
	}
}
