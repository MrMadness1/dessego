//go:build cgo
// +build cgo

package character

import (
	"bytes"
	"database/sql"
	"github.com/danmrichards/dessego/internal/service/msg"
	"github.com/danmrichards/dessego/internal/service/replay"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"testing"
)

func TestExistingTablesGainIndexesAndReplayListsOmitPayload(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(filepath.Clean(filepath.Join(original, "../../.."))); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	// Start with the old table-only schema, as existing deployments do.
	for _, service := range []string{"msg", "replay"} {
		ddl, err := os.ReadFile("internal/service/" + service + "/ddl.sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(string(ddl)); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if _, err := msg.NewSQLiteService(db, zerolog.Nop()); err != nil {
			t.Fatal(err)
		}
		if _, err := replay.NewSQLiteService(db, zerolog.Nop()); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"message_block_legacy_character", "replay_block_legacy"} {
		var count int
		if err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?", name).Scan(&count); err != nil || count != 1 {
			t.Fatalf("index %s count %d: %v", name, count, err)
		}
	}
	rs, err := replay.NewSQLiteService(db, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	want := &replay.Replay{ID: 123, CharacterID: "synthetic", BlockID: 40070, PosX: 1, PosY: 2, PosZ: 3, AngX: 4, AngY: 5, AngZ: 6, MsgID: 7, MainMsgID: 8, AddMsgCateID: 9, Data: bytes.Repeat([]byte("payload"), 4096)}
	if err = rs.Add(want); err != nil {
		t.Fatal(err)
	}
	list, err := rs.List(40070, 1, replay.NonLegacy)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || len(list[0].Data) != 0 || !bytes.Equal(list[0].Header(), want.Header()) {
		t.Fatalf("header changed or payload loaded: %v", list)
	}
	got, err := rs.Get(123)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Data, want.Data) || !bytes.Equal(got.Header(), want.Header()) {
		t.Fatal("full replay changed")
	}
}
