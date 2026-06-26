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

func TestExecWithExitCode(t *testing.T) {
	e := &Executor{Logger: newTestLogger()}
	res := e.Exec("sh -c 'exit 42'")
	if res.ExitCode != 42 {
		t.Fatalf("expected exit 42, got %d", res.ExitCode)
	}
	if res.Err == nil {
		t.Fatal("expected non-nil err for non-zero exit")
	}
}

func TestExecQuotedArgsWithSpaces(t *testing.T) {
	e := &Executor{Logger: newTestLogger()}
	// Single-quoted arg with spaces must be passed as one argument to echo.
	res := e.Exec(`sh -c 'echo "hello world" >/dev/null'`)
	if res.Err != nil || res.ExitCode != 0 {
		t.Fatalf("expected success with quoted args, got code=%d err=%v", res.ExitCode, res.Err)
	}
}

func TestExecMalformedQuoting(t *testing.T) {
	e := &Executor{Logger: newTestLogger()}
	res := e.Exec(`echo "unclosed`)
	if res.Err == nil {
		t.Fatal("expected parse error for unclosed quote")
	}
	if res.ExitCode != 1 {
		t.Fatalf("expected exit 1 on parse error, got %d", res.ExitCode)
	}
}

func TestExecNilLogger(t *testing.T) {
	e := &Executor{}
	res := e.Exec("true")
	if res.ExitCode != 0 || res.Err != nil {
		t.Fatalf("nil logger should not break Exec; got %+v", res)
	}
}
