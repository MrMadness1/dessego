package sos

import (
	"sync"
	"sync/atomic"
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

func TestManagerRejectsExpiredAndCompetingSummons(t *testing.T) {
	for _, monk := range []bool{false, true} {
		m := NewManager(zerolog.Nop())
		sign := &SOS{CharacterID: "helper", BlockID: 40070, Updated: time.Now().Add(-time.Hour)}
		m.Add(sign)
		summon := func(room string) bool {
			if monk {
				return m.Monk(room)
			}
			return m.Summon(sign.ID, room)
		}
		if summon("expired") {
			t.Fatal("accepted expired sign without prior List")
		}
		m.Check("helper")
		if summon("") {
			t.Fatal("accepted empty room")
		}
		if !summon("first") || !summon("first") {
			t.Fatal("failed fresh or same-room retry")
		}
		if summon("second") {
			t.Fatal("overwrote pending room")
		}
		if got := m.Check("helper"); got != "first" {
			t.Fatalf("room %q", got)
		}
		if !summon("cancelled") {
			t.Fatal("fresh summon failed")
		}
		m.Delete("helper")
		if got := m.Check("helper"); got != "" {
			t.Fatalf("cancelled summon delivered %q", got)
		}
	}
}
func TestManagerExpiresPendingRooms(t *testing.T) {
	m := NewManager(zerolog.Nop())
	s := &SOS{CharacterID: "helper", BlockID: 40070, Updated: time.Now()}
	m.Add(s)
	m.Summon(s.ID, "room")
	s.Updated = time.Now().Add(-time.Hour)
	if got := m.Check("helper"); got != "" {
		t.Fatalf("expired pending delivered %q", got)
	}
	if len(m.List(40070, 0)) != 0 || len(m.List(40070, -1)) != 0 {
		t.Fatal("invalid count")
	}
}

func TestOnlyOneConcurrentHostWinsPendingSummon(t *testing.T) {
	m := NewManager(zerolog.Nop())
	s := &SOS{CharacterID: "helper", BlockID: 40070, Updated: time.Now()}
	m.Add(s)
	var wg sync.WaitGroup
	var winners int32
	for _, room := range []string{"first", "second"} {
		wg.Add(1)
		go func(room string) {
			defer wg.Done()
			if m.Summon(s.ID, room) {
				atomic.AddInt32(&winners, 1)
			}
		}(room)
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("winners %d", winners)
	}
	if m.Monk("other-room") {
		t.Fatal("Old Monk stole blue pending summon")
	}
	if got := m.Check("helper"); got != "first" && got != "second" {
		t.Fatalf("room %q", got)
	}
}
