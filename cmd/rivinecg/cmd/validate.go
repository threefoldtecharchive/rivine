package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/threefoldtech/rivine/cmd/rivinecg/pkg/config"
)

// root validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate content (e.g. configs)",
}

var validateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "validate blockchain config file",
	Args:  cobra.ExactArgs(0),
	Run:   validateConfigFile,
}

func validateConfigFile(cmd *cobra.Command, args []string) {
	_, err := config.ImportAndValidateConfig(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid config file: %v", err)
		os.Exit(1)
		return
	}
	fmt.Println("Ok")
}

func init() {
	validateConfigCmd.Flags().StringVarP(
		&filePath, "config", "c", "blockchaincfg.yaml",
		"file path of the config, ecoding is based on the file extension, can be yaml or json")
	validateCmd.AddCommand(
		validateConfigCmd,
	)
}
