package cmd

import (
	"github.com/spf13/cobra"
)

// version is set at build time with -ldflags "-X .../cmd.version=v1.2.3".
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "kubin",
	Short:   "Kubin CLI - Create and share Kubernetes cluster snapshots",
	Version: version,
	// Runtime errors aren't usage mistakes, so don't print the help text.
	// Execute prints the error once through cobra.CheckErr.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
}
