package cmd

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/sirupsen/logrus"
)

// fakeWaiter returns a channel per PID. Behavior is configured per call:
//   - err != nil → first Wait returns it
//   - fire == true → channel is closed immediately (PID is "gone")
//   - block == true → channel only closes when ctx is canceled
//   - perPID overrides individual PIDs (key = pid, value = "fire" | "block")
type fakeWaiter struct {
	err    error
	fire   bool
	block  bool
	perPID map[int]string

	mu        sync.Mutex
	calledFor []int
}

func (f *fakeWaiter) Wait(ctx context.Context, pid int, _ int64) (<-chan struct{}, error) {
	f.mu.Lock()
	f.calledFor = append(f.calledFor, pid)
	f.mu.Unlock()

	if f.err != nil {
		return nil, f.err
	}
	mode := ""
	if v, ok := f.perPID[pid]; ok {
		mode = v
	} else if f.fire {
		mode = "fire"
	} else if f.block {
		mode = "block"
	}

	ch := make(chan struct{})
	switch mode {
	case "fire":
		close(ch)
	case "block":
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

type watcherOpts struct {
	command string
	timeout int64
	maxWait time.Duration
	waitFor string
}

func withCommand(s string) func(*watcherOpts)        { return func(o *watcherOpts) { o.command = s } }
func withMaxWait(d time.Duration) func(*watcherOpts) { return func(o *watcherOpts) { o.maxWait = d } }
func withWaitFor(s string) func(*watcherOpts)        { return func(o *watcherOpts) { o.waitFor = s } }

func runWatcher(t *testing.T, ctx context.Context, pids []int, w Waiter, e Executor, opts ...func(*watcherOpts)) (int, error) {
	t.Helper()
	o := watcherOpts{command: "echo hi", timeout: 10, maxWait: 0, waitFor: waitForAll}
	for _, fn := range opts {
		fn(&o)
	}
	return watcher(ctx, pids, o.command, o.timeout, o.maxWait, o.waitFor, w, e, newTestLogger())
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
			if err := checkPid(tc.pid, l); (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestWatcherSinglePidHappyPath(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{}
	code, err := runWatcher(t, context.Background(), []int{1234}, w, e)
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !e.called {
		t.Fatal("executor not called")
	}
}

func TestWatcherPropagatesExitCode(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{res: executor.Result{ExitCode: 42, Err: errors.New("boom")}}
	code, _ := runWatcher(t, context.Background(), []int{1234}, w, e)
	if code != 42 {
		t.Fatalf("code = %d, want 42", code)
	}
}

func TestWatcherInvalidPID(t *testing.T) {
	code, err := runWatcher(t, context.Background(), []int{0}, &fakeWaiter{}, &fakeExecutor{})
	if err == nil || code != 1 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestWatcherNoPIDs(t *testing.T) {
	code, err := runWatcher(t, context.Background(), nil, &fakeWaiter{}, &fakeExecutor{})
	if err == nil || code != 1 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestWatcherEmptyCommand(t *testing.T) {
	e := &fakeExecutor{}
	code, err := runWatcher(t, context.Background(), []int{1234}, &fakeWaiter{fire: true}, e, withCommand(""))
	if err == nil || code != 1 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if e.called {
		t.Fatal("executor must not run")
	}
}

func TestWatcherInvalidWaitFor(t *testing.T) {
	code, err := runWatcher(t, context.Background(), []int{1234}, &fakeWaiter{}, &fakeExecutor{}, withWaitFor("bogus"))
	if err == nil || code != 1 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestWatcherWaiterError(t *testing.T) {
	w := &fakeWaiter{err: errors.New("waiter exploded")}
	e := &fakeExecutor{}
	code, err := runWatcher(t, context.Background(), []int{1234}, w, e)
	if err == nil || code != 1 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if e.called {
		t.Fatal("executor must not run")
	}
}

func TestWatcherSignalCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &fakeWaiter{block: true}
	e := &fakeExecutor{}

	done := make(chan struct{})
	var (
		code int
		err  error
	)
	go func() {
		code, err = runWatcher(t, ctx, []int{1234}, w, e)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watcher did not return")
	}
	if code != exitCodeInterrupted {
		t.Fatalf("code=%d, want %d", code, exitCodeInterrupted)
	}
	if err == nil || e.called {
		t.Fatalf("err=%v called=%v", err, e.called)
	}
}

func TestWatcherMaxWaitTimeout(t *testing.T) {
	w := &fakeWaiter{block: true}
	e := &fakeExecutor{}
	code, err := runWatcher(t, context.Background(), []int{1234}, w, e, withMaxWait(50*time.Millisecond))
	if code != exitCodeTimeout {
		t.Fatalf("code = %d, want %d", code, exitCodeTimeout)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if e.called {
		t.Fatal("executor must not run on timeout")
	}
}

func TestWatcherMultiPidAllExited(t *testing.T) {
	w := &fakeWaiter{fire: true}
	e := &fakeExecutor{res: executor.Result{ExitCode: 0}}
	code, err := runWatcher(t, context.Background(), []int{1111, 2222, 3333}, w, e)
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if !e.called {
		t.Fatal("executor not called")
	}
	w.mu.Lock()
	if len(w.calledFor) != 3 {
		t.Fatalf("Wait calls = %d, want 3", len(w.calledFor))
	}
	w.mu.Unlock()
}

func TestWatcherMultiPidAllOneBlocking(t *testing.T) {
	// All-mode + one PID never exits → must block until ctx cancellation.
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	w := &fakeWaiter{
		perPID: map[int]string{
			1111: "fire",
			2222: "block",
		},
	}
	e := &fakeExecutor{}
	code, err := runWatcher(t, ctx, []int{1111, 2222}, w, e)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if code != exitCodeTimeout && code != exitCodeInterrupted {
		t.Fatalf("unexpected code %d", code)
	}
	if e.called {
		t.Fatal("executor must not run when not all PIDs exited")
	}
}

func TestWatcherMultiPidAnyMode(t *testing.T) {
	// Any-mode: one PID fires immediately, the other blocks. Should return promptly.
	w := &fakeWaiter{
		perPID: map[int]string{
			1111: "fire",
			2222: "block",
		},
	}
	e := &fakeExecutor{}
	start := time.Now()
	code, err := runWatcher(t, context.Background(), []int{1111, 2222}, w, e, withWaitFor(waitForAny))
	elapsed := time.Since(start)
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("any-mode took %s, expected to be quick", elapsed)
	}
	if !e.called {
		t.Fatal("executor must run when any PID exits")
	}
}
