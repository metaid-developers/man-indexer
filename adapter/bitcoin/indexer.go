package bitcoin

import (
	"encoding/hex"
	"errors"
	"fmt"
	"manindexer/common"
	"manindexer/database"
	"manindexer/mrc20"
	"manindexer/pin"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var PopCutNum int = 0

type Indexer struct {
	ChainParams *chaincfg.Params
	Block       interface{}
	PopCutNum   int
	DbAdapter   *database.Db
	ChainName   string
}

func init() {
	PopCutNum = common.Config.Btc.PopCutNum
}
func (indexer *Indexer) GetCurHeight() (height int64) {
	return
}
func (indexer *Indexer) GetAddress(pkScript []byte) (address string) {
	_, addresses, _, _ := txscript.ExtractPkScriptAddrs(pkScript, indexer.ChainParams)
	if len(addresses) > 0 {
		address = addresses[0].String()
	}
	return
}
func (indexer *Indexer) CatchPins(blockHeight int64) (pinInscriptions []*pin.PinInscription, txInList []string) {
	chain := BitcoinChain{}
	blockMsg, err := chain.GetBlock(blockHeight)
	if err != nil {
		return
	}
	indexer.Block = blockMsg
	block := blockMsg.(*wire.MsgBlock)

	timestamp := block.Header.Timestamp.Unix()
	blockHash := block.BlockHash().String()
	merkleRoot := block.Header.MerkleRoot.String()
	for i, tx := range block.Transactions {
		for _, in := range tx.TxIn {
			id := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
			txInList = append(txInList, id)
		}
		if !tx.HasWitness() {
			continue
		}
		txPins := indexer.CatchPinsByTx(tx, blockHeight, timestamp, blockHash, merkleRoot, i)
		if len(txPins) > 0 {
			pinInscriptions = append(pinInscriptions, txPins...)
		}
	}
	return
}

func (indexer *Indexer) CatchMempoolPins(txList []interface{}) (pinInscriptions []*pin.PinInscription, txInList []string) {
	timestamp := time.Now().Unix()
	blockHash := "none"
	merkleRoot := "none"
	for i, item := range txList {
		tx := item.(*wire.MsgTx)
		for _, in := range tx.TxIn {
			id := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
			txInList = append(txInList, id)
		}
		if !tx.HasWitness() {
			continue
		}
		txPins := indexer.CatchPinsByTx(tx, -1, timestamp, blockHash, merkleRoot, i)
		if len(txPins) > 0 {
			pinInscriptions = append(pinInscriptions, txPins...)
		}
	}
	return
}

func (indexer *Indexer) CatchTransfer(idMap map[string]struct{}) (trasferMap map[string]*pin.PinTransferInfo) {
	trasferMap = make(map[string]*pin.PinTransferInfo)
	block := indexer.Block.(*wire.MsgBlock)
	for _, tx := range block.Transactions {
		for _, in := range tx.TxIn {
			id := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
			if _, ok := idMap[id]; ok {
				info, err := indexer.GetOWnerAddress(id, tx)
				if err == nil && info != nil {
					trasferMap[id] = info
				}
			}
		}
	}
	return
}
func (indexer *Indexer) CatchNativeMrc20Transfer(blockHeight int64, utxoList []*mrc20.Mrc20Utxo, mrc20TransferPinTx map[string]struct{}) (savelist []*mrc20.Mrc20Utxo) {
	pointMap := make(map[string][]*mrc20.Mrc20Utxo)
	//keyMap := make(map[string]*mrc20.Mrc20Utxo) //key point-tickid
	for _, u := range utxoList {
		if u.MrcOption == "deploy" {
			continue
		}
		pointMap[u.TxPoint] = append(pointMap[u.TxPoint], u)
	}
	block := indexer.Block.(*wire.MsgBlock)
	t := block.Header.Timestamp.Unix()
	for _, tx := range block.Transactions {
		//if have data transfer
		_, ok := mrc20TransferPinTx[tx.TxHash().String()]
		if ok {
			continue
		}
		list := indexer.createMrc20NativeTransfer(tx, blockHeight, t, pointMap)
		if len(list) > 0 {
			savelist = append(savelist, list...)
		}
	}
	// for _, u := range keyMap {
	// 	savelist = append(savelist, u)
	// }
	return
}
func (indexer *Indexer) CatchMempoolNativeMrc20Transfer(txList []interface{}, utxoList []*mrc20.Mrc20Utxo, mrc20TransferPinTx map[string]struct{}) (savelist []*mrc20.Mrc20Utxo) {
	pointMap := make(map[string][]*mrc20.Mrc20Utxo)
	//keyMap := make(map[string]*mrc20.Mrc20Utxo) //key point-tickid
	for _, u := range utxoList {
		if u.MrcOption == "deploy" {
			continue
		}
		pointMap[u.TxPoint] = append(pointMap[u.TxPoint], u)
	}
	t := time.Now().Unix()
	for _, item := range txList {
		tx := item.(*wire.MsgTx)
		//if have data transfer
		_, ok := mrc20TransferPinTx[tx.TxHash().String()]
		if ok {
			continue
		}
		list := indexer.createMrc20NativeTransfer(tx, -1, t, pointMap)
		if len(list) > 0 {
			savelist = append(savelist, list...)
		}
	}
	// for _, u := range keyMap {
	// 	savelist = append(savelist, u)
	// }
	return
}
func (indexer *Indexer) createMrc20NativeTransfer(tx *wire.MsgTx, blockHeight int64, blockTime int64, pointMap map[string][]*mrc20.Mrc20Utxo) (mrc20Utxolist []*mrc20.Mrc20Utxo) {
	keyMap := make(map[string]*mrc20.Mrc20Utxo)
	for _, in := range tx.TxIn {
		id := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
		if v, ok := pointMap[id]; ok {
			for _, utxo := range v {
				send := *utxo
				send.Status = -1
				mrc20Utxolist = append(mrc20Utxolist, &send)
				key := fmt.Sprintf("%s-%s", send.Mrc20Id, send.TxPoint)
				_, find := keyMap[key]
				if find {
					//keyMap[key].AmtChange += send.AmtChange
					keyMap[key].AmtChange = keyMap[key].AmtChange.Add(send.AmtChange)
				} else {
					recive := *utxo
					recive.MrcOption = "native-transfer"
					recive.FromAddress = recive.ToAddress
					recive.ToAddress = indexer.GetAddress(tx.TxOut[0].PkScript)
					recive.BlockHeight = blockHeight
					recive.TxPoint = fmt.Sprintf("%s:%d", tx.TxHash().String(), 0)
					recive.Timestamp = blockTime
					recive.Chain = "btc"
					keyMap[key] = &recive
				}
			}
		}
	}
	for _, u := range keyMap {
		mrc20Utxolist = append(mrc20Utxolist, u)
	}
	return
}

func (indexer *Indexer) GetOWnerAddress(inputId string, tx *wire.MsgTx) (info *pin.PinTransferInfo, err error) {
	//fmt.Println("tx:", tx.TxHash().String(), inputId)
	info = &pin.PinTransferInfo{}
	firstInputId := fmt.Sprintf("%s:%d", tx.TxIn[0].PreviousOutPoint.Hash, tx.TxIn[0].PreviousOutPoint.Index)
	if len(tx.TxIn) == 1 || firstInputId == inputId {
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(tx.TxOut[0].PkScript, indexer.ChainParams)
		if len(addresses) > 0 {
			info.Address = addresses[0].String()
		} else if class.String() == "nulldata" {
			info.Address = hex.EncodeToString(tx.TxOut[0].PkScript)
		}
		info.Location = fmt.Sprintf("%s:%d:%d", tx.TxHash().String(), 0, 0)
		info.Offset = 0
		info.Output = fmt.Sprintf("%s:%d", tx.TxHash().String(), 0)
		info.OutputValue = tx.TxOut[0].Value
		return
	}
	totalOutputValue := int64(0)
	for _, out := range tx.TxOut {
		totalOutputValue += out.Value
	}
	inputValue := int64(0)
	for _, in := range tx.TxIn {
		id := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash, in.PreviousOutPoint.Index)
		if id == inputId {
			break
		}
		value, err1 := GetValueByTx(in.PreviousOutPoint.Hash.String(), int(in.PreviousOutPoint.Index))
		if err1 != nil {
			err = errors.New("get value error")
			return
		}
		inputValue += value
		if inputValue > totalOutputValue {
			return
		}
	}
	outputValue := int64(0)
	for i, out := range tx.TxOut {
		outputValue += out.Value
		if outputValue > inputValue {
			class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, indexer.ChainParams)
			if len(addresses) > 0 {
				info.Address = addresses[0].String()
			} else if class.String() == "nulldata" {
				info.Address = hex.EncodeToString(out.PkScript)
			}
			info.Output = fmt.Sprintf("%s:%d", tx.TxHash().String(), i)
			//count offset
			info.Location = fmt.Sprintf("%s:%d", info.Output, out.Value-(outputValue-inputValue))
			info.Offset = uint64(i)
			info.OutputValue = out.Value
			break
		}
	}

	return
}
func (indexer *Indexer) CatchPinsByTx(msgTx *wire.MsgTx, blockHeight int64, timestamp int64, blockHash string, merkleRoot string, txIndex int) (pinInscriptions []*pin.PinInscription) {
	//No witness data
	// if !msgTx.HasWitness() {
	// 	return nil
	// }
	chain := BitcoinChain{}
	haveOpReturn := false
	for i, out := range msgTx.TxOut {
		class, _, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, indexer.ChainParams)
		//fmt.Println(class.String())
		if class.String() == "nonstandard" {
			pinInscription := indexer.ParseOpReturnPin(out.PkScript)
			if pinInscription == nil {
				continue
			}
			address, outIdx, locationIdx := indexer.GetOpReturnPinOwner(msgTx)
			if address == "" {
				continue
			}
			txHash := msgTx.TxHash().String()
			id := fmt.Sprintf("%si%d", txHash, outIdx)
			metaId := common.GetMetaIdByAddress(address)
			contentTypeDetect := common.DetectContentType(&pinInscription.ContentBody)
			pop := ""
			if merkleRoot != "" && blockHash != "" {
				pop, _ = common.GenPop(id, merkleRoot, blockHash)
			}
			popLv, _ := pin.PopLevelCount(indexer.ChainName, pop)
			creator := chain.GetCreatorAddress(msgTx.TxIn[0].PreviousOutPoint.Hash.String(), msgTx.TxIn[0].PreviousOutPoint.Index, indexer.ChainParams)
			pinInscriptions = append(pinInscriptions, &pin.PinInscription{
				//Pin:                pinInscription,
				ChainName:          indexer.ChainName,
				Id:                 id,
				MetaId:             metaId,
				Number:             0,
				Address:            address,
				InitialOwner:       address,
				CreateAddress:      creator,
				CreateMetaId:       common.GetMetaIdByAddress(creator),
				Timestamp:          timestamp,
				GenesisHeight:      blockHeight,
				GenesisTransaction: txHash,
				Output:             fmt.Sprintf("%s:%d", txHash, outIdx),
				OutputValue:        msgTx.TxOut[outIdx].Value,
				TxInIndex:          uint32(i - 1),
				Offset:             uint64(outIdx),
				TxIndex:            txIndex,
				Operation:          pinInscription.Operation,
				Location:           fmt.Sprintf("%s:%d:%d", txHash, outIdx, locationIdx),
				Path:               pinInscription.Path,
				OriginalPath:       pinInscription.Path,
				ParentPath:         pinInscription.ParentPath,
				Encryption:         pinInscription.Encryption,
				Version:            pinInscription.Version,
				ContentType:        pinInscription.ContentType,
				ContentTypeDetect:  contentTypeDetect,
				ContentBody:        pinInscription.ContentBody,
				ContentLength:      pinInscription.ContentLength,
				ContentSummary:     getContentSummary(pinInscription, id, contentTypeDetect),
				Pop:                pop,
				PopLv:              popLv,
				DataValue:          pin.RarityScoreBinary(indexer.ChainName, pop),
				Mrc20MintId:        []string{},
			})
			haveOpReturn = true
			break
		}
	}
	if haveOpReturn {
		return
	}
	for i, v := range msgTx.TxIn {
		index, input := i, v
		//Witness length error
		if len(input.Witness) <= 1 {
			continue
		}
		//Witness length error,Taproot
		if len(input.Witness) == 2 && input.Witness[len(input.Witness)-1][0] == txscript.TaprootAnnexTag {
			continue
		}
		// If Taproot Annex data exists, take the last element of the witness as the script data, otherwise,
		// take the penultimate element of the witness as the script data
		var witnessScript []byte
		if input.Witness[len(input.Witness)-1][0] == txscript.TaprootAnnexTag {
			witnessScript = input.Witness[len(input.Witness)-1]
		} else {
			witnessScript = input.Witness[len(input.Witness)-2]
		}
		// Parse script and get pin content
		pinInscription := indexer.ParsePin(witnessScript)
		if pinInscription == nil {
			continue
		}
		address, outIdx, locationIdx := indexer.GetPinOwner(msgTx, index)
		id := fmt.Sprintf("%si%d", msgTx.TxHash().String(), outIdx)
		metaId := common.GetMetaIdByAddress(address)
		contentTypeDetect := common.DetectContentType(&pinInscription.ContentBody)
		pop := ""
		if merkleRoot != "" && blockHash != "" {
			if merkleRoot == "none" && blockHash == "none" {
				pop = "none"
			} else {
				pop, _ = common.GenPop(id, merkleRoot, blockHash)
			}

		}
		popLv, _ := pin.PopLevelCount(indexer.ChainName, pop)
		creator := chain.GetCreatorAddress(v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index, indexer.ChainParams)
		pinInscriptions = append(pinInscriptions, &pin.PinInscription{
			//Pin:                pinInscription,
			ChainName:          indexer.ChainName,
			Id:                 id,
			MetaId:             metaId,
			Number:             0,
			Address:            address,
			InitialOwner:       address,
			CreateAddress:      creator,
			CreateMetaId:       common.GetMetaIdByAddress(creator),
			Timestamp:          timestamp,
			GenesisHeight:      blockHeight,
			GenesisTransaction: msgTx.TxHash().String(),
			Output:             fmt.Sprintf("%s:%d", msgTx.TxHash().String(), outIdx),
			OutputValue:        msgTx.TxOut[outIdx].Value,
			TxInIndex:          uint32(index),
			Offset:             uint64(outIdx),
			TxIndex:            txIndex,
			Operation:          pinInscription.Operation,
			Location:           fmt.Sprintf("%s:%d:%d", msgTx.TxHash().String(), outIdx, locationIdx),
			Path:               pinInscription.Path,
			OriginalPath:       pinInscription.Path,
			ParentPath:         pinInscription.ParentPath,
			Encryption:         pinInscription.Encryption,
			Version:            pinInscription.Version,
			ContentType:        pinInscription.ContentType,
			ContentTypeDetect:  contentTypeDetect,
			ContentBody:        pinInscription.ContentBody,
			ContentLength:      pinInscription.ContentLength,
			ContentSummary:     getContentSummary(pinInscription, id, contentTypeDetect),
			Pop:                pop,
			PopLv:              popLv,
			DataValue:          pin.RarityScoreBinary(indexer.ChainName, pop),
			Mrc20MintId:        []string{},
		})
	}
	return
}
func getParentPath(path string) (parentPath string) {
	arr := strings.Split(path, "/")
	if len(arr) < 3 {
		return
	}
	parentPath = strings.Join(arr[0:len(arr)-1], "/")
	return
}
func getContentSummary(pinode *pin.PersonalInformationNode, id string, contentTypeDetect string) (content string) {
	if contentTypeDetect[0:4] != "text" {
		return fmt.Sprintf("/content/%s", id)
	} else {
		c := string(pinode.ContentBody)
		if len(c) > 150 {
			return c[0:150]
		} else {
			return string(pinode.ContentBody)
		}
	}
}

