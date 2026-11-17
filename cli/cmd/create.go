package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/3nd3r1/kubin/cli/pkg/kube"
	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/snapshot"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a snapshot of your current Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		client, err := kube.NewKubeClient()
		if err != nil {
			return err
		}

		path := snapshot.DefaultFileName(time.Now())
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return err
		}
		defer func() {
			err = errors.Join(err, f.Close())
			if err != nil {
				_ = os.Remove(path)
			}
		}()

		log.Info("Creating snapshot...")
		manifest, err := snapshot.NewManager(client, version).Create(cmd.Context(), f)
		if err != nil {
			return fmt.Errorf("create snapshot: %w", err)
		}

		total := 0
		for _, k := range manifest.Resources {
			total += k.Count
		}
		log.Info("Snapshot created", "file", path, "cluster", manifest.Cluster.Name, "resources", total)
		return nil
	},
}
