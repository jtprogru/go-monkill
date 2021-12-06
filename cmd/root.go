package cmd

import (
	"github.com/spf13/cobra"
)

const Version string = "v0.0.1"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "monkill",
	Short: "Monkill for watching a process will stopped",
	Long:  `Monitor when process with PID will killed or stopped and run what you need.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
