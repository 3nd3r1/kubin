package upload

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadSendsTarball(t *testing.T) {
	var got []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/snapshots", r.URL.Path)
		assert.Equal(t, "application/gzip", r.Header.Get("Content-Type"))
		assert.Equal(t, "kubin/test", r.Header.Get("User-Agent"))
		got, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"abc","url":"http://kubin/s/abc"}`))
	}))
	defer srv.Close()

	res, err := New(srv.URL+"/", "kubin/test").Upload(context.Background(), strings.NewReader("tarball"), 7)
	require.NoError(t, err)
	assert.Equal(t, Result{ID: "abc", URL: "http://kubin/s/abc"}, res)
	assert.Equal(t, "tarball", string(got))
}

func TestUploadErrors(t *testing.T) {
	refused := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid snapshot: unsupported format_version 2"}`))
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
		{"server message", refused.URL, "refused the snapshot: invalid snapshot: unsupported format_version 2 (400 Bad Request)"},
		{"no message", proxyErr.URL, "refused the snapshot: 502 Bad Gateway"},
		{"not kubin", html.URL, "is it a kubin server?"},
		{"unreachable", down.URL, "can't reach the kubin server at " + down.URL},
		{"no scheme", "kubin.example.com", "must be a URL like"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.server, "").Upload(context.Background(), strings.NewReader("x"), 1)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
