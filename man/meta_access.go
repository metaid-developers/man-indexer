package man

import (
	"manindexer/metaaccess"
	"manindexer/pin"
)

type MetaAccess struct{}

var validator = MetaAccessValidator{}

func (ma *MetaAccess) PinHandle(pinList []*pin.PinInscription) {
	var controlList []*metaaccess.AccessControl
	var passList []*metaaccess.AccessPassData
	for _, pinNode := range pinList {
		switch pinNode.Path {
		case "/protocols/metaaccess/accesscontrol":
			data, err := ma.AccessControlHandle(pinNode)
			if err == nil {
				controlList = append(controlList, &data)
			}
		case "/protocols/metaaccess/accesspass":
			data, err := ma.AccessPassHandle(pinNode)
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
