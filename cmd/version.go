// Package cmd contains all commands

package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// versionCmd – show current version of this package
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of go-monkill",
	Long:  `All software has versions. This is go-monkill's`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logrus.New()
		if Verbose {
			l.SetLevel(logrus.DebugLevel)
		} else {
			l.SetLevel(logrus.InfoLevel)
		}
		fmt.Printf("%s\nVersion: %s\n", rootCmd.Long, Version)
		l.Debug("version ", Version, " command 'version' was called")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
