package cli

import (
	"bufio"
	"fmt"
	"manindexer/inscribe/mrc20_service"
	"manindexer/man"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// ./man-cli mrc20op deploy
// ./man-cli mrc20op mint {tickId} {feeRate}
// ./man-cli mrc20op transfer {tickId} {to} {amount} {feeRate}

var mrc20OperationCmd = &cobra.Command{
	Use:   "mrc20op",
	Short: "MRC20 related operation commands,support deploy, mint, transfer",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkWallet(); err != nil {
			return
		}
		if err := checkManDbAdapter(); err != nil {
			return
		}

		if len(args) < 1 {
			fmt.Println("mrc20op command required")
			return
		}
		switch args[0] {
		case "deploy":
			tick := ""
			tokenName := ""
			decimals := ""
			amtPerMint := ""
			mintCount := ""
			premineCount := ""
			beginBlock := ""
			endBlock := ""
			pinCheckCreator := ""
			pinCheckPath := ""
			pinCheckCount := ""
			pinCheckLvl := ""
			feeRate := int64(0)
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter tick (2-24 characters): ")
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			tick = strings.TrimSpace(input)
			if len(tick) < 2 || len(tick) > 24 {
				fmt.Println("tick length should be 2-24 characters")
				return
			}

			fmt.Print("Enter tokenName: ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			tokenName = strings.TrimSpace(input)

			fmt.Print("Enter decimals (0-12, default 0): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			decimals = strings.TrimSpace(input)

			fmt.Print("Enter amtPerMint ([1, 1e12]): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			amtPerMint = strings.TrimSpace(input)
			if amtPerMint == "" {
				fmt.Println("amtPerMint is required")
				return
			}

			fmt.Print("Enter mintCount ([1, 1e12]): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			mintCount = strings.TrimSpace(input)
			if mintCount == "" {
				fmt.Println("mintCount is required")
				return
			}

			fmt.Print("Enter premineCount (optional, default to 0, [0, mintCount]): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			premineCount = strings.TrimSpace(input)

			fmt.Print("Enter Begin block(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			beginBlock = strings.TrimSpace(input)

			fmt.Print("Enter End block(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			endBlock = strings.TrimSpace(input)

			fmt.Print("Enter pinCheck-Creator(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			pinCheckCreator = strings.TrimSpace(input)

			fmt.Print("Enter pinCheck-Path(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			pinCheckPath = strings.TrimSpace(input)

			fmt.Print("Enter pinCheck-Count(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			pinCheckCount = strings.TrimSpace(input)

			fmt.Print("Enter pinCheck-Lvl(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			pinCheckLvl = strings.TrimSpace(input)

			fmt.Print("Enter payCheck-payTo(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			payCheckPayTo := strings.TrimSpace(input)
			fmt.Print("Enter payCheck-payAmount(optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			payCheckPayAmount := strings.TrimSpace(input)

			fmt.Print("Enter FeeRate: ")
			input, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			feeRate, _ = strconv.ParseInt(strings.TrimSpace(input), 10, 64)

			mrc20opDeploy(tick, tokenName,
				decimals, amtPerMint, mintCount, premineCount, beginBlock, endBlock,
				pinCheckCreator, pinCheckPath, pinCheckCount, pinCheckLvl, payCheckPayTo, payCheckPayAmount, feeRate)
			break
		case "mint":
			if len(args) < 3 {
				fmt.Println("mrc20op mint {tickId} {feeRate}")
				return
			}
			tickId := args[1]
			feeRate, _ := strconv.ParseInt(args[2], 10, 64)
			mrc20opMint(tickId, feeRate)
			break
		case "transfer":
			if len(args) < 5 {
				fmt.Println("mrc20op transfer {tickId} {to} {amount} {feeRate}")
				return
			}
			tickId := args[1]
			to := args[2]
			amount := args[3]
			feeRate, _ := strconv.ParseInt(args[4], 10, 64)
			mrc20opTransfer(tickId, to, amount, feeRate)
			return
		}
	},
}

func mrc20opDeploy(tick, tokenName string,
	decimals, amtPerMint, mintCount, premineCount, beginBlock, endBlock string,
	pinCheckCreator, pinCheckPath, pinCheckCount, pinCheckLvl string, payTo, payAmount string, feeRate int64) {
	var (
		commitTxId, revealTxId string = "", ""
		fee                    int64  = 0
		err                    error
		opRep                  *mrc20_service.Mrc20OpRequest
		payload                string = ""
		fetchCommitUtxoFunc    mrc20_service.FetchCommitUtxoFunc
	)
	payload, _, _ = mrc20_service.MakeDeployPayloadForIdCoins(
		tick, tokenName, "", payTo, payAmount,
		mintCount, amtPerMint, premineCount, beginBlock, endBlock, decimals,
		pinCheckCreator, pinCheckPath, pinCheckCount, pinCheckLvl)

	tickInfo, _ := man.DbAdapter.GetMrc20TickInfo("", strings.ToUpper(tick))
	if tickInfo.Mrc20Id != "" {
		fmt.Printf("Mrc20 tick:%s already exist\n", tickInfo.Tick)
		return
	}

	opRep = &mrc20_service.Mrc20OpRequest{
		Net:                     getNetParams(),
		MetaIdFlag:              getMetaIdFlag(),
		Op:                      "deploy",
		OpPayload:               payload,
		DeployPinOutAddress:     wallet.GetAddress(),
		DeployPremineOutAddress: wallet.GetAddress(),
		Mrc20OutValue:           546,
		ChangeAddress:           wallet.GetAddress(),
	}

	fetchCommitUtxoFunc = func(needAmount int64) ([]*mrc20_service.CommitUtxo, error) {
		return wallet.GetBtcUtxos(needAmount)
	}

	commitTxId, revealTxId, fee, err = mrc20_service.Mrc20Deploy(opRep, feeRate, fetchCommitUtxoFunc, broadcastTx)
	if err != nil {
		fmt.Printf("Mrc20 deploy err:%s\n", err.Error())
		return
	}
	fmt.Printf("Mrc20 deploy success\n")
	fmt.Printf("Fee:%d\n", fee)
	fmt.Printf("CommitTx:%s\n", commitTxId)
	fmt.Printf("RevealTxId:%s\n", revealTxId)
}

func mrc20opMint(tickId string, feeRate int64) {
	var (
		commitTxId, revealTxId string = "", ""
		fee                    int64  = 0
		err                    error
		opRep                  *mrc20_service.Mrc20OpRequest
		payload                string                   = fmt.Sprintf(`{"id":"%s"}`, tickId)
		mintPins               []*mrc20_service.MintPin = make([]*mrc20_service.MintPin, 0)
		payTos                 []*mrc20_service.PayTo   = make([]*mrc20_service.PayTo, 0)
		changeAddress          string                   = wallet.GetAddress()
		fetchCommitUtxoFunc    mrc20_service.FetchCommitUtxoFunc
	)

	mintPins, payTos, err = wallet.GetShovels(wallet.GetAddress(), tickId)
	if err != nil {
		fmt.Printf("Mrc20 mint err:%s\n", err.Error())
		return
	}

	payload = fmt.Sprintf(`{"id":"%s", "vout":"%d"}`, tickId, len(mintPins)+1)

	opRep = &mrc20_service.Mrc20OpRequest{
		Net:           getNetParams(),
		MetaIdFlag:    getMetaIdFlag(),
		Op:            "mint",
		OpPayload:     payload,
		MintPins:      mintPins,
		PayTos:        payTos,
		Mrc20OutValue: 546,
		Mrc20OutAddressList: []string{
			wallet.GetAddress(),
			wallet.GetAddress(),
		},
		ChangeAddress: changeAddress,
	}
	fetchCommitUtxoFunc = func(needAmount int64) ([]*mrc20_service.CommitUtxo, error) {
		return wallet.GetBtcUtxos(needAmount)
	}

	commitTxId, revealTxId, fee, err = mrc20_service.Mrc20Mint(opRep, feeRate, fetchCommitUtxoFunc, broadcastTx)
	if err != nil {
		fmt.Printf("Mrc20 mint err:%s\n", err.Error())
		return
	}
	fmt.Printf("Mrc20 mint success\n")
	fmt.Printf("Fee:%d\n", fee)
	fmt.Printf("CommitTx:%s\n", commitTxId)
	fmt.Printf("RevealTxId:%s\n", revealTxId)
}

func mrc20opTransfer(tickId, to, amount string, feeRate int64) {
	var (
		commitTxId, revealTxId string = "", ""
		fee                    int64  = 0
		err                    error
		toPkScript, _                 = mrc20_service.AddressToPkScript(getNetParams(), to)
		changeAddress          string = wallet.GetAddress()
		opRep                  *mrc20_service.Mrc20OpRequest
		transferMrc20s         []*mrc20_service.TransferMrc20 = make([]*mrc20_service.TransferMrc20, 0)
		mrc20Outs              []*mrc20_service.Mrc20OutInfo  = []*mrc20_service.Mrc20OutInfo{
			{
				Amount:   amount,
				Address:  to,
				PkScript: toPkScript,
				OutValue: 546,
			},
		}
		payload             string = ""
		fetchCommitUtxoFunc mrc20_service.FetchCommitUtxoFunc
	)

	transferMrc20s, err = wallet.GetMrc20Utxos(wallet.GetAddress(), tickId, amount)
	if err != nil {
		fmt.Printf("Mrc20 transfer err:%s\n", err.Error())
		return
	}

	payload, err = mrc20_service.MakeTransferPayload(tickId, transferMrc20s, mrc20Outs)
	if err != nil {
		fmt.Printf("Mrc20 transfer err:%s\n", err.Error())
		return
	}
	opRep = &mrc20_service.Mrc20OpRequest{
		Net:            getNetParams(),
		MetaIdFlag:     getMetaIdFlag(),
		Op:             "transfer",
		OpPayload:      payload,
		TransferMrc20s: transferMrc20s,
		Mrc20Outs:      mrc20Outs,
		ChangeAddress:  changeAddress,
	}

	fetchCommitUtxoFunc = func(needAmount int64) ([]*mrc20_service.CommitUtxo, error) {
		return wallet.GetBtcUtxos(needAmount)
	}

	commitTxId, revealTxId, fee, err = mrc20_service.Mrc20Transfer(opRep, feeRate, fetchCommitUtxoFunc, broadcastTx)
	if err != nil {
		fmt.Printf("Mrc20 transfer err:%s\n", err.Error())
		return
	}
	fmt.Printf("Mrc20 transfer success\n")
	fmt.Printf("Fee:%d\n", fee)
	fmt.Printf("CommitTx:%s\n", commitTxId)
	fmt.Printf("RevealTxId:%s\n", revealTxId)
}
