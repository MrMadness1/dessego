package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestDecrypter_Decrypt(t *testing.T) {
	d, err := NewDecrypter(DefaultAESKey)
	if err != nil {
		t.Fatal(err)
	}

	plain := []byte("ver=100&characterID=foobar&index=1")
	enc, err := testEnc(plain)
	if err != nil {
		t.Fatal(err)
	}

	if dec, err := d.Decrypt(enc); err != nil || !reflect.DeepEqual(plain, dec) {
		t.Fatalf("expected %q got %q", plain, dec)
	}
}

func testEnc(plain []byte) ([]byte, error) {
	// CBC mode works on blocks so plaintexts may need to be padded to the
	// next whole block. For an example of such padding, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
	// assume that the plaintext is already of the correct length.
	bl := (len(plain)/aes.BlockSize + 1) * aes.BlockSize

	// Padding delta.
	d := bl - len(plain)
	for i := 0; i < d; i++ {
		plain = append(plain, byte(d))
	}

	block, err := aes.NewCipher([]byte(DefaultAESKey))
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	enc := make([]byte, aes.BlockSize+len(plain))
	iv := enc[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("read iv: %w", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(enc[aes.BlockSize:], plain)

	return enc, nil
}

func TestDecryptRejectsMalformedRequests(t *testing.T) {
	d, _ := NewDecrypter(DefaultAESKey)
	for n := 0; n < 65; n++ {
		if n >= 32 && (n-16)%16 == 0 {
			continue
		}
		if _, err := d.Decrypt(make([]byte, n)); err == nil {
			t.Fatalf("accepted length %d", n)
		}
	}
	block, _ := aes.NewCipher([]byte(DefaultAESKey))
	for _, pad := range []byte{0, 17, 255, 2} {
		plain := make([]byte, 16)
		plain[15] = pad
		enc := make([]byte, 32)
		cipher.NewCBCEncrypter(block, enc[:16]).CryptBlocks(enc[16:], plain)
		if _, err := d.Decrypt(enc); err == nil {
			t.Fatalf("accepted invalid padding %d", pad)
		}
	}
	for _, plain := range [][]byte{nil, []byte("1234567890123456"), []byte("a")} {
		enc, err := testEnc(plain)
		if err != nil {
			t.Fatal(err)
		}
		dec, err := d.Decrypt(enc)
		if err != nil || string(dec) != string(plain) {
			t.Fatalf("valid plaintext: %q, %v", dec, err)
		}
	}
}
