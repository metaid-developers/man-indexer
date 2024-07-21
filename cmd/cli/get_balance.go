package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Display balance for a given address",
	Long:  `Display balance for a given address`,
	Run: func(cmd *cobra.Command, args []string) {
		printBalance()
	},
}

func printBalance() {
	address, balance, err := getBalance()
	if err == nil {
		fmt.Printf("address: %s,balance is: %d\n", address, balance)
	} else {
		fmt.Println("get balance failed.", err)
	}
}
func getBalance() (address string, balance int64, err error) {
	address, err = CreateLegacyWallet(WALLETNAME)
	if err != nil {
		return
	}
	addressList := []string{address}
	utxoList, err := GetUtxo(addressList)
	if err != nil {
		return
	}
	for _, utxo := range utxoList {
		balance += utxo.Amount
	}
	return
}
