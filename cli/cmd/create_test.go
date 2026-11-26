package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/kube/kubetest"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/tokens"
	"github.com/3nd3r1/kubin/cli/pkg/userconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeClient() kube.Client {
	return kubetest.New(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web"}},
	)
}

// firstEntry returns the name of the first file in a gzipped tar.
func firstEntry(t *testing.T, data []byte) string {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	hdr, err := tar.NewReader(gz).Next()
	require.NoError(t, err)
	return hdr.Name
}

func TestCreateUploadsAndPrintsOnlyTheURL(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	var stderr bytes.Buffer
	log.SetOutput(&stderr)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	var uploaded []byte
	var rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		uploaded, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"abc","url":"http://kubin.test/s/abc","expires_at":"2026-12-01T10:00:00Z","delete_token":"super-secret-token"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, TempDir: tmp}, fakeClient(), &stdout)
	require.NoError(t, err)

	assert.Empty(t, rawQuery, "no expires_in query param when --expire is unset")
	assert.Equal(t, "http://kubin.test/s/abc\n", stdout.String())
	assert.Contains(t, stderr.String(), "Expires 2026-12-01 10:00 UTC. Delete it with: kubin delete http://kubin.test/s/abc")
	assert.NotContains(t, stdout.String(), "super-secret-token")
	assert.NotContains(t, stderr.String(), "super-secret-token")

	tok, ok, err := tokens.Get("http://kubin.test/s/abc")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "super-secret-token", tok)

	assert.Equal(t, "manifest.json", firstEntry(t, uploaded))
	left, _ := os.ReadDir(tmp)
	assert.Empty(t, left, "temporary snapshot is removed after upload")
}

func TestCreateWithExpireSendsQueryParam(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	var gotExpire string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotExpire = r.URL.Query().Get("expires_in")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"abc","url":"http://kubin.test/s/abc","expires_at":"2026-12-01T10:00:00Z","delete_token":"tok-exp"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, TempDir: tmp, Expire: "24h"}, fakeClient(), &stdout)
	require.NoError(t, err)
	assert.Equal(t, "24h", gotExpire)
	assert.Equal(t, "http://kubin.test/s/abc\n", stdout.String())
}

func TestCreateExpireRejectedByServerShowsServerMessage(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"expires_in 1000h0m0s exceeds maximum 720h0m0s"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, TempDir: tmp, Expire: "1000h"}, fakeClient(), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expires_in 1000h0m0s exceeds maximum 720h0m0s")
	assert.Empty(t, stdout.String())
}

func TestCreateOutputWritesNoToken(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	dir := t.TempDir()
	out := filepath.Join(dir, "cluster.tar.gz")
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Output: out, Expire: "24h"}, fakeClient(), &stdout)
	require.NoError(t, err)

	path, err := tokens.Path()
	require.NoError(t, err)
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "--output must not create a token file")
}

func TestParseExpire(t *testing.T) {
	for _, in := range []string{"", "30m", "24h", "1h30m"} {
		assert.NoError(t, parseExpire(in), in)
	}
	for _, in := range []string{"0s", "-1h", "tomorrow"} {
		assert.Error(t, parseExpire(in), in)
	}
}

func TestCreateRemovesTempFileWhenUploadFails(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, TempDir: tmp}, fakeClient(), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "can't reach the kubin server")
	assert.Empty(t, stdout.String())
	left, _ := os.ReadDir(tmp)
	assert.Empty(t, left)
}

func TestCreateOutputWritesFileAndSkipsUpload(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "cluster.tar.gz")
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, Output: out}, fakeClient(), &stdout)
	require.NoError(t, err)

	assert.False(t, called, "no upload with --output")
	assert.Empty(t, stdout.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Equal(t, "manifest.json", firstEntry(t, data))
	entries, _ := os.ReadDir(dir)
	assert.Len(t, entries, 1, "no staging file left next to the output")
}

func TestCreateRefusesSnapshotOverMaxSize(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer srv.Close()

	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, TempDir: tmp, MaxSize: "100"}, fakeClient(), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "over --max-size 100")
	assert.False(t, called, "nothing is uploaded")
	left, _ := os.ReadDir(tmp)
	assert.Empty(t, left)
}

