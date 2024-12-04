package man

import (
	"manindexer/basicprotocols/metaaccess"
	"manindexer/pin"
)

type MetaAccess struct{}

var validator = MetaAccessValidator{}

func (ma *MetaAccess) PinHandle(pinList []*pin.PinInscription, mempool bool) {
	var controlList []*metaaccess.AccessControl
	var passList []*metaaccess.AccessPassData
	for _, pinNode := range pinList {
		switch pinNode.Path {
		case "/metaaccess/accesscontrol":
			data, err := ma.AccessControlHandle(pinNode)
			//fmt.Println(err)
			if err == nil {
				if mempool {
					data.Mempool = 1
				} else {
					data.Mempool = 0
				}
				controlList = append(controlList, &data)
			}
		case "/metaaccess/accesspass":
			data, err := ma.AccessPassHandle(pinNode)
			//fmt.Println(err)
			if err == nil {
				for _, item := range data {
					passList = append(passList, &item)
				}
			}
		}
	}
	if len(controlList) > 0 {
		DbAdapter.BatchSaveAccesscontrol(controlList)
	}
	if len(passList) > 0 {
		DbAdapter.BatchSaveAccessPass(passList)
	}
}
func (ma *MetaAccess) AccessControlHandle(pinNode *pin.PinInscription) (data metaaccess.AccessControl, err error) {
	return validator.AccessControl(pinNode)
}

func (ma *MetaAccess) AccessPassHandle(pinNode *pin.PinInscription) (data []metaaccess.AccessPassData, err error) {
	return validator.AccessPass(pinNode)
}
