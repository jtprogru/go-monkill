package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewLoggerVerboseLevel(t *testing.T) {
	t.Cleanup(resetGlobals)
	Verbose = true
	Logfile = ""

	l, closer, err := newLogger()
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	defer func() { _ = closer.Close() }()

	if l.GetLevel() != logrus.DebugLevel {
		t.Fatalf("level = %v, want Debug", l.GetLevel())
	}
}

func TestNewLoggerInfoLevelByDefault(t *testing.T) {
	t.Cleanup(resetGlobals)
	Verbose = false
	Logfile = ""

	l, closer, err := newLogger()
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	defer func() { _ = closer.Close() }()

	if l.GetLevel() != logrus.InfoLevel {
		t.Fatalf("level = %v, want Info", l.GetLevel())
	}
}

func TestNewLoggerWritesJSONToFile(t *testing.T) {
	t.Cleanup(resetGlobals)
	dir := t.TempDir()
	path := filepath.Join(dir, "monkill.log")

	Verbose = false
	Logfile = path

	l, closer, err := newLogger()
	if err != nil {
		t.Fatalf("newLogger: %v", err)
	}
	// Avoid polluting the test runner's stderr.
	l.SetOutput(io.Discard)

	l.Info("hello-from-test")
	if err := closer.Close(); err != nil {
		t.Fatalf("closer.Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read logfile: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `"msg":"hello-from-test"`) {
		t.Fatalf("logfile missing JSON entry; got: %s", got)
	}
	if !strings.Contains(got, `"level":"info"`) {
		t.Fatalf("logfile missing level field; got: %s", got)
	}
}

func TestNewLoggerErrorOnUnopenableFile(t *testing.T) {
	t.Cleanup(resetGlobals)
	// A directory cannot be opened with O_WRONLY|O_APPEND.
	Verbose = false
	Logfile = t.TempDir()

	if _, _, err := newLogger(); err == nil {
		t.Fatal("expected error when logfile path is a directory")
	}
}

func resetGlobals() {
	Verbose = false
	Logfile = ""
}
