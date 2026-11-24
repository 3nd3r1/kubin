package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// commit is set at build time with -ldflags "-X .../cmd.commit=<sha>".
var commit = "none"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the kubin version and commit",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), versionLine())
	},
}

func versionLine() string {
	return fmt.Sprintf("kubin %s (commit %s)", version, commit)
}

// resolveBuild fills in a version or commit that ldflags left at its default
// from the Go build info: the main module version unless it is "(devel)", and
// the first 12 characters of vcs.revision.
func resolveBuild(v, c string, info *debug.BuildInfo) (string, string) {
	if info == nil {
		return v, c
	}
	if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	if c == "none" {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				c = s.Value
				if len(c) > 12 {
					c = c[:12]
				}
			}
		}
	}
	return v, c
}

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		version, commit = resolveBuild(version, commit, info)
	}
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}} (commit " + commit + ")\n")
}
