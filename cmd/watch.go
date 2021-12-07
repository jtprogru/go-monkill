package cmd

import (
	"os"

	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/jtprogru/go-monkill/pkg/waiter"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `Monitor when process with PID will killed or stopped and run what you need.

For example:

go-monkill watch --pid=12345 --command="ping jtprog.ru -c 4""
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := zerolog.New(os.Stderr)
		return watcher(WatcherConfig.pid, WatcherConfig.command, waiter.Waiter{}, executor.Executor{}, l)
	},
}

// WatcherConfig provides config for watcher
var WatcherConfig struct {
	pid     int    // Specified PID for process
	command string // Specified command for running
}

// Add command watchCmd to rootCmd
//
// &WatcherConfig pid as PID for monitoring – defined in flag --pid
// &WatcherConfig command as command for running - defined in flag --command
func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().IntVar(&WatcherConfig.pid, "pid", -1, "PID for watching")
	watchCmd.PersistentFlags().StringVar(&WatcherConfig.command, "command", "ping jtprog.ru -c 2", "Command for running")
}

// Waiter interface
type Waiter interface {
	Wait(pid int) (<-chan struct{}, error)
}

// Executor interface
type Executor interface {
	Exec(command string) error
}

func watcher(pid int, command string, w Waiter, e Executor, l zerolog.Logger) error {
	l.Info().Int("pid", pid).Str("command", command).Msg("Arguments readed")
	ch, err := w.Wait(pid)
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
