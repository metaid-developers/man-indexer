package metaaccess

const (
	ErrOperation      = "operation  error"
	ErrPinContent     = "pin content  error"
	ErrGetContronlPin = "get control pin  error"
	ErrGetPassTx      = "get pass transaction  error"
)

type AccessControl struct {
	PinId         string                  `json:"pinId"`
	Address       string                  `json:"address"`
	MetaId        string                  `json:"metaId"`
	ControlPins   []string                `json:"controlPins"`
	ControlPath   string                  `json:"controlPath"`
	ManDomain     string                  `json:"manDomain"`
	ManPubkey     string                  `json:"manPubkey"`
	CreatorPubkey string                  `json:"creatorPubkey"`
	EncryptedKey  string                  `json:"encryptedKey"`
	HoldCheck     *AccessControlHoldCheck `json:"holdCheck"`
	PayCheck      *AccessControlPayCheck  `json:"payCheck"`
	Mempool       int                     `json:"mempool"`
}
type AccessControlHoldCheck struct {
	AccType string `json:"type"` //"chainCoin" or "mrc20"
	Ticker  string `json:"ticker"`
	Amount  string `json:"amount"`
}
type AccessControlPayCheck struct {
	AccType     string `json:"type"`
	Ticker      string `json:"ticker"`
	Amount      string `json:"amount"`
	PayTo       string `json:"payTo"`
	ValidPeriod string `json:"validPeriod"`
}
type AccessPass struct {
	AccessControlID string `json:"accessControlID"`
}
type AccessPassData struct {
	PinId          string `json:"pinId"`
	CreatorAddress string `json:"creatorAddress"`
	BuyerAddress   string `json:"buyerAddress"`
	CreatorMetaId  string `json:"creatorMetaId"`
	BuyerMetaId    string `json:"buyerMetaId"`
	ControlId      string `json:"controlId"`
	ControlPath    string `json:"controlPath"`
	ContentPinId   string `json:"contentPinId"`
	CheckMode      string `json:"checkMode"`
	ValidPeriod    int64  `json:"validPeriod"`
	ManPubkey      string `json:"manPubkey"`
	CreatorPubkey  string `json:"creatorPubkey"`
	EncryptedKey   string `json:"encryptedKey"`
}
