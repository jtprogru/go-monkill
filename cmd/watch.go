package cmd

import (
	"github.com/jtprogru/go-monkill/pkg/executor"
	"github.com/jtprogru/go-monkill/pkg/waiter"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `Monitor when process with PID will killed or stopped and run what you need.

For example:

monkill watch --pid=12345 --command="rm -f /tmp/12345.log"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := zerolog.New(os.Stderr)
		return watcher(WatcherConfig.pid, WatcherConfig.command, waiter.Waiter{}, executor.Executor{}, l)
	},
}

var WatcherConfig struct {
	pid     int
	command string
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().IntVar(&WatcherConfig.pid, "pid", -1, "PID for watching")
	watchCmd.PersistentFlags().StringVar(&WatcherConfig.command, "command", "ping jtprog.ru -c 2", "Command for running")
}

type Waiter interface {
	Wait(pid int) (<-chan struct{}, error)
}

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
