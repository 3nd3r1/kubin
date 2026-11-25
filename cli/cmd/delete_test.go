package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitSnapshotURL(t *testing.T) {
	tests := []struct {
		in        string
		wantBase  string
		wantID    string
		wantCanon string
		wantErr   bool
	}{
		{
			in:        "https://kubin.example.com/s/7fK2q",
			wantBase:  "https://kubin.example.com",
			wantID:    "7fK2q",
			wantCanon: "https://kubin.example.com/s/7fK2q",
		},
		{
			in:        "https://example.com/sub/s/7fK2q/",
			wantBase:  "https://example.com/sub",
			wantID:    "7fK2q",
			wantCanon: "https://example.com/sub/s/7fK2q",
		},
		{
			in:        "https://example.com/s/outer/s/inner",
			wantBase:  "https://example.com/s/outer",
			wantID:    "inner",
			wantCanon: "https://example.com/s/outer/s/inner",
		},
		{
			in:        "http://localhost:8080/s/abc?ns=default",
			wantBase:  "http://localhost:8080",
			wantID:    "abc",
			wantCanon: "http://localhost:8080/s/abc",
		},
		{in: "not-a-url", wantErr: true},
		{in: "https://kubin.example.com/abc", wantErr: true},
		{in: "https://kubin.example.com/s/", wantErr: true},
		{in: "https://kubin.example.com/s/abc/pods", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			base, id, canon, err := splitSnapshotURL(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBase, base)
			assert.Equal(t, tt.wantID, id)
			assert.Equal(t, tt.wantCanon, canon)
		})
	}
}

func TestDeleteWithStoredToken(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	var gotMethod, gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	snapURL := srv.URL + "/s/abc123"
	require.NoError(t, tokens.Set(snapURL, "stored-secret"))

	var stderr bytes.Buffer
	log.SetOutput(&stderr)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	err := runDelete(context.Background(), snapURL, deleteOptions{})
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/api/v1/snapshots/abc123", gotPath)
	assert.Equal(t, "Bearer stored-secret", gotAuth)
	assert.NotContains(t, stderr.String(), "stored-secret")

	_, ok, err := tokens.Get(snapURL)
	require.NoError(t, err)
	assert.False(t, ok, "stored token should be removed after delete")
}

func TestDeleteWithTokenFlag(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	snapURL := srv.URL + "/s/abc123"
	require.NoError(t, tokens.Set(snapURL, "old-token"))

	err := runDelete(context.Background(), snapURL, deleteOptions{Token: "flag-token"})
	require.NoError(t, err)
	assert.Equal(t, "Bearer flag-token", gotAuth)

	_, ok, err := tokens.Get(snapURL)
	require.NoError(t, err)
	assert.False(t, ok, "stored token should be removed after delete")
}

func TestDeleteNotFoundRemovesStoredTokenAndExitsZero(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"snapshot not found"}`))
	}))
	defer srv.Close()

	snapURL := srv.URL + "/s/gone"
	require.NoError(t, tokens.Set(snapURL, "tok-gone"))

	var stderr bytes.Buffer
	log.SetOutput(&stderr)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	err := runDelete(context.Background(), snapURL, deleteOptions{})
	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "not found")
	assert.NotContains(t, stderr.String(), "tok-gone")

	_, ok, err := tokens.Get(snapURL)
	require.NoError(t, err)
	assert.False(t, ok, "stored token should be removed on 404")
}

func TestDeleteWrongTokenReturnsErrorAndKeepsStoredToken(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"invalid delete token"}`))
	}))
	defer srv.Close()

	snapURL := srv.URL + "/s/abc"
	require.NoError(t, tokens.Set(snapURL, "real-token"))

	err := runDelete(context.Background(), snapURL, deleteOptions{Token: "bad-token"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid delete token")
	assert.NotContains(t, err.Error(), "bad-token")

	tok, ok, err := tokens.Get(snapURL)
	require.NoError(t, err)
	assert.True(t, ok, "stored token should not be removed on 403")
	assert.Equal(t, "real-token", tok)
}

func TestDeleteMissingTokenReturnsError(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	defer srv.Close()

	err := runDelete(context.Background(), srv.URL+"/s/abc", deleteOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no delete token stored")
	assert.Contains(t, err.Error(), "--token")
	assert.False(t, called)
}
