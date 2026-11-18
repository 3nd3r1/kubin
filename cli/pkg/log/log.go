// Package log prints progress and warnings for people reading a terminal.
// Lines go to stderr as "message key=value ...", so stdout stays free for
// the snapshot link. Set LOG_LEVEL=debug to also see debug lines.
package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// Logger prints lines with a fixed set of leading fields.
type Logger struct {
	out    io.Writer
	debug  bool
	fields []any
	mu     *sync.Mutex
}

var std = &Logger{
	out:   os.Stderr,
	debug: strings.EqualFold(os.Getenv("LOG_LEVEL"), "debug"),
	mu:    &sync.Mutex{},
}

// SetOutput redirects the default logger, for tests.
func SetOutput(w io.Writer) { std.out = w }

// With returns a logger that adds fields to every line.
func With(fields ...any) *Logger { return std.With(fields...) }

// WithError returns a logger that adds the error to every line.
func WithError(err error) *Logger { return std.With("error", err) }

func Debug(msg string, fields ...any) { std.Debug(msg, fields...) }
func Info(msg string, fields ...any)  { std.Info(msg, fields...) }
func Warn(msg string, fields ...any)  { std.Warn(msg, fields...) }
func Error(msg string, fields ...any) { std.Error(msg, fields...) }

// Fatal prints an error line and exits with status 1.
func Fatal(msg string, fields ...any) {
	std.Error(msg, fields...)
	os.Exit(1)
}

func (l *Logger) With(fields ...any) *Logger {
	c := *l
	c.fields = append(append([]any{}, l.fields...), fields...)
	return &c
}

func (l *Logger) Debug(msg string, fields ...any) {
	if l.debug {
		l.print("debug: ", msg, fields)
	}
}
func (l *Logger) Info(msg string, fields ...any)  { l.print("", msg, fields) }
func (l *Logger) Warn(msg string, fields ...any)  { l.print("warning: ", msg, fields) }
func (l *Logger) Error(msg string, fields ...any) { l.print("error: ", msg, fields) }

func (l *Logger) print(prefix, msg string, fields []any) {
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(msg)
	all := append(append([]any{}, l.fields...), fields...)
	for i := 0; i < len(all); i += 2 {
		key := fmt.Sprint(all[i])
		val := "<missing>"
		if i+1 < len(all) {
			val = fmt.Sprint(all[i+1])
		}
		if strings.ContainsAny(val, " \t\"=") || val == "" {
			val = fmt.Sprintf("%q", val)
		}
		fmt.Fprintf(&b, " %s=%s", key, val)
	}
	b.WriteByte('\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.out, b.String())
}
