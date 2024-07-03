package mrc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"manindexer/common"
)

type Mrc20Builder struct {
	Net        *chaincfg.Params
	MetaIdData *MetaIdData

	//deploy
	deployPremineOutAddress string
	deployPinOutAddress     string

	//mint
	MintPins            []*MintPin
	mrc20OutValue       int64
	mrc20OutAddressList []string

	//transfer
	TransferMrc20s []*TransferMrc20
	Mrc20Outs      []*Mrc20OutInfo

	FeeRate            int64
	op                 string
	mrc20ChangeAddress string

	revealTaprootDataInputIndex uint32
	TxCtxData                   *inscriptionTxCtxData
	RevealPsbtBuilder           *common.PsbtBuilder
	revealTx                    *wire.MsgTx
}

type MintPin struct {
	PinId           string
	PinUtxoTxId     string
	PinUtxoIndex    uint32
	PinUtxoOutValue int64
	PrivateKeyHex   string
	Address         string
	RedeemScript    string
	PkScript        string
	OutRaw          string
}

type TransferMrc20 struct {
	PrivateKeyHex string
	Address       string
	RedeemScript  string
	PkScript      string
	OutRaw        string
	UtxoTxId      string
	UtxoIndex     uint32
	UtxoOutValue  int64
	Mrc20Amount   string
	Mrc20TickerId string
}

type MetaIdData struct {
	MetaIDFlag  string
	Operation   string
	Path        string
	Content     []byte
	Encryption  string
	Version     string
	ContentType string
}

type inscriptionTxCtxData struct {
	privateKey              *btcec.PrivateKey
	InscriptionScript       []byte
	CommitTxAddressPkScript []byte
	ControlBlockWitness     []byte
	recoveryPrivateKeyWIF   string
	RecoveryPrivateKeyHex   string
	revealTxPrevOutput      *wire.TxOut
}

type CalInput struct {
	OutTxId    string
	OutIndex   uint32
	OutAddress string
}

func NewMrc20BuilderFromPsbtRaws(net *chaincfg.Params, revealPsbtRaw string) (*Mrc20Builder, error) {
	var (
		revealPsbtBuilder *common.PsbtBuilder
		revealTx          *wire.MsgTx
		err               error
	)

	revealPsbtBuilder, err = common.NewPsbtBuilder(net, revealPsbtRaw)
	if err != nil {
		return nil, err
	}

	return &Mrc20Builder{
		Net:               net,
		RevealPsbtBuilder: revealPsbtBuilder,
		revealTx:          revealTx,
	}, nil
}

