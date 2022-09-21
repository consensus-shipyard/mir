package libp2p

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/types"
)

func TestLibp2pReconnect(t *testing.T) {
	nodeIDs := []types.NodeID{types.NewNodeIDFromInt(1), types.NewNodeIDFromInt(2)}
	logger := logging.ConsoleDebugLogger

	ctx := context.Background()

	tr := NewLocalLibp2pTransport(nodeIDs, logger, 10000)

	a, err := tr.Link(nodeIDs[0])
	require.NoError(t, err)

	err = a.Start()
	require.NoError(t, err)

	b, err := tr.Link(nodeIDs[1])
	require.NoError(t, err)

	err = b.Start()
	require.NoError(t, err)

	a.syncConnect(ctx, tr.Nodes())

	msg1 := messagepb.Message{}
	err = a.Send(nodeIDs[1], &msg1)
	require.NoError(t, err)

	err = b.host.Network().ClosePeer(a.host.ID())
	require.NoError(t, err)

	err = b.host.Network().ClosePeer(a.host.ID())
	require.NoError(t, err)

	fmt.Println(a.host.Network().ConnsToPeer(b.host.ID()))

	n := len(b.host.Network().Peers())
	require.Equal(t, 0, n)

	time.Sleep(5 * time.Second)

	err = a.Send(nodeIDs[1], &msg1)
	require.Error(t, err)

	a.Stop()
	b.Stop()
}
