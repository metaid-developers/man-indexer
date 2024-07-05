package man

import (
	"fmt"
	"manindexer/adapter"
	"manindexer/adapter/bitcoin"
	"manindexer/adapter/microvisionchain"
	"manindexer/common"
	"manindexer/database"
	"manindexer/database/mongodb"
	"manindexer/database/pebbledb"
	"manindexer/database/postgresql"
	"manindexer/pin"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/IceflowRE/go-multiprogressbar"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/schollz/progressbar/v3"
)

var (
	ChainAdapter   map[string]adapter.Chain
	IndexerAdapter map[string]adapter.Indexer
	DbAdapter      database.Db
	ChainParams    *chaincfg.Params
	//Number          int64    = 0
	MaxHeight       map[string]int64
	CurBlockHeight  map[string]int64
	BaseFilter      []string = []string{"/info", "/file", "/flow", "ft"}
	SyncBaseFilter  map[string]struct{}
	ProtocolsFilter map[string]struct{}
	OptionLimit     []string = []string{"create", "modify", "revoke", "hide"}
	BarMap          map[string]*progressbar.ProgressBar
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
	ChainAdapter = make(map[string]adapter.Chain)
	IndexerAdapter = make(map[string]adapter.Indexer)
	MaxHeight = make(map[string]int64)
	CurBlockHeight = make(map[string]int64)
	ProtocolsFilter = make(map[string]struct{})
	SyncBaseFilter = make(map[string]struct{})
	BarMap = make(map[string]*progressbar.ProgressBar)
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
	chainList := strings.Split(chainType, ",")
	ChainParams = &chaincfg.MainNetParams
	if test == "1" {
		ChainParams = &chaincfg.TestNet3Params
	}
	if test == "2" {
		ChainParams = &chaincfg.RegressionNetParams
	}
	for _, chain := range chainList {
		switch chain {
		case "btc":
			ChainAdapter[chain] = &bitcoin.BitcoinChain{}
			IndexerAdapter[chain] = &bitcoin.Indexer{
				ChainParams: ChainParams,
				PopCutNum:   common.Config.Btc.PopCutNum,
				DbAdapter:   &DbAdapter,
				ChainName:   chain,
			}
			//BarMap[chain] = progressbar.Default(1, "[BTC]")
			bestHeight := ChainAdapter[chain].GetBestHeight()
			common.InitHeightFile("./btc_del_mempool_height.txt", bestHeight)
		case "mvc":
			ChainAdapter[chain] = &microvisionchain.MicroVisionChain{}
			IndexerAdapter[chain] = &microvisionchain.Indexer{
				ChainParams: ChainParams,
				PopCutNum:   common.Config.Mvc.PopCutNum,
				DbAdapter:   &DbAdapter,
				ChainName:   chain,
			}

			//BarMap[chain] = progressbar.Default(1, "[MVC]")
			bestHeight := ChainAdapter["mvc"].GetBestHeight()
			common.InitHeightFile("./mvc_del_mempool_height.txt", bestHeight)
		}
	}

}
func ZmqRun() {
	//zmq
	for chain, indexer := range IndexerAdapter {
		if chain != "btc" {
			continue
		}
		s := make(chan []*pin.PinInscription)
		go indexer.ZmqRun(s)
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
	for k, chain := range ChainAdapter {
		bestHeight := chain.GetBestHeight()
		localLastHeight, err := common.GetLocalLastHeight(fmt.Sprintf("./%s_del_mempool_height.txt", k))
		if err != nil {
			return
		}
		if localLastHeight >= bestHeight {
			return
		}
		for i := localLastHeight; i <= bestHeight; i++ {
			DeleteMempoolData(i, k)
			common.UpdateLocalLastHeight(fmt.Sprintf("./%s_del_mempool_height.txt", k), i)
		}
	}
}
func DeleteMempoolData(bestHeight int64, chainName string) {
	list := IndexerAdapter[chainName].GetBlockTxHash(bestHeight)
	DbAdapter.DeleteMempoolInscription(list)
}
func getSyncHeight(chainName string) (from, to int64) {
	if MaxHeight[chainName] <= 0 {
		var err error
		MaxHeight[chainName], err = DbAdapter.GetMaxHeight(chainName)
		if err != nil {
			return
		}
	}
	bestHeight := ChainAdapter[chainName].GetBestHeight()
	if MaxHeight[chainName] >= bestHeight {
		return
	}
	/*
		if Number <= 0 {
			Number = DbAdapter.GetMaxNumber()
		}
	*/
	initialHeight := ChainAdapter[chainName].GetInitialHeight()
	if MaxHeight[chainName] < initialHeight {
		from = initialHeight
	} else {
		from = MaxHeight[chainName]
	}
	to = bestHeight
	return
}

func IndexerRun() {
	size := int64(0)
	fromMap := make(map[string]int64)
	toMap := make(map[string]int64)
	mpb := multiprogressbar.New()
	mbpIndex := make(map[string]int)
	i := 0
	for chainName := range ChainAdapter {
		from, to := getSyncHeight(chainName)
		if from >= to {
			continue
		}
		if to-from > size {
			size = to - from
		}
		fromMap[chainName] = from
		toMap[chainName] = to
		BarMap[chainName] = progressbar.Default(to-from, "["+chainName+"]")
		//BarMap[chainName].ChangeMax64(to)
		//BarMap[chainName].Set64(from)
		mpb.Add(BarMap[chainName])
		mbpIndex[chainName] = i
		i += 1
	}
	if size > 0 {
		for i := int64(0); i <= size; i++ {
			var nextStr []string
			for chainName := range ChainAdapter {
				if _, ok := fromMap[chainName]; !ok {
					continue
				}
				if toMap[chainName]-fromMap[chainName] < i {
					continue
				}
				next := fromMap[chainName] + i
				timestamp, err := ChainAdapter[chainName].GetBlockTime(next)
				if err != nil {
					return
				}
				nextStr = append(nextStr, fmt.Sprintf("%d_%s_%d", timestamp, chainName, next))
			}
			if len(nextStr) > 0 {
				sort.Strings(nextStr)
				for _, str := range nextStr {
					arr := strings.Split(str, "_")
					nextHeight, _ := strconv.ParseInt(arr[2], 10, 64)
					chainName := arr[1]
					DoIndexerRun(chainName, nextHeight)
					mpb.Get(mbpIndex[chainName]).Add(1)
				}
			}
		}
	}

}
func DoIndexerRun(chainName string, height int64) (err error) {
	//bar := progressbar.Default(to - from)
	//for i := from + 1; i <= to; i++ {
	//bar.Add(1)
	MaxHeight[chainName] = height
	pinList, protocolsData, metaIdData, pinTreeData, updatedData, mrc20List, followData, infoAdditional, _ := GetSaveData(chainName, height)

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
	//Handle MRC20 last.
	if len(mrc20List) > 0 {
		Mrc20Handle(mrc20List)
	}
	//}
	//bar.Finish()
	return
}
func GetSaveData(chainName string, blockHeight int64) (
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
	pins, txInList := IndexerAdapter[chainName].CatchPins(blockHeight)
	//check transfer
	transferCheck, err := DbAdapter.GetPinListByOutPutList(txInList)
	if err == nil && len(transferCheck) > 0 {
		idMap := make(map[string]struct{})
		for _, t := range transferCheck {
			idMap[t.Output] = struct{}{}
		}
		trasferMap := IndexerAdapter[chainName].CatchTransfer(idMap)
		DbAdapter.UpdateTransferPin(trasferMap)
	}

	//pin validator
	mrc20TransferPinTx := make(map[string]struct{})
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
		if len(pinNode.Path) > 10 && pinNode.Path[0:10] == "/ft/mrc20/" {
			mrc20List = append(mrc20List, pinNode)
			if pinNode.Path == "/ft/mrc20/transfer" {
				mrc20TransferPinTx[pinNode.GenesisTransaction] = struct{}{}
			}
		}
	}
	//check mrc20 transfer
	mrc20transferCheck, err := DbAdapter.GetMrc20UtxoByOutPutList(txInList)
	if err == nil && len(mrc20transferCheck) > 0 {
		mrc20TrasferList := IndexerAdapter[chainName].CatchNativeMrc20Transfer(blockHeight, mrc20transferCheck, mrc20TransferPinTx)
		if len(mrc20TrasferList) > 0 {
			DbAdapter.UpdateMrc20Utxo(mrc20TrasferList)
		}
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
			if pinNode.ChainName != "btc" {
				continue
			}
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
			if m.ChainName != "btc" {
				continue
			}
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
	if metaIdInfo.ChainName == "" {
		metaIdInfo.ChainName = pinNode.ChainName
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
