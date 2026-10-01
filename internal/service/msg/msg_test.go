package msg

import (
	"reflect"
	"testing"
)

func TestNewBloodMsgFromBytes(t *testing.T) {
	exp := &BloodMsg{
		ID:           1,
		CharacterID:  "test",
		BlockID:      -1,
		PosX:         1.2,
		PosY:         2.3,
		PosZ:         3.4,
		AngX:         5.6,
		AngY:         6.7,
		AngZ:         7.8,
		MsgID:        9,
		MainMsgID:    10,
		AddMsgCateID: 11,
		Rating:       12,
		Legacy:       1,
	}

	bm, err := NewBloodMsgFromBytes(exp.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(exp, bm) {
		t.Fatalf("expected: %+v\ngot: %+v", exp, bm)
	}
}

func TestBloodMessageParserRejectsTruncation(t *testing.T) {
	data := (BloodMsg{CharacterID: "test", AngX: 1, AngZ: 3}).Bytes()
	for n := 0; n < len(data); n++ {
		if _, err := NewBloodMsgFromBytes(data[:n]); err == nil {
			t.Fatalf("accepted truncation %d", n)
		}
	}
	if _, err := NewBloodMsgFromBytes([]byte("no terminator")); err == nil {
		t.Fatal("accepted missing terminator")
	}
}
