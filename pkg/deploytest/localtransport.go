package deploytest

import (
	"github.com/filecoin-project/mir/pkg/net"
	t "github.com/filecoin-project/mir/pkg/types"
)

type LocalTransportLayer interface {
	Link(source t.NodeID) net.Transport
}
