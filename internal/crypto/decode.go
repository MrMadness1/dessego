package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// DefaultAESKey is the AES encryption key used by Demon's Souls.
//
// Clearly, they were very security conscious...
const DefaultAESKey = "11111111222222223333333344444444"

// Decrypter is an AEA request decrypter.
type Decrypter struct {
	c cipher.Block
}

// NewDecrypter returns a new request decrypter.
func NewDecrypter(key string) (d *Decrypter, err error) {
	d = &Decrypter{}
	d.c, err = aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	return d, nil
}

// Decrypt returns a byte slice containing the decrypted contents of the AES
// encrypted enc.
func (d *Decrypter) Decrypt(enc []byte) ([]byte, error) {
	if len(enc) < 2*aes.BlockSize || (len(enc)-aes.BlockSize)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid encrypted request length")
	}
	// Block-chain mode decrypter, pulling out the initialisation vector
	// (IV) from the body.
	cbc := cipher.NewCBCDecrypter(d.c, enc[:aes.BlockSize])

	// Decrypt everything in the body, following the IV.
	dec := make([]byte, len(enc[aes.BlockSize:]))
	cbc.CryptBlocks(dec, enc[aes.BlockSize:])

	// Trim the trailing characters.
	trim := int(dec[len(dec)-1])
	if trim < 1 || trim > aes.BlockSize || trim > len(dec) {
		return nil, fmt.Errorf("invalid request padding")
	}
	for _, b := range dec[len(dec)-trim:] {
		if int(b) != trim {
			return nil, fmt.Errorf("invalid request padding")
		}
	}
	dec = dec[:len(dec)-trim]

	return dec, nil
}
