package man

import (
	"fmt"
	"log"
	"manindexer/adapter"
	"manindexer/adapter/bitcoin"
	"manindexer/common"
	"manindexer/database"
	"manindexer/database/mongodb"
	"manindexer/database/pebbledb"
	"manindexer/database/postgresql"
	"manindexer/pin"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/schollz/progressbar/v3"
)

var (
	ChainAdapter   adapter.Chain
	IndexerAdapter adapter.Indexer
	DbAdapter      database.Db
	//Number          int64    = 0
	MaxHeight       int64    = 0
	CurBlockHeight  int64    = 0
	BaseFilter      []string = []string{"/info", "/file", "/flow"}
	SyncBaseFilter  map[string]struct{}
	ProtocolsFilter map[string]struct{}
	OptionLimit     []string = []string{"create", "modify", "revoke"}
)

const (
	StatusBlockHeightLower      = -101
	StatusPinIsTransfered       = -102
	StatusModifyPinIdNotExist   = -201
	StatusModifyPinAddrNotExist = -202
	StatusModifyPinAddrDenied   = -203
	StatusModifyPinIsModifyed   = -204
	StatusModifyPinOptIsInit    = -205
	//Revoke
	StatusRevokePinIdNotExist   = -301
	StatusRevokePinAddrNotExist = -302
	StatusRevokePinAddrDenied   = -303
	StatusRevokePinIsRevoked    = -304
	StatusRevokePinOptIsInit    = -305
)

