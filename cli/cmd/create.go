package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/collector"
	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/redact"
	"github.com/3nd3r1/kubin/cli/pkg/snapshot"
	"github.com/3nd3r1/kubin/cli/pkg/tokens"
	"github.com/3nd3r1/kubin/cli/pkg/upload"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
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
	// KeepEnvValues keeps literal env var values in pod specs.
	KeepEnvValues bool
	// Context is the kubeconfig context to use. Empty means the current
	// one.
	Context string
	// Namespaces limits collection to these namespaces; Exclude leaves
	// namespaces out.
	Namespaces []string
	Exclude    []string
	// NoLogs leaves container logs out.
	NoLogs bool
	// MaxSize is the largest snapshot uploaded, like "64Mi". Empty means
	// no limit. --output files are not limited.
	MaxSize string
	// Expire is the snapshot lifetime sent to the server, like "24h".
	// Empty lets the server apply its default. Ignored with --output.
	Expire string
}

// defaultMaxSize matches the server's default SERVER_MAX_UPLOAD_BYTES.
const defaultMaxSize = "64Mi"

var createOpts createOptions

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Snapshot the current cluster and print a link to it",
	Long: `Snapshot the cluster in the current kubeconfig context, upload it to a
kubin server and print the link. Only the link goes to stdout.

Use --output to keep the snapshot as a file instead of uploading it.`,
	Example: `  kubin create --server https://kubin.example.com
  kubin create --server https://kubin.example.com --expire 24h
  KUBIN_SERVER=http://localhost:8080 kubin create
  kubin create --namespace shop --namespace payments
  kubin create --exclude-namespace kube-system --no-logs
  kubin create --context prod --output cluster.tar.gz`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if createOpts.LogTail < 1 {
			return errors.New("--log-tail must be at least 1")
		}
		if createOpts.Output == "" && createOpts.Server == "" {
			return errors.New("no kubin server: pass --server or set KUBIN_SERVER, or use --output to write a file")
		}
		if _, err := parseSize(createOpts.MaxSize); err != nil {
			return err
		}
		if createOpts.Output == "" {
			if err := parseExpire(createOpts.Expire); err != nil {
				return err
			}
		}
		client, err := kube.NewKubeClient(createOpts.Context)
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
	createCmd.Flags().BoolVar(&createOpts.KeepEnvValues, "keep-env-values", false,
		"keep env var values in pod specs; by default they are dropped because they can hold credentials")
	createCmd.Flags().StringVar(&createOpts.Context, "context", "",
		"kubeconfig context to use (default the current context)")
	createCmd.Flags().StringSliceVarP(&createOpts.Namespaces, "namespace", "n", nil,
		"collect only this namespace; repeat or comma-separate for more. Cluster-scoped kinds like nodes are still collected")
	createCmd.Flags().StringSliceVar(&createOpts.Exclude, "exclude-namespace", nil,
		"leave this namespace out; repeat or comma-separate for more")
	createCmd.Flags().BoolVar(&createOpts.NoLogs, "no-logs", false,
		"leave container logs out of the snapshot")
	createCmd.Flags().StringVar(&createOpts.MaxSize, "max-size", defaultMaxSize,
		`don't upload a snapshot bigger than this, like 64Mi or 200M; "0" means no limit`)
	createCmd.Flags().StringVar(&createOpts.Expire, "expire", "",
		"snapshot lifetime, like 24h or 30m (default is set by the server)")
}

// parseExpire checks a --expire duration. Empty means use the server default.
func parseExpire(s string) error {
	if s == "" {
		return nil
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return fmt.Errorf("--expire %q: want a positive duration like 24h or 30m", s)
	}
	return nil
}

// parseSize reads a --max-size value. 0 means no limit.
func parseSize(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	q, err := resource.ParseQuantity(s)
	if err != nil || q.Sign() < 0 {
		return 0, fmt.Errorf("--max-size %q: want a size like 64Mi or 200M", s)
	}
	return q.Value(), nil
}

// runCreate collects the cluster and either writes the tarball to
// opts.Output or uploads it and prints the link to stdout.
func runCreate(ctx context.Context, opts createOptions, client kube.Client, stdout io.Writer) error {
	logs := collector.DefaultLogOptions()
	logs.TailLines = opts.LogTail
	maxSize, err := parseSize(opts.MaxSize)
	if err != nil {
		return err
	}
	mgr := snapshot.NewManager(client, version, snapshot.Options{
		Scope:    collector.Scope{Namespaces: opts.Namespaces, Exclude: opts.Exclude},
		Logs:     logs,
		SkipLogs: opts.NoLogs,
		Redact:   redact.Options{KeepEnvValues: opts.KeepEnvValues},
	})

	if opts.Output != "" {
		manifest, err := writeSnapshot(ctx, mgr, opts.Output)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("Wrote %d resources from %s to %s", total(manifest), manifest.Cluster.Name, opts.Output))
		return nil
	}

	if err := parseExpire(opts.Expire); err != nil {
		return err
	}

	f, err := os.CreateTemp(opts.TempDir, "kubin-snapshot-*.tar.gz")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	log.Info("Collecting the cluster...")
	manifest, err := mgr.Create(ctx, f)
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	if maxSize > 0 {
		info, err := os.Stat(f.Name())
		if err != nil {
			return err
		}
		if info.Size() > maxSize {
			return fmt.Errorf("snapshot is %s, over --max-size %s: narrow it with --namespace, --exclude-namespace or --no-logs, or raise --max-size",
				humanSize(info.Size()), opts.MaxSize)
		}
	}

	log.Info(fmt.Sprintf("Uploading %d resources from %s to %s...", total(manifest), manifest.Cluster.Name, opts.Server))
	res, err := upload.New(opts.Server, "kubin/"+version).File(ctx, f.Name(), opts.Expire)
	if err != nil {
		return err
	}
	if res.DeleteToken != "" {
		if err := tokens.Set(res.URL, res.DeleteToken); err != nil {
			log.Warn("Could not save delete token", "error", err)
		}
	}
	if !res.ExpiresAt.IsZero() {
		log.Info(fmt.Sprintf("Expires %s. Delete it with: kubin delete %s",
			res.ExpiresAt.UTC().Format("2006-01-02 15:04 UTC"), res.URL))
	} else if res.DeleteToken != "" {
		log.Info(fmt.Sprintf("Delete it with: kubin delete %s", res.URL))
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
	log.Info("Collecting the cluster...")
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

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func total(m snapshot.Manifest) int {
	n := 0
	for _, k := range m.Resources {
		n += k.Count
	}
	return n
}
