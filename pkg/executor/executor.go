package executor

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

type Executor struct{}

func (e Executor) Exec(command string) error {
	cmds := strings.Split(command, " ")
	if len(cmds) == 0 {
		return errors.New("command not specified")
	}
	sCmd := cmds[0]
	var args []string
	if len(cmds) > 1 {
		args = cmds[1:]
	}
	cmd := exec.Command(sCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
