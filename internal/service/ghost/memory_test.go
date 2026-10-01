package ghost

import (
	"github.com/rs/zerolog"
	"reflect"
	"testing"
	"time"
)

func TestMemory_Get(t *testing.T) {
	gt := time.Date(2020, 11, 17, 13, 0, 0, 0, time.UTC)

	gm := Memory{
		ghosts: map[string]*Ghost{
			"test234": {
				BlockID:     123,
				CharacterID: "test234",
				timestamp:   gt,
			},
		},
	}

	tcs := []struct {
		name        string
		characterID string
		blockID     int32
		expGhosts   []*Ghost
	}{
		{
			name:        "non-matching character",
			characterID: "test123",
			blockID:     123,
			expGhosts: []*Ghost{{
				BlockID:     123,
				CharacterID: "test234",
				timestamp:   gt,
			}},
		},
		{
			name:        "matching character",
			characterID: "test234",
			blockID:     123,
			expGhosts:   []*Ghost{},
		},
		{
			name:        "incorrect block",
			characterID: "test123",
			blockID:     456,
			expGhosts:   []*Ghost{},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ghosts := gm.Get(tc.characterID, tc.blockID, 10)
			if !reflect.DeepEqual(ghosts, tc.expGhosts) {
				t.Fatalf("expected: %d ghosts got: %d", len(tc.expGhosts), len(ghosts))
			}
		})
	}
}

func TestMemory_ClearBefore(t *testing.T) {
	gt := time.Date(2020, 11, 17, 13, 0, 0, 0, time.UTC)

	gm := Memory{
		ghosts: map[string]*Ghost{
			"test234": {
				BlockID:     123,
				CharacterID: "test234",
				timestamp:   gt,
			},
			"test456": {
				BlockID:     123,
				CharacterID: "test456",
				timestamp:   gt.Add(-1 * time.Minute),
			},
			"test678": {
				BlockID:     123,
				CharacterID: "test678",
				timestamp:   gt.Add(-30 * time.Second),
			},
		},
	}

	gm.ClearBefore(gt.Add(-35 * time.Second))

	expGhosts := map[string]*Ghost{
		"test234": {
			BlockID:     123,
			CharacterID: "test234",
			timestamp:   gt,
		},
		"test678": {
			BlockID:     123,
			CharacterID: "test678",
			timestamp:   gt.Add(-30 * time.Second),
		},
	}

	if !reflect.DeepEqual(expGhosts, gm.ghosts) {
		t.Fatalf("expected %d ghosts got %d", len(expGhosts), len(gm.ghosts))
	}
}

func TestMemoryExpiresGhostsThroughPublicSet(t *testing.T) {
	gm := NewMemory(zerolog.Nop())
	now := time.Now()
	for _, id := range []string{"old-a", "old-b", "fresh"} {
		g := NewGhost(123, id, nil)
		if id != "fresh" {
			g.timestamp = now.Add(-time.Hour)
		}
		gm.Set(id, g)
	}
	gm.ClearBefore(now.Add(-30 * time.Second))
	if got := gm.Get("viewer", 123, 10); len(got) != 1 || got[0].CharacterID != "fresh" {
		t.Fatalf("ghost cleanup: %v", got)
	}
	gm.ClearBefore(now.Add(-30 * time.Second))
	for _, n := range []int{-1, 0} {
		if len(gm.Get("viewer", 123, n)) != 0 {
			t.Fatal("nonpositive count returned ghosts")
		}
	}
}
