package example

import (
	"manindexer/common"
	"testing"
)

func TestMetaIDGenerate(t *testing.T) {
	address := "1FtUEic4XeMoTMaAfTQqrSZdoAiCAGHoa4"
	metaId := common.GetMetaIdByAddress(address)
	t.Logf("MetaId: %s", metaId)
}
