package pin

const (
	ProtocolID string = "746573746964" //testid(HEX16)
	//ProtocolID    string = "6d6574616964" //metaid
	CompliantPath string = "info;file;protocols;nft;ft;mrc20;follow"
)

type PinInscription struct {
	Id                 string `json:"id"`
	Number             int64  `json:"number"`
	MetaId             string `json:"metaid"`
	Address            string `json:"address"`
	CreateAddress      string `json:"creator"`
	CreateMetaId       string `json:"createMetaId"`
	InitialOwner       string `json:"initialOwner"`
	Output             string `json:"output"`
	OutputValue        int64  `json:"outputValue"`
	Timestamp          int64  `json:"timestamp"`
	GenesisFee         int64  `json:"genesisFee"`
	GenesisHeight      int64  `json:"genesisHeight"`
	GenesisTransaction string `json:"genesisTransaction"`
	TxIndex            int    `json:"txIndex"`
	TxInIndex          uint32 `json:"txInIndex"`
	Offset             uint64 `json:"offset"`
	Location           string `json:"location"`
	Operation          string `json:"operation"`
	Path               string `json:"path"`
	ParentPath         string `json:"parentPath"`
	OriginalPath       string `json:"originalPath"`
	Encryption         string `json:"encryption"`
	Version            string `json:"version"`
	ContentType        string `json:"contentType"`
	ContentTypeDetect  string `json:"contentTypeDetect"`
	ContentBody        []byte `json:"contentBody"`
	ContentLength      uint64 `json:"contentLength"`
	ContentSummary     string `json:"contentSummary"`
	Status             int    `json:"status"`
	OriginalId         string `json:"originalId"`
	IsTransfered       bool   `json:"isTransfered"`
	Preview            string `json:"preview"`
	Content            string `json:"content"`
	Pop                string `json:"pop"`
	PopLv              int    `json:"popLv"`
	ChainName          string `json:"chainName"`
	DataValue          int    `json:"dataValue"`
	//Mrc20Minted        bool   `json:"mrc20Minted"`  //true Consumed
	//Mrc20MintPin       string `json:"mrc20MintPin"` //mrc20 mint pin id
	Mrc20MintId []string `json:"mrc20MintId"`
}
type PinTransferInfo struct {
	Address     string `json:"address"`
	Output      string `json:"output"`
	OutputValue int64  `json:"outputValue"`
	Offset      uint64 `json:"offset"`
	Location    string `json:"location"`
}
type PersonalInformationNode struct {
	Operation     string `json:"operation"`
	Path          string `json:"path"`
	Encryption    string `json:"encryption"`
	Version       string `json:"cersion"`
	ContentType   string `json:"contentType"`
	ContentBody   []byte `json:"contentBody"`
	ContentLength uint64 `json:"contentLength"`
	ParentPath    string `json:"parentPath"`
	Protocols     bool   `json:"protocols"`
}
type FollowData struct {
	MetaId        string `json:"metaId"`
	FollowMetaId  string `json:"followMetaId"`
	FollowTime    int64  `json:"followTime"`
	FollowPinId   string `json:"followPinId"`
	UnFollowPinId string `json:"unFollowPinId"`
	Status        bool   `json:"status"`
}
type MetaIdInfo struct {
	ChainName     string `json:"chainName"`
	Number        int64  `json:"number"`
	MetaId        string `json:"metaid"`
	Name          string `json:"name"`
	NameId        string `json:"nameId"`
	Address       string `json:"address"`
	Avatar        string `json:"avatar"`
	AvatarId      string `json:"avatarId"`
	Bio           string `json:"bio"`
	BioId         string `json:"bioId"`
	SoulbondToken string `json:"soulbondToken"`
	IsInit        bool   `json:"isInit"`
	FollowCount   int64  `json:"followCount"`
	Pdv           int64  `json:"pdv"`
	Fdv           int64  `json:"fdv"`
}
type MetaIdDataValue struct {
	ChainName string `json:"chainName"`
	MetaId    string `json:"metaid"`
	Address   string `json:"address"`
	Pdv       int64  `json:"pdv"`
	Fdv       int64  `json:"fdv"`
}
type MetaIdInfoAdditional struct {
	MetaId    string `json:"metaId"`
	InfoKey   string `json:"infoKey"`
	InfoValue string `json:"infoValue"`
	PinId     string `json:"pinId"`
}
type PinTreeCatalog struct {
	RootTxId string `json:"rootTxId"`
	TreePath string `json:"treePath"`
}

type PinMsg struct {
	Content       string `json:"content"`
	Number        int64  `json:"number"`
	Operation     string `json:"operation"`
	Height        int64  `json:"height"`
	Id            string `json:"id"`
	Type          string `json:"type"`
	Path          string `json:"path"`
	MetaId        string `json:"metaid"`
	Pop           string `json:"pop"`
	ChainName     string `json:"chainName"`
	Address       string `json:"address"`
	CreateAddress string `json:"creator"`
	InitialOwner  string `json:"initialOwner"`
}

type BlockMsg struct {
	BlockHash      string   `json:"blockHash"`
	Target         string   `json:"target"`
	Timestamp      string   `json:"timestamp"`
	Size           int64    `json:"size"`
	Weight         int64    `json:"weight"`
	TransactionNum int      `json:"transactionNum"`
	Transaction    []string `json:"transaction"`
}
type PinCount struct {
	Block  int64 `json:"block"`
	Pin    int64 `json:"Pin"`
	MetaId int64 `json:"metaId"`
	App    int64 `json:"app"`
}
type MemPoolTrasferPin struct {
	PinId       string `json:"pinId"`
	FromAddress string `json:"fromAddress"`
	ToAddress   string `json:"toAddress"`
	InTime      int64  `json:"inTime"`
	TxHash      string `json:"txHash"`
}
