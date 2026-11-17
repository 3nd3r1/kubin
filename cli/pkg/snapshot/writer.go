package snapshot

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	coreGroup     = "core"
	clusterScoped = "_cluster"
	manifestName  = "manifest.json"
)

// Writer stages a snapshot in a temporary directory and then writes it as a
// gzipped tar in the v1 layout. Call Cleanup when done.
type Writer struct {
	dir    string
	counts map[KindCount]int
}

func NewWriter() (*Writer, error) {
	dir, err := os.MkdirTemp("", "kubin-snapshot-*")
	if err != nil {
		return nil, err
	}
	return &Writer{dir: dir, counts: map[KindCount]int{}}, nil
}

// AddResource stages one object. It strips metadata.managedFields and fails
// on names that aren't safe path segments or on the same object twice.
func (w *Writer) AddResource(r collector.Resource) error {
	group := r.Group
	if group == "" {
		group = coreGroup
	}
	ns := r.Namespace
	if ns == "" {
		ns = clusterScoped
	}
	segments := []string{group, r.Version, r.Kind, ns, r.Name}
	for _, s := range segments {
		if err := checkSegment(s); err != nil {
			return fmt.Errorf("%s %s/%s: %w", r.Kind, r.Namespace, r.Name, err)
		}
	}

	rel := path.Join("resources", path.Join(segments...)+".json")
	full := filepath.Join(w.dir, filepath.FromSlash(rel))
	if _, err := os.Stat(full); err == nil {
		return fmt.Errorf("%s: collected twice", rel)
	}

	unstructured.RemoveNestedField(r.Object, "metadata", "managedFields")
	data, err := json.MarshalIndent(r.Object, "", "  ")
	if err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return err
	}

	w.counts[KindCount{Group: group, Version: r.Version, Kind: r.Kind}]++
	return nil
}

func checkSegment(s string) error {
	switch {
	case s == "":
		return errors.New("empty path segment")
	case s == "." || s == "..":
		return fmt.Errorf("unsafe path segment %q", s)
	case strings.ContainsAny(s, "/\\\x00"):
		return fmt.Errorf("unsafe path segment %q", s)
	}
	return nil
}

// Counts returns the staged objects per kind, sorted by group, version, kind.
func (w *Writer) Counts() []KindCount {
	out := make([]KindCount, 0, len(w.counts))
	for k, n := range w.counts {
		k.Count = n
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Group != b.Group {
			return a.Group < b.Group
		}
		if a.Version != b.Version {
			return a.Version < b.Version
		}
		return a.Kind < b.Kind
	})
	return out
}

// Finish writes the manifest and every staged file to out as a gzipped tar.
// The manifest is the first entry and the rest follow sorted by path.
func (w *Writer) Finish(m Manifest, out io.Writer) (err error) {
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	defer func() {
		// Closing writes the tar footer and gzip trailer. A failure here
		// means a truncated archive, so it must reach the caller.
		err = errors.Join(err, tw.Close(), gz.Close())
	}()

	manifest, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry(tw, manifestName, manifest, m.CreatedAt); err != nil {
		return err
	}

	var files []string
	err = filepath.WalkDir(w.dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, p := range files {
		rel, err := filepath.Rel(w.dir, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := writeEntry(tw, filepath.ToSlash(rel), data, m.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func writeEntry(tw *tar.Writer, name string, data []byte, mod time.Time) error {
	hdr := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Mode:     0o644,
		Size:     int64(len(data)),
		ModTime:  mod,
		Format:   tar.FormatPAX,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// Cleanup removes the staging directory.
func (w *Writer) Cleanup() error {
	return os.RemoveAll(w.dir)
}
