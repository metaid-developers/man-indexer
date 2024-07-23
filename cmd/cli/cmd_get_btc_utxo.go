package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getBtcUtxoCmd = &cobra.Command{
	Use:   "utxo",
	Short: "Display utxo list",
	Long:  `Display utxo list for default address`,
	Run: func(cmd *cobra.Command, args []string) {
		printBtcUtxo()
	},
}

func printBtcUtxo() {
	address, list, err := getBtcUtxoList()
	if err == nil {
		fmt.Println("address:", address)
		for _, utxo := range list {
			fmt.Printf("%+v\n", utxo)
		}
	} else {
		fmt.Println("get balance failed.", err)
	}
}
func getBtcUtxoList() (address string, utxoList []*Utxo, err error) {
	address, err = CreateLegacyWallet(WALLETNAME)
	if err != nil {
		return
	}
	addressList := []string{address}
	utxoList, err = GetUtxo(addressList)
	return
}
