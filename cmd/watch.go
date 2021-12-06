package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"strings"
)

var pid int
var command string

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "A brief description of your command",
	Long: `Monitor when process with PID will killed or stopped and run what you need.

For example:

monkill watch --pid=12345 --command="rm -f /tmp/12345.log"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return watch()
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().IntVar(&pid, "pid", -1, "PID for watching")
	watchCmd.PersistentFlags().StringVar(&command, "command", "ping jtprog.ru -c 2", "Command for running")
}

func watch() error {
	F := true
	for F {
		if err := procPIDcmdline(pid); err != nil {
			fmt.Println(err)
			return err
		} else {
			F = false
			if err := runNeededCommand(command); err != nil {
				log.Fatal(err)
				return err
			}
			return nil
		}
	}
	return nil
}

func procPIDcmdline(p int) error {
	procCmdline := fmt.Sprintf("/proc/%d/cmdline", p)
	if _, err := os.Stat(procCmdline); errors.Is(err, os.ErrNotExist) {
		return errors.New(fmt.Sprintf("Process with PID %d was not found", p))
	}
	return nil
}

func runNeededCommand(commnd string) error {
	cmdd, err := exec.LookPath(strings.Split(commnd, " ")[0])
	args := strings.Split(commnd, " ")[:]
	if err != nil {
		log.Fatal("installing fortune is in your future")
		return err
	}
	cmd := &exec.Cmd{
		Path:   cmdd,
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stdout,
	}
	fmt.Println(cmd.String())
	if err := cmd.Run(); err != nil {
		log.Fatalf("[FATAL] Error: %v\n", err)
		return err
	}
	return nil
}
