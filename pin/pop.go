package pin

import (
	"manindexer/common"
	"strings"
)

func PopLevelCount(chainName, pop string) (lv int, lastStr string) {
	PopCutNum := 1000
	switch chainName {
	case "btc":
		PopCutNum = common.Config.Btc.PopCutNum
	case "mvc":
		PopCutNum = common.Config.Mvc.PopCutNum
	}
	cnt := len(pop) - len(strings.TrimLeft(pop, "0"))
	if cnt <= PopCutNum {
		lv = -1
		lastStr = pop[PopCutNum:]
		return
	} else {
		lv = cnt - PopCutNum
		lastStr = pop[PopCutNum:]
		return
	}
}
