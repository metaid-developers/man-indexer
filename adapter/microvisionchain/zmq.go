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
func (indexer *Indexer) ZmqRun1(chanMsg chan pin.MempollChanMsg) {
	q, _ := zmq.NewSocket(zmq.SUB)
	defer q.Close()
	err := q.Connect(common.Config.Mvc.ZmqHost)
	if err != nil {
		log.Println("Mvc ZmqRun:", err)
	}
	q.SetTcpKeepalive(1)
	q.SetTcpKeepaliveIdle(60)
	q.SetTcpKeepaliveIntvl(1)
	q.SetSubscribe("rawtx")
	for {
		msg, _ := q.RecvMessage(0)
		var msgTx wire.MsgTx
		if err := msgTx.Deserialize(bytes.NewReader([]byte(msg[1]))); err != nil {
			continue
		}

		pinInscriptions := indexer.CatchPinsByTx(&msgTx, 0, 0, "", "", 0)
		if len(pinInscriptions) > 0 {
			chanMsg <- pin.MempollChanMsg{PinList: pinInscriptions, Tx: msgTx}
		}
		//PIN transfer check
		tansferList, err := indexer.TransferCheck(&msgTx)
		if err == nil && len(tansferList) > 0 {
			chanMsg <- pin.MempollChanMsg{PinList: tansferList, Tx: msgTx}
		}
	}
}

func keepTcpKeepalive() {
	var subscriber *zmq.Socket
	var err error
	reconnectInterval := 10 * time.Second
	ticker := time.NewTicker(reconnectInterval)
	defer ticker.Stop()
	reconnect := func() {
		if subscriber != nil {
			subscriber.Close()
		}
		subscriber, err = zmq.NewSocket(zmq.SUB)
		if err != nil {
			log.Fatal(err)
		}

		err = subscriber.Connect(common.Config.Mvc.ZmqHost)
		if err != nil {
			log.Printf("Error connecting: %v", err)
			subscriber.Close()
			subscriber = nil
		} else {
			err = subscriber.SetSubscribe("")
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	reconnect()
	go func() {
		for {
			msg, err := subscriber.Recv(0)
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				reconnect()
			} else {
				log.Printf("Received message: %s", msg)
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			log.Println("Checking connection...")
			reconnect()
		}
	}
}
func (indexer *Indexer) ZmqRun(chanMsg chan pin.MempollChanMsg) {
	context, _ := zmq.NewContext()
	subscriber, _ := context.NewSocket(zmq.SUB)
	defer subscriber.Close()
	//subscriber.SetSubscribe("rawtx")
	subscriber.SetSubscribe("rawtx")
	//Enable the keepalive property
	err := subscriber.SetTcpKeepalive(1)
	if err != nil {
		log.Println("SetTcpKeepalive err,", err)
	}
	//If there is no data exchange within 0.5 seconds, check the connection
	err = subscriber.SetTcpKeepaliveIdle(60)
	if err != nil {
		log.Println("SetTcpKeepaliveIdle err,", err)
	}
	//The packet sending interval during the check is 2 seconds
	err = subscriber.SetTcpKeepaliveIntvl(1)
	if err != nil {
		log.Println("SetTcpKeepaliveIntvl err,", err)
	}
	subscriber.SetRcvhwm(20000)
	subscriber.SetRcvbuf(1024 * 200)
	err = subscriber.Connect(common.Config.Mvc.ZmqHost)
	if err != nil {
		log.Println("Connect to MVC ZMQ error", err)
		return
	} else {
		log.Println("MVC ZMQ connected")
	}

	//go keepTcpKeepalive()
	for {
		//recvmsg, err := subscriber.RecvMessage(0)
		recvmsg, err := subscriber.Recv(0)
		//recvmsg, err := subscriber.RecvMessage(zmq.DONTWAIT)
		if err != nil {
			log.Println("MVC ZMQ RecvMessage Err,", err)
			continue
		} else {
			//s, _ := subscriber.GetEvents()
			//log.Println("Recive MVC ZMQ message", len(recvmsg), s.String())
			if recvmsg == "rawtx" || len(recvmsg) < 10 {
				continue
			}
			var msgTx wire.MsgTx
			if err := msgTx.Deserialize(bytes.NewReader([]byte(recvmsg))); err != nil {
				continue
			}
			//newHash, _ := GetNewHash(&msgTx)
			//log.Println("TxHash:", newHash)
			pinInscriptions := indexer.CatchPinsByTx(&msgTx, 0, 0, "", "", 0)
			if len(pinInscriptions) > 0 {
				chanMsg <- pin.MempollChanMsg{PinList: pinInscriptions, Tx: msgTx}
			}
			//PIN transfer check
			tansferList, err := indexer.TransferCheck(&msgTx)
			if err == nil && len(tansferList) > 0 {
				chanMsg <- pin.MempollChanMsg{PinList: tansferList, Tx: msgTx}
			}
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