func (m *Mrc20Builder) buildEmptyRevealPsbt() error {
	var (
		revealPsbtBuilder     *common.PsbtBuilder
		inputs                []common.Input      = make([]common.Input, 0)
		inSigners             []*common.InputSign = make([]*common.InputSign, 0)
		outputs               []common.Output     = make([]common.Output, 0)
		taprootDataInputIndex uint32              = 0
		err                   error
	)
	if m.MintPins != nil && len(m.MintPins) != 0 {
		for i, v := range m.MintPins {
			in := common.Input{
				OutTxId:  v.PinUtxoTxId,
				OutIndex: v.PinUtxoIndex,
			}
			inputs = append(inputs, in)
			taprootDataInputIndex++

			utxoType := common.Witness
			addressClass, err := CheckAddressClass(m.Net, v.Address)
			if err != nil {
				return err
			}
			if addressClass == txscript.WitnessV1TaprootTy {
				utxoType = common.Taproot
			} else if addressClass == txscript.PubKeyHashTy {
				utxoType = common.NonWitness
				if v.OutRaw == "" {
					return errors.New("outRaw is empty")
				}
			} else if addressClass == txscript.ScriptHashTy {
				//if v.ReeemScript == "" {
				//	return errors.New("redeemScript is empty")
				//}
			}
			fmt.Printf("addressClass:%v\n", addressClass)

			inSigner := &common.InputSign{
				UtxoType:     utxoType,
				Index:        i,
				OutRaw:       v.OutRaw,
				PkScript:     v.PkScript,
				RedeemScript: v.RedeemScript,
				Amount:       uint64(v.PinUtxoOutValue),
				SighashType:  txscript.SigHashAll,
				PriHex:       v.PrivateKeyHex,
				//MultiSigScript: "",
				//PreSigScript:   "",
			}
			inSigners = append(inSigners, inSigner)

		}
	}
	if m.TransferMrc20s != nil && len(m.TransferMrc20s) != 0 {
		for i, v := range m.TransferMrc20s {
			in := common.Input{
				OutTxId:  v.UtxoTxId,
				OutIndex: v.UtxoIndex,
			}
			inputs = append(inputs, in)
			taprootDataInputIndex++

			utxoType := common.Witness
			addressClass, err := CheckAddressClass(m.Net, v.Address)
			if err != nil {
				return err
			}
			if addressClass == txscript.WitnessV1TaprootTy {
				utxoType = common.Taproot
			} else if addressClass == txscript.PubKeyHashTy {
				utxoType = common.NonWitness
				if v.OutRaw == "" {
					return errors.New("outRaw is empty")
				}
			} else if addressClass == txscript.ScriptHashTy {
				//if v.ReeemScript == "" {
				//	return errors.New("redeemScript is empty")
				//}
			}
			fmt.Printf("addressClass:%v\n", addressClass)

			inSigner := &common.InputSign{
				UtxoType:     utxoType,
				Index:        i,
				OutRaw:       v.OutRaw,
				PkScript:     v.PkScript,
				RedeemScript: v.RedeemScript,
				Amount:       uint64(v.UtxoOutValue),
				SighashType:  txscript.SigHashAll,
				PriHex:       v.PrivateKeyHex,
				//MultiSigScript: "",
				//PreSigScript:   "",
			}
			inSigners = append(inSigners, inSigner)
		}
	}

	if m.op == "mint" {
		for _, v := range m.MintPins {
			out := common.Output{
				Address: v.Address,
				Amount:  uint64(v.PinUtxoOutValue),
				//Script:  "",
			}
			outputs = append(outputs, out)
		}
		for _, mrc20OutAddress := range m.mrc20OutAddressList {
			mrc20Out := common.Output{
				Address: mrc20OutAddress,
				Amount:  uint64(m.mrc20OutValue),
				//Script:  "",
			}
			outputs = append(outputs, mrc20Out)
		}
	} else if m.op == "transfer" {
		out := common.Output{
			Address: m.mrc20ChangeAddress,
			Amount:  uint64(546),
			//Script:  "",
		}
		outputs = append(outputs, out)
		for _, v := range m.Mrc20Outs {
			out := common.Output{
				Address: v.Address,
				Amount:  uint64(v.OutValue),
				//Script:  "",
			}
			outputs = append(outputs, out)
		}
	} else if m.op == "deploy" {
		outPin := common.Output{
			Address: m.deployPinOutAddress,
			Amount:  uint64(m.mrc20OutValue),
		}
		outMrc20Premine := common.Output{
			Address: m.deployPremineOutAddress,
			Amount:  uint64(m.mrc20OutValue),
		}
		outputs = append(outputs, outPin)
		outputs = append(outputs, outMrc20Premine)
	}

	emptyTxId := "0000000000000000000000000000000000000000000000000000000000000000"
	taprootDataIn := common.Input{
		OutTxId:  emptyTxId,
		OutIndex: 0,
	}

	inputs = append(inputs, taprootDataIn)

	revealPsbtBuilder, err = common.CreatePsbtBuilder(m.Net, inputs, outputs)
	if err != nil {
		return err
	}
	m.RevealPsbtBuilder = revealPsbtBuilder

	taprootDataInSigner := &common.InputSign{
		UtxoType: common.Taproot,
		Index:    int(taprootDataInputIndex),
		//OutRaw:         "",
		PkScript:            hex.EncodeToString(m.TxCtxData.CommitTxAddressPkScript),
		RedeemScript:        hex.EncodeToString(m.TxCtxData.InscriptionScript),
		ControlBlockWitness: hex.EncodeToString(m.TxCtxData.ControlBlockWitness),
		Amount:              uint64(m.CalRevealPsbtFee(m.FeeRate)),
		SighashType:         txscript.SigHashAll,
		PriHex:              "",
		//MultiSigScript: "",
		//PreSigScript:   "",
	}
	inSigners = append(inSigners, taprootDataInSigner)

	err = revealPsbtBuilder.UpdateAndAddInputWitness(inSigners)
	if err != nil {
		return err
	}

	m.RevealPsbtBuilder = revealPsbtBuilder
	m.revealTaprootDataInputIndex = taprootDataInputIndex
	m.TxCtxData.revealTxPrevOutput = &wire.TxOut{
		PkScript: m.TxCtxData.CommitTxAddressPkScript,
		Value:    m.CalRevealPsbtFee(m.FeeRate),
	}
	return nil
}

