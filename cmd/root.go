// Package cmd contains all commands

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Version, Commit and BuildDate are populated at build time via -ldflags by goreleaser.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Verbose enables debug-level logging.
var Verbose bool

// Logfile is the optional path to an external log file (JSON-formatted).
var Logfile string

var rootCmd = &cobra.Command{
	Use:   "go-monkill",
	Short: "go-monkill watches a process by PID and runs a command after it exits",
	Long:  `Monitor when process with PID will finish or be killed and run what you need.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose (debug-level) output")
	rootCmd.PersistentFlags().StringVar(&Logfile, "logfile", "", "path to a log file (JSON format); empty disables file logging")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// newLogger constructs a configured *logrus.Logger respecting --verbose and --logfile flags.
// Returned closer must be called by the caller (typically via defer) to release the log file.
func newLogger() (*logrus.Logger, io.Closer, error) {
	l := logrus.New()
	if Verbose {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}
	l.SetOutput(os.Stderr)

	if Logfile == "" {
		return l, io.NopCloser(nil), nil
	}

	f, err := os.OpenFile(Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open logfile %q: %w", Logfile, err)
	}

	// Text to stderr, JSON to file.
	l.AddHook(&fileHook{w: f, formatter: &logrus.JSONFormatter{}})
	return l, f, nil
}

// fileHook duplicates log entries to a file using a dedicated formatter.
type fileHook struct {
	w         io.Writer
	formatter logrus.Formatter
}

func (h *fileHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *fileHook) Fire(entry *logrus.Entry) error {
	b, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.w.Write(b)
	return err
}
