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
	if cfg.SnapshotTTL != 168*time.Hour {
		t.Errorf("SnapshotTTL = %v, want 168h", cfg.SnapshotTTL)
	}
	if cfg.SnapshotMaxTTL != 720*time.Hour {
		t.Errorf("SnapshotMaxTTL = %v, want 720h", cfg.SnapshotMaxTTL)
	}
	if cfg.CleanupInterval != 10*time.Minute {
		t.Errorf("CleanupInterval = %v, want 10m", cfg.CleanupInterval)
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

func TestLoadExpiryValidation(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"zero ttl", map[string]string{"SERVER_SNAPSHOT_TTL": "0s"}},
		{"negative ttl", map[string]string{"SERVER_SNAPSHOT_TTL": "-1h"}},
		{"zero max ttl", map[string]string{"SERVER_SNAPSHOT_MAX_TTL": "0s"}},
		{"negative max ttl", map[string]string{"SERVER_SNAPSHOT_MAX_TTL": "-1h"}},
		{"ttl above max ttl", map[string]string{"SERVER_SNAPSHOT_TTL": "48h", "SERVER_SNAPSHOT_MAX_TTL": "24h"}},
		{"zero cleanup interval", map[string]string{"SERVER_CLEANUP_INTERVAL": "0s"}},
		{"negative cleanup interval", map[string]string{"SERVER_CLEANUP_INTERVAL": "-5m"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if _, err := Load(); err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}
}
