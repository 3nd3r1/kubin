// Package tokens stores snapshot delete tokens under the user's config directory.
package tokens

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Path returns the path to the token file: <os.UserConfigDir()>/kubin/snapshots.json.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(dir, "kubin", "snapshots.json"), nil
}

// Set stores token under snapshotURL.
func Set(snapshotURL, token string) error {
	path, err := Path()
	if err != nil {
		return err
	}
	entries, err := load(path)
	if err != nil {
		return err
	}
	entries[normalizeURL(snapshotURL)] = token
	return save(path, entries)
}

// Get returns the stored token for snapshotURL, or ("", false, nil) if none exists.
func Get(snapshotURL string) (string, bool, error) {
	path, err := Path()
	if err != nil {
		return "", false, err
	}
	entries, err := load(path)
	if err != nil {
		return "", false, err
	}
	tok := entries[normalizeURL(snapshotURL)]
	return tok, tok != "", nil
}

// Delete removes any stored token for snapshotURL.
func Delete(snapshotURL string) error {
	path, err := Path()
	if err != nil {
		return err
	}
	entries, err := load(path)
	if err != nil {
		return err
	}
	key := normalizeURL(snapshotURL)
	if _, ok := entries[key]; !ok {
		return nil
	}
	delete(entries, key)
	return save(path, entries)
}

func normalizeURL(u string) string {
	return strings.TrimRight(u, "/")
}

func load(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return map[string]string{}, nil
	}
	var entries map[string]string
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse token file %s: %w", path, err)
	}
	if entries == nil {
		entries = map[string]string{}
	}
	return entries, nil
}

func save(path string, entries map[string]string) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	f, err := os.CreateTemp(dir, ".snapshots-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp token file: %w", err)
	}
	tmp := f.Name()
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()

	if err := os.Chmod(tmp, 0o600); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
