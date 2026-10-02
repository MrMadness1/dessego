package transport

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/danmrichards/dessego/internal/transport/encoding/form"
)

// RequestDecrypter is the interface that wraps the basic Decrypt method.
//
// Decrypt returns a byte slice containing the decrypted contents of the given
// input byte slice.
type RequestDecrypter interface {
	Decrypt([]byte) ([]byte, error)
}

// DecodeRequest decodes the bytes from a request body into the target, v.
func DecodeRequest(rd RequestDecrypter, data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return InvalidDecodeTargetError{Type: reflect.TypeOf(v)}
	}

	// Demon's Souls sends it's request body as an AES encrypted version of
	// a standard HTTP form POST.
	req, err := rd.Decrypt(data)
	if err != nil {
		return fmt.Errorf("decrypt request: %w", err)
	}

	// Native PS3 clients can append unused bytes after Base64 padding. Strip
	// those tails before the form parser interprets them as URL syntax.
	vals, err := url.ParseQuery(trimBase64Tails(string(req)))
	if err != nil {
		return fmt.Errorf("parse request vals: %w", err)
	}

	if err = form.NewDecoder(vals).Decode(v); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}

	return nil
}

// trimBase64Tails handles the two opaque Base64 fields used by the game.
// The existing game Base64 decoder stops at padding; bytes after it are not
// part of the payload. Do not relax URL parsing for ordinary fields or for
// malformed data before padding.
func trimBase64Tails(body string) string {
	fields := strings.Split(body, "&")
	for i, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || (parts[0] != "replayData" && parts[0] != "NPRoomID") {
			continue
		}
		value := parts[1]
		padding := strings.IndexByte(value, '=')
		if padding < 0 {
			continue
		}
		prefix, err := url.QueryUnescape(value[:padding])
		if err != nil || prefix == "" {
			continue
		}
		valid := true
		for _, c := range prefix {
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
				(c >= '0' && c <= '9') || c == '+' || c == '/' || c == ' ') {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		end := padding
		for end < len(value) && value[end] == '=' {
			end++
		}
		fields[i] = parts[0] + "=" + value[:end]
	}
	return strings.Join(fields, "&")
}
