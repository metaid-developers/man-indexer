package man

import (
	"errors"
	"fmt"
	"manindexer/pin"
	"strings"
)

var compliantPath map[string]struct{}

func init() {
	arr := strings.Split(pin.CompliantPath, ";")
	compliantPath = make(map[string]struct{}, len(arr))
	for _, path := range arr {
		compliantPath[path] = struct{}{}
	}
}
func validator(pinode *pin.PinInscription) (err error) {
	limitCheck := false
	for _, limit := range OptionLimit {
		if pinode.Operation == limit {
			limitCheck = true
			break
		}
	}
	if !limitCheck {
		err = fmt.Errorf("option %s error", pinode.Operation)
		return
	}
	switch pinode.Operation {
	case "modify":
		if len(pinode.Path) <= 1 || pinode.Path[0:1] != "@" {
			err = errors.New("modify path error")
			return
		}
	case "revoke":
		if len(pinode.Path) <= 1 || pinode.Path[0:1] != "@" {
			err = errors.New("revoke path error")
			return
		}
	case "create":
		pathArr := strings.Split(pinode.Path, "/")
		if len(pinode.Path) < 1 || len(pathArr) < 2 {
			err = errors.New("create path length error")
			return
		}
		if _, ok := compliantPath[pathArr[1]]; !ok {
			err = errors.New("root path  error")
			return
		}
	case "init":
		if pinode.Path != "/" {
			err = errors.New("init operation path  error")
			return
		}
	}
	return
}
