package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/snapshot"
	"github.com/3nd3r1/kubin/cli/pkg/upload"
	"github.com/spf13/cobra"
)

type createOptions struct {
	// Server is the kubin server to upload to.
	Server string
	// Output writes the tarball to this file instead of uploading it.
	Output string
	// TempDir holds the snapshot until it is uploaded. Empty means
	// os.TempDir.
	TempDir string
	// LogTail is the most lines kept per container log.
	LogTail int
}

var createOpts createOptions

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Snapshot the current cluster and print a link to it",
	Long: `Snapshot the cluster in the current kubeconfig context, upload it to a
kubin server and print the link. Only the link goes to stdout.

Use --output to keep the snapshot as a file instead of uploading it.`,
	Example: `  kubin create --server https://kubin.example.com
  KUBIN_SERVER=http://localhost:8080 kubin create
  kubin create --output cluster.tar.gz`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if createOpts.LogTail < 1 {
			return errors.New("--log-tail must be at least 1")
		}
		if createOpts.Output == "" && createOpts.Server == "" {
			return errors.New("no kubin server: pass --server or set KUBIN_SERVER, or use --output to write a file")
		}
		client, err := kube.NewKubeClient()
		if err != nil {
			return err
		}
		return runCreate(cmd.Context(), createOpts, client, cmd.OutOrStdout())
	},
}

func init() {
	createCmd.Flags().StringVar(&createOpts.Server, "server", os.Getenv("KUBIN_SERVER"),
		"kubin server to upload to (env KUBIN_SERVER)")
	createCmd.Flags().StringVarP(&createOpts.Output, "output", "o", "",
		"write the snapshot to this file and don't upload it")
	createCmd.Flags().IntVar(&createOpts.LogTail, "log-tail", collector.DefaultLogTailLines,
		"keep at most this many lines of each container log")
}

// runCreate collects the cluster and either writes the tarball to
// opts.Output or uploads it and prints the link to stdout.
func runCreate(ctx context.Context, opts createOptions, client kube.Client, stdout io.Writer) error {
	logs := collector.DefaultLogOptions()
	logs.TailLines = opts.LogTail
	mgr := snapshot.NewManager(client, version, logs)

	if opts.Output != "" {
		manifest, err := writeSnapshot(ctx, mgr, opts.Output)
		if err != nil {
			return err
		}
		log.Info("Snapshot written", "file", opts.Output, "cluster", manifest.Cluster.Name, "resources", total(manifest))
		return nil
	}

	f, err := os.CreateTemp(opts.TempDir, "kubin-snapshot-*.tar.gz")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	log.Info("Collecting snapshot...")
	manifest, err := mgr.Create(ctx, f)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}

	log.Info("Uploading snapshot...", "server", opts.Server, "cluster", manifest.Cluster.Name, "resources", total(manifest))
	res, err := upload.New(opts.Server, "kubin/"+version).File(ctx, f.Name())
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, res.URL)
	return err
}

// writeSnapshot writes the snapshot next to path and renames it into place,
// so a failure never leaves a partial file or clobbers an existing one.
func writeSnapshot(ctx context.Context, mgr *snapshot.Manager, path string) (_ snapshot.Manifest, err error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".kubin-snapshot-*")
	if err != nil {
		return snapshot.Manifest{}, err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}
	}()
	log.Info("Collecting snapshot...")
	manifest, err := mgr.Create(ctx, f)
	if err != nil {
		return snapshot.Manifest{}, fmt.Errorf("create snapshot: %w", err)
	}
	if err := f.Close(); err != nil {
		return snapshot.Manifest{}, err
	}
	if err := os.Chmod(f.Name(), 0o644); err != nil {
		return snapshot.Manifest{}, err
	}
	return manifest, os.Rename(f.Name(), path)
}

func total(m snapshot.Manifest) int {
	n := 0
	for _, k := range m.Resources {
		n += k.Count
	}
	return n
}
