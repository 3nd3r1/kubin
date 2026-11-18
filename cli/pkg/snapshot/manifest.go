package snapshot

import (
	"path"
	"time"
)

// FormatVersion is the snapshot format this CLI writes. See
// docs/snapshot-format.md.
const FormatVersion = 3

type Manifest struct {
	FormatVersion int         `json:"format_version"`
	CreatedAt     time.Time   `json:"created_at"`
	CLIVersion    string      `json:"cli_version"`
	Cluster       Cluster     `json:"cluster"`
	Resources     []KindCount `json:"resources"`
	Logs          Logs        `json:"logs"`
	// Redactions names the redact rules that ran on every object.
	Redactions []string `json:"redactions"`
}

type Cluster struct {
	Context       string `json:"context"`
	Name          string `json:"name"`
	ServerVersion string `json:"server_version"`
}

type KindCount struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Count   int    `json:"count"`
}

// Logs says which log files the snapshot holds and the limits they were cut
// to.
type Logs struct {
	TailLines int       `json:"tail_lines"`
	MaxBytes  int       `json:"max_bytes"`
	Files     []LogFile `json:"files"`
}

type LogFile struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
	Previous  bool   `json:"previous"`
	Truncated bool   `json:"truncated"`
}

// path is where the log lives in the archive.
func (f LogFile) path() string {
	name := f.Container + ".log"
	if f.Previous {
		name = f.Container + ".previous.log"
	}
	return path.Join("logs", f.Namespace, f.Pod, name)
}
