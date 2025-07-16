package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "kubin",
    Short: "Kubin CLI - Create and share Kubernetes cluster snapshots",
}

func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}

func init() {
    rootCmd.AddCommand(createCmd)
}
