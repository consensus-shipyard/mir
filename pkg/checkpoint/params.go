package checkpoint

import (
	"time"

	"google.golang.org/protobuf/proto"

	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/stdtypes"
)

// ModuleParams represents the state associated with a single instance of the checkpoint protocol
// (establishing a single stable checkpoint).
type ModuleParams struct {

	// The ID of the node executing this instance of the protocol.
	OwnID t.NodeID

	// The IDs of nodes to execute this instance of the checkpoint protocol.
	// Note that it is the Membership of Epoch e-1 that constructs the Membership for Epoch e.
	// (As the starting checkpoint for e is the "finishing" checkpoint for e-1.)
	Membership *trantorpbtypes.Membership

	// EpochConfig to which this checkpoint belongs
	// It contains:.
	// - the Epoch the checkpoint's associated sequence number (SeqNr) is part of.
	// - Sequence number associated with this checkpoint protocol instance.
	//	 This checkpoint encompasses SeqNr sequence numbers,
	//	 i.e., SeqNr is the first sequence number *not* encompassed by this checkpoint.
	//	 One can imagine that the checkpoint represents the state of the system just before SeqNr,
	//	 i.e., "between" SeqNr-1 and SeqNr.
	//   among others
	EpochConfig *trantorpbtypes.EpochConfig

	// LeaderPolicy serialization data.
	LeaderPolicyData []byte

	// Time interval for repeated retransmission of checkpoint messages.
	ResendPeriod time.Duration
}

func (mp *ModuleParams) ToBytes() ([]byte, error) {
	params := checkpointpbtypes.InstanceParams{
		Membership:       mp.Membership,
		ResendPeriod:     mp.ResendPeriod,
		LeaderPolicyData: mp.LeaderPolicyData,
		EpochConfig:      mp.EpochConfig,
	}

	return proto.Marshal(params.Pb())
}
