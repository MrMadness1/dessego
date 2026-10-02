package transport

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"testing"

	gamecrypto "github.com/danmrichards/dessego/internal/crypto"
	dsbase64 "github.com/danmrichards/dessego/internal/transport/encoding/base64"
)

// Synthetic tails reproduce native-PS3 failures without publishing player data.
func TestDecodeRequestNativeBase64Tails(t *testing.T) {
	for i, tail := range []string{"z%\x60\xa1\x99\xf2C8??8?P", "\xa3\xe0\xb0??%HT\x0f/8??8?", "\xe9\xea\xac;\x03\xb6?\ni???8?", "H\x8bs;\x13\x0c\xa5+\xafa\xce8?08?08", ";unused", "\x17U?\xbf%>\x84M?\xc18???\xd0", "3\xc4;\xc0"} {
		for _, key := range []string{"replayData", "NPRoomID"} {
			t.Run(fmt.Sprintf("%s/tail%d", key, i), func(t *testing.T) {
				var got struct {
					Replay  string "form:\"replayData\""
					Room    string "form:\"NPRoomID\""
					Version int    "form:\"ver\""
				}
				body := key + "=QUJDRA==" + tail + "&ver=100&"
				if err := DecodeRequest(testRequestDecrypter(t), encryptTestRequest(t, body), &got); err != nil {
					t.Fatal(err)
				}
				value := got.Replay
				if key == "NPRoomID" {
					value = got.Room
				}
				if value != "QUJDRA==" || got.Version != 100 {
					t.Fatalf("value=%q version=%d", value, got.Version)
				}
				decoded, err := dsbase64.StdEncoding.DecodeString(value)
				if err != nil || string(decoded) != "ABCD" {
					t.Fatalf("payload=%q err=%v", decoded, err)
				}
			})
		}
	}
}
func TestDecodeRequestKeepsStrictFormParsing(t *testing.T) {
	for _, body := range []string{"characterID=bad%HT", "characterID=a;b", "replayData=bad%HT", "NPRoomID=bad%HT==tail", "replayData=a;b==tail", "ver=not-a-number"} {
		t.Run(body, func(t *testing.T) {
			var got struct {
				Version   int    "form:\"ver\""
				Character string "form:\"characterID\""
				Replay    string "form:\"replayData\""
				Room      string "form:\"NPRoomID\""
			}
			if err := DecodeRequest(testRequestDecrypter(t), encryptTestRequest(t, body), &got); err == nil {
				t.Fatal("accepted malformed request")
			}
		})
	}
}
func TestDecodeRequestPreservesValidForms(t *testing.T) {
	var got struct {
		Character string "form:\"characterID\""
		Room      string "form:\"NPRoomID\""
		Values    []int  "form:\"sosList\""
		Version   int    "form:\"ver\""
	}
	body := "characterID=a%26b%3Bc%25&NPRoomID=%2B%2FAA%3D%3D&sosList=1&sosList=2&ver=100"
	if err := DecodeRequest(testRequestDecrypter(t), encryptTestRequest(t, body), &got); err != nil {
		t.Fatal(err)
	}
	if got.Character != "a&b;c%" || got.Room != "+/AA==" || len(got.Values) != 2 || got.Values[0] != 1 || got.Values[1] != 2 || got.Version != 100 {
		t.Fatalf("unexpected decode: %+v", got)
	}
	if err := DecodeRequest(testRequestDecrypter(t), encryptTestRequest(t, "NPRoomID=room&ver=100"), &got); err != nil || got.Room != "room" {
		t.Fatalf("room=%q err=%v", got.Room, err)
	}
}
func TestTrimBase64TailsBoundaries(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"both and repeated", "NPRoomID=QUJDRA==%HT;&replayData=QUJDRA==;tail&NPRoomID=QQ==%HT&ver=100", "NPRoomID=QUJDRA==&replayData=QUJDRA==&NPRoomID=QQ==&ver=100"},
		{"encoded alphabet", "NPRoomID=%2B%2FAA==%HT&ver=100", "NPRoomID=%2B%2FAA==&ver=100"},
		{"raw plus", "replayData=+/AA==;tail&ver=100", "replayData=+/AA==&ver=100"},
		{"encoded padding unchanged", "NPRoomID=QUJDRA%3D%3D%HT&ver=100", "NPRoomID=QUJDRA%3D%3D%HT&ver=100"},
		{"ampersand remains delimiter", "replayData=QUJDRA==unused&characterID=bad%HT", "replayData=QUJDRA==&characterID=bad%HT"},
		{"ordinary field unchanged", "characterID=QUJDRA==%HT", "characterID=QUJDRA==%HT"},
		{"exact key only", "nproomid=QUJDRA==%HT", "nproomid=QUJDRA==%HT"},
		{"empty prefix unchanged", "replayData===%HT", "replayData===%HT"},
		{"unpadded unchanged", "replayData=QUJDRA&ver=100", "replayData=QUJDRA&ver=100"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := trimBase64Tails(tc.input); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestDecodeRequestRejectsInvalidTargetsAndCrypto(t *testing.T) {
	d := testRequestDecrypter(t)
	var nilTarget *struct{}
	for _, target := range []interface{}{nil, nilTarget, struct{}{}} {
		if err := DecodeRequest(d, encryptTestRequest(t, "ver=100"), target); err == nil {
			t.Fatal("accepted invalid target")
		}
	}
	var got struct {
		Version int `form:"ver"`
	}
	badPadding := encryptTestRequest(t, "ver=100")
	badPadding[aes.BlockSize-1] ^= 1
	for _, body := range [][]byte{nil, make([]byte, 17), make([]byte, 33), badPadding} {
		if err := DecodeRequest(d, body, &got); err == nil {
			t.Fatal("accepted invalid encrypted request")
		}
	}
}

func testRequestDecrypter(t *testing.T) RequestDecrypter {
	t.Helper()
	d, err := gamecrypto.NewDecrypter(gamecrypto.DefaultAESKey)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func encryptTestRequest(t *testing.T, body string) []byte {
	t.Helper()
	plain := []byte(body)
	padding := aes.BlockSize - len(plain)%aes.BlockSize
	plain = append(plain, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, err := aes.NewCipher([]byte(gamecrypto.DefaultAESKey))
	if err != nil {
		t.Fatal(err)
	}
	enc := make([]byte, aes.BlockSize+len(plain))
	cipher.NewCBCEncrypter(block, enc[:aes.BlockSize]).CryptBlocks(enc[aes.BlockSize:], plain)
	return enc
}
