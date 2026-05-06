package waiter

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func newTestLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(&strings.Builder{})
	l.SetLevel(logrus.DebugLevel)
	return l
}

func TestWaitDetectsProcessExit(t *testing.T) {
	cmd := exec.Command("sleep", "0.3")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}

	w := &Waiter{Logger: newTestLogger()}
	ch, err := w.Wait(cmd.Process.Pid, 50)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}

	// Reap the child so the process actually disappears from the table.
	if err := cmd.Wait(); err != nil {
		t.Fatalf("cmd.Wait: %v", err)
	}

	select {
	case <-ch:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("Wait channel never fired after process exit")
	}
}

func TestWaitMissingPID(t *testing.T) {
	w := &Waiter{Logger: newTestLogger()}
	// PID 999999 is almost certainly not running on a test runner.
	if _, err := w.Wait(999999, 50); err == nil {
		t.Fatal("expected error for non-existent PID, got nil")
	}
}

func TestWaitNilLogger(t *testing.T) {
	cmd := exec.Command("sleep", "0.1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	defer func() { _ = cmd.Wait() }()

	w := &Waiter{}
	if _, err := w.Wait(cmd.Process.Pid, 50); err != nil {
		t.Fatalf("nil logger should not break Wait: %v", err)
	}
}
