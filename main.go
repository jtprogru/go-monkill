/*
	Package go-monkill
	A very simple utility that allows you to run the desired command or script
	as soon as a certain process with a known PID completes correctly or with an error.

	For example:
		go-monkill watch --pid 12345 --command "ping jtprog.ru -c 4"
*/

package main

import "github.com/jtprogru/go-monkill/cmd"

func main() {
	cmd.Execute()
}
