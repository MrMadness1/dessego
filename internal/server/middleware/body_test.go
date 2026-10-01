package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDiscardRequestBodyWaitsForSplitPS3POST(t *testing.T) {
	s := httptest.NewServer(DiscardRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "bootstrap")
	})))
	defer s.Close()
	c, err := net.DialTimeout("tcp", s.Listener.Addr().String(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := fmt.Fprint(c, "POST /demons-souls-us/ss.info HTTP/1.1\r\nHost: c.demons-souls.com:18000\r\nConnection: close\r\nContent-Length: 32\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	var early [1]byte
	if n, err := c.Read(early[:]); n != 0 || err == nil {
		t.Fatal("server replied before the POST body arrived")
	} else if timeout, ok := err.(net.Error); !ok || !timeout.Timeout() {
		t.Fatalf("expected read timeout while waiting for body, got %v", err)
	}
	c.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := c.Write(make([]byte, 32)); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(c)
	response, err := http.ReadResponse(reader, &http.Request{Method: http.MethodPost})
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "bootstrap" {
		t.Fatalf("unexpected response: %d %q", response.StatusCode, body)
	}
	if _, err := reader.ReadByte(); err != io.EOF {
		t.Fatalf("expected clean connection close, got %v", err)
	}
}

func TestDiscardRequestBodyRejectsInvalidBodies(t *testing.T) {
	for _, tc := range []struct {
		name     string
		data     []byte
		declared int64
		status   int
	}{
		{"oversized", make([]byte, maxIgnoredRequestBodyBytes+1), maxIgnoredRequestBodyBytes + 1, http.StatusRequestEntityTooLarge},
		{"truncated", []byte("short"), 32, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			h := DiscardRequestBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			r.Body = &incompleteBody{Reader: bytes.NewReader(tc.data), incomplete: int64(len(tc.data)) < tc.declared}
			r.ContentLength = tc.declared
			w := httptest.NewRecorder()
			h(w, r)
			if called || w.Code != tc.status {
				t.Fatalf("handler called=%v, status=%d; want %d", called, w.Code, tc.status)
			}
		})
	}
}

type incompleteBody struct {
	*bytes.Reader
	incomplete bool
}

func (b *incompleteBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	if err == io.EOF && b.incomplete {
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}
func (b *incompleteBody) Close() error { return nil }
