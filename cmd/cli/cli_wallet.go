package cli

import (
	"fmt"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/mongo"
	"manindexer/inscribe/mrc20_service"
	"manindexer/man"
	"manindexer/mrc20"
	"strings"
)

var (
	wallet *CliWallet
)

type CliWallet struct {
	walletName string `json:"walletName"`
	privateKey string `json:"private_key"`
	address    string `json:"address"`
}

type WalletUtxo struct {
	TxId         string `json:"txId"`
	Vout         uint32 `json:"vout"`
	Shatoshi     int64  `json:"shatoshi"`
	ScriptPubKey string `json:"scriptPubKey"`
	Address      string `json:"address"`
}

func (c *CliWallet) toString() string {
	return "privateKey: " + c.privateKey + ", address: " + c.address
}

func (c *CliWallet) GetPrivateKey() string {
	return c.privateKey
}

func (c *CliWallet) GetAddress() string {
	return c.address
}

func (c *CliWallet) GetBtcUtxos(needAmount int64) ([]*mrc20_service.CommitUtxo, error) {
	utxos, err := GetUtxo([]string{c.GetAddress()})
	if err != nil {
		return nil, err
	}
	list := make([]*mrc20_service.CommitUtxo, 0)
	totalAmount := int64(0)
	for _, u := range utxos {
		//check pin
		outPoint := fmt.Sprintf("%s:%d", u.Txid, int64(u.Vout))
		pinInfo, err := man.DbAdapter.GetPinByOutput(outPoint)
		if err != nil && !strings.Contains(err.Error(), mongo.ErrNoDocuments.Error()) {
			return nil, err
		}
		if pinInfo != nil && pinInfo.Operation != "hide" {
			continue
		}

		//check mrc20
		_, total, _ := man.DbAdapter.GetHistoryByTx(u.Txid, int64(u.Vout), 0, 100)
		if err != nil && !strings.Contains(err.Error(), mongo.ErrNoDocuments.Error()) {
			return nil, err
		}
		if total > 0 {
			continue
		}
		pkScript, err := mrc20_service.AddressToPkScript(getNetParams(), c.GetAddress())
		if err != nil {
			return nil, err
		}

		amountDe := decimal.NewFromInt(u.Amount)
		item := &mrc20_service.CommitUtxo{
			PrivateKeyHex: wallet.GetPrivateKey(),
			PkScript:      pkScript,
			Address:       wallet.GetAddress(),
			UtxoTxId:      u.Txid,
			UtxoIndex:     uint32(u.Vout),
			UtxoOutValue:  amountDe.IntPart(),
		}
		list = append(list, item)
		totalAmount += item.UtxoOutValue
		if needAmount > 0 && totalAmount >= needAmount {
			break
		}
	}
	if totalAmount < needAmount {
		fmt.Printf("utxo amount not enough, need: %d, total: %d\n", needAmount, totalAmount)
		return nil, fmt.Errorf("utxo amount not enough")
	}
	return list, nil
}

func (c *CliWallet) GetMrc20Utxos(address, tickId, needAmount string) ([]*mrc20_service.TransferMrc20, error) {
	return getMrc20Utxos(address, tickId, needAmount)
}

func (c *CliWallet) GetShovels(address, tickId string) ([]*mrc20_service.MintPin, []*mrc20_service.PayTo, error) {
	return getShovels(address, tickId)
}

func (c *CliWallet) GetBalance() (int64, error) {
	addressList := []string{c.GetAddress()}
	balance := int64(0)
	utxoList, err := GetUtxo(addressList)
	if err != nil {
		return 0, err
	}
	for _, utxo := range utxoList {
		balance += utxo.Amount
	}
	return balance, nil
}

func (c *CliWallet) GetMrc20Balance() ([]mrc20.Mrc20Balance, error) {
	address := wallet.GetAddress()
	list, total, err := man.DbAdapter.GetMrc20BalanceByAddress(address, 0, 100)
	if err != nil {
		fmt.Printf("Failed to get mrc20 balance: %s\n", err)
		return nil, err
	}
	fmt.Printf("Total: %d\n", total)
	//for _, v := range list {
	//	fmt.Printf("TickId: %s, Balance: %s\n", v.Id, v.Balance)
	//}
	return list, nil
}
