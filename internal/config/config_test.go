package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"PORT", "ADDR", "MAX_PAGE_SIZE", "AUTH_TOKEN", "RATE_LIMIT"} {
		os.Unsetenv(k)
	}
	cfg := Load()
	if cfg.Addr != ":8080" {
		t.Fatalf("addr = %s", cfg.Addr)
	}
	if cfg.MaxPageSize != 100 {
		t.Fatalf("max page size = %d", cfg.MaxPageSize)
	}
	if cfg.AuthToken != "" {
		t.Fatalf("auth token should be empty")
	}
	if cfg.RateLimit != 0 {
		t.Fatalf("rate limit should be 0")
	}
}

func TestLoadWithEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("AUTH_TOKEN", "secret")
	os.Setenv("RATE_LIMIT", "60")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("AUTH_TOKEN")
		os.Unsetenv("RATE_LIMIT")
	}()
	cfg := Load()
	if cfg.Addr != ":9090" || cfg.AuthToken != "secret" || cfg.RateLimit != 60 {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestLoadAddrOverridesPort(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("ADDR", "127.0.0.1:7070")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("ADDR")
	}()
	if cfg := Load(); cfg.Addr != "127.0.0.1:7070" {
		t.Fatalf("addr = %s", cfg.Addr)
	}
}

func TestLoadInvalidRateLimitFallsBack(t *testing.T) {
	os.Setenv("RATE_LIMIT", "abc")
	defer os.Unsetenv("RATE_LIMIT")
	if cfg := Load(); cfg.RateLimit != 0 {
		t.Fatalf("rate limit = %d, want 0", cfg.RateLimit)
	}
}
