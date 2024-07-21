package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"manindexer/man"
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
	address := wallet.GetAddress()
	list, total, err := man.DbAdapter.GetMrc20BalanceByAddress(address, 0, 100)
	if err != nil {
		fmt.Printf("Failed to get mrc20 balance: %s\n", err)
		return
	}
	fmt.Printf("Total: %d\n", total)
	for _, v := range list {
		fmt.Printf("TickId: %s, Balance: %s\n", v.Id, v.Balance)
	}
}
