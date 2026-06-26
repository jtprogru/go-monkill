package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/sirupsen/logrus"
)

func swapChildRunner(t *testing.T, fn childRunner) {
	t.Helper()
	old := runForChild
	runForChild = fn
	t.Cleanup(func() { runForChild = old })
}

func TestRunnerNoArgs(t *testing.T) {
	code, err := runner(context.Background(), nil, "", "", "", &fakeExecutor{}, newTestLogger())
	if err == nil || code != 1 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestRunnerSuccessFiresOnSuccessAndOnExit(t *testing.T) {
	swapChildRunner(t, func(_ context.Context, _ []string, _ *logrus.Logger) (int, error) {
		return 0, nil
	})
	e := &fakeExecutor{}
	calls := []string{}
	e.res = executor.Result{ExitCode: 0}
	hookExec := &recordingExecutor{}

	code, err := runner(context.Background(), []string{"true"}, "post-success", "should-not-run", "always", hookExec, newTestLogger())
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	calls = hookExec.calls
	if len(calls) != 2 {
		t.Fatalf("expected 2 hook calls (on-success + on-exit), got %d: %v", len(calls), calls)
	}
	if calls[0] != "post-success" || calls[1] != "always" {
		t.Fatalf("hook order wrong: %v", calls)
	}
}

func TestRunnerFailureFiresOnFailureAndOnExit(t *testing.T) {
	swapChildRunner(t, func(_ context.Context, _ []string, _ *logrus.Logger) (int, error) {
		return 7, errors.New("boom")
	})
	hookExec := &recordingExecutor{}

	code, err := runner(context.Background(), []string{"false"}, "should-not-run", "alert", "cleanup", hookExec, newTestLogger())
	if code != 7 {
		t.Fatalf("code=%d, want 7", code)
	}
	if err == nil {
		t.Fatal("expected child err to bubble up")
	}
	if got := hookExec.calls; len(got) != 2 || got[0] != "alert" || got[1] != "cleanup" {
		t.Fatalf("hook calls = %v, want [alert cleanup]", got)
	}
}

func TestRunnerFailureWithoutOnFailureStillFiresOnExit(t *testing.T) {
	swapChildRunner(t, func(_ context.Context, _ []string, _ *logrus.Logger) (int, error) {
		return 2, errors.New("nope")
	})
	hookExec := &recordingExecutor{}

	_, _ = runner(context.Background(), []string{"false"}, "", "", "cleanup", hookExec, newTestLogger())
	if got := hookExec.calls; len(got) != 1 || got[0] != "cleanup" {
		t.Fatalf("hook calls = %v, want [cleanup]", got)
	}
}

func TestRunnerHookFailureDoesNotMaskChildExit(t *testing.T) {
	swapChildRunner(t, func(_ context.Context, _ []string, _ *logrus.Logger) (int, error) {
		return 0, nil
	})
	hookExec := &recordingExecutor{
		failNext: 1,
		res:      executor.Result{ExitCode: 99, Err: errors.New("hook broke")},
	}

	code, err := runner(context.Background(), []string{"true"}, "broken-hook", "", "", hookExec, newTestLogger())
	if code != 0 || err != nil {
		t.Fatalf("hook failure must not change child exit code/err: code=%d err=%v", code, err)
	}
}

func TestRunnerNoHooks(t *testing.T) {
	swapChildRunner(t, func(_ context.Context, _ []string, _ *logrus.Logger) (int, error) {
		return 0, nil
	})
	hookExec := &recordingExecutor{}
	code, err := runner(context.Background(), []string{"true"}, "", "", "", hookExec, newTestLogger())
	if code != 0 || err != nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
	if len(hookExec.calls) != 0 {
		t.Fatalf("no hooks expected, got %v", hookExec.calls)
	}
}

func TestDefaultChildRunnerSuccess(t *testing.T) {
	code, err := defaultChildRunner(context.Background(), []string{"true"}, newTestLogger())
	if code != 0 || err != nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestDefaultChildRunnerFailure(t *testing.T) {
	code, err := defaultChildRunner(context.Background(), []string{"false"}, newTestLogger())
	if code != 1 || err == nil {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestDefaultChildRunnerMissingBinary(t *testing.T) {
	code, err := defaultChildRunner(context.Background(), []string{"/nonexistent/binary/xyz"}, newTestLogger())
	if code != 1 || err == nil {
		t.Fatalf("expected exit 1 + err, got code=%d err=%v", code, err)
	}
}

// recordingExecutor records every Exec call and optionally returns failures.
type recordingExecutor struct {
	calls    []string
	failNext int
	res      executor.Result
}

func (r *recordingExecutor) Exec(command string) executor.Result {
	r.calls = append(r.calls, command)
	if r.failNext > 0 {
		r.failNext--
		return r.res
	}
	return executor.Result{}
}
