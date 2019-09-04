package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rivinecg",
	Short: "a rivine blockchain generator tool",
}

// Execute executes the command line logic as specified in this package
// as driven by the arguments and flags passed by the user.
//
// See specific command files for documentation about the individual commands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		versionCmd,
		generateCmd,
		validateCmd,
	)
}
