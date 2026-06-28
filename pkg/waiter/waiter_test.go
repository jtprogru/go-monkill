package waiter

import (
	"context"
	"errors"
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
	ch, err := w.Wait(context.Background(), cmd.Process.Pid, 50)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		t.Fatalf("cmd.Wait: %v", err)
	}

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("Wait channel never fired after process exit")
	}
}

func TestWaitMissingPID(t *testing.T) {
	w := &Waiter{Logger: newTestLogger()}
	if _, err := w.Wait(context.Background(), 999999, 50); err == nil {
		t.Fatal("expected error for non-existent PID, got nil")
	}
}

func TestWaitContextCancel(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	w := &Waiter{Logger: newTestLogger()}
	ch, err := w.Wait(ctx, cmd.Process.Pid, 50)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}

	time.Sleep(80 * time.Millisecond)
	cancel()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("Wait did not return after context cancel")
	}
}

func TestProcessStartTimeStableThenGone(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	pid := cmd.Process.Pid

	first, err := processStartTime(pid)
	if err != nil {
		t.Fatalf("processStartTime (first): %v", err)
	}
	second, err := processStartTime(pid)
	if err != nil {
		t.Fatalf("processStartTime (second): %v", err)
	}
	if first != second {
		t.Fatalf("start time not stable for a live process: %d != %d", first, second)
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("kill: %v", err)
	}
	_ = cmd.Wait() // reap so the PID leaves the process table

	// Poll briefly: the kernel may take a moment to drop the entry.
	deadline := time.Now().Add(2 * time.Second)
	for {
		_, err = processStartTime(pid)
		if errors.Is(err, errProcessGone) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected errProcessGone after kill, got %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestWaitNilLogger(t *testing.T) {
	cmd := exec.Command("sleep", "0.1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	defer func() { _ = cmd.Wait() }()

	w := &Waiter{}
	if _, err := w.Wait(context.Background(), cmd.Process.Pid, 50); err != nil {
		t.Fatalf("nil logger should not break Wait: %v", err)
	}
}
