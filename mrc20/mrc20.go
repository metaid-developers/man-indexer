package mrc20

import (
	"github.com/shopspring/decimal"
)

const (
	ErrDeployContent        = "deploy content format error, it needs to be a JSON string"
	ErrDeployTickLength     = "the length must be between 2 and 24"
	ErrDeployTickNameLength = "the length must be between 1 and 48"
	ErrDeployNum            = "incorrect deployment parameters"
	ErrDeployTxGet          = "failed to retrieve transaction information"
	ErrDeployTickExists     = "tick already exists"
	ErrCrossChain           = "cross-chain operations are currently not allowed"
	ErrMintTickNotExists    = "tick not exists"
	ErrMintLimit            = "minting limit reached"
	ErrMintHeight           = "current block height is too low"
	ErrMintPopNull          = "shovel is none"
	ErrMintPopDiff          = "pop level check failed"
	ErrMintCreator          = "creator check failed"
	ErrMintPathCheck        = "shovel path check failed"
	ErrMintCountCheck       = "shovel count check failed"
	ErrMintTickIdNull       = "tickId is null"
	ErrMintDecimals         = "decimals error"
	ErrMintPinIdNull        = "pin is null"
	ErrMintPinOwner         = "not have the right to use this PIN"
	ErrTranferReqData       = "transfer data error"
	ErrTranferBalnceErr     = "transfer balance error"
	ErrTranferBalnceLess    = "insufficient balance for transfer"
)

type Mrc20Utxo struct {
	Tick        string          `json:"tick"`
	Mrc20Id     string          `json:"mrc20Id"`
	TxPoint     string          `json:"txPoint"`
	PointValue  int64           `json:"pointValue"`
	PinId       string          `json:"pinId"`
	PinContent  string          `json:"pinContent"`
	Verify      bool            `json:"verify"`
	BlockHeight int64           `json:"blockHeight"`
	MrcOption   string          `json:"mrcOption"`
	FromAddress string          `json:"fromAddress"`
	ToAddress   string          `json:"toAddress"`
	Msg         string          `json:"msg"`
	AmtChange   decimal.Decimal `json:"amtChange"`
	Status      int             `json:"status"`
	Chain       string          `json:"chain"`
	Index       int             `json:"index"`
	Timestamp   int64           `json:"timestamp"`
}
type Mrc20DeployQual struct {
	Creator string `json:"creator"`
	Lv      string `json:"lvl"`
	Path    string `json:"path"`
	Count   string `json:"count"`
}
type Mrc20Deploy struct {
	Tick         string          `json:"tick"`
	TokenName    string          `json:"tokenName"`
	Decimals     string          `json:"decimals"`
	AmtPerMint   string          `json:"amtPerMint"`
	MintCount    string          `json:"mintCount"`
	Blockheight  string          `json:"blockheight"`
	Metadata     string          `json:"metadata"`
	DeployType   string          `json:"type"`
	PremineCount string          `json:"premineCount"`
	Qual         Mrc20DeployQual `json:"qual"`
}

type Mrc20DeployInfo struct {
	Tick         string          `json:"tick"`
	TokenName    string          `json:"tokenName"`
	Decimals     string          `json:"decimals"`
	AmtPerMint   string          `json:"amtPerMint"`
	MintCount    int64           `json:"mintCount"`
	Blockheight  string          `json:"blockheight"`
	Metadata     string          `json:"metadata"`
	DeployType   string          `json:"type"`
	PremineCount int64           `json:"premineCount"`
	Qual         Mrc20DeployQual `json:"qual"`
	TotalMinted  int64           `json:"totalMinted"`
	Mrc20Id      string          `json:"mrc20Id"`
	PinNumber    int64           `json:"pinNumber"`
	Chain        string          `json:"chain"`
	Holders      int64           `json:"holders"`
	TxCount      int64           `json:"txCount"`
	MetaId       string          `json:"metaId"`
	Address      string          `json:"address"`
	DeployTime   int64           `json:"deployTime"`
}
type Mrc20Shovel struct {
	Id           string `json:"id"`
	Mrc20MintPin string `json:"mrc20MintPin"`
}
type Mrc20MintData struct {
	Id string `json:"id"`
	//Pin string `json:"pin"`
}
type Mrc20TranferData struct {
	Amount string `json:"amount"`
	Vout   int    `json:"vout"`
	Id     string `json:"id"`
}
type Mrc20Balance struct {
	Id      string          `json:"id"`
	Name    string          `json:"name"`
	Balance decimal.Decimal `json:"balance"`
}
