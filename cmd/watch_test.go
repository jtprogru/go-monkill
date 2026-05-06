package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/sirupsen/logrus"
)

type fakeWaiter struct {
	err  error
	fire bool
	// block: when true, channel never fires unless ctx cancels
	block bool
}

func (f *fakeWaiter) Wait(ctx context.Context, pid int, timeout int64) (<-chan struct{}, error) {
	if f.err != nil {
		return nil, f.err
	}
	ch := make(chan struct{})
	switch {
	case f.fire:
		close(ch)
	case f.block:
		go func() {
			<-ctx.Done()
			close(ch)
		}()
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
	code, err := watcher(context.Background(), 1234, "echo hi", 10, 0, w, e, newTestLogger())
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
	code, err := watcher(context.Background(), 1234, "echo hi", 10, 0, w, e, newTestLogger())
	if code != 42 {
		t.Fatalf("exit code = %d, want 42", code)
	}
	if err == nil {
		t.Fatal("expected err to be propagated")
	}
}

func TestWatcherInvalidPID(t *testing.T) {
	code, err := watcher(context.Background(), 0, "echo hi", 10, 0, &fakeWaiter{}, &fakeExecutor{}, newTestLogger())
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
	code, err := watcher(context.Background(), 1234, "", 10, 0, w, e, newTestLogger())
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
	code, err := watcher(context.Background(), 1234, "echo hi", 10, 0, w, e, newTestLogger())
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

func TestWatcherCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &fakeWaiter{block: true}
	e := &fakeExecutor{}

	done := make(chan struct{})
	var (
		code int
		err  error
	)
	go func() {
		code, err = watcher(ctx, 1234, "echo hi", 10, 0, w, e, newTestLogger())
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not return after context cancel")
	}

	if code != exitCodeInterrupted {
		t.Fatalf("exit code = %d, want %d", code, exitCodeInterrupted)
	}
	if err == nil {
		t.Fatal("expected ctx.Err to be returned")
	}
	if e.called {
		t.Fatal("executor must not run when watcher is canceled")
	}
}

func TestWatcherMaxWaitTimeout(t *testing.T) {
	w := &fakeWaiter{block: true}
	e := &fakeExecutor{}
	code, err := watcher(context.Background(), 1234, "echo hi", 10, 50*time.Millisecond, w, e, newTestLogger())
	if code != exitCodeTimeout {
		t.Fatalf("exit code = %d, want %d", code, exitCodeTimeout)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if e.called {
		t.Fatal("executor must not run when max-wait elapses")
	}
}

func TestWatcherMaxWaitNotReached(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{res: executor.Result{ExitCode: 0}}
	code, err := watcher(context.Background(), 1234, "echo hi", 10, time.Hour, w, e, newTestLogger())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !e.called {
		t.Fatal("executor should run when max-wait did not expire")
	}
}
