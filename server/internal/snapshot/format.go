package snapshot

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/3nd3r1/kubin/server/internal/store"
)

// MinFormatVersion and MaxFormatVersion bound the snapshot formats this
// server reads. See docs/snapshot-format.md. v2 and v3 only add files and
// manifest fields to v1, so one reader covers all three.
const (
	MinFormatVersion = 1
	MaxFormatVersion = 3
)

const (
	manifestName  = "manifest.json"
	clusterScoped = "_cluster"
	// tarBlock is what each entry header costs in the archive. Counting it
	// keeps a flood of empty files under the unpacked size limit.
	tarBlock = 512
)

// InvalidError means the upload isn't a valid snapshot. Its message is safe
// to show to the client.
type InvalidError struct {
	msg string
}

func (e *InvalidError) Error() string { return e.msg }

func invalid(format string, args ...any) error {
	return &InvalidError{msg: fmt.Sprintf(format, args...)}
}

// Manifest is the part of manifest.json the server reads.
type Manifest struct {
	FormatVersion int       `json:"format_version"`
	CreatedAt     time.Time `json:"created_at"`
}

// archive reads a snapshot entry by entry and enforces the reading rules
// from docs/snapshot-format.md.
type archive struct {
	tr          *tar.Reader
	maxUnpacked int64
	unpacked    int64
	seen        map[string]bool
}

// openArchive reads the gzip stream and the manifest, which has to be the
// first entry and name a known format version.
func openArchive(r io.Reader, maxUnpacked int64) (*archive, Manifest, []byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, Manifest{}, nil, invalid("not a gzip stream: %v", err)
	}
	a := &archive{tr: tar.NewReader(gz), maxUnpacked: maxUnpacked, seen: map[string]bool{}}

	hdr, err := a.next()
	if err == io.EOF {
		return nil, Manifest{}, nil, invalid("empty archive")
	}
	if err != nil {
		return nil, Manifest{}, nil, err
	}
	if hdr.Name != manifestName {
		return nil, Manifest{}, nil, invalid("first entry is %q, want %s", hdr.Name, manifestName)
	}
	raw, err := a.read(hdr)
	if err != nil {
		return nil, Manifest{}, nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, Manifest{}, nil, invalid("%s: %v", manifestName, err)
	}
	if m.FormatVersion < MinFormatVersion || m.FormatVersion > MaxFormatVersion {
		return nil, Manifest{}, nil, invalid("unsupported format_version %d, this server reads %d to %d",
			m.FormatVersion, MinFormatVersion, MaxFormatVersion)
	}
	return a, m, raw, nil
}

// next returns the next entry header after checking its type, path and the
// size budget. It returns io.EOF at the end of the archive.
func (a *archive) next() (*tar.Header, error) {
	hdr, err := a.tr.Next()
	if err == io.EOF {
		return nil, io.EOF
	}
	if err != nil {
		return nil, invalid("broken archive: %v", err)
	}
	if hdr.Typeflag != tar.TypeReg {
		return nil, invalid("%s: only regular files are allowed", hdr.Name)
	}
	if a.seen[hdr.Name] {
		return nil, invalid("%s: appears twice", hdr.Name)
	}
	a.seen[hdr.Name] = true
	a.unpacked += tarBlock + hdr.Size
	if hdr.Size < 0 || a.unpacked > a.maxUnpacked {
		return nil, invalid("archive unpacks to more than %d bytes", a.maxUnpacked)
	}
	return hdr, nil
}

func (a *archive) read(hdr *tar.Header) ([]byte, error) {
	buf := make([]byte, hdr.Size)
	if _, err := io.ReadFull(a.tr, buf); err != nil {
		return nil, invalid("%s: %v", hdr.Name, err)
	}
	return buf, nil
}

// entry is one file after the manifest. Resource is nil for log files, which
// stay in the blob store and aren't indexed.
type entry struct {
	Resource *store.Resource
}

// nextEntry reads and checks the next entry after the manifest.
func (a *archive) nextEntry() (entry, error) {
	hdr, err := a.next()
	if err != nil {
		return entry{}, err
	}
	if hdr.Name == manifestName {
		return entry{}, invalid("%s: appears twice", manifestName)
	}
	parts := strings.Split(hdr.Name, "/")
	for _, p := range parts {
		if err := checkSegment(p); err != nil {
			return entry{}, invalid("%s: %v", hdr.Name, err)
		}
	}

	switch {
	case parts[0] == "resources" && len(parts) == 6:
		name, ok := strings.CutSuffix(parts[5], ".json")
		if !ok || checkSegment(name) != nil {
			return entry{}, invalid("%s: resource files are <name>.json", hdr.Name)
		}
		ns := parts[4]
		if ns == clusterScoped {
			ns = ""
		}
		raw, err := a.read(hdr)
		if err != nil {
			return entry{}, err
		}
		var obj bytes.Buffer
		if err := json.Compact(&obj, raw); err != nil {
			return entry{}, invalid("%s: %v", hdr.Name, err)
		}
		if obj.Len() == 0 || obj.Bytes()[0] != '{' {
			return entry{}, invalid("%s: not a JSON object", hdr.Name)
		}
		return entry{Resource: &store.Resource{
			Group: parts[1], Version: parts[2], Kind: parts[3],
			Namespace: ns, Name: name, Object: obj.Bytes(),
		}}, nil

	case parts[0] == "logs" && len(parts) == 4:
		if c, ok := strings.CutSuffix(parts[3], ".log"); !ok || checkSegment(c) != nil {
			return entry{}, invalid("%s: log files are <container>.log", hdr.Name)
		}
		// Drain it so the size budget covers what was actually read.
		if _, err := io.Copy(io.Discard, a.tr); err != nil {
			return entry{}, invalid("%s: %v", hdr.Name, err)
		}
		return entry{}, nil
	}
	return entry{}, invalid("%s: not part of the snapshot layout", hdr.Name)
}

func checkSegment(s string) error {
	switch {
	case s == "":
		return errors.New("empty path segment")
	case s == "." || s == "..":
		return fmt.Errorf("unsafe path segment %q", s)
	case strings.ContainsAny(s, "\\\x00"):
		return fmt.Errorf("unsafe path segment %q", s)
	}
	return nil
}
