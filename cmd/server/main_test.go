package main

import (
	"os"
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	const key = "DESSEGO_TEST_SETTING"
	oldValue, hadValue := os.LookupEnv(key)
	defer func() {
		if hadValue {
			_ = os.Setenv(key, oldValue)
			return
		}
		_ = os.Unsetenv(key)
	}()

	if err := os.Unsetenv(key); err != nil {
		t.Fatal(err)
	}
	if got := envOrDefault(key, "fallback"); got != "fallback" {
		t.Fatalf("envOrDefault() = %q, want fallback", got)
	}

	if err := os.Setenv(key, "configured"); err != nil {
		t.Fatal(err)
	}
	if got := envOrDefault(key, "fallback"); got != "configured" {
		t.Fatalf("envOrDefault() = %q, want configured", got)
	}
}