func InitAdapter(chainType, dbType, test, server string) {

	ProtocolsFilter = make(map[string]struct{})
	SyncBaseFilter = make(map[string]struct{})
	syncConfig := common.Config.Sync
	if len(syncConfig.SyncProtocols) > 0 {
		for _, f := range BaseFilter {
			SyncBaseFilter[f] = struct{}{}
		}
		for _, protocol := range syncConfig.SyncProtocols {
			p := strings.ToLower("/protocols/" + protocol)
			ProtocolsFilter[p] = struct{}{}
		}
	}

	switch dbType {
	case "mongo":
		DbAdapter = &mongodb.Mongodb{}
	case "pg":
		DbAdapter = &postgresql.Postgresql{}
	case "pb":
		DbAdapter = &pebbledb.Pebble{}
	}
	DbAdapter.InitDatabase()
	switch chainType {
	case "btc":
		ChainAdapter = &bitcoin.BitcoinChain{}
		chainParams := &chaincfg.MainNetParams
		if test == "1" {
			chainParams = &chaincfg.TestNet3Params
		}
		if test == "2" {
			chainParams = &chaincfg.RegressionNetParams
		}
		IndexerAdapter = &bitcoin.Indexer{
			ChainParams: chainParams,
			PopCutNum:   common.Config.Btc.PopCutNum,
			DbAdapter:   &DbAdapter,
		}
		// case "mvc":
		// 	ChainAdapter = &mvc.MvcChain{}
		// 	IndexerAdapter = &mvc.Indexer{}
	}
	bestHeight := ChainAdapter.GetBestHeight()
	common.InitHeightFile("./del_mempool_height.txt", bestHeight)
}
func ZmqRun() {
	//zmq
	s := make(chan []*pin.PinInscription)
	go IndexerAdapter.ZmqRun(s)
	//go IndexerAdapter.ZmqHashblock()
	for x := range s {
		for _, pinNode := range x {
			if !pinNode.IsTransfered {
				handleMempoolPin(pinNode)
			} else if pinNode.IsTransfered {
				handleMempoolTransferPin(pinNode)
			}
		}
	}
}
func handleMempoolPin(pinNode *pin.PinInscription) {
	if pinNode.Operation == "modify" || pinNode.Operation == "revoke" {
		pinNode.OriginalId = strings.Replace(pinNode.Path, "@", "", -1)
		originalPins, err := DbAdapter.GetPinListByIdList([]string{pinNode.OriginalId})
		if err == nil && len(originalPins) > 0 {
			pinNode.OriginalPath = originalPins[0].OriginalPath
		}
	}
	pinNode.Timestamp = time.Now().Unix()
	pinNode.Number = -1
	pinNode.ContentTypeDetect = common.DetectContentType(&pinNode.ContentBody)
	if len(ProtocolsFilter) > 0 && pinNode.Path != "" {
		p := strings.ToLower(pinNode.Path)
		if _, protCheck := ProtocolsFilter[p]; protCheck {
			DbAdapter.BatchAddProtocolData([]*pin.PinInscription{pinNode})
		}
	}
	DbAdapter.AddMempoolPin(pinNode)
}
func handleMempoolTransferPin(pinNode *pin.PinInscription) {
	transferPin := pin.MemPoolTrasferPin{
		PinId:       pinNode.Id,
		FromAddress: pinNode.CreateAddress,
		ToAddress:   pinNode.Address,
		InTime:      pinNode.Timestamp,
		TxHash:      pinNode.GenesisTransaction,
	}
	DbAdapter.AddMempoolTransfer(&transferPin)
}
func CheckNewBlock() {
	bestHeight := ChainAdapter.GetBestHeight()
	localLastHeight, err := common.GetLocalLastHeight("./del_mempool_height.txt")
	if err != nil {
		return
	}
	if localLastHeight >= bestHeight {
		return
	}
	for i := localLastHeight; i <= bestHeight; i++ {
		DeleteMempoolData(i)
		common.UpdateLocalLastHeight("./del_mempool_height.txt", i)
	}
}
func DeleteMempoolData(bestHeight int64) {
	list := IndexerAdapter.GetBlockTxHash(bestHeight)
	DbAdapter.DeleteMempoolInscription(list)
}
func getSyncHeight() (from, to int64) {
	if MaxHeight <= 0 {
		var err error
		MaxHeight, err = DbAdapter.GetMaxHeight()
		if err != nil {
			return
		}
	}
	bestHeight := ChainAdapter.GetBestHeight()
	if MaxHeight >= bestHeight {
		return
	}
	/*
		if Number <= 0 {
			Number = DbAdapter.GetMaxNumber()
		}
	*/
	initialHeight := ChainAdapter.GetInitialHeight()
	if MaxHeight < initialHeight {
		from = initialHeight
	} else {
		from = MaxHeight
	}
	to = bestHeight
	return
}
func IndexerRun(from, to int64) (err error) {
	if from == 0 && to == 0 {
		from, to = getSyncHeight()
		if from >= to {
			return
		}
	}
	log.Println("from:", from, ",to:", to)
	bar := progressbar.Default(to - from)
	for i := from + 1; i <= to; i++ {
		bar.Add(1)
		MaxHeight = i
		pinList, protocolsData, metaIdData, pinTreeData, updatedData, _, followData, infoAdditional, _ := GetSaveData(i)
		if len(metaIdData) > 0 {
			err = DbAdapter.BatchUpsertMetaIdInfo(metaIdData)
			//metaIdData = metaIdData[0:0]
		}
		if len(pinList) > 0 {
			DbAdapter.BatchAddPins(pinList)
		}

		if len(pinTreeData) > 0 {
			DbAdapter.BatchAddPinTree(pinTreeData)
		}
		if len(protocolsData) > 0 {
			DbAdapter.BatchAddProtocolData(protocolsData)
		}
		if len(updatedData) > 0 {
			DbAdapter.BatchUpdatePins(updatedData)
		}
		if len(followData) > 0 {
			DbAdapter.BatchUpsertFollowData(followData)
		}
		if len(infoAdditional) > 0 {
			DbAdapter.BatchUpsertMetaIdInfoAddition(infoAdditional)
		}
	}
	bar.Finish()
	return
}
func GetSaveData(blockHeight int64) (
	pinList []interface{},
	protocolsData []*pin.PinInscription,
	metaIdData map[string]*pin.MetaIdInfo,
	pinTreeData []interface{},
	updatedData []*pin.PinInscription,
	mrc20List []*pin.PinInscription,
	followData []*pin.FollowData,
	infoAdditional []*pin.MetaIdInfoAdditional,
	err error) {
	metaIdData = make(map[string]*pin.MetaIdInfo)
	pins, txInList := IndexerAdapter.CatchPins(blockHeight)
	//check transfer
	transferCheck, err := DbAdapter.GetPinListByOutPutList(txInList)
	if err == nil && len(transferCheck) > 0 {
		idMap := make(map[string]struct{})
		for _, t := range transferCheck {
			idMap[t.Output] = struct{}{}
		}
		trasferMap := IndexerAdapter.CatchTransfer(idMap)
		DbAdapter.UpdateTransferPin(trasferMap)
	}
	for _, pinNode := range pins {
		err := validator(pinNode)
		if err != nil {
			continue
		}
		//save all data or protocols data
		s := handleProtocolsData(pinNode)
		if s == -1 {
			continue
		} else if s == 1 {
			protocolsData = append(protocolsData, pinNode)
		}
		pinList = append(pinList, pinNode)
		//mrc20 pin
		// if common.Config.Mrc20 == 1 && len(pinNode.Path) > 7 && pinNode.Path[0:7] == "/mrc20/" {
		// 	mrc20List = append(mrc20List, pinNode)
		// }
	}

	handlePathAndOperation(&pinList, &metaIdData, &pinTreeData, &updatedData, &followData, &infoAdditional)
	createPinNumber(&pinList)
	createMetaIdNumber(metaIdData)
	return
}

