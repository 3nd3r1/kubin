package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxUnpackedBytes != 1<<30 {
		t.Errorf("MaxUnpackedBytes = %d, want %d", cfg.MaxUnpackedBytes, int64(1<<30))
	}
	if cfg.UploadTimeout != 5*time.Minute {
		t.Errorf("UploadTimeout = %v, want 5m", cfg.UploadTimeout)
	}
	if cfg.UploadRate != 30 {
		t.Errorf("UploadRate = %d, want 30", cfg.UploadRate)
	}
	if cfg.UploadBurst != 10 {
		t.Errorf("UploadBurst = %d, want 10", cfg.UploadBurst)
	}
	if len(cfg.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v, want empty", cfg.TrustedProxies)
	}
}

func TestLoadTrustedProxies(t *testing.T) {
	t.Setenv("SERVER_TRUSTED_PROXIES", "10.0.0.0/8, 172.16.0.0/12, 2001:db8::/32")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.TrustedProxies) != 3 {
		t.Fatalf("got %d prefixes, want 3", len(cfg.TrustedProxies))
	}
	if cfg.TrustedProxies[0].String() != "10.0.0.0/8" || cfg.TrustedProxies[2].String() != "2001:db8::/32" {
		t.Errorf("unexpected prefixes: %v", cfg.TrustedProxies)
	}
}

func TestLoadTrustedProxiesInvalid(t *testing.T) {
	t.Setenv("SERVER_TRUSTED_PROXIES", "not-a-cidr")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}
