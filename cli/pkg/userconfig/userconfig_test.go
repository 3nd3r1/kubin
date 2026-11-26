package userconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMissingReturnsZero(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	cfg, err := Load()
	require.NoError(t, err)
	assert.False(t, cfg.PublicUploadConfirmed)
}

func TestSaveAndLoadPermissions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	require.NoError(t, Save(Config{PublicUploadConfirmed: true}))

	path, err := Path()
	require.NoError(t, err)

	dirInfo, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.PublicUploadConfirmed)
}
