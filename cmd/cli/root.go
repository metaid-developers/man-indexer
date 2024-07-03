package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var CfgFile = ""

var rootCmd = &cobra.Command{
	Use:   "man-cli",
	Short: "MAN-CLI is a tool to interact with metaid-v2",
	Long:  "This is a MAN-CLI, which is a tool to interact with metaid-v2 in bitcoin chain",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func initConfig() {
	//if err := viper.ReadInConfig(); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(initWalletCmd)
	rootCmd.AddCommand(getBalanceCmd)
	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "config.json", "config file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
