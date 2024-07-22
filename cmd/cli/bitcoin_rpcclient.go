package cli

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"manindexer/common"
	"math"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
)

type Utxo struct {
	Address       string
	Txid          string
	Vout          float64
	Amount        int64
	Confirmations float64
	Spendable     bool
}

var (
	client *rpcclient.Client
)

func InitBtcRpc(walletName string) {
	btc := common.Config.Btc
	rpcConfig := &rpcclient.ConnConfig{
		Host:                 btc.RpcHost + walletName,
		User:                 btc.RpcUser,
		Pass:                 btc.RpcPass,
		HTTPPostMode:         btc.RpcHTTPPostMode, // Bitcoin core only supports HTTP POST mode
		DisableTLS:           btc.RpcDisableTLS,   // Bitcoin core does not provide TLS by default
		DisableAutoReconnect: true,
		DisableConnectOnNew:  true,
	}
	var err error
	client, err = rpcclient.New(rpcConfig, nil)
	if err != nil {
		panic(err)
	}
}

func GetNewAddress(accountName string) (string, error) {
	newAddress, err := client.GetNewAddressType(accountName, "bech32")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error creating new address: %v", err))
	}
	return newAddress.EncodeAddress(), nil
}
func DumpPrivKeyHex(newAddress string) (string, error) {
	addr, err := btcutil.DecodeAddress(newAddress, getNetParams())
	if err != nil {
		return "", err
	}
	dumpPrivKey, err := client.DumpPrivKey(addr)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error dumping private key: %v", err))

	}
	return hex.EncodeToString(dumpPrivKey.PrivKey.Serialize()), nil
}

func CreateLegacyWallet(walletName string) (address string, err error) {
	client.RawRequest("createwallet", []json.RawMessage{
		json.RawMessage(fmt.Sprintf("\"%s\"", walletName)),
		json.RawMessage("false"), // disable_private_keys
		json.RawMessage("false"), // blank
		json.RawMessage("\"\""),  // passphrase
		json.RawMessage("false"), // avoid_reuse
		json.RawMessage("false"), // descriptors
		json.RawMessage("false"), // load_on_startup
	})
	//InitBtcRpc("/wallet/" + walletName)
	result, err := client.RawRequest("getaddressesbylabel", []json.RawMessage{
		json.RawMessage(fmt.Sprintf("\"%s\"", walletName)), // label
	})
	if err == nil {
		var list map[string]interface{}
		err = json.Unmarshal(result, &list)
		if err == nil && len(list) > 0 {
			var addressList []string
			for k := range list {
				addressList = append(addressList, k)
			}
			sort.Strings(addressList)
			address = addressList[0]
			return
		}
	}
	address, err = GetNewAddress(walletName)
	return
}
func GetMempool(account string) (amt btcutil.Amount, err error) {
	return client.GetUnconfirmedBalance(account)
}
func GetUtxo(addressList []string) (list []*Utxo, err error) {
	listStr := joinWithQuotes(addressList, ", ")
	result, err := client.RawRequest("listunspent", []json.RawMessage{
		json.RawMessage("1"),                 //minconf
		json.RawMessage("9999999"),           //maxconf
		json.RawMessage(`[` + listStr + `]`), //address list
	})
	if err != nil {
		return
	}

	var utxos []map[string]interface{}
	err = json.Unmarshal(result, &utxos)
	if err != nil {
		return
	}
	for _, utxo := range utxos {
		u := Utxo{
			Address:       utxo["address"].(string),
			Txid:          utxo["txid"].(string),
			Vout:          utxo["vout"].(float64),
			Amount:        int64(math.Round(utxo["amount"].(float64) * 1e8)),
			Confirmations: utxo["confirmations"].(float64),
			Spendable:     utxo["spendable"].(bool),
		}
		list = append(list, &u)
	}
	return
}
func joinWithQuotes(slice []string, separator string) string {
	quotedSlice := make([]string, len(slice))
	for i, str := range slice {
		quotedSlice[i] = fmt.Sprintf(`"%s"`, str)
	}
	return strings.Join(quotedSlice, separator)
}

func broadcastTx(txHex string) (string, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", err
	}
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return "", err
	}
	txHash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		return "", nil
	}
	return txHash.String(), nil
}
