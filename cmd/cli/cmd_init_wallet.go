package cli

import (
	"fmt"
	"manindexer/man"

	"github.com/spf13/cobra"
)

const WALLETNAME = "metaID_MAN_wallet"

var (
	configFileName = "config.json"
)

var initWalletCmd = &cobra.Command{
	Use:   "init-wallet",
	Short: "Init Wallet for CLI",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		initCliWallet()
	},
}

func initCliWallet() {
	address, err := CreateLegacyWallet(WALLETNAME)
	if err == nil {
		fmt.Println("Wallet initialized successfully. Available address is:", address)
	} else {
		fmt.Println("Wallet initialization failed.", err)
	}

	privateKey, err := DumpPrivKeyHex(address)
	if err != nil {
		fmt.Println("Wallet initialization failed.", err)
		return
	}
	wallet = &CliWallet{
		walletName: WALLETNAME,
		privateKey: privateKey,
		address:    address,
	}
}

func checkWallet() error {
	if wallet == nil {
		address, err := CreateLegacyWallet(WALLETNAME)
		if err == nil {
		} else {
			fmt.Println("Wallet initialization failed.", err)
		}
		if address == "" {
			return fmt.Errorf("wallet is not initialized")
		}

		privateKey, err := DumpPrivKeyHex(address)
		if err != nil {
			fmt.Println("Wallet initialization failed.", err)
			return fmt.Errorf("wallet initialization failed:%s", err)
		}
		wallet = &CliWallet{
			walletName: WALLETNAME,
			privateKey: privateKey,
			address:    address,
		}

	}
	if wallet.GetAddress() == "" {
		return fmt.Errorf("wallet address is not initialized")
	}
	if wallet.GetPrivateKey() == "" {
		return fmt.Errorf("wallet private key is not initialized")
	}
	return nil
}

func checkManDbAdapter() error {
	if man.DbAdapter == nil {
		return fmt.Errorf("MAN DB adapter is not initialized")
	}
	return nil
}
