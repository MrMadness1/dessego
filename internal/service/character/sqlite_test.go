//go:build cgo
// +build cgo

package character

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSQLiteUpdatePlayerGrade(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ddl, err := os.ReadFile("character.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(ddl)); err != nil {
		t.Fatal(err)
	}
	s := &SQLiteService{db: db}
	id := "player' OR 1=1 --"
	for _, player := range []string{id, "other-player"} {
		if _, err := db.Exec("INSERT INTO character (id, sessions, msg_rating) VALUES (?, 7, 9)", player); err != nil {
			t.Fatal(err)
		}
	}
	grades := []MultiplayerGrade{GradeS, GradeA, GradeB, GradeC, GradeD}
	for i, grade := range grades {
		for n := 0; n <= i; n++ {
			if err := s.UpdatePlayerGrade(id, grade); err != nil {
				t.Fatalf("%s: %v", grade, err)
			}
		}
	}
	if err := s.UpdatePlayerGrade(id, GradeUnknown); err != nil {
		t.Fatalf("unrated session: %v", err)
	}
	if err := s.UpdatePlayerGrade(id, MultiplayerGrade("grade_s = 99; --")); err == nil {
		t.Fatal("invalid grade accepted")
	}
	for _, player := range []string{id, "other-player"} {
		var got [7]int
		if err := db.QueryRow("SELECT grade_s, grade_a, grade_b, grade_c, grade_d, sessions, msg_rating FROM character WHERE id = ?", player).Scan(&got[0], &got[1], &got[2], &got[3], &got[4], &got[5], &got[6]); err != nil {
			t.Fatal(err)
		}
		want := [7]int{0, 0, 0, 0, 0, 7, 9}
		if player == id {
			want = [7]int{1, 2, 3, 4, 5, 7, 9}
		}
		if got != want {
			t.Fatalf("%q: got %v, want %v", player, got, want)
		}
	}
}
