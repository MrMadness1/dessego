package sos

import (
	"encoding/binary"
	"testing"
)

func TestSOSBytesIncludesRatingsAndSessions(t *testing.T) {
	s := SOS{ID: 7, CharacterID: "helper", BlockID: 60070, Ratings: []int{1, 2, 3, 4, 5}, TotalSessions: 6, PlayerInfo: "info"}
	data := s.Bytes()
	// ID, terminated character ID, block, six floats, three message fields,
	// reserved field, five ratings, reserved field, sessions, terminated info,
	// two tendency values and the phantom flag.
	want := 4 + len(s.CharacterID) + 1 + 4 + 24 + 12 + 4 + 20 + 4 + 4 + len(s.PlayerInfo) + 1 + 8 + 1
	if len(data) != want {
		t.Fatalf("sign record length = %d, want %d", len(data), want)
	}
	offset := 4 + len(s.CharacterID) + 1 + 4 + 24 + 12 + 4
	for i, rating := range s.Ratings {
		if got := binary.LittleEndian.Uint32(data[offset+4*i : offset+4*i+4]); got != uint32(rating) {
			t.Fatalf("rating %d = %d, want %d", i, got, rating)
		}
	}
	offset += 20 + 4
	if got := binary.LittleEndian.Uint32(data[offset : offset+4]); got != uint32(s.TotalSessions) {
		t.Fatalf("sessions = %d, want %d", got, s.TotalSessions)
	}
}