func TestParseSize(t *testing.T) {
	for in, want := range map[string]int64{"": 0, "0": 0, "64Mi": 64 << 20, "200M": 200_000_000, "1500": 1500} {
		got, err := parseSize(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"lots", "-1Mi"} {
		_, err := parseSize(in)
		assert.Error(t, err, in)
	}
}

func withFakeDefaultServer(t *testing.T, url string) {
	t.Helper()
	prev := defaultServer
	defaultServer = url
	t.Cleanup(func() { defaultServer = prev })
}

func TestDefaultServerConstant(t *testing.T) {
	assert.Equal(t, "https://kubin.ender.fi", defaultServer)
}

func TestCreateDefaultServerPromptsOnceThenRemembers(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	var uploads int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		uploads++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pub1","url":"https://kubin.ender.fi/s/pub1"}`))
	}))
	defer srv.Close()
	withFakeDefaultServer(t, srv.URL)

	// First run: asks [y/N], accepts "yes", stores answer.
	var stdout, stderr bytes.Buffer
	err := runCreate(context.Background(), createOptions{
		TempDir:    t.TempDir(),
		Stdin:      strings.NewReader("YES\n"),
		Stderr:     &stderr,
		IsTerminal: func() bool { return true },
	}, fakeClient(), &stdout)
	require.NoError(t, err)
	assert.Equal(t, 1, uploads)
	assert.Equal(t, "https://kubin.ender.fi/s/pub1\n", stdout.String())
	assert.Contains(t, stderr.String(), "Collected kind-kubin (https://127.0.0.1:6443): 1 namespaces, 2 resources, 0 logs,")
	assert.Contains(t, stderr.String(), "[y/N]")
	assert.Contains(t, stderr.String(), "--server or KUBIN_SERVER")

	cfg, err := userconfig.Load()
	require.NoError(t, err)
	assert.True(t, cfg.PublicUploadConfirmed)

	// Second run: skips prompt even without a terminal, still prints summary.
	stdout.Reset()
	stderr.Reset()
	err = runCreate(context.Background(), createOptions{
		TempDir:    t.TempDir(),
		Stdin:      strings.NewReader(""),
		Stderr:     &stderr,
		IsTerminal: func() bool { return false },
	}, fakeClient(), &stdout)
	require.NoError(t, err)
	assert.Equal(t, 2, uploads)
	assert.Equal(t, "https://kubin.ender.fi/s/pub1\n", stdout.String())
	assert.Contains(t, stderr.String(), "Collected kind-kubin (https://127.0.0.1:6443):")
	assert.NotContains(t, stderr.String(), "[y/N]")
}

func TestCreateDefaultServerDeclineUploadsNothingAndCleansTemp(t *testing.T) {
	for name, input := range map[string]string{
		"no":    "n\n",
		"empty": "\n",
		"eof":   "",
		"other": "maybe\n",
	} {
		t.Run(name, func(t *testing.T) {
			cfgDir := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", cfgDir)
			t.Setenv("HOME", cfgDir)

			called := false
			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			defer srv.Close()
			withFakeDefaultServer(t, srv.URL)

			tmp := t.TempDir()
			var stdout, stderr bytes.Buffer
			err := runCreate(context.Background(), createOptions{
				TempDir:    tmp,
				Stdin:      strings.NewReader(input),
				Stderr:     &stderr,
				IsTerminal: func() bool { return true },
			}, fakeClient(), &stdout)
			require.Error(t, err)
			assert.False(t, called)
			assert.Empty(t, stdout.String())
			assert.Contains(t, stderr.String(), "[y/N]")

			left, _ := os.ReadDir(tmp)
			assert.Empty(t, left, "temp snapshot must be cleaned up on decline")

			cfg, err := userconfig.Load()
			require.NoError(t, err)
			assert.False(t, cfg.PublicUploadConfirmed)
		})
	}
}

func TestCreateDefaultServerNonInteractiveRefusesWithoutYes(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	defer srv.Close()
	withFakeDefaultServer(t, srv.URL)

	tmp := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runCreate(context.Background(), createOptions{
		TempDir:    tmp,
		Stdin:      strings.NewReader("y\n"),
		Stderr:     &stderr,
		IsTerminal: func() bool { return false },
	}, fakeClient(), &stdout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--yes")
	assert.Contains(t, err.Error(), "--server")
	assert.Contains(t, err.Error(), "KUBIN_SERVER")
	assert.False(t, called)
	assert.Empty(t, stdout.String())
	left, _ := os.ReadDir(tmp)
	assert.Empty(t, left)
}

func TestCreateDefaultServerYesUploadsWithoutPromptOrSavingConfig(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"yes1","url":"https://kubin.ender.fi/s/yes1"}`))
	}))
	defer srv.Close()
	withFakeDefaultServer(t, srv.URL)

	var stdout, stderr bytes.Buffer
	err := runCreate(context.Background(), createOptions{
		TempDir:    t.TempDir(),
		Yes:        true,
		Stdin:      strings.NewReader(""),
		Stderr:     &stderr,
		IsTerminal: func() bool { return false },
	}, fakeClient(), &stdout)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "https://kubin.ender.fi/s/yes1\n", stdout.String())
	assert.Contains(t, stderr.String(), "Collected kind-kubin")
	assert.NotContains(t, stderr.String(), "[y/N]")

	path, err := userconfig.Path()
	require.NoError(t, err)
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "--yes must not write config file")
}

func TestCreateExplicitServerSkipsSummaryAndPrompt(t *testing.T) {
	cfgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgDir)
	t.Setenv("HOME", cfgDir)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"exp1","url":"http://kubin.test/s/exp1"}`))
	}))
	defer srv.Close()
	// Even if explicit --server matches defaultServer, no summary or prompt runs.
	withFakeDefaultServer(t, srv.URL)

	var stdout, stderr bytes.Buffer
	err := runCreate(context.Background(), createOptions{
		Server:     srv.URL,
		TempDir:    t.TempDir(),
		Stdin:      strings.NewReader(""),
		Stderr:     &stderr,
		IsTerminal: func() bool { return false },
	}, fakeClient(), &stdout)
	require.NoError(t, err)
	assert.Equal(t, "http://kubin.test/s/exp1\n", stdout.String())
	assert.Empty(t, stderr.String())
}
