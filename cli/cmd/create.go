package cmd

import (
	"github.com/3nd3r1/kubin/cli/pkg/log"
	"github.com/3nd3r1/kubin/cli/pkg/snapshot"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a snapshot of your current Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := snapshot.NewManager()
		if err != nil {
			return err
		}

        log.Info("Creating snapshot...")
		if err := manager.CreateSnapshot(cmd.Context()); err != nil {
            log.WithError(err).Error("Failed to create snapshot")
			return err
		}

        log.Info("Snapshot created")
		return nil
	},
}
