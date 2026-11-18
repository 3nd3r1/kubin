package log

import (
	"bytes"
	"errors"
	"sync"
	"testing"
)

func TestPrintsPlainLines(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{out: &buf, mu: &sync.Mutex{}}

	l.Info("Uploading...")
	l.Warn("skipped logs", "pod", "kube-system/coredns", "reason", "not found")
	l.With("error", errors.New("boom")).Error("failed")
	l.Debug("hidden")

	want := "Uploading...\n" +
		"warning: skipped logs pod=kube-system/coredns reason=\"not found\"\n" +
		"error: failed error=boom\n"
	if got := buf.String(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}
