package game

import (
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danmrichards/dessego/internal/service/character"
	"github.com/rs/zerolog"
)

type tendencyTestCharacters struct{ Characters }

func (tendencyTestCharacters) WorldTendency(n int) ([]character.WorldTendency, error) {
	return []character.WorldTendency{{WB1: -10, WB2: 20, WB3: -30, WB4: 40, WB5: -50, WB6: 60, WB7: -70}}, nil
}

func (tendencyTestCharacters) Stats(id string) (*character.Stats, error) {
	return &character.Stats{GradeS: 1, GradeA: 2, GradeB: 3, GradeC: 4, GradeD: 5, Sessions: 6}, nil
}

func (tendencyTestCharacters) MsgRating(id string) (int, error) { return 7, nil }

func TestCharacterGradeResponsesHaveSuccessFlag(t *testing.T) {
	s := &Server{cs: tendencyTestCharacters{}, rd: sosPlaintextDecrypter{}, l: zerolog.Nop()}
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		command byte
		values  []int32
	}{
		{"multiplayer", s.characterMPGradeHandler(), 0x28, []int32{1, 2, 3, 4, 5, 6}},
		{"blood message", s.characterBloodMsgGradeHandler(), 0x29, []int32{7}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.handler(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("NPID=helper&ver=100")))
			if w.Code != http.StatusOK {
				t.Fatalf("status %d", w.Code)
			}
			data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Body.String()))
			if err != nil {
				t.Fatal(err)
			}
			want := 6 + 4*len(tc.values)
			if len(data) != want || data[0] != tc.command || binary.LittleEndian.Uint32(data[1:5]) != uint32(want) {
				t.Fatalf("grade frame length = %d, want %d", len(data), want)
			}
			if data[5] != 1 {
				t.Fatalf("success flag = %d, want 1", data[5])
			}
			for i, want := range tc.values {
				if got := int32(binary.LittleEndian.Uint32(data[6+i*4:])); got != want {
					t.Fatalf("field %d = %d, want %d", i, got, want)
				}
			}
		})
	}
}

func TestWorldTendencyResponseHasSevenPairs(t *testing.T) {
	s := &Server{cs: tendencyTestCharacters{}, rd: sosPlaintextDecrypter{}, l: zerolog.Nop()}
	w := httptest.NewRecorder()
	s.worldTendencyHandler()(w, httptest.NewRequest(http.MethodPost, "/cgi-bin/getQWCData.spd", strings.NewReader("maxNum=10&ver=100")))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.Body.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 61 || data[0] != 0x0e || binary.LittleEndian.Uint32(data[1:5]) != 61 {
		t.Fatalf("world tendency frame length = %d, want 61", len(data))
	}
	for i, want := range []int32{-10, 0, 20, 0, -30, 0, 40, 0, -50, 0, 60, 0, -70, 0} {
		if got := int32(binary.LittleEndian.Uint32(data[5+i*4:])); got != want {
			t.Fatalf("field %d = %d, want %d", i, got, want)
		}
	}
}