func (indexer *Indexer) ParseOpReturnPin(pkScript []byte) (pinode *pin.PersonalInformationNode) {
	// Parse pins content from witness script
	tokenizer := txscript.MakeScriptTokenizer(0, pkScript)
	for tokenizer.Next() {
		// Check inscription envelop header: OP_FALSE(0x00), OP_IF(0x63), PROTOCOL_ID
		if tokenizer.Opcode() == txscript.OP_RETURN {
			if !tokenizer.Next() || hex.EncodeToString(tokenizer.Data()) != common.Config.ProtocolID {
				return
			}
			pinode = indexer.parseOpReturnOnePin(&tokenizer)
		}
	}
	return
}
func (indexer *Indexer) parseOpReturnOnePin(tokenizer *txscript.ScriptTokenizer) *pin.PersonalInformationNode {
	// Find any pushed data in the script. This includes OP_0, but not OP_1 - OP_16.
	var infoList [][]byte
	for tokenizer.Next() {
		infoList = append(infoList, tokenizer.Data())
	}
	// Error occurred
	if err := tokenizer.Err(); err != nil {
		return nil
	}
	if len(infoList) < 1 {
		return nil
	}

	pinode := pin.PersonalInformationNode{}
	pinode.Operation = strings.ToLower(string(infoList[0]))
	if pinode.Operation == "init" {
		pinode.Path = "/"
		return &pinode
	}
	if len(infoList) < 6 && pinode.Operation != "revoke" {
		return nil
	}
	if pinode.Operation == "revoke" && len(infoList) < 5 {
		return nil
	}
	pinode.Path = strings.ToLower(string(infoList[1]))
	pinode.ParentPath = getParentPath(pinode.Path)
	encryption := "0"
	if infoList[2] != nil {
		encryption = string(infoList[2])
	}
	pinode.Encryption = encryption
	version := "0"
	if infoList[3] != nil {
		version = string(infoList[3])
	}
	pinode.Version = version
	contentType := "application/json"
	if infoList[4] != nil {
		contentType = strings.ToLower(string(infoList[4]))
	}
	pinode.ContentType = contentType
	var body []byte
	for i := 5; i < len(infoList); i++ {
		body = append(body, infoList[i]...)
	}
	pinode.ContentBody = body
	pinode.ContentLength = uint64(len(body))
	return &pinode
}
func (indexer *Indexer) GetOpReturnPinOwner(tx *wire.MsgTx) (address string, outIdx int, locationIdx int64) {
	for i, out := range tx.TxOut {
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, indexer.ChainParams)
		if class.String() != "nonstandard" && len(addresses) > 0 {
			outIdx = i
			address = addresses[0].String()
			locationIdx = 0
		}
	}
	return
}
func (indexer *Indexer) GetPinOwner(tx *wire.MsgTx, inIdx int) (address string, outIdx int, locationIdx int64) {
	if len(tx.TxIn) == 1 || len(tx.TxOut) == 1 || inIdx == 0 {
		_, addresses, _, _ := txscript.ExtractPkScriptAddrs(tx.TxOut[0].PkScript, indexer.ChainParams)
		if len(addresses) > 0 {
			address = addresses[0].String()
		}
		return
	}
	inputValue := int64(0)
	for i, in := range tx.TxIn {
		if i == inIdx {
			break
		}
		value, err := GetValueByTx(in.PreviousOutPoint.Hash.String(), int(in.PreviousOutPoint.Index))
		if err != nil {
			return
		}
		inputValue += value
	}
	outputValue := int64(0)
	for x, out := range tx.TxOut {
		outputValue += out.Value
		if outputValue > inputValue {
			locationIdx = outputValue - inputValue
			_, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, indexer.ChainParams)
			if len(addresses) > 0 {
				address = addresses[0].String()
				outIdx = x
			}
			break
		}
	}
	return
}
func (indexer *Indexer) ParsePins(witnessScript []byte) (pins []*pin.PersonalInformationNode) {
	// Parse pins content from witness script
	tokenizer := txscript.MakeScriptTokenizer(0, witnessScript)
	for tokenizer.Next() {
		// Check inscription envelop header: OP_FALSE(0x00), OP_IF(0x63), PROTOCOL_ID
		if tokenizer.Opcode() == txscript.OP_FALSE {
			if !tokenizer.Next() || tokenizer.Opcode() != txscript.OP_IF {
				return
			}
			if !tokenizer.Next() || hex.EncodeToString(tokenizer.Data()) != common.Config.ProtocolID {
				return
			}
			pinode := indexer.parseOnePin(&tokenizer)
			if pinode != nil {
				pins = append(pins, pinode)
			}
		}
	}
	return
}
func (indexer *Indexer) ParsePin(witnessScript []byte) (pinode *pin.PersonalInformationNode) {
	// Parse pins content from witness script
	tokenizer := txscript.MakeScriptTokenizer(0, witnessScript)
	for tokenizer.Next() {
		// Check inscription envelop header: OP_FALSE(0x00), OP_IF(0x63), PROTOCOL_ID
		if tokenizer.Opcode() == txscript.OP_FALSE {
			if !tokenizer.Next() || tokenizer.Opcode() != txscript.OP_IF {
				return
			}
			if !tokenizer.Next() || hex.EncodeToString(tokenizer.Data()) != common.Config.ProtocolID {
				return
			}
			pinode = indexer.parseOnePin(&tokenizer)
		}
	}
	return
}
func (indexer *Indexer) parseOnePin(tokenizer *txscript.ScriptTokenizer) *pin.PersonalInformationNode {
	// Find any pushed data in the script. This includes OP_0, but not OP_1 - OP_16.
	var infoList [][]byte
	for tokenizer.Next() {
		if tokenizer.Opcode() == txscript.OP_ENDIF {
			break
		}
		infoList = append(infoList, tokenizer.Data())
		if len(tokenizer.Data()) > 520 {
			//log.Errorf("data is longer than 520")
			return nil
		}
	}
	// No OP_ENDIF
	if tokenizer.Opcode() != txscript.OP_ENDIF {
		return nil
	}
	// Error occurred
	if err := tokenizer.Err(); err != nil {
		return nil
	}
	if len(infoList) < 1 {
		return nil
	}

	pinode := pin.PersonalInformationNode{}
	pinode.Operation = strings.ToLower(string(infoList[0]))
	if pinode.Operation == "init" {
		pinode.Path = "/"
		return &pinode
	}
	if len(infoList) < 6 && pinode.Operation != "revoke" {
		return nil
	}
	if pinode.Operation == "revoke" && len(infoList) < 5 {
		return nil
	}
	pinode.Path = strings.ToLower(string(infoList[1]))
	pinode.ParentPath = getParentPath(pinode.Path)
	encryption := "0"
	if infoList[2] != nil {
		encryption = string(infoList[2])
	}
	pinode.Encryption = encryption
	version := "0"
	if infoList[3] != nil {
		version = string(infoList[3])
	}
	pinode.Version = version
	contentType := "application/json"
	if infoList[4] != nil {
		contentType = strings.ToLower(string(infoList[4]))
	}
	pinode.ContentType = contentType
	var body []byte
	for i := 5; i < len(infoList); i++ {
		body = append(body, infoList[i]...)
	}
	pinode.ContentBody = body
	pinode.ContentLength = uint64(len(body))
	return &pinode
}
func (indexer *Indexer) GetBlockTxHash(blockHeight int64) (txhashList []string) {
	chain := BitcoinChain{}
	blockMsg, err := chain.GetBlock(blockHeight)
	if err != nil {
		return
	}
	block := blockMsg.(*wire.MsgBlock)
	for _, tx := range block.Transactions {
		for i := range tx.Copy().TxOut {
			var pinId strings.Builder
			pinId.WriteString(tx.TxHash().String())
			pinId.WriteString("i")
			pinId.WriteString(strconv.Itoa(i))
			txhashList = append(txhashList, pinId.String())
		}
	}
	return
}