func handleProtocolsData(pinNode *pin.PinInscription) int {
	if len(ProtocolsFilter) > 0 && pinNode.Path != "" {
		p := strings.ToLower(pinNode.Path)
		_, baseCheck := SyncBaseFilter[p]
		_, protCheck := ProtocolsFilter[p]
		if !common.Config.Sync.SyncAllData && !protCheck && !baseCheck {
			return -1 //save nothing
		} else if protCheck {
			//add to protocols data
			return 1
		}
	}
	return 0
}
func createPinNumber(pinList *[]interface{}) {
	if len(*pinList) > 0 {
		maxNumber := DbAdapter.GetMaxNumber()
		for _, p := range *pinList {
			pinNode := p.(*pin.PinInscription)
			pinNode.Number = maxNumber
			maxNumber += 1
			if pinNode.MetaId == "" {
				pinNode.MetaId = common.GetMetaIdByAddress(pinNode.Address)
			}
		}
	}
}
func createMetaIdNumber(metaIdData map[string]*pin.MetaIdInfo) {
	if len(metaIdData) > 0 {
		maxMetaIdNumber := DbAdapter.GetMaxMetaIdNumber()
		for _, m := range metaIdData {
			if m.Number == 0 {
				m.Number = maxMetaIdNumber
				maxMetaIdNumber += 1
			}
		}
	}
}
func handlePathAndOperation(
	pinList *[]interface{},
	metaIdData *map[string]*pin.MetaIdInfo,
	pinTreeData *[]interface{},
	updatedData *[]*pin.PinInscription,
	followData *[]*pin.FollowData,
	infoAdditional *[]*pin.MetaIdInfoAdditional) {
	var modifyPinIdList []string
	newPinMap := make(map[string]*pin.PinInscription)
	for _, p := range *pinList {

		pinNode := p.(*pin.PinInscription)
		if pinNode.MetaId == "" {
			pinNode.MetaId = common.GetMetaIdByAddress(pinNode.Address)
		}
		metaIdInfoParse(pinNode, "", metaIdData)
		switch pinNode.Operation {
		case "modify":
			updatePin := *pinNode
			updatePin.Status = 1
			updatePin.OriginalId = strings.Replace(pinNode.Path, "@", "", -1)
			modifyPinIdList = append(modifyPinIdList, updatePin.OriginalId)
			pinNode.OriginalId = updatePin.OriginalId
			newPinMap[updatePin.Id] = &updatePin
		case "revoke":
			updatePin := *pinNode
			updatePin.Status = -1
			updatePin.OriginalId = strings.Replace(pinNode.Path, "@", "", -1)
			modifyPinIdList = append(modifyPinIdList, updatePin.OriginalId)
			pinNode.OriginalId = updatePin.OriginalId
			newPinMap[updatePin.Id] = &updatePin
		}

		path := pinNode.Path
		// if len(path) > 5 && path[0:5] == "/info" {
		// 	metaIdInfo := metaIdInfoParse(pinNode, "")
		// 	*metaIdData = append(*metaIdData, metaIdInfo)
		// }
		pathArray := strings.Split(path, "/")
		if len(pathArray) > 1 && path != "/" {
			path = strings.Join(pathArray[0:len(pathArray)-1], "/")
		}
		pinTree := pin.PinTreeCatalog{RootTxId: common.GetMetaIdByAddress(pinNode.Address), TreePath: path}
		*pinTreeData = append(*pinTreeData, pinTree)
		//follow
		if pinNode.Path == "/follow" {
			*followData = append(*followData, creatFollowData(pinNode, true))
		}
		//infoAdditional
		additional := createInfoAdditional(pinNode, pinNode.Path)
		if additional != (pin.MetaIdInfoAdditional{}) {
			*infoAdditional = append(*infoAdditional, &additional)
		}
	}
	if len(modifyPinIdList) <= 0 {
		return
	}
	originalPins, err := DbAdapter.GetPinListByIdList(modifyPinIdList)
	if err != nil {
		return
	}
	originalPinMap := make(map[string]*pin.PinInscription)
	for _, mp := range originalPins {
		originalPinMap[mp.Id] = mp
	}
	statusMap := getModifyPinStatus(newPinMap, originalPinMap)
	for _, p := range *pinList {
		pinNode := p.(*pin.PinInscription)
		if pinNode.OriginalId == "" {
			pinNode.OriginalId = pinNode.Id
		}
		if pinNode.Operation == "modify" || pinNode.Operation == "revoke" {
			if v, ok := statusMap[pinNode.Id]; ok {
				pinNode.Status = v
			}
			if pinNode.Status >= 0 {
				*updatedData = append(*updatedData, newPinMap[pinNode.Id])
			}
			_, check := originalPinMap[pinNode.OriginalId]
			if check {
				pinNode.OriginalPath = originalPinMap[pinNode.OriginalId].OriginalPath
			}
			if pinNode.Operation == "modify" && pinNode.Status >= 0 && check {
				if len(originalPinMap[pinNode.OriginalId].OriginalPath) > 5 && originalPinMap[pinNode.OriginalId].OriginalPath[0:5] == "/info" {
					metaIdInfoParse(pinNode, originalPinMap[pinNode.OriginalId].OriginalPath, metaIdData)
				}
			}
			//unfollow
			if pinNode.Operation == "revoke" && pinNode.OriginalPath == "/follow" {
				*followData = append(*followData, creatFollowData(pinNode, false))
			}
			//infoAdditional
			if pinNode.Operation == "modify" {
				additional := createInfoAdditional(pinNode, pinNode.OriginalPath)
				if additional != (pin.MetaIdInfoAdditional{}) {
					*infoAdditional = append(*infoAdditional, &additional)
				}
			}

		} else {
			metaIdInfoParse(pinNode, "", metaIdData)
		}
	}
}
func createInfoAdditional(pinNode *pin.PinInscription, path string) (addition pin.MetaIdInfoAdditional) {
	if len(path) > 7 && path[0:6] == "/info/" {
		infoPathArr := strings.Split(path, "/")
		if len(infoPathArr) < 3 || infoPathArr[2] == "name" || infoPathArr[2] == "avatar" || infoPathArr[2] == "bio" {
			return
		}
		addition = pin.MetaIdInfoAdditional{
			MetaId:    pinNode.MetaId,
			InfoKey:   infoPathArr[2],
			InfoValue: string(pinNode.ContentBody),
			PinId:     pinNode.Id,
		}
	}
	return
}
func creatFollowData(pinNode *pin.PinInscription, follow bool) (followData *pin.FollowData) {
	if pinNode.MetaId == "" {
		pinNode.MetaId = common.GetMetaIdByAddress(pinNode.Address)
	}
	followData = &pin.FollowData{}
	if follow {
		followData.MetaId = string(pinNode.ContentBody)
		followData.FollowMetaId = pinNode.MetaId
		followData.FollowPinId = pinNode.Id
		followData.FollowTime = pinNode.Timestamp
		followData.Status = true
	} else {
		followData.FollowPinId = strings.Replace(pinNode.Path, "@", "", -1)
		followData.UnFollowPinId = pinNode.Id
		followData.Status = false
	}
	return
}
func getModifyPinStatus(curPinMap map[string]*pin.PinInscription, originalPinMap map[string]*pin.PinInscription) (statusMap map[string]int) {
	statusMap = make(map[string]int)
	for cid, np := range curPinMap {
		id := np.OriginalId
		if np.Operation == "modify" {
			if _, ok := originalPinMap[id]; !ok {
				statusMap[cid] = StatusModifyPinIdNotExist
				continue
			}
			if np.Address != originalPinMap[id].Address {
				statusMap[cid] = StatusModifyPinAddrDenied
				continue
			}
			if originalPinMap[id].Status == 1 {
				statusMap[cid] = StatusModifyPinIsModifyed
				continue
			}
			if originalPinMap[id].Operation == "init" {
				statusMap[cid] = StatusModifyPinOptIsInit
				continue
			}
		} else if np.Operation == "revoke" {
			if _, ok := originalPinMap[id]; !ok {
				statusMap[cid] = StatusRevokePinIdNotExist
				continue
			}
			if np.Address != originalPinMap[id].Address {
				statusMap[cid] = StatusRevokePinAddrDenied
				continue
			}
			if originalPinMap[id].Status == -1 {
				statusMap[cid] = StatusRevokePinIsRevoked
				continue
			}
			if originalPinMap[id].Operation == "init" {
				statusMap[cid] = StatusRevokePinOptIsInit
				continue
			}
			if len(originalPinMap[id].Path) > 5 && originalPinMap[id].Path[0:5] == "/info" {
				statusMap[cid] = StatusRevokePinOptIsInit
				continue
			}
		}
		if np.GenesisHeight <= originalPinMap[id].GenesisHeight {
			statusMap[cid] = StatusBlockHeightLower
			continue
		} else if originalPinMap[id].IsTransfered {
			statusMap[cid] = StatusPinIsTransfered
			continue
		}
	}
	return
}

