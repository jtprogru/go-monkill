// Package cmd contains all commands

package cmd

import (
	"errors"
	"os"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/jtprogru/go-monkill/pkg/waiter"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// watchCmd – run watch func
//
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `Monitor when process with PID will killed or stopped and run what you need.

For example:

go-monkill watch --pid 12345 --command "ping jtprog.ru -c 4"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := zerolog.New(os.Stderr)
		l.Level(zerolog.TraceLevel)
		return watcher(WatcherConfig.pid, WatcherConfig.command, WatcherConfig.timeout, &waiter.Waiter{}, &executor.Executor{}, l)
	},
}

// Verbose flag
//var Verbose bool

// defaultPid is -1 for
var defaultPid int = -1

// defaultTimeOut is 250 milliseconds
var defaultTimeOut int64 = 250

// WatcherConfig provides config for watcher
var WatcherConfig struct {
	pid     int    // Specified PID for process
	command string // Specified command for running
	timeout int64  // Specified timeout for sleep
}

// Add command watchCmd to rootCmd
//
// &WatcherConfig pid as PID for monitoring – defined in flag --pid
// &WatcherConfig command as command for running - defined in flag --command
// &WatcherConfig timeout as timeout for check process - defined in flag --timeout
func init() {
	rootCmd.AddCommand(watchCmd)
	// TODO: Implement verbose log output by flag --verbose
	// rootCmd.InheritedFlags().BoolVar(&Verbose, "verbose", false, "Enable debug logging")
	// TODO: Implement output to logfile by flag --logfile
	// rootCmd.InheritedFlags().StringVar(&WatcherConfig.logfile, "logfile", "/tmp/go-monkill.log", "Enable debug logging")
	watchCmd.PersistentFlags().IntVar(&WatcherConfig.pid, "pid", defaultPid, "PID for watching")
	watchCmd.PersistentFlags().StringVar(&WatcherConfig.command, "command", "ping jtprog.ru -c 2", "Command for running")
	watchCmd.PersistentFlags().Int64Var(&WatcherConfig.timeout, "timeout", defaultTimeOut, "Set timeout for check status of process")
}

// Waiter interface for monitor process PID every timeout milliseconds
type Waiter interface {
	Wait(pid int, timeout int64) (<-chan struct{}, error)
}

// Executor interface
type Executor interface {
	Exec(command string) error
}

// watcher – run Waiter.Wait
// &WatcherConfig pid as PID for monitoring – defined in flag --pid
// &WatcherConfig command as command for running - defined in flag --command
// &WatcherConfig timeout as timeout for watch - defined in flag --timeout
func watcher(pid int, command string, timeout int64, w Waiter, e Executor, l zerolog.Logger) error {
	if err := checkPID(pid, l); err != nil {
		return err
	}
	l.Info().Int("pid", pid).Str("command", command).Msg("Arguments readed")
	ch, err := w.Wait(pid, timeout)
	if err != nil {
		l.Error().Err(err).Msg("Break execution. Error on watch process")
		return err
	}
	<-ch
	l.Info().Int("pid", pid).Msg("Process finished, run command")
	err = e.Exec(command)
	if err != nil {
		l.Error().Err(err).Msg("Break execution. Error on start command")
		return err
	}
	return nil
}

func checkPID(pid int, l zerolog.Logger) error {
	if pid == -1 || pid == 0 {
		l.Fatal().Int("pid", pid).Msg("PID was not defined")
		return errors.New("PID was not defined")
	}
	if pid == 1 {
		l.Fatal().Int("pid", pid).Msg("PID was defined as 1 - this is PID for init process")
		return errors.New("PID was defined as 1 - this is PID for init process")
	}
	l.Info().Int("pid", pid).Msgf("PID was defined as %d", pid)
	return nil
}
