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
	"testing"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sversion "k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func fakeClient() kube.Client {
	cs := fake.NewClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web"}},
	)
	cs.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &k8sversion.Info{GitVersion: "v1.33.1"}
	return kube.NewFromClientset(cs, "kind-kubin", "kind-kubin")
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
	var uploaded []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploaded, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"abc","url":"http://kubin.test/s/abc"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	var stdout bytes.Buffer
	err := runCreate(context.Background(), createOptions{Server: srv.URL, TempDir: tmp}, fakeClient(), &stdout)
	require.NoError(t, err)

	assert.Equal(t, "http://kubin.test/s/abc\n", stdout.String())
	assert.Equal(t, "manifest.json", firstEntry(t, uploaded))
	left, _ := os.ReadDir(tmp)
	assert.Empty(t, left, "temporary snapshot is removed after upload")
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
