package tokens

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setTempConfigEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	return dir
}

func TestSetGetDeleteAndPermissions(t *testing.T) {
	setTempConfigEnv(t)

	tok, ok, err := Get("https://kubin.test/s/abc")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, tok)

	require.NoError(t, Set("https://kubin.test/s/abc", "tok-1"))
	require.NoError(t, Set("https://kubin.test/s/def/", "tok-2"))

	path, err := Path()
	require.NoError(t, err)

	dirInfo, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temp file left behind")

	tok, ok, err = Get("https://kubin.test/s/abc/")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "tok-1", tok)

	tok, ok, err = Get("https://kubin.test/s/def")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "tok-2", tok)

	require.NoError(t, Delete("https://kubin.test/s/abc"))
	_, ok, err = Get("https://kubin.test/s/abc")
	require.NoError(t, err)
	assert.False(t, ok)

	// Other entries remain intact.
	tok, ok, err = Get("https://kubin.test/s/def")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "tok-2", tok)

	// Deleting a missing entry is a no-op.
	require.NoError(t, Delete("https://kubin.test/s/nonexistent"))
}
