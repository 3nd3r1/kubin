package upload

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadSendsTarball(t *testing.T) {
	var got []byte
	var rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/snapshots", r.URL.Path)
		rawQuery = r.URL.RawQuery
		assert.Equal(t, "application/gzip", r.Header.Get("Content-Type"))
		assert.Equal(t, "kubin/test", r.Header.Get("User-Agent"))
		got, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"abc","url":"http://kubin/s/abc","expires_at":"2026-12-01T10:00:00Z","delete_token":"secret-tok"}`))
	}))
	defer srv.Close()

	res, err := New(srv.URL+"/", "kubin/test").Upload(context.Background(), strings.NewReader("tarball"), 7, "")
	require.NoError(t, err)
	assert.Empty(t, rawQuery)
	assert.Equal(t, Result{
		ID:          "abc",
		URL:         "http://kubin/s/abc",
		ExpiresAt:   time.Date(2026, 12, 1, 10, 0, 0, 0, time.UTC),
		DeleteToken: "secret-tok",
	}, res)
	assert.Equal(t, "tarball", string(got))
}

func TestUploadSendsExpireQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("expires_in")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"abc","url":"http://kubin/s/abc"}`))
	}))
	defer srv.Close()

	_, err := New(srv.URL, "kubin/test").Upload(context.Background(), strings.NewReader("tarball"), 7, "24h")
	require.NoError(t, err)
	assert.Equal(t, "24h", gotQuery)
}

func TestUploadErrors(t *testing.T) {
	refused := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"expires_in 1000h0m0s exceeds maximum 720h0m0s"}`))
	}))
	defer refused.Close()
	html := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html>hello</html>`))
	}))
	defer html.Close()
	proxyErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxyErr.Close()
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()

	for _, tt := range []struct {
		name, server, want string
	}{
		{"server message", refused.URL, "refused the snapshot: expires_in 1000h0m0s exceeds maximum 720h0m0s (400 Bad Request)"},
		{"no message", proxyErr.URL, "refused the snapshot: 502 Bad Gateway"},
		{"not kubin", html.URL, "is it a kubin server?"},
		{"unreachable", down.URL, "can't reach the kubin server at " + down.URL},
		{"no scheme", "kubin.example.com", "must be a URL like"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.server, "").Upload(context.Background(), strings.NewReader("x"), 1, "")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}

func TestDeleteSuccess(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	err := New(srv.URL+"/", "kubin/test").Delete(context.Background(), "abc123", "tok-secret")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/api/v1/snapshots/abc123", gotPath)
	assert.Equal(t, "Bearer tok-secret", gotAuth)
	assert.Equal(t, "kubin/test", gotUA)
}

func TestDeleteErrors(t *testing.T) {
	forbidden := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"invalid delete token"}`))
	}))
	defer forbidden.Close()

	notFound := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"snapshot not found"}`))
	}))
	defer notFound.Close()

	proxyErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxyErr.Close()

	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()

	err := New(forbidden.URL, "").Delete(context.Background(), "abc", "wrong")
	require.ErrorIs(t, err, ErrForbidden)
	assert.NotContains(t, err.Error(), "wrong")

	err = New(notFound.URL, "").Delete(context.Background(), "abc", "tok")
	require.ErrorIs(t, err, ErrNotFound)

	err = New(proxyErr.URL, "").Delete(context.Background(), "abc", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "502 Bad Gateway")

	err = New(down.URL, "").Delete(context.Background(), "abc", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't reach the kubin server")

	err = New("kubin.example.com", "").Delete(context.Background(), "abc", "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a URL like")
}
