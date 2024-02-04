package blockchain

import (
	"math"
	"strconv"
	"time"

	"github.com/filecoin-project/mir/pkg/blockchain/bcm"
	"github.com/filecoin-project/mir/pkg/blockchain/broadcast"
	"github.com/filecoin-project/mir/pkg/blockchain/miner"
	"github.com/filecoin-project/mir/pkg/blockchain/synchronizer"
	"github.com/filecoin-project/mir/pkg/eventmangler"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/timer"
	t "github.com/filecoin-project/mir/pkg/types"
)

type System struct {
	modules modules.Modules
}

func New(
	ownId t.NodeID,
	application modules.Module,
	disableMangle bool,
	dropRate float64,
	minDelay float64,
	maxDelay float64,
	numberOfNodes int,
	logger logging.Logger,
	transportLogger logging.Logger,
) *System {
	// determine "other" nodes for this node
	nodes := make(map[t.NodeID]*trantorpbtypes.NodeIdentity, numberOfNodes)
	allNodeIds := make([]t.NodeID, numberOfNodes)
	otherNodes := make([]t.NodeID, numberOfNodes-1)
	otherNodesIndex := 0
	for i := 0; i < numberOfNodes; i++ {
		nodeIdStr := strconv.Itoa(i)
		nodeId := t.NodeID(nodeIdStr)
		allNodeIds[i] = nodeId
		if nodeId != ownId {
			otherNodes[otherNodesIndex] = nodeId
			otherNodesIndex++
		}
		nodes[nodeId] = &trantorpbtypes.NodeIdentity{Id: nodeId, Addr: "/ip4/127.0.0.1/tcp/1000" + nodeIdStr, Key: nil, Weight: "1"}
	}
	membership := &trantorpbtypes.Membership{Nodes: nodes}

	// Instantiate network transport module and establish connections.
	transport, err := grpc.NewTransport(ownId, membership.Nodes[ownId].Addr, transportLogger)
	if err != nil {
		panic(err)
	}
	if err := transport.Start(); err != nil {
		panic(err)
	}
	transport.Connect(membership)

	modules := modules.Modules{
		"transport":    transport,
		"bcm":          bcm.NewBCM(logging.Decorate(logger, "BCM:\t")),
		"miner":        miner.NewMiner(ownId, 0.2, logging.Decorate(logger, "Miner:\t")),
		"broadcast":    broadcast.NewBroadcast(otherNodes, !disableMangle, logging.Decorate(logger, "Comm:\t")),
		"synchronizer": synchronizer.NewSynchronizer(ownId, otherNodes, logging.Decorate(logger, "Sync:\t")),
		"devnull":      modules.NullPassive{}, // for messages that are actually destined for the interceptor
		"application":  application,
	}

	if !disableMangle {
		minDelayDuration := time.Duration(int64(math.Round(minDelay * float64(time.Second))))
		maxDelayDuration := time.Duration(int64(math.Round(maxDelay * float64(time.Second))))
		manglerModule, err := eventmangler.NewModule(
			eventmangler.ModuleConfig{Self: "mangler", Dest: "transport", Timer: "timer"},
			&eventmangler.ModuleParams{MinDelay: minDelayDuration, MaxDelay: maxDelayDuration, DropRate: float32(dropRate)},
		)
		if err != nil {
			panic(err)
		}

		modules["timer"] = timer.New()
		modules["mangler"] = manglerModule
	}

	return &System{modules: modules}
}

func (s *System) Modules() modules.Modules {
	return s.modules
}
