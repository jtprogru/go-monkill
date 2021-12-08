// Package cmd contains all commands

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd – show current version of this package
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of monkill",
	Long:  `All software has versions. This is monkill's`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s\nVersion: %s\n", rootCmd.Long, Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
