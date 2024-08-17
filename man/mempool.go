package man

import (
	"manindexer/mrc20"
)

type ManMempool struct{}

func (mm *ManMempool) CheckMempool(chainName string) {
	//check mrc20 native transaction
	list, err := ChainAdapter[chainName].GetMempoolTransactionList()
	if err != nil {
		return
	}
	mm.CheckMempoolHadle(chainName, list)
}
func (mm *ManMempool) CheckMempoolHadle(chainName string, list []interface{}) {
	pins, txInList := IndexerAdapter[chainName].CatchMempoolPins(list)
	mrc20TransferPinTx := make(map[string]struct{})
	var mrc20TrasferList []*mrc20.Mrc20Utxo
	var mrc20MintList []*mrc20.Mrc20Utxo
	mrc20Validator := Mrc20Validator{}

	for _, pinNode := range pins {
		err := ManValidator(pinNode)
		if err != nil {
			continue
		}
		//mrc20 pin
		if len(pinNode.Path) > 10 && pinNode.Path[0:10] == "/ft/mrc20/" {
			if pinNode.Path == "/ft/mrc20/transfer" {
				mrc20TransferPinTx[pinNode.GenesisTransaction] = struct{}{}
				transferPinList, _ := CreateMrc20TransferUtxo(pinNode, &mrc20Validator, true)
				if len(transferPinList) > 0 {
					mrc20TrasferList = append(mrc20TrasferList, transferPinList...)
				}
			}
			//else if pinNode.Path == "/ft/mrc20/mint" {
			// mintUtxo, err := CreateMrc20MintPin(pinNode, &mrc20Validator, true)
			// if err == nil {
			// 	mintUtxo.OperationTx = pinNode.GenesisTransaction
			// 	mrc20MintList = append(mrc20MintList, &mintUtxo)
			// }

			//}
		}
	}
	//check mrc20 native transaction
	mrc20transferCheck, err := DbAdapter.GetMrc20UtxoByOutPutList(txInList, true)
	if err == nil && len(mrc20transferCheck) > 0 {
		mrc20NativeTrasferList := IndexerAdapter[chainName].CatchMempoolNativeMrc20Transfer(list, mrc20transferCheck, mrc20TransferPinTx)
		if len(mrc20NativeTrasferList) > 0 {
			DbAdapter.UpdateMrc20Utxo(mrc20NativeTrasferList, true)
		}
	}
	if len(mrc20TrasferList) > 0 {
		DbAdapter.UpdateMrc20Utxo(mrc20TrasferList, true)
	}
	if len(mrc20MintList) > 0 {
		DbAdapter.UpdateMrc20Utxo(mrc20MintList, true)
	}

}
