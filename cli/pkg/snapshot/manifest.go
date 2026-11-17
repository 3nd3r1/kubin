package snapshot

import "time"

// FormatVersion is the snapshot format this CLI writes. See
// docs/snapshot-format.md.
const FormatVersion = 1

type Manifest struct {
	FormatVersion int         `json:"format_version"`
	CreatedAt     time.Time   `json:"created_at"`
	CLIVersion    string      `json:"cli_version"`
	Cluster       Cluster     `json:"cluster"`
	Resources     []KindCount `json:"resources"`
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