func (m *Mrc20Builder) CalRevealPsbtFee(feeRate int64) int64 {
	var (
		tx          *wire.MsgTx = m.RevealPsbtBuilder.PsbtUpdater.Upsbt.UnsignedTx
		txTotalSize int         = tx.SerializeSize()
		txBaseSize  int         = tx.SerializeSizeStripped()
		txFee       int64       = 0
		weight      int64       = 0
		vSize       int64       = 0

		emptySegwitWitenss   = wire.TxWitness{make([]byte, 71), make([]byte, 33)}
		emptyNestSignature   = make([]byte, 23)
		emptylegacySignature = make([]byte, 107)
		emptyTaprootWitness  = wire.TxWitness{make([]byte, 64)}
		revealOutValues      = int64(0)
	)

	if m.op == "mint" {
		for _, v := range m.MintPins {
			addressClass, err := CheckAddressClass(m.Net, v.Address)
			if err != nil {
				fmt.Printf("CheckAddressClass err:%s\n", err.Error())
				continue
			}
			if addressClass == txscript.WitnessV1TaprootTy {
				txTotalSize += emptyTaprootWitness.SerializeSize()
			} else if addressClass == txscript.PubKeyHashTy {
				txBaseSize += 40 + wire.VarIntSerializeSize(uint64(len(emptylegacySignature))) + len(emptylegacySignature)
			} else if addressClass == txscript.ScriptHashTy {
				txBaseSize += 40 + wire.VarIntSerializeSize(uint64(len(emptyNestSignature))) + len(emptyNestSignature)
				txTotalSize += emptySegwitWitenss.SerializeSize()
			} else {
				txTotalSize += emptySegwitWitenss.SerializeSize()
			}
		}
		for _, v := range m.mrc20OutAddressList {
			revealOutValues += m.mrc20OutValue
			_ = v
		}
	} else if m.op == "transfer" {
		for _, v := range m.TransferMrc20s {
			addressClass, err := CheckAddressClass(m.Net, v.Address)
			if err != nil {
				fmt.Printf("CheckAddressClass err:%s\n", err.Error())
				continue
			}
			if addressClass == txscript.WitnessV1TaprootTy {
				txTotalSize += emptyTaprootWitness.SerializeSize()
			} else if addressClass == txscript.PubKeyHashTy {
				txBaseSize += 40 + wire.VarIntSerializeSize(uint64(len(emptylegacySignature))) + len(emptylegacySignature)
			} else if addressClass == txscript.ScriptHashTy {
				txBaseSize += 40 + wire.VarIntSerializeSize(uint64(len(emptyNestSignature))) + len(emptyNestSignature)
				txTotalSize += emptySegwitWitenss.SerializeSize()
			} else {
				txTotalSize += emptySegwitWitenss.SerializeSize()
			}
		}
		for _, v := range m.Mrc20Outs {
			revealOutValues += v.OutValue
		}
	}

	emptySignature := make([]byte, 64)
	emptyControlBlockWitness := make([]byte, 33)
	txTotalSize += wire.TxWitness{emptySignature, m.TxCtxData.InscriptionScript, emptyControlBlockWitness}.SerializeSize()

	weight = int64(txBaseSize*3 + txTotalSize)
	vSize = (weight + (blockchain.WitnessScaleFactor - 1)) / blockchain.WitnessScaleFactor
	txFee = vSize * feeRate
	fmt.Printf("weight:%d, vSize:%d, txFee:%d\n", weight, vSize, txFee)
	fmt.Printf("revealOutValues:%d, totalMinerFee:%d\n", revealOutValues, txFee+revealOutValues)
	return txFee + revealOutValues
}

