package cmd

import (
	"bytes"
	"runtime/debug"
	"testing"
)

func TestVersionPrintsLineWithLdflagValues(t *testing.T) {
	oldV, oldC := version, commit
	t.Cleanup(func() { version, commit = oldV, oldC })
	version, commit = "v1.2.3", "0123456789ab"

	var out bytes.Buffer
	versionCmd.SetOut(&out)
	versionCmd.Run(versionCmd, nil)
	if want := "kubin v1.2.3 (commit 0123456789ab)\n"; out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}

func TestVersionFlagMatchesCommand(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetArgs([]string{"--version"})
	t.Cleanup(func() { rootCmd.SetOut(nil); rootCmd.SetArgs(nil) })
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if want := versionLine() + "\n"; out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}

func TestResolveBuild(t *testing.T) {
	info := &debug.BuildInfo{
		Main:     debug.Module{Version: "v0.4.0"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789abcdef0123"}},
	}
	devel := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}
	cases := []struct {
		name         string
		v, c         string
		info         *debug.BuildInfo
		wantV, wantC string
	}{
		{"from build info", "dev", "none", info, "v0.4.0", "0123456789ab"},
		{"ldflags win", "v1.0.0", "abc", info, "v1.0.0", "abc"},
		{"devel and no revision", "dev", "none", devel, "dev", "none"},
		{"no build info", "dev", "none", nil, "dev", "none"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, c := resolveBuild(tc.v, tc.c, tc.info)
			if v != tc.wantV || c != tc.wantC {
				t.Fatalf("got %q %q, want %q %q", v, c, tc.wantV, tc.wantC)
			}
		})
	}
}
