package microvisionchain

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"manindexer/common"
	"manindexer/pin"
	"strings"
	"time"

	"github.com/btcsuite/btcd/wire"
	zmq "github.com/pebbe/zmq4"
)

func (indexer *Indexer) ZmqHashblock() {
	q, _ := zmq.NewSocket(zmq.SUB)
	defer q.Close()
	q.Connect("tcp://127.0.0.1:28337")
	q.SetSubscribe("hashblock")

	for {
		msg, err := q.RecvBytes(0)
		if err == nil {
			blockHeightBytes := msg[4:8]
			blockHeight := binary.LittleEndian.Uint32(blockHeightBytes)
			fmt.Println("Received block height:", blockHeight)
		}
	}
}
func (indexer *Indexer) ZmqRun(chanMsg chan []*pin.PinInscription) {
	q, _ := zmq.NewSocket(zmq.SUB)
	defer q.Close()
	err := q.Connect(common.Config.Mvc.ZmqHost)
	if err != nil {
		log.Println("Mvc ZmqRun:", err)
	}
	q.SetSubscribe("rawtx")
	for {
		msg, _ := q.RecvMessage(0)
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader([]byte(msg[1]))); err != nil {
			continue
		}
		pinInscriptions := indexer.CatchPinsByTx(&msgTx, 0, 0, "", "", 0)
		if len(pinInscriptions) > 0 {
			chanMsg <- pinInscriptions
		}
		//PIN transfer check
		tansferList, err := indexer.TransferCheck(&msgTx)
		if err == nil && len(tansferList) > 0 {
			chanMsg <- tansferList
		}
	}
}
func (indexer *Indexer) TransferCheck(tx *wire.MsgTx) (transferPinList []*pin.PinInscription, err error) {
	var outputList []string
	for _, in := range tx.TxIn {
		output := fmt.Sprintf("%s:%d", in.PreviousOutPoint.Hash.String(), in.PreviousOutPoint.Index)
		outputList = append(outputList, output)
	}
	pinList, err := (*indexer.DbAdapter).GetPinListByOutPutList(outputList)
	if err != nil {
		return
	}
	timeNow := time.Now().Unix()
	for _, pinNode := range pinList {
		arr := strings.Split(pinNode.Output, ":")
		if len(arr) < 2 {
			continue
		}
		//idx, _ := strconv.Atoi(arr[1])
		transferPin := pin.PinInscription{
			Id:                 pinNode.Id,
			CreateAddress:      pinNode.Address,
			Timestamp:          timeNow,
			GenesisTransaction: tx.TxHash().String(),
			IsTransfered:       true,
		}
		fmt.Println(pinNode.Output)
		info, err := indexer.GetOWnerAddress(pinNode.Output, tx)
		//transferPin.Address, _, _ = indexer.GetPinOwner(tx, idx)
		if err != nil {
			continue
		}
		transferPin.Address = info.Address
		transferPinList = append(transferPinList, &transferPin)
	}
	return
}
