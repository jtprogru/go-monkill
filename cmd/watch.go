// Package cmd contains all commands

package cmd

import (
	"errors"
	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/jtprogru/go-monkill/pkg/waiter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// watchCmd – command for run watch func
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `Monitor when process with PID will killed or stopped and run what you need.

For example:

go-monkill watch --pid 12345 --command "ping jtprog.ru -c 4"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logrus.New()
		if Verbose {
			l.SetLevel(logrus.DebugLevel)
		} else {
			l.SetLevel(logrus.InfoLevel)
		}

		return watcher(
			WatcherConfig.pid,
			WatcherConfig.command,
			WatcherConfig.timeout,
			&waiter.Waiter{},
			&executor.Executor{},
			l)
	},
}

// defaultPid is -1 for
var defaultPid = -1

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
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().IntVar(
		&WatcherConfig.pid,
		"pid",
		defaultPid,
		"PID for watching")
	watchCmd.PersistentFlags().StringVar(
		&WatcherConfig.command,
		"command",
		"ping jtprog.ru -c 2",
		"Command for running")
	watchCmd.PersistentFlags().Int64Var(
		&WatcherConfig.timeout,
		"timeout",
		defaultTimeOut,
		"Set timeout for check status of process")
}

// Waiter interface for monitor process PID every timeout milliseconds
type Waiter interface {
	Wait(pid int, timeout int64) (<-chan struct{}, error)
}

// Executor interface
type Executor interface {
	Exec(command string) error
}

// watcher – func run Waiter.Wait
// &WatcherConfig pid as PID for monitoring – defined in flag --pid
// &WatcherConfig command as command for running - defined in flag --command
// &WatcherConfig timeout as timeout for watch - defined in flag --timeout
func watcher(pid int, command string, timeout int64, w Waiter, e Executor, l *logrus.Logger) error {

	l.Debug("pid ", pid, " timeout ", timeout, " watcher was started")
	if err := checkPid(pid, l); err != nil {
		return err
	}
	l.Info("pid ", pid, " command ", command, " arguments readed")
	ch, err := w.Wait(pid, timeout)
	if err != nil {
		l.Error("break execution. Error on watch process ", err)
		return err
	}
	<-ch
	l.Info("pid ", pid, " process finished, run command")
	err = e.Exec(command)
	if err != nil {
		l.Error("break execution. Error on start command ", err)
		return err
	}
	return nil
}

// checkPid – func check correctness defined PID
func checkPid(pid int, l *logrus.Logger) error {
	if pid < 1 {
		l.WithFields(logrus.Fields{"pid": pid}).Debug("PID was not defined")
		return errors.New("PID was not defined")
	}
	if pid == 1 {
		l.WithFields(logrus.Fields{"pid": pid}).Debug("PID was defined as 1 - this is PID for init process")
		return errors.New("PID was defined as 1 - this is PID for init process")
	}
	l.WithFields(logrus.Fields{"pid": pid}).Info("PID was defined")
	return nil
}
