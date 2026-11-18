// Package snapshottest builds snapshot tarballs for tests.
package snapshottest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

// File is one tar entry. Type defaults to a regular file.
type File struct {
	Name string
	Body string
	Type byte
}

// Tarball gzips files into a tar in the order given.
func Tarball(t testing.TB, files ...File) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, f := range files {
		typ := f.Type
		if typ == 0 {
			typ = tar.TypeReg
		}
		hdr := &tar.Header{Name: f.Name, Typeflag: typ, Mode: 0o644, Format: tar.FormatPAX}
		if typ == tar.TypeReg {
			hdr.Size = int64(len(f.Body))
		}
		if typ == tar.TypeSymlink {
			hdr.Linkname = "manifest.json"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if typ == tar.TypeReg {
			if _, err := tw.Write([]byte(f.Body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Manifest is a v2 manifest.json matching FixtureFiles.
const Manifest = `{
  "format_version": 2,
  "created_at": "2026-11-17T10:00:00Z",
  "cli_version": "v0.1.0",
  "cluster": {"context": "kind-kubin", "name": "kind-kubin", "server_version": "v1.33.1"},
  "resources": [
    {"group": "core", "version": "v1", "kind": "Namespace", "count": 2},
    {"group": "core", "version": "v1", "kind": "Pod", "count": 2}
  ],
  "logs": {
    "tail_lines": 1000,
    "max_bytes": 1048576,
    "files": [
      {"namespace": "default", "pod": "web-0", "container": "nginx", "previous": false, "truncated": false},
      {"namespace": "default", "pod": "web-0", "container": "nginx", "previous": true, "truncated": true}
    ]
  }
}`

// FixtureFiles is a small snapshot as the CLI writes it: two namespaces
// (cluster-scoped), two pods, and the current and previous log of one
// container.
func FixtureFiles() []File {
	return []File{
		{Name: "manifest.json", Body: Manifest},
		{Name: "logs/default/web-0/nginx.log", Body: "started\n"},
		{Name: "logs/default/web-0/nginx.previous.log", Body: "panic: boom\n"},
		{Name: "resources/core/v1/Namespace/_cluster/default.json", Body: `{
  "apiVersion": "v1", "kind": "Namespace", "metadata": {"name": "default"}
}`},
		{Name: "resources/core/v1/Namespace/_cluster/kube-system.json", Body: `{
  "apiVersion": "v1", "kind": "Namespace", "metadata": {"name": "kube-system"}
}`},
		{Name: "resources/core/v1/Pod/default/web-0.json", Body: `{
  "apiVersion": "v1", "kind": "Pod",
  "metadata": {"name": "web-0", "namespace": "default"},
  "spec": {"containers": [{"name": "nginx", "image": "nginx:1.27"}]},
  "status": {"phase": "Running"}
}`},
		{Name: "resources/core/v1/Pod/kube-system/coredns-1.json", Body: `{
  "apiVersion": "v1", "kind": "Pod",
  "metadata": {"name": "coredns-1", "namespace": "kube-system"},
  "status": {"phase": "Running"}
}`},
	}
}

// Fixture is FixtureFiles as a tarball.
func Fixture(t testing.TB) []byte {
	return Tarball(t, FixtureFiles()...)
}
