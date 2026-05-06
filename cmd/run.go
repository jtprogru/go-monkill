// Package cmd contains all commands

package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- command...",
	Short: "Run a command and trigger hooks based on its exit code",
	Long: `Run a child command, capture its exit code, then optionally trigger one
or more hooks. The utility exits with the child's exit code.

Hooks (any combination):
  --on-success <cmd>   run when the child exits with code 0
  --on-failure <cmd>   run when the child exits with a non-zero code
  --on-exit    <cmd>   always run, in addition to --on-success / --on-failure

For example:

go-monkill run --on-success "deploy.sh" --on-failure "alert.sh" -- ./tests.sh
go-monkill run --on-exit "cleanup.sh" -- ./build.sh release
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l, closer, err := newLogger()
		if err != nil {
			return err
		}
		defer func() { _ = closer.Close() }()

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		exitCode, err := runner(
			ctx,
			args,
			RunConfig.onSuccess,
			RunConfig.onFailure,
			RunConfig.onExit,
			&executor.Executor{Logger: l},
			l,
		)
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return err
	},
}

// RunConfig holds parsed flags for the run command.
var RunConfig struct {
	onSuccess string
	onFailure string
	onExit    string
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(
		&RunConfig.onSuccess,
		"on-success", "",
		"command to run if the child exits with code 0",
	)
	runCmd.Flags().StringVar(
		&RunConfig.onFailure,
		"on-failure", "",
		"command to run if the child exits with a non-zero code",
	)
	runCmd.Flags().StringVar(
		&RunConfig.onExit,
		"on-exit", "",
		"command to run after the child exits, regardless of code",
	)
}

// childRunner runs a command and returns its exit code; it's an injection
// point so tests can avoid spawning real processes.
type childRunner func(ctx context.Context, args []string, l *logrus.Logger) (int, error)

// defaultChildRunner spawns args as a real subprocess inheriting std{in,out,err}.
func defaultChildRunner(ctx context.Context, args []string, l *logrus.Logger) (int, error) {
	l.Debugf("run start: %v", args)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
	}
	return exitCode, err
}

// runForChild allows tests to inject a fake child runner.
var runForChild childRunner = defaultChildRunner

// runner orchestrates child execution + hook dispatch. Returns the child's
// exit code; hook errors are logged but do not affect the returned code.
func runner(
	ctx context.Context,
	args []string,
	onSuccess, onFailure, onExit string,
	e Executor,
	l *logrus.Logger,
) (int, error) {
	if len(args) == 0 {
		return 1, errors.New("no command provided")
	}

	exitCode, runErr := runForChild(ctx, args, l)

	switch {
	case exitCode == 0:
		l.Infof("child exited with code 0")
		runHook(onSuccess, "on-success", e, l)
	default:
		l.Warnf("child exited with code %d: %v", exitCode, runErr)
		runHook(onFailure, "on-failure", e, l)
	}
	runHook(onExit, "on-exit", e, l)

	return exitCode, runErr
}

func runHook(command, label string, e Executor, l *logrus.Logger) {
	if command == "" {
		return
	}
	l.Infof("running %s hook: %s", label, command)
	res := e.Exec(command)
	if res.Err != nil {
		l.Errorf("%s hook failed (exit=%d): %v", label, res.ExitCode, res.Err)
	}
}
