package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/threefoldtech/rivine/build"
	rivinecli "github.com/threefoldtech/rivine/pkg/client"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print version information.",
	Run: rivinecli.Wrap(func() {
		fmt.Printf("Rivine Blockchain Generator Client v%s\r\n",
			build.Version.String(),
		)

		fmt.Println()
		fmt.Printf("Go Version   v%s\r\n", runtime.Version()[2:])
		fmt.Printf("GOOS         %s\r\n", runtime.GOOS)
		fmt.Printf("GOARCH       %s\r\n", runtime.GOARCH)
	}),
}