func metaIdInfoParse(pinNode *pin.PinInscription, path string, metaIdData *map[string]*pin.MetaIdInfo) {
	var metaIdInfo *pin.MetaIdInfo
	var ok bool
	var err error
	metaIdInfo, ok = (*metaIdData)[pinNode.Address]
	if !ok {
		metaIdInfo, _, err = DbAdapter.GetMetaIdInfo(pinNode.Address, false, "")
		if err != nil {
			return
		}
	}
	if metaIdInfo == nil {
		metaIdInfo = &pin.MetaIdInfo{MetaId: common.GetMetaIdByAddress(pinNode.Address), Address: pinNode.Address}
	}
	if path == "" {
		path = pinNode.Path
	}

	if metaIdInfo.MetaId == "" {
		metaIdInfo.MetaId = pinNode.Id
	}
	switch path {
	case "/info/name":
		metaIdInfo.Name = string(pinNode.ContentBody)
		metaIdInfo.NameId = pinNode.Id
	case "/info/avatar":
		metaIdInfo.Avatar = fmt.Sprintf("/content/%s", pinNode.Id)
		metaIdInfo.AvatarId = pinNode.Id
	case "/info/bio":
		metaIdInfo.Bio = string(pinNode.ContentBody)
		metaIdInfo.BioId = pinNode.Id
	}
	(*metaIdData)[pinNode.Address] = metaIdInfo
}
