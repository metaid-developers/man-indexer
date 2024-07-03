package cli

import "github.com/spf13/cobra"

var getBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Display balance for a given address",
	Long:  `Display balance for a given address`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