func (m *Mrc20Builder) completeRevealPsbt(commitTxId string, commitTxOutIndex uint32) error {
	var (
		commitPreOutPoint *wire.OutPoint
		txHash            *chainhash.Hash
		err               error
	)
	txHash, err = chainhash.NewHashFromStr(commitTxId)
	if err != nil {
		return err
	}
	commitPreOutPoint = wire.NewOutPoint(txHash, commitTxOutIndex)
	m.RevealPsbtBuilder.PsbtUpdater.Upsbt.UnsignedTx.TxIn[m.revealTaprootDataInputIndex].PreviousOutPoint = *commitPreOutPoint
	return nil
}

func (m *Mrc20Builder) signRevealPsbt(mintPins []*MintPin, transferMrc20s []*TransferMrc20, taprootInSigner *common.InputSign) error {
	var (
		revealSigners        []*common.InputSign = make([]*common.InputSign, 0)
		revealTaprootSigners []*common.InputSign = make([]*common.InputSign, 0)
		err                  error
	)
	if mintPins == nil {
		mintPins = m.MintPins
	}
	if transferMrc20s == nil {
		transferMrc20s = m.TransferMrc20s
	}
	if len(mintPins) == 0 && len(transferMrc20s) == 0 {
		return errors.New("empty mintPins and transferMrc20s")
	}

	for i, v := range mintPins {
		inSigner := &common.InputSign{
			UtxoType: common.Witness,
			Index:    i,
			//OutRaw:         "",
			PkScript:     v.PkScript,
			RedeemScript: "",
			Amount:       uint64(v.PinUtxoOutValue),
			SighashType:  txscript.SigHashAll,
			PriHex:       v.PrivateKeyHex,
			//MultiSigScript: "",
			//PreSigScript:   "",
		}
		revealSigners = append(revealSigners, inSigner)
	}

	for i, v := range transferMrc20s {
		inSigner := &common.InputSign{
			UtxoType: common.Witness,
			Index:    i,
			//OutRaw:         "",
			PkScript:     v.PkScript,
			RedeemScript: "",
			Amount:       uint64(v.UtxoOutValue),
			SighashType:  txscript.SigHashAll,
			PriHex:       v.PrivateKeyHex,
			//MultiSigScript: "",
			//PreSigScript:   "",
		}
		revealSigners = append(revealSigners, inSigner)

	}

	err = m.RevealPsbtBuilder.UpdateAndSignInput(revealSigners)
	if err != nil {
		return err
	}

	if taprootInSigner == nil {
		taprootInSigner = &common.InputSign{
			UtxoType: common.Taproot,
			Index:    int(m.revealTaprootDataInputIndex),
			//OutRaw:         "",
			PkScript:            hex.EncodeToString(m.TxCtxData.CommitTxAddressPkScript),
			RedeemScript:        hex.EncodeToString(m.TxCtxData.InscriptionScript),
			ControlBlockWitness: hex.EncodeToString(m.TxCtxData.ControlBlockWitness),
			Amount:              uint64(m.CalRevealPsbtFee(m.FeeRate)),
			SighashType:         txscript.SigHashAll,
			PriHex:              m.TxCtxData.RecoveryPrivateKeyHex,
			//MultiSigScript: "",
			//PreSigScript:   "",
		}
		revealTaprootSigners = append(revealTaprootSigners, taprootInSigner)
	}

	err = m.RevealPsbtBuilder.UpdateAndSignTaprootInput(revealTaprootSigners)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mrc20Builder) ExtractRevealTransaction() (string, string, error) {
	var (
		commitTxHex string
		revealTxHex string
		err         error
	)

	revealTxHex, err = m.RevealPsbtBuilder.ExtractPsbtTransaction()
	if err != nil {
		return "", "", err
	}
	return commitTxHex, revealTxHex, nil
}

