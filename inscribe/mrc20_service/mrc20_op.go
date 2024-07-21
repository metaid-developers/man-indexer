package mrc20_service

import (
	"github.com/btcsuite/btcd/chaincfg"
)

type Mrc20OpRequest struct {
	Net         *chaincfg.Params
	MetaIdFlag  string
	Op          string // mint, transfer
	OpPayload   string
	CommitUtxos []*CommitUtxo

	//deploy
	DeployPremineOutAddress string
	DeployPinOutAddress     string

	//mint
	MintPins            []*MintPin
	Mrc20OutValue       int64
	Mrc20OutAddressList []string

	//transfer
	TransferMrc20s []*TransferMrc20
	Mrc20Outs      []*Mrc20OutInfo
	ChangeAddress  string
}

type FetchCommitUtxoFunc func(needAmount int64) ([]*CommitUtxo, error)

func Mrc20Deploy(opRep *Mrc20OpRequest, feeRate int64, fetchUtxos FetchCommitUtxoFunc) (string, string, int64, error) {
	var (
		err          error
		mrc20Builder *Mrc20Builder
		fee          int64 = 0

		content                = opRep.OpPayload
		path                   = "/ft/mrc20/deploy"
		metaIdData *MetaIdData = &MetaIdData{
			MetaIDFlag:  opRep.MetaIdFlag,
			Operation:   "create",
			Path:        path,
			Content:     []byte(content),
			Encryption:  "",
			Version:     "",
			ContentType: "application/json",
		}
		commitTx, revealTx string = "", ""
	)
	mrc20Builder = &Mrc20Builder{
		Net:            opRep.Net,
		MetaIdData:     metaIdData,
		MintPins:       opRep.MintPins,
		TransferMrc20s: opRep.TransferMrc20s,
		FeeRate:        feeRate,
		op:             opRep.Op,

		mrc20OutValue: opRep.Mrc20OutValue,

		deployPinOutAddress:     opRep.DeployPinOutAddress,
		deployPremineOutAddress: opRep.DeployPremineOutAddress,
	}

	txCtxData, err := createMetaIdTxCtxData(opRep.Net, mrc20Builder.MetaIdData)
	if err != nil {
		return "", "", 0, err
	}
	mrc20Builder.TxCtxData = txCtxData

	err = mrc20Builder.buildEmptyRevealPsbt()
	if err != nil {
		return "", "", 0, err
	}
	fee = mrc20Builder.CalRevealPsbtFee(feeRate)

	commitUtxos, err := fetchUtxos(fee)
	if err != nil {
		return "", "", 0, err
	}

	err = mrc20Builder.buildCommitPsbt(commitUtxos)
	if err != nil {
		return "", "", 0, err
	}

	err = mrc20Builder.SignAll()
	if err != nil {
		return "", "", 0, err
	}

	commitTxHex, revealTxHex, err := mrc20Builder.ExtractAllPsbtTransaction()
	if err != nil {
		return "", "", 0, err
	}

	mrc20Builder.commitTxRaw = commitTxHex
	mrc20Builder.revealTxRaw = revealTxHex
	commitTx, revealTx, err = mrc20Builder.Inscribe()

	return commitTx, revealTx, fee, nil
}

