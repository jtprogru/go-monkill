package executor

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func newTestLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(&strings.Builder{})
	l.SetLevel(logrus.DebugLevel)
	return l
}

func TestExec(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		wantCode int
		wantErr  bool
	}{
		{"empty command", "", 1, true},
		{"only whitespace", "   \t  ", 1, true},
		{"successful command", "true", 0, false},
		{"failing command", "false", 1, true},
		{"nonexistent binary", "/nonexistent/binary/xyz", 1, true},
	}

	e := &Executor{Logger: newTestLogger()}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := e.Exec(tc.command)
			if (res.Err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", res.Err, tc.wantErr)
			}
			if res.ExitCode != tc.wantCode {
				t.Fatalf("exit code = %d, want %d", res.ExitCode, tc.wantCode)
			}
		})
	}
}

// TestExecWithExitCode verifies arbitrary exit-code propagation.
// Quoted argument support depends on the shell-style parser added in a
// later commit; once that lands, replace this with: `sh -c 'exit 42'`.
func TestExecWithExitCode(t *testing.T) {
	t.Skip("re-enable once command parser supports quoted arguments")
}

func TestExecNilLogger(t *testing.T) {
	e := &Executor{}
	res := e.Exec("true")
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("nil logger should not break Exec; got %+v", res)
	}
}
