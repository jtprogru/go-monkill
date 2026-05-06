// Package cmd contains all commands

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/jtprogru/go-monkill/pkg/waiter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// exitCodeInterrupted is returned when the watcher is canceled by a signal
// before the watched process(es) terminate (128+SIGINT, the shell convention).
const exitCodeInterrupted = 130

// exitCodeTimeout is returned when --max-wait elapses before the watched
// process(es) terminate (matches the convention used by timeout(1)).
const exitCodeTimeout = 124

// Wait-for modes for multi-PID monitoring.
const (
	waitForAll = "all"
	waitForAny = "any"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch one or more PIDs and run a command after they terminate",
	Long: `Monitor when process(es) with given PID(s) will be killed or stopped and run what you need.

For example:

go-monkill watch --pid 12345 --command "ping jtprog.ru -c 4"
go-monkill watch --pid 12345 --pid 67890 --wait-for any --command "echo first done"
go-monkill watch --pid 100,200,300 --wait-for all --command "cleanup.sh"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l, closer, err := newLogger()
		if err != nil {
			return err
		}
		defer func() { _ = closer.Close() }()

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		exitCode, err := watcher(
			ctx,
			WatcherConfig.pids,
			WatcherConfig.command,
			WatcherConfig.timeout,
			WatcherConfig.maxWait,
			WatcherConfig.waitFor,
			&waiter.Waiter{Logger: l},
			&executor.Executor{Logger: l},
			l,
		)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return err
	},
}

var defaultTimeOut int64 = 250

// WatcherConfig holds parsed flags for the watch command.
var WatcherConfig struct {
	pids    []int
	command string
	timeout int64
	maxWait time.Duration
	waitFor string
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().IntSliceVar(
		&WatcherConfig.pids,
		"pid",
		nil,
		"PID(s) to watch. Repeat the flag or pass a comma-separated list",
	)
	watchCmd.PersistentFlags().StringVar(&WatcherConfig.command, "command", "", "command to run after the process exits")
	watchCmd.PersistentFlags().Int64Var(
		&WatcherConfig.timeout,
		"timeout",
		defaultTimeOut,
		"poll interval in milliseconds",
	)
	watchCmd.PersistentFlags().DurationVar(
		&WatcherConfig.maxWait,
		"max-wait",
		0,
		"give up waiting after this duration (e.g. 30s, 5m); 0 = unlimited",
	)
	watchCmd.PersistentFlags().StringVar(
		&WatcherConfig.waitFor,
		"wait-for",
		waitForAll,
		"with multiple PIDs, run the command after 'all' have exited or 'any' first one",
	)
}

// Waiter watches a PID and signals when it terminates.
type Waiter interface {
	Wait(ctx context.Context, pid int, timeout int64) (<-chan struct{}, error)
}

// Executor runs an external command and returns its exit status.
type Executor interface {
	Exec(command string) executor.Result
}

// watcher orchestrates a Waiter + Executor and returns the exit code of the
// executed command (0 when no command ran or it succeeded).
func watcher(
	ctx context.Context,
	pids []int,
	command string,
	timeout int64,
	maxWait time.Duration,
	waitFor string,
	w Waiter,
	e Executor,
	l *logrus.Logger,
) (int, error) {
	l.Debugf("watcher start: pids=%v timeout=%dms maxWait=%s waitFor=%s command=%q",
		pids, timeout, maxWait, waitFor, command)

	if len(pids) == 0 {
		return 1, errors.New("at least one --pid must be provided")
	}
	for _, pid := range pids {
		if err := checkPid(pid, l); err != nil {
			return 1, err
		}
	}
	if command == "" {
		return 1, errors.New("--command must not be empty")
	}
	if waitFor != waitForAll && waitFor != waitForAny {
		return 1, fmt.Errorf("--wait-for must be %q or %q, got %q", waitForAll, waitForAny, waitFor)
	}

	if maxWait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, maxWait)
		defer cancel()
	}

	if err := waitForPids(ctx, pids, waitFor, timeout, w, l); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			l.Warnf("max-wait %s elapsed before pids=%v exited (mode=%s); skipping command",
				maxWait, pids, waitFor)
			return exitCodeTimeout, err
		}
		if errors.Is(err, context.Canceled) {
			l.Warnf("interrupted before pids=%v exited (mode=%s); skipping command", pids, waitFor)
			return exitCodeInterrupted, err
		}
		l.Errorf("waiter failed: %v", err)
		return 1, err
	}

	l.Infof("pids=%v finished (mode=%s), running command: %s", pids, waitFor, command)
	res := e.Exec(command)
	if res.Err != nil {
		l.Errorf("command failed (exit=%d): %v", res.ExitCode, res.Err)
		return res.ExitCode, res.Err
	}
	return res.ExitCode, nil
}

// waitForPids fans out to one Waiter.Wait per PID and aggregates according to
// mode: "all" returns when every PID has terminated; "any" returns when the
// first PID terminates. Either ctx cancellation or a Wait setup error aborts.
func waitForPids(
	ctx context.Context,
	pids []int,
	mode string,
	timeout int64,
	w Waiter,
	_ *logrus.Logger,
) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	chans := make([]<-chan struct{}, 0, len(pids))
	for _, pid := range pids {
		ch, err := w.Wait(childCtx, pid, timeout)
		if err != nil {
			return fmt.Errorf("waiter setup for pid=%d: %w", pid, err)
		}
		chans = append(chans, ch)
	}

	if mode == waitForAny {
		first := make(chan struct{})
		var once sync.Once
		for _, ch := range chans {
			go func(c <-chan struct{}) {
				select {
				case <-c:
					once.Do(func() { close(first) })
				case <-childCtx.Done():
				}
			}(ch)
		}
		select {
		case <-first:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// waitForAll
	for _, ch := range chans {
		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func checkPid(pid int, l *logrus.Logger) error {
	if pid < 1 {
		l.WithFields(logrus.Fields{"pid": pid}).Debug("PID was not defined")
		return fmt.Errorf("invalid PID %d", pid)
	}
	if pid == 1 {
		l.WithFields(logrus.Fields{"pid": pid}).Debug("PID 1 is the init process and cannot be watched")
		return errors.New("PID was defined as 1 - this is PID for init process")
	}
	l.WithFields(logrus.Fields{"pid": pid}).Debug("PID accepted")
	return nil
}
