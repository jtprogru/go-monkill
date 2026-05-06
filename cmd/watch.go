// Package cmd contains all commands

package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/jtprogru/go-monkill/pkg/waiter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// exitCodeInterrupted is returned when the watcher is canceled by a signal
// before the watched process terminates (128+SIGINT, the shell convention).
const exitCodeInterrupted = 130

// exitCodeTimeout is returned when --max-wait elapses before the watched
// process terminates (matches the convention used by timeout(1)).
const exitCodeTimeout = 124

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a PID and run a command after it terminates",
	Long: `Monitor when process with PID will be killed or stopped and run what you need.

For example:

go-monkill watch --pid 12345 --command "ping jtprog.ru -c 4"
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
			WatcherConfig.pid,
			WatcherConfig.command,
			WatcherConfig.timeout,
			WatcherConfig.maxWait,
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

var defaultPid = -1
var defaultTimeOut int64 = 250

// WatcherConfig holds parsed flags for the watch command.
var WatcherConfig struct {
	pid     int
	command string
	timeout int64
	maxWait time.Duration
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().IntVar(&WatcherConfig.pid, "pid", defaultPid, "PID to watch")
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
	pid int,
	command string,
	timeout int64,
	maxWait time.Duration,
	w Waiter,
	e Executor,
	l *logrus.Logger,
) (int, error) {
	l.Debugf("watcher start: pid=%d timeout=%dms maxWait=%s command=%q", pid, timeout, maxWait, command)

	if err := checkPid(pid, l); err != nil {
		return 1, err
	}
	if command == "" {
		return 1, errors.New("--command must not be empty")
	}

	if maxWait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, maxWait)
		defer cancel()
	}

	ch, err := w.Wait(ctx, pid, timeout)
	if err != nil {
		l.Errorf("waiter failed: %v", err)
		return 1, err
	}

	select {
	case <-ch:
		// proceed to run the command
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			l.Warnf("max-wait %s elapsed before pid=%d exited; skipping command", maxWait, pid)
			return exitCodeTimeout, ctx.Err()
		}
		l.Warnf("interrupted before pid=%d exited; skipping command", pid)
		return exitCodeInterrupted, ctx.Err()
	}

	l.Infof("pid=%d finished, running command: %s", pid, command)
	res := e.Exec(command)
	if res.Err != nil {
		l.Errorf("command failed (exit=%d): %v", res.ExitCode, res.Err)
		return res.ExitCode, res.Err
	}
	return res.ExitCode, nil
}

func checkPid(pid int, l *logrus.Logger) error {
	if pid < 1 {
		l.WithFields(logrus.Fields{"pid": pid}).Debug("PID was not defined")
		return errors.New("PID was not defined (use --pid)")
	}
	if pid == 1 {
		l.WithFields(logrus.Fields{"pid": pid}).Debug("PID 1 is the init process and cannot be watched")
		return errors.New("PID was defined as 1 - this is PID for init process")
	}
	l.WithFields(logrus.Fields{"pid": pid}).Debug("PID accepted")
	return nil
}
