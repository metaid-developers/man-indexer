package cli

import (
	"fmt"
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

	mrc20Balances, err := wallet.GetMrc20Balance()
	if err != nil {
		fmt.Printf("get mrc20 balance failed: %v\n", err)
		return
	}

	for _, v := range mrc20Balances {
		fmt.Printf("Mrc20Id:%s, TokenName: %s, Balance: %s\n", v.Id, v.Name, v.Balance.String())
	}
}
