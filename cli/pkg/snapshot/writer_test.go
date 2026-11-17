package snapshot

import (
	"errors"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestWriter(t *testing.T) *Writer {
	t.Helper()
	w, err := NewWriter()
	require.NoError(t, err)
	t.Cleanup(func() { _ = w.Cleanup() })
	return w
}

func res(namespace, name string) collector.Resource {
	return collector.Resource{
		Version: "v1", Kind: "Pod", Namespace: namespace, Name: name,
		Object: map[string]any{"metadata": map[string]any{"name": name}},
	}
}

func TestAddResourceRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "a/b", `a\b`, "../../etc/passwd"} {
		err := newTestWriter(t).AddResource(res("default", name))
		assert.Error(t, err, "name %q", name)
	}
	assert.Error(t, newTestWriter(t).AddResource(res("..", "web")), "namespace ..")
}

func TestAddResourceRejectsDuplicates(t *testing.T) {
	w := newTestWriter(t)
	require.NoError(t, w.AddResource(res("default", "web")))
	assert.ErrorContains(t, w.AddResource(res("default", "web")), "collected twice")
}

type failingWriter struct{ after int }

func (f *failingWriter) Write(p []byte) (int, error) {
	if f.after <= 0 {
		return 0, errors.New("disk full")
	}
	f.after--
	return len(p), nil
}

func TestFinishReturnsWriteErrors(t *testing.T) {
	w := newTestWriter(t)
	require.NoError(t, w.AddResource(res("default", "web")))
	m := Manifest{FormatVersion: FormatVersion, CreatedAt: time.Now()}

	// gzip buffers, so after its header most bytes only reach out when it is
	// closed. Failing at each of the first writes covers both the error from
	// writing an entry and the one from closing.
	for after := 0; after < 3; after++ {
		err := w.Finish(m, &failingWriter{after: after})
		assert.ErrorContains(t, err, "disk full", "failing after %d writes", after)
	}
}
