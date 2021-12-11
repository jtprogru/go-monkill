// Package cmd contains all commands

package cmd

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

// versionCmd – show current version of this package
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of monkill",
	Long:  `All software has versions. This is monkill's`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := zerolog.New(os.Stderr)
		if Verbose {
			l.Level(zerolog.DebugLevel)
		} else {
			l.Level(zerolog.InfoLevel)
		}
		fmt.Printf("%s\nVersion: %s\n", rootCmd.Long, Version)
		l.Debug().Str("version", Version).Msg("Command 'version' was called")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