func createMetaIdTxCtxData(net *chaincfg.Params, metaIdData *MetaIdData) (*inscriptionTxCtxData, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	inscriptionBuilder := txscript.NewScriptBuilder().
		AddData(schnorr.SerializePubKey(privateKey.PubKey())).
		AddOp(txscript.OP_CHECKSIG).
		AddOp(txscript.OP_FALSE).
		AddOp(txscript.OP_IF).
		AddData([]byte(metaIdData.MetaIDFlag)). //<metaid_flag>
		AddData([]byte(metaIdData.Operation))   //<operation>

	inscriptionBuilder.AddData([]byte(metaIdData.Path)) //<path>
	if metaIdData.Encryption == "" {
		inscriptionBuilder.AddOp(txscript.OP_0)
	} else {
		inscriptionBuilder.AddData([]byte(metaIdData.Encryption)) //<Encryption>
	}

	if metaIdData.Version == "" {
		inscriptionBuilder.AddOp(txscript.OP_0)
	} else {
		inscriptionBuilder.AddData([]byte(metaIdData.Version)) //<version>
	}

	if metaIdData.ContentType == "" {
		inscriptionBuilder.AddOp(txscript.OP_0)
	} else {
		inscriptionBuilder.AddData([]byte(metaIdData.ContentType)) //<content-type>
	}
	maxChunkSize := 520
	bodySize := len(metaIdData.Content)
	for i := 0; i < bodySize; i += maxChunkSize {
		end := i + maxChunkSize
		if end > bodySize {
			end = bodySize
		}
		inscriptionBuilder.AddFullData(metaIdData.Content[i:end]) //<payload>
	}

	inscriptionScript, err := inscriptionBuilder.Script()
	if err != nil {
		return nil, err
	}
	inscriptionScript = append(inscriptionScript, txscript.OP_ENDIF)

	proof := &txscript.TapscriptProof{
		TapLeaf:  txscript.NewBaseTapLeaf(schnorr.SerializePubKey(privateKey.PubKey())),
		RootNode: txscript.NewBaseTapLeaf(inscriptionScript),
	}

	controlBlock := proof.ToControlBlock(privateKey.PubKey())
	controlBlockWitness, err := controlBlock.ToBytes()
	if err != nil {
		return nil, err
	}

	tapHash := proof.RootNode.TapHash()
	commitTxAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootOutputKey(privateKey.PubKey(), tapHash[:])), net)
	if err != nil {
		return nil, err
	}
	commitTxAddressPkScript, err := txscript.PayToAddrScript(commitTxAddress)
	if err != nil {
		return nil, err
	}

	recoveryPrivateKeyWIF, err := btcutil.NewWIF(txscript.TweakTaprootPrivKey(*privateKey, tapHash[:]), net, true)
	if err != nil {
		return nil, err
	}

	recoveryPrivateKeyHex := hex.EncodeToString(privateKey.Serialize())

	return &inscriptionTxCtxData{
		privateKey:              privateKey,
		InscriptionScript:       inscriptionScript,
		CommitTxAddressPkScript: commitTxAddressPkScript,
		ControlBlockWitness:     controlBlockWitness,
		recoveryPrivateKeyWIF:   recoveryPrivateKeyWIF.String(),
		RecoveryPrivateKeyHex:   recoveryPrivateKeyHex,
	}, nil
}

// address to pkScript
func AddressToPkScript(net *chaincfg.Params, address string) (string, error) {
	addr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return "", err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	return pkScript, nil
}

func CheckAddressClass(net *chaincfg.Params, address string) (txscript.ScriptClass, error) {
	addr, err := btcutil.DecodeAddress(address, net)
	if err != nil {
		return txscript.NonStandardTy, err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return txscript.NonStandardTy, err
	}
	scriptClass, _, _, err := txscript.ExtractPkScriptAddrs(pkScriptByte, net)
	if err != nil {
		return txscript.NonStandardTy, err
	}
	return scriptClass, nil
}