func Mrc20Mint(opRep *Mrc20OpRequest, feeRate int64, fetchUtxos FetchCommitUtxoFunc) (string, string, int64, error) {
	var (
		err          error
		mrc20Builder *Mrc20Builder
		fee          int64 = 0

		content                = opRep.OpPayload
		path                   = "/ft/mrc20/mint"
		metaIdData *MetaIdData = &MetaIdData{
			MetaIDFlag:  opRep.MetaIdFlag,
			Operation:   "hide",
			Path:        path,
			Content:     []byte(content),
			Encryption:  "",
			Version:     "",
			ContentType: "application/json",
		}
		commitTx, revealTx string = "", ""
	)
	mrc20Builder = &Mrc20Builder{
		Net:         opRep.Net,
		MetaIdData:  metaIdData,
		CommitUtxos: opRep.CommitUtxos,
		MintPins:    opRep.MintPins,
		//TransferMrc20s: opRep.TransferMrc20s,
		FeeRate: feeRate,
		op:      opRep.Op,

		mrc20OutValue:       opRep.Mrc20OutValue,
		mrc20OutAddressList: opRep.Mrc20OutAddressList,
	}

	txCtxData, err := createMetaIdTxCtxData(opRep.Net, mrc20Builder.MetaIdData)
	if err != nil {
		return "", "", 0, err
	}
	mrc20Builder.TxCtxData = txCtxData

	err = mrc20Builder.buildEmptyRevealPsbt()
	if err != nil {
		return "", "", 0, err
	}
	fee = mrc20Builder.CalRevealPsbtFee(feeRate)

	commitUtxos, err := fetchUtxos(fee)
	if err != nil {
		return "", "", 0, err
	}

	err = mrc20Builder.buildCommitPsbt(commitUtxos)
	if err != nil {
		return "", "", 0, err
	}

	err = mrc20Builder.SignAll()
	if err != nil {
		return "", "", 0, err
	}

	commitTxHex, revealTxHex, err := mrc20Builder.ExtractAllPsbtTransaction()
	if err != nil {
		return "", "", 0, err
	}

	mrc20Builder.commitTxRaw = commitTxHex
	mrc20Builder.revealTxRaw = revealTxHex
	commitTx, revealTx, err = mrc20Builder.Inscribe()

	return commitTx, revealTx, fee, nil
}

func Mrc20Transfer(opRep *Mrc20OpRequest, feeRate int64, fetchUtxos FetchCommitUtxoFunc) (string, string, int64, error) {
	var (
		err          error
		mrc20Builder *Mrc20Builder
		fee          int64 = 0

		content                = opRep.OpPayload
		path                   = "/ft/mrc20/transfer"
		metaIdData *MetaIdData = &MetaIdData{
			MetaIDFlag:  opRep.MetaIdFlag,
			Operation:   "hide",
			Path:        path,
			Content:     []byte(content),
			Encryption:  "",
			Version:     "",
			ContentType: "application/json",
		}

		commitTx, revealTx string = "", ""
	)
	mrc20Builder = &Mrc20Builder{
		Net:                opRep.Net,
		MetaIdData:         metaIdData,
		CommitUtxos:        opRep.CommitUtxos,
		TransferMrc20s:     opRep.TransferMrc20s,
		Mrc20Outs:          opRep.Mrc20Outs,
		FeeRate:            feeRate,
		op:                 opRep.Op,
		mrc20ChangeAddress: opRep.ChangeAddress,

		//mrc20OutValue:       opRep.Mrc20OutValue,
		//mrc20OutAddressList: opRep.Mrc20OutAddressList,
	}

	txCtxData, err := createMetaIdTxCtxData(opRep.Net, mrc20Builder.MetaIdData)
	if err != nil {
		return "", "", 0, err
	}
	mrc20Builder.TxCtxData = txCtxData

	err = mrc20Builder.buildEmptyRevealPsbt()
	if err != nil {
		return "", "", 0, err
	}
	fee = mrc20Builder.CalRevealPsbtFee(feeRate)

	commitUtxos, err := fetchUtxos(fee)
	if err != nil {
		return "", "", 0, err
	}

	err = mrc20Builder.buildCommitPsbt(commitUtxos)
	if err != nil {
		return "", "", 0, err
	}

	err = mrc20Builder.SignAll()
	if err != nil {
		return "", "", 0, err
	}

	commitTxHex, revealTxHex, err := mrc20Builder.ExtractAllPsbtTransaction()
	if err != nil {
		return "", "", 0, err
	}

	mrc20Builder.commitTxRaw = commitTxHex
	mrc20Builder.revealTxRaw = revealTxHex

	commitTx, revealTx, err = mrc20Builder.Inscribe()

	return commitTx, revealTx, fee, nil
}
