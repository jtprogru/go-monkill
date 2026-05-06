// Package cmd contains all commands

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-monkill",
	Long:  `All software has versions. This is go-monkill's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("go-monkill %s\ncommit: %s\nbuilt:  %s\n", Version, Commit, BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
