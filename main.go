//go:build linux || darwin

/*
Package go-monkill is a small utility that runs the desired command or
script as soon as a process with a known PID terminates (subcommand
"watch") or runs a child process and dispatches hooks based on its
exit code (subcommand "run").

Supported platforms: Linux and macOS. Windows is intentionally not
supported — see README for details.
*/

package main

import "github.com/jtprogru/go-monkill/cmd"

func main() {
	cmd.Execute()
}
