package snapshot

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

// readArchive returns the entry names in order and their contents.
func readArchive(t *testing.T, data []byte) ([]string, map[string][]byte) {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	tr := tar.NewReader(gz)
	var names []string
	files := map[string][]byte{}
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		assert.Equal(t, byte(tar.TypeReg), hdr.Typeflag, "%s is a regular file", hdr.Name)
		b, err := io.ReadAll(tr)
		require.NoError(t, err)
		names = append(names, hdr.Name)
		files[hdr.Name] = b
	}
	return names, files
}

func TestCreateWritesV1Layout(t *testing.T) {
	managed := []metav1.ManagedFieldsEntry{{Manager: "kubectl", Operation: metav1.ManagedFieldsOperationApply}}
	cs := fake.NewClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "prod"}},
		// Same name in two namespaces: both must be kept.
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "web", ManagedFields: managed}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "prod", Name: "web", ManagedFields: managed}},
	)
	cs.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{GitVersion: "v1.33.1"}

	mgr := NewManager(kube.NewFromClientset(cs, "kind-kubin", "kind-kubin"), "v0.0.0-test")
	created := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	mgr.now = func() time.Time { return created }

	var buf bytes.Buffer
	_, err := mgr.Create(context.Background(), &buf)
	require.NoError(t, err)

	names, files := readArchive(t, buf.Bytes())
	require.NotEmpty(t, names)
	assert.Equal(t, "manifest.json", names[0], "manifest is the first entry")
	assert.True(t, sort.StringsAreSorted(names[1:]), "entries after the manifest are sorted: %v", names)
	assert.Equal(t, []string{
		"manifest.json",
		"resources/core/v1/Namespace/_cluster/default.json",
		"resources/core/v1/Namespace/_cluster/prod.json",
		"resources/core/v1/Pod/default/web.json",
		"resources/core/v1/Pod/prod/web.json",
	}, names)

	var m Manifest
	require.NoError(t, json.Unmarshal(files["manifest.json"], &m))
	assert.Equal(t, Manifest{
		FormatVersion: 1,
		CreatedAt:     created,
		CLIVersion:    "v0.0.0-test",
		Cluster:       Cluster{Context: "kind-kubin", Name: "kind-kubin", ServerVersion: "v1.33.1"},
		Resources: []KindCount{
			{Group: "core", Version: "v1", Kind: "Namespace", Count: 2},
			{Group: "core", Version: "v1", Kind: "Pod", Count: 2},
		},
	}, m)

	var pod map[string]any
	require.NoError(t, json.Unmarshal(files["resources/core/v1/Pod/prod/web.json"], &pod))
	assert.Equal(t, "v1", pod["apiVersion"])
	assert.Equal(t, "Pod", pod["kind"])
	meta := pod["metadata"].(map[string]any)
	assert.Equal(t, "prod", meta["namespace"])
	assert.NotContains(t, meta, "managedFields")
}

func TestDefaultFileName(t *testing.T) {
	got := DefaultFileName(time.Date(2026, 11, 17, 10, 4, 5, 0, time.FixedZone("EET", 7200)))
	assert.Equal(t, "kubin-snapshot-20261117-080405.tar.gz", got)
}
