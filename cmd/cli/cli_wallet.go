package cli

var (
	wallet *CliWallet
)

type CliWallet struct {
	account    string `json:"account"`
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

func (c *CliWallet) GetUtxos() []*WalletUtxo {
	return nil
}

func (c *CliWallet) GetMrc20Utxos() {

}

func (c *CliWallet) GetPins() {

}

func (c *CliWallet) GetBalance() {

}

func (c *CliWallet) GetMrc20Balance() {

}
