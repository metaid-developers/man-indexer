package example

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/chaincfg"
	"manindexer/adapter/bitcoin"
	"manindexer/pin"
	"testing"
)

func init() {

}

func TestPinParseFromWitnessScript(t *testing.T) {
	var (
		witnessScript      string = "2063cc4129f5d7e454848eaa939e88554df07306b39b0ff7bdf0b64335ac7c30eaac0063066d6574616964066372656174650a2f696e666f2f6e616d65013005312e302e300a746578742f706c61696e0b4d65746149442d4465767368"
		witnessScriptBytes []byte
		err                error

		netParams = &chaincfg.TestNet3Params
		pinNode   *pin.PersonalInformationNode
		indexer   *bitcoin.Indexer
	)
	witnessScriptBytes, err = hex.DecodeString(witnessScript)
	if err != nil {
		t.Fatal(err)
	}
	// Parse the witness script
	indexer = &bitcoin.Indexer{ChainParams: netParams}
	pinNode = indexer.ParsePin(witnessScriptBytes)
	t.Logf("Operation:%s", pinNode.Operation)
	t.Logf("Path:%s", pinNode.Path)
	t.Logf("Encryption:%s", pinNode.Encryption)
	t.Logf("Version:%s", pinNode.Version)
	t.Logf("ContentType:%s", pinNode.ContentType)
	t.Logf("ContentBody:%s", string(pinNode.ContentBody))
	t.Logf("ContentLength:%d", pinNode.ContentLength)
	t.Logf("ParentPath:%s", pinNode.ParentPath)
	t.Logf("Protocols:%t", pinNode.Protocols)
}
