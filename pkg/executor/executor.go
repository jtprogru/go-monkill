// Package executor runs an external command and reports its exit status.
package executor

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// Executor runs commands.
type Executor struct {
	Logger *logrus.Logger
}

// Result describes the outcome of running a command.
type Result struct {
	ExitCode int
	Err      error
}

// Exec parses the command string with simple whitespace splitting and runs it,
// streaming stdout/stderr to the parent process. The exit code of the child
// process is returned in Result.ExitCode.
func (e *Executor) Exec(command string) Result {
	cmds := strings.Fields(command)
	if len(cmds) == 0 {
		return Result{ExitCode: 1, Err: errors.New("command not specified")}
	}

	bin, args := cmds[0], cmds[1:]
	e.debugf("exec %q args=%v", bin, args)

	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

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
	if exitCode == 0 {
		e.infof("command %q finished with exit code 0", command)
	} else {
		e.infof("command %q finished with exit code %d", command, exitCode)
	}
	return Result{ExitCode: exitCode, Err: err}
}

func (e *Executor) debugf(format string, args ...interface{}) {
	if e.Logger == nil {
		return
	}
	e.Logger.Debugf(format, args...)
}

func (e *Executor) infof(format string, args ...interface{}) {
	if e.Logger == nil {
		return
	}
	e.Logger.Infof(format, args...)
}
