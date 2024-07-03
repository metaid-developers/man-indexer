package cli

import (
	"github.com/spf13/cobra"
)

var initWalletCmd = &cobra.Command{
	Use:   "init-wallet",
	Short: "Init Wallet",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func CreateCliConfigFile() {
	//config := map[string]string{
	//	"mnemonics": mnemonics,
	//	"path":      path,
	//}
	//
	//file, err := json.MarshalIndent(config, "", "  ")
	//if err != nil {
	//	fmt.Println("Error creating JSON:", err)
	//	return
	//}
	//
	//err = ioutil.WriteFile("cli-config.json", file, 0644)
	//if err != nil {
	//	fmt.Println("Error writing file:", err)
	//	return
	//}
	//
	//fmt.Println("cli-config.json created successfully")
}
