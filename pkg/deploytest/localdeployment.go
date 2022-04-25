package deploytest

import (
	"fmt"

	t "github.com/filecoin-project/mir/pkg/types"
)

func LocalAddresses(nodeIDs []t.NodeID, basePort int) map[t.NodeID]string {
	addrs := make(map[t.NodeID]string)
	for i := range nodeIDs {
		addrs[t.NewNodeIDFromInt(i)] = fmt.Sprintf("127.0.0.1:%d", basePort+i)
	}
	return addrs
}
