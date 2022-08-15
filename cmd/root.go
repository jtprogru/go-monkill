// Package cmd contains all commands

package cmd

import (
	"github.com/spf13/cobra"
)

// Version of release
const Version string = "v0.2.1"

// Verbose flag
var Verbose bool

// Logfile filepath
//var Logfile string

// rootCmd – default root command
// rootCmd by default show help message
var rootCmd = &cobra.Command{
	Use:   "go-monkill",
	Short: "go-monkill for watching a process will finish or be killed",
	Long:  `Monitor when process with PID will finish or be killed and run what you need.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	go cobra.CheckErr(rootCmd.Execute())
}
