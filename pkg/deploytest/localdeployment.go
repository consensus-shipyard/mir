package deploytest

import (
	"fmt"
	"strconv"

	t "github.com/filecoin-project/mir/pkg/types"
)

func LocalAddresses(nodeIDs []t.NodeID, basePort int) map[t.NodeID]string {
	addrs := make(map[t.NodeID]string)
	for _, i := range nodeIDs {
		p, err := strconv.Atoi(string(i))
		if err != nil {
			panic(fmt.Errorf("could not convert node ID: %w", err))
		}
		addrs[i] = fmt.Sprintf("127.0.0.1:%d", basePort+p)
	}
	return addrs
}
