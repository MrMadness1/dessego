package sos

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestManagerDeliversSummonRoom(t *testing.T) {
	for _, monk := range []bool{false, true} {
		name := "blue phantom"
		if monk {
			name = "Old Monk"
		}
		t.Run(name, func(t *testing.T) {
			m := NewManager(zerolog.Nop())
			sign := &SOS{CharacterID: "helper", BlockID: 40070, Updated: time.Now()}
			m.Add(sign)
			if m.Summon(sign.ID+1, "wrong-room") {
				t.Fatal("accepted missing sign")
			}
			var matched bool
			if monk {
				matched = m.Monk("test-room")
			} else {
				matched = m.Summon(sign.ID, "test-room")
			}
			if !matched {
				t.Fatal("failed to match active sign")
			}
			if room := m.Check("helper"); room != "test-room" {
				t.Fatalf("room = %q", room)
			}
			if room := m.Check("helper"); room != "" {
				t.Fatalf("delivered room twice: %q", room)
			}
		})
	}
}
