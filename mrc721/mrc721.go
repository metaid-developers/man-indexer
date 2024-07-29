package mrc721

const (
	ErrOperation           = "operation is error"
	ErrPinContent          = "pin content is error"
	ErrTotalSupply         = "totalSupply  must be between -1 and 1e12"
	ErrRoyaltyRate         = "royaltyRate  must be between 0 and 20"
	ErrCollectionExist     = "collectionName already exists"
	ErrCollectionNotExist  = "collectionName not exists"
	ErrTotalSupplyEexceeds = "exceeds total supply"
)

type Mrc721CollectionDesc struct {
	Name        string `json:"name"`
	TotalSupply int64  `json:"totalsupply"`
	RoyaltyRate int    `json:"royaltyrate"`
	Desc        string `json:"desc"`
	Website     string `json:"website"`
	Cover       string `json:"cover"`
	Metadata    string `json:"metadata"`
}

type Mrc721CollectionDescPin struct {
	CollectionName string `json:"collectionname"`
	Name           string `json:"name"`
	TotalSupply    int64  `json:"totalsupply"`
	RoyaltyRate    int    `json:"royaltyrate"`
	Desc           string `json:"desc"`
	Website        string `json:"website"`
	Cover          string `json:"cover"`
	Metadata       string `json:"metadata"`
	PinId          string `json:"pinid"`
	Address        string `json:"address"`
	MetaId         string `json:"metaid"`
	CreateTime     int64  `json:"createtime"`
	TotalNum       int64  `json:"totalnum"`
}

type Mrc721ItemDescList struct {
	Items []Mrc721ItemDesc `json:"items"`
}

type Mrc721ItemDesc struct {
	PinId    string `json:"pinid"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Cover    string `json:"cover"`
	Metadata string `json:"metadata"`
}

type Mrc721ItemDescPin struct {
	CollectionPinId string `json:"collectionPinId"`
	CollectionName  string `json:"collectionName"`
	ItemPinId       string `json:"itemPinId"`
	DescPinId       string `json:"descPinId"`
	Name            string `json:"name"`
	Desc            string `json:"desc"`
	Cover           string `json:"cover"`
	Metadata        string `json:"metaData"`
	CreateTime      int64  `json:"createTime"`
	Address         string `json:"address"`
	Content         []byte `json:"content"`
	MetaId          string `json:"metaId"`
	DescAdded       bool   `json:"descadded"`
}
