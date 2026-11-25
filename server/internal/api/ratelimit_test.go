package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/server/internal/snapshot"
)

type stubUploader struct {
	calls       int
	lastTTL     time.Duration
	deleteCalls int
}

func (u *stubUploader) Upload(_ context.Context, r io.Reader, ttl time.Duration) (snapshot.UploadResult, error) {
	u.calls++
	u.lastTTL = ttl
	_, _ = io.Copy(io.Discard, r)
	return snapshot.UploadResult{
		ID:          "testid123456789012345678",
		ExpiresAt:   time.Date(2026, 12, 2, 12, 0, 0, 0, time.UTC),
		DeleteToken: "stub-delete-token",
	}, nil
}

func (u *stubUploader) Delete(_ context.Context, _, _ string) error {
	u.deleteCalls++
	return nil
}

type failOnReadBody struct {
	readCalled bool
}

func (b *failOnReadBody) Read(_ []byte) (int, error) {
	b.readCalled = true
	return 0, io.ErrUnexpectedEOF
}

func TestUploadRateLimit(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	up := &stubUploader{}
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      up,
		MaxUploadBytes: 1 << 20,
		UploadRate:     30, // 1 per 2 minutes
		UploadBurst:    2,
		Now:            func() time.Time { return now },
	})

	post := func(remoteAddr string, body io.Reader) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", body)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	// First two uploads from client A fit in the burst.
	for i := range 2 {
		rec := post("198.51.100.10:12345", bytes.NewReader([]byte("ok")))
		if rec.Code != http.StatusCreated {
			t.Fatalf("upload %d: got %d %s, want 201", i+1, rec.Code, rec.Body.String())
		}
	}

	// Third upload from client A is rejected before reading the body.
	unread := &failOnReadBody{}
	rec := post("198.51.100.10:54321", unread)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("got %d %s, want 429", rec.Code, rec.Body.String())
	}
	if unread.readCalled {
		t.Error("rate-limited request body was read")
	}
	if up.calls != 2 {
		t.Errorf("uploader called %d times, want 2", up.calls)
	}
	retryAfter, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	if err != nil || retryAfter != 120 {
		t.Errorf("Retry-After = %q (err %v), want 120", rec.Header().Get("Retry-After"), err)
	}
	var errResp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("invalid JSON error: %v", err)
	}
	if !strings.Contains(errResp.Error, "too many uploads") || !strings.Contains(errResp.Error, "2m") {
		t.Errorf("error = %q, want mention of too many uploads and 2m", errResp.Error)
	}

	// Another client is unaffected.
	if rec := post("198.51.100.11:11111", bytes.NewReader([]byte("ok"))); rec.Code != http.StatusCreated {
		t.Fatalf("other client got %d %s, want 201", rec.Code, rec.Body.String())
	}

	// After 2 minutes, client A gets a token back.
	now = now.Add(2 * time.Minute)
	if rec := post("198.51.100.10:9999", bytes.NewReader([]byte("ok"))); rec.Code != http.StatusCreated {
		t.Fatalf("after refill got %d %s, want 201", rec.Code, rec.Body.String())
	}
}

func TestUploadRateLimitIPv6Prefix(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      &stubUploader{},
		MaxUploadBytes: 1 << 20,
		UploadRate:     30,
		UploadBurst:    1,
		Now:            func() time.Time { return now },
	})

	post := func(remoteAddr string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", bytes.NewReader([]byte("ok")))
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := post("[2001:db8:abcd:12::1]:1000"); code != http.StatusCreated {
		t.Fatalf("first IPv6 upload got %d, want 201", code)
	}
	// Different host in the same /64 shares the bucket.
	if code := post("[2001:db8:abcd:12::beef]:2000"); code != http.StatusTooManyRequests {
		t.Fatalf("same /64 got %d, want 429", code)
	}
	// Different /64 has its own bucket.
	if code := post("[2001:db8:abcd:13::1]:3000"); code != http.StatusCreated {
		t.Fatalf("different /64 got %d, want 201", code)
	}
}

func TestUploadRateLimitDisabledWhenZero(t *testing.T) {
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      &stubUploader{},
		MaxUploadBytes: 1 << 20,
		UploadRate:     0,
		UploadBurst:    1,
	})
	for i := range 20 {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", bytes.NewReader([]byte("ok")))
		req.RemoteAddr = "198.51.100.10:12345"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("upload %d got %d, want 201", i+1, rec.Code)
		}
	}
}

func TestUploadRateLimitTrustedProxies(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      &stubUploader{},
		MaxUploadBytes: 1 << 20,
		UploadRate:     30,
		UploadBurst:    1,
		TrustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
		Now:            func() time.Time { return now },
	})

	post := func(remoteAddr, xff string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", bytes.NewReader([]byte("ok")))
		req.RemoteAddr = remoteAddr
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}

	// From a trusted proxy, right-most non-trusted IP in XFF is the client.
	if code := post("10.1.2.3:4000", "203.0.113.5, 10.9.9.9"); code != http.StatusCreated {
		t.Fatalf("first proxied upload got %d, want 201", code)
	}
	// Spoofing a different left-most IP does not bypass the limit for 203.0.113.5.
	if code := post("10.1.2.4:5000", "198.51.100.7, 203.0.113.5, 10.0.0.1"); code != http.StatusTooManyRequests {
		t.Fatalf("spoofed left-most XFF got %d, want 429", code)
	}
	// A different client behind the same proxy is allowed.
	if code := post("10.1.2.3:4001", "203.0.113.6"); code != http.StatusCreated {
		t.Fatalf("second client behind proxy got %d, want 201", code)
	}

	// From an untrusted address, X-Forwarded-For is ignored.
	if code := post("192.0.2.50:4000", "203.0.113.99"); code != http.StatusCreated {
		t.Fatalf("first direct upload got %d, want 201", code)
	}
	if code := post("192.0.2.50:4001", "203.0.113.100"); code != http.StatusTooManyRequests {
		t.Fatalf("untrusted XFF should be ignored: got %d, want 429", code)
	}
}

func TestUploadLimiterEvictsIdleClients(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	lim := newUploadLimiter(30, 10, nil, func() time.Time { return now })

	for i := range 100 {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", nil)
		req.RemoteAddr = "198.51.100." + strconv.Itoa(i+1) + ":1234"
		if ok, _ := lim.allow(req); !ok {
			t.Fatalf("client %d unexpectedly rejected", i)
		}
	}
	if got := lim.len(); got != 100 {
		t.Fatalf("len = %d, want 100", got)
	}

	// Advance past the idle TTL; the next request triggers eviction.
	now = now.Add(time.Hour)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	if ok, _ := lim.allow(req); !ok {
		t.Fatal("fresh client rejected")
	}
	if got := lim.len(); got != 1 {
		t.Fatalf("len after eviction = %d, want 1", got)
	}
}
