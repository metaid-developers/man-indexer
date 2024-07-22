package cli

import (
	"github.com/spf13/cobra"
)

var getMrc20BalanceCmd = &cobra.Command{
	Use:   "mrc20balance",
	Short: "Display mrc20 balance for a given address",
	Long:  `Display mrc20 balance for a given address`,
	Run: func(cmd *cobra.Command, args []string) {
		getMrc20Balance()
	},
}

func getMrc20Balance() {
	if err := checkWallet(); err != nil {
		return
	}
	if err := checkManDbAdapter(); err != nil {
		return
	}

}
