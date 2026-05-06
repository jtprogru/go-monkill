package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/sirupsen/logrus"
)

type fakeWaiter struct {
	err  error
	fire bool
}

func (f *fakeWaiter) Wait(pid int, timeout int64) (<-chan struct{}, error) {
	if f.err != nil {
		return nil, f.err
	}
	ch := make(chan struct{})
	if f.fire {
		close(ch)
	}
	return ch, nil
}

type fakeExecutor struct {
	called bool
	cmd    string
	res    executor.Result
}

func (f *fakeExecutor) Exec(command string) executor.Result {
	f.called = true
	f.cmd = command
	return f.res
}

func newTestLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(&strings.Builder{})
	l.SetLevel(logrus.DebugLevel)
	return l
}

func TestCheckPid(t *testing.T) {
	l := newTestLogger()
	cases := []struct {
		name    string
		pid     int
		wantErr bool
	}{
		{"negative", -1, true},
		{"zero", 0, true},
		{"init", 1, true},
		{"valid", 1234, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPid(tc.pid, l)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestWatcherHappyPath(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{res: executor.Result{ExitCode: 0}}
	code, err := watcher(1234, "echo hi", 10, w, e, newTestLogger())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !e.called {
		t.Fatal("executor was not called")
	}
	if e.cmd != "echo hi" {
		t.Fatalf("executor got %q", e.cmd)
	}
}

func TestWatcherPropagatesExitCode(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{res: executor.Result{ExitCode: 42, Err: errors.New("boom")}}
	code, err := watcher(1234, "echo hi", 10, w, e, newTestLogger())
	if code != 42 {
		t.Fatalf("exit code = %d, want 42", code)
	}
	if err == nil {
		t.Fatal("expected err to be propagated")
	}
}

func TestWatcherInvalidPID(t *testing.T) {
	code, err := watcher(0, "echo hi", 10, &fakeWaiter{}, &fakeExecutor{}, newTestLogger())
	if err == nil {
		t.Fatal("expected error for invalid PID")
	}
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestWatcherEmptyCommand(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{}
	code, err := watcher(1234, "", 10, w, e, newTestLogger())
	if err == nil {
		t.Fatal("expected error for empty command")
	}
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if e.called {
		t.Fatal("executor must not be called when command is empty")
	}
}

func TestWatcherWaiterError(t *testing.T) {
	w := &fakeWaiter{err: errors.New("waiter exploded")}
	e := &fakeExecutor{}
	code, err := watcher(1234, "echo hi", 10, w, e, newTestLogger())
	if err == nil {
		t.Fatal("expected waiter error to bubble up")
	}
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if e.called {
		t.Fatal("executor must not run when waiter fails")
	}
}
