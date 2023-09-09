package granite

import (
	granitepbtypes "github.com/filecoin-project/mir/pkg/pb/granitepb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type RoundNr uint64

// InstanceUID is used to uniquely identify an instance of Granite.
// It is used to prevent cross-instance signature replay attack and should be unique across all executions.
type InstanceUID uint64

type StepMsgStore[Value any] map[RoundNr]map[t.NodeID]Value

func (sms StepMsgStore[Value]) EnsureRound(roundNr RoundNr) {
	if _, ok := sms[roundNr]; !ok {
		sms[roundNr] = make(map[t.NodeID]Value)
	}
}

func (sms StepMsgStore[Value]) StoreMessage(roundNr RoundNr, nodeID t.NodeID, message Value) {
	sms.EnsureRound(roundNr)
	sms[roundNr][nodeID] = message
}

func (sms StepMsgStore[Value]) Get(round RoundNr, nodeID t.NodeID) (Value, bool) {
	if nodes, ok := sms[round]; ok {
		if val, ok := nodes[nodeID]; ok {
			return val, ok
		}
	}

	var val Value
	return val, false
}

type MsgStore struct {
	//Storing all messages as it is relevant for validation
	Msgs map[MsgType]StepMsgStore[*granitepbtypes.ConsensusMsg]
}

func NewMsgStore() *MsgStore {
	msgStore := make(map[MsgType]StepMsgStore[*granitepbtypes.ConsensusMsg])
	msgStore[CONVERGE] = make(StepMsgStore[*granitepbtypes.ConsensusMsg])
	msgStore[PROPOSE] = make(StepMsgStore[*granitepbtypes.ConsensusMsg])
	msgStore[PREPARE] = make(StepMsgStore[*granitepbtypes.ConsensusMsg])
	msgStore[COMMIT] = make(StepMsgStore[*granitepbtypes.ConsensusMsg])

	return &MsgStore{
		Msgs: msgStore,
	}
}

// Get method for the MsgStore
func (mst *MsgStore) Get(msgType MsgType, round RoundNr, nodeID t.NodeID) (*granitepbtypes.ConsensusMsg, bool) {
	if msg, ok := mst.Msgs[msgType].Get(round, nodeID); ok {
		return msg, ok
	}
	return nil, false
}

func (mst *MsgStore) StoreMessage(msgType MsgType, round RoundNr, nodeID t.NodeID, message *granitepbtypes.ConsensusMsg) {
	mst.Msgs[msgType].StoreMessage(round, nodeID, message)
}

func (mst *MsgStore) RemoveMessage(msgType MsgType, round RoundNr, nodeID t.NodeID, key string) {
	delete(mst.Msgs[msgType][round], nodeID)
}

type MsgType int32 //Protos do not support uint8 but this should be uint8

const (
	CONVERGE MsgType = iota
	PROPOSE
	PREPARE
	COMMIT
)
