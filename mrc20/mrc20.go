package mrc20

const (
	ErrDeployContent     = "deploy content format error, it needs to be a JSON string"
	ErrDeployTickLength  = "no less than 4 characters"
	ErrDeployTickExists  = "tick already exists"
	ErrMintTickNotExists = "tick not exists"
	ErrMintLimit         = "minting capacity reached"
	ErrMintHeight        = "current block height is too low"
	ErrMintPopNull       = "shovel is none"
	ErrMintPopDiff       = "pop difficulty check failed"
	ErrMintTickIdNull    = "tickId is null"
	ErrMintPinIdNull     = "pin is null"
	ErrMintPinOwner      = "not have the right to use this PIN"
	ErrTranferReqData    = "transfer data error"
	ErrTranferBalnceErr  = "transfer balance error"
	ErrTranferBalnceLess = "insufficient balance for transfer"
)

type Mrc20Pin struct {
	Tick        string `json:"tick"`
	Mrc20Id     string `json:"mrc20Id"`
	PinNum      int64  `json:"pinNum"`
	PinId       string `json:"pinId"`
	Verify      bool   `json:"verify"`
	BlockHeight int64  `json:"blockHeight"`
	MrcOption   string `json:"mrcOption"`
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	PinTxHash   string `json:"pinTxHash"`
	Content     string `json:"content"`
	ErrorMsg    string `json:"errorMsg"`
	AmtChange   int64  `json:"amtChange"`
}
type Mrc20DeployQual struct {
	Lv string `json:"lvl"`
}
type Mrc20Deploy struct {
	Tick        string          `json:"tick"`
	TokenName   string          `json:"tokenName"`
	Decimals    string          `json:"decimals"`
	AmtPerMint  string          `json:"amtPerMint"`
	MintCount   string          `json:"mintCount"`
	Blockheight string          `json:"blockheight"`
	Metadata    string          `json:"metadata"`
	DeployType  string          `json:"type"`
	Qual        Mrc20DeployQual `json:"qual"`
}

type Mrc20DeployInfo struct {
	Tick        string          `json:"tick"`
	TokenName   string          `json:"tokenName"`
	Decimals    string          `json:"decimals"`
	AmtPerMint  string          `json:"amtPerMint"`
	MintCount   string          `json:"mintCount"`
	Blockheight string          `json:"blockheight"`
	Metadata    string          `json:"metadata"`
	DeployType  string          `json:"type"`
	Qual        Mrc20DeployQual `json:"qual"`
	TotalMinted int64           `json:"totalMinted"`
	Mrc20Id     string          `json:"mrc20Id"`
	PinNumber   int64           `json:"pinNumber"`
}
type Mrc20Shovel struct {
	Shovel   string `json:"Shovel"`
	UsePinId string `json:"usePinId"`
}
type Mrc20MintData struct {
	Id  string `json:"id"`
	Pin string `json:"pin"`
}
type Mrc20TranferData struct {
	Amount string `json:"amount"`
	Addr   string `json:"addr"`
	Id     string `json:"id"`
}
