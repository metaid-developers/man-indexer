package bitcoin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"manindexer/common"
	"manindexer/pin"

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
	q.Connect(common.Config.Btc.ZmqHost)
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
	}
}
