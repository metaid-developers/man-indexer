package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

var getVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "version subcommand show git version info.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.0.1")
	},
}
