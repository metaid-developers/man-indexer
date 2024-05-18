package bitcoin

import (
	"encoding/hex"
	"errors"
	"fmt"
	"manindexer/common"
	"manindexer/pin"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var PopCutNum int = 0

type Indexer struct {
	ChainParams *chaincfg.Params
	Block       interface{}
	PopCutNum   int
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
			id := fmt.Sprintf("%si%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
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
func (indexer *Indexer) CatchTransfer(idMap map[string]struct{}) (addressMap map[string]string) {
	addressMap = make(map[string]string)
	block := indexer.Block.(*wire.MsgBlock)
	for _, tx := range block.Transactions {
		for _, in := range tx.TxIn {
			id := fmt.Sprintf("%si%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
			if _, ok := idMap[id]; ok {
				newAddress, err := indexer.getOWnerAddress(id, tx)
				if err == nil && newAddress != "" {
					addressMap[id] = newAddress
				} else {
					addressMap[id] = "errAddress"
				}
			}
		}
	}
	return
}
func (indexer *Indexer) getOWnerAddress(inputId string, tx *wire.MsgTx) (address string, err error) {
	//fmt.Println("tx:", tx.TxHash().String(), inputId)
	firstInputId := fmt.Sprintf("%si%d", tx.TxIn[0].PreviousOutPoint.Hash, tx.TxIn[0].PreviousOutPoint.Index)
	if len(tx.TxIn) == 1 || firstInputId == inputId {
		class, addresses, _, _ := txscript.ExtractPkScriptAddrs(tx.TxOut[0].PkScript, indexer.ChainParams)
		if len(addresses) > 0 {
			address = addresses[0].String()
		} else if class.String() == "nulldata" {
			address = hex.EncodeToString(tx.TxOut[0].PkScript)
		}
		return
	}
	totalOutputValue := int64(0)
	for _, out := range tx.TxOut {
		totalOutputValue += out.Value
	}
	inputValue := int64(0)
	for _, in := range tx.TxIn {
		id := fmt.Sprintf("%si%d", in.PreviousOutPoint.Hash, in.PreviousOutPoint.Index)
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
	for _, out := range tx.TxOut {
		outputValue += out.Value
		//fmt.Println(out.Value)
		if outputValue > inputValue {
			class, addresses, _, _ := txscript.ExtractPkScriptAddrs(out.PkScript, indexer.ChainParams)
			if len(addresses) > 0 {
				address = addresses[0].String()
			} else if class.String() == "nulldata" {
				address = hex.EncodeToString(out.PkScript)
			}
			break
		}
	}

	return
}
func (indexer *Indexer) CatchPinsByTx(msgTx *wire.MsgTx, blockHeight int64, timestamp int64, blockHash string, merkleRoot string, txIndex int) (pinInscriptions []*pin.PinInscription) {
	//No witness data
	if !msgTx.HasWitness() {
		return nil
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
		address, outIdx := indexer.getPinOwner(msgTx, index)
		id := fmt.Sprintf("%si%d", msgTx.TxHash().String(), outIdx)
		rootTxId := ""
		metaId := ""
		if pinInscription.Operation == "init" {
			rootTxId = msgTx.TxHash().String()
			metaId = id
		}
		contentTypeDetect := common.DetectContentType(&pinInscription.ContentBody)
		pop := ""
		if merkleRoot != "" && blockHash != "" {
			pop, _ = common.GenPop(id, merkleRoot, blockHash)
		}

		pinInscriptions = append(pinInscriptions, &pin.PinInscription{
			//Pin:                pinInscription,
			Id:                 id,
			RootTxId:           rootTxId,
			MetaId:             metaId,
			Number:             0,
			Address:            address,
			CreateAddress:      address,
			Timestamp:          timestamp,
			GenesisHeight:      blockHeight,
			GenesisTransaction: msgTx.TxHash().String(),
			Output:             fmt.Sprintf("%s:%d", msgTx.TxHash().String(), outIdx),
			OutputValue:        msgTx.TxOut[outIdx].Value,
			TxInIndex:          uint32(index),
			TxInOffset:         uint64(0),
			TxIndex:            txIndex,
			Operation:          pinInscription.Operation,
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
func (indexer *Indexer) getPinOwner(tx *wire.MsgTx, inIdx int) (address string, outIdx int) {
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
			if !tokenizer.Next() || hex.EncodeToString(tokenizer.Data()) != pin.ProtocolID {
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
			if !tokenizer.Next() || hex.EncodeToString(tokenizer.Data()) != pin.ProtocolID {
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
	if len(infoList) < 6 {
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
func (indexer *Indexer) PopLevelCount(pop string) string {
	cnt := len(pop) - len(strings.TrimLeft(pop, "0"))
	if cnt <= PopCutNum {
		return "--"
	} else {
		return fmt.Sprintf("Lv%d", cnt-PopCutNum)
	}
}
