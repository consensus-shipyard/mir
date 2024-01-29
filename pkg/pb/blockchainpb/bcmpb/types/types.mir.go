// Code generated by Mir codegen. DO NOT EDIT.

package bcmpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb"
	types3 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb/types"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	types2 "github.com/filecoin-project/mir/pkg/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() bcmpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb bcmpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *bcmpb.Event_NewBlock:
		return &Event_NewBlock{NewBlock: NewBlockFromPb(pb.NewBlock)}
	case *bcmpb.Event_NewChain:
		return &Event_NewChain{NewChain: NewChainFromPb(pb.NewChain)}
	case *bcmpb.Event_GetBlockRequest:
		return &Event_GetBlockRequest{GetBlockRequest: GetBlockRequestFromPb(pb.GetBlockRequest)}
	case *bcmpb.Event_GetBlockResponse:
		return &Event_GetBlockResponse{GetBlockResponse: GetBlockResponseFromPb(pb.GetBlockResponse)}
	case *bcmpb.Event_GetChainRequest:
		return &Event_GetChainRequest{GetChainRequest: GetChainRequestFromPb(pb.GetChainRequest)}
	case *bcmpb.Event_GetChainResponse:
		return &Event_GetChainResponse{GetChainResponse: GetChainResponseFromPb(pb.GetChainResponse)}
	case *bcmpb.Event_GetHeadToCheckpointChainRequest:
		return &Event_GetHeadToCheckpointChainRequest{GetHeadToCheckpointChainRequest: GetHeadToCheckpointChainRequestFromPb(pb.GetHeadToCheckpointChainRequest)}
	case *bcmpb.Event_GetHeadToCheckpointChainResponse:
		return &Event_GetHeadToCheckpointChainResponse{GetHeadToCheckpointChainResponse: GetHeadToCheckpointChainResponseFromPb(pb.GetHeadToCheckpointChainResponse)}
	case *bcmpb.Event_RegisterCheckpoint:
		return &Event_RegisterCheckpoint{RegisterCheckpoint: RegisterCheckpointFromPb(pb.RegisterCheckpoint)}
	case *bcmpb.Event_InitBlockchain:
		return &Event_InitBlockchain{InitBlockchain: InitBlockchainFromPb(pb.InitBlockchain)}
	}
	return nil
}

type Event_NewBlock struct {
	NewBlock *NewBlock
}

func (*Event_NewBlock) isEvent_Type() {}

func (w *Event_NewBlock) Unwrap() *NewBlock {
	return w.NewBlock
}

func (w *Event_NewBlock) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.NewBlock == nil {
		return &bcmpb.Event_NewBlock{}
	}
	return &bcmpb.Event_NewBlock{NewBlock: (w.NewBlock).Pb()}
}

func (*Event_NewBlock) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_NewBlock]()}
}

type Event_NewChain struct {
	NewChain *NewChain
}

func (*Event_NewChain) isEvent_Type() {}

func (w *Event_NewChain) Unwrap() *NewChain {
	return w.NewChain
}

func (w *Event_NewChain) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.NewChain == nil {
		return &bcmpb.Event_NewChain{}
	}
	return &bcmpb.Event_NewChain{NewChain: (w.NewChain).Pb()}
}

func (*Event_NewChain) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_NewChain]()}
}

type Event_GetBlockRequest struct {
	GetBlockRequest *GetBlockRequest
}

func (*Event_GetBlockRequest) isEvent_Type() {}

func (w *Event_GetBlockRequest) Unwrap() *GetBlockRequest {
	return w.GetBlockRequest
}

func (w *Event_GetBlockRequest) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.GetBlockRequest == nil {
		return &bcmpb.Event_GetBlockRequest{}
	}
	return &bcmpb.Event_GetBlockRequest{GetBlockRequest: (w.GetBlockRequest).Pb()}
}

func (*Event_GetBlockRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_GetBlockRequest]()}
}

type Event_GetBlockResponse struct {
	GetBlockResponse *GetBlockResponse
}

func (*Event_GetBlockResponse) isEvent_Type() {}

func (w *Event_GetBlockResponse) Unwrap() *GetBlockResponse {
	return w.GetBlockResponse
}

func (w *Event_GetBlockResponse) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.GetBlockResponse == nil {
		return &bcmpb.Event_GetBlockResponse{}
	}
	return &bcmpb.Event_GetBlockResponse{GetBlockResponse: (w.GetBlockResponse).Pb()}
}

func (*Event_GetBlockResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_GetBlockResponse]()}
}

type Event_GetChainRequest struct {
	GetChainRequest *GetChainRequest
}

func (*Event_GetChainRequest) isEvent_Type() {}

func (w *Event_GetChainRequest) Unwrap() *GetChainRequest {
	return w.GetChainRequest
}

func (w *Event_GetChainRequest) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.GetChainRequest == nil {
		return &bcmpb.Event_GetChainRequest{}
	}
	return &bcmpb.Event_GetChainRequest{GetChainRequest: (w.GetChainRequest).Pb()}
}

func (*Event_GetChainRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_GetChainRequest]()}
}

type Event_GetChainResponse struct {
	GetChainResponse *GetChainResponse
}

func (*Event_GetChainResponse) isEvent_Type() {}

func (w *Event_GetChainResponse) Unwrap() *GetChainResponse {
	return w.GetChainResponse
}

func (w *Event_GetChainResponse) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.GetChainResponse == nil {
		return &bcmpb.Event_GetChainResponse{}
	}
	return &bcmpb.Event_GetChainResponse{GetChainResponse: (w.GetChainResponse).Pb()}
}

func (*Event_GetChainResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_GetChainResponse]()}
}

type Event_GetHeadToCheckpointChainRequest struct {
	GetHeadToCheckpointChainRequest *GetHeadToCheckpointChainRequest
}

func (*Event_GetHeadToCheckpointChainRequest) isEvent_Type() {}

func (w *Event_GetHeadToCheckpointChainRequest) Unwrap() *GetHeadToCheckpointChainRequest {
	return w.GetHeadToCheckpointChainRequest
}

func (w *Event_GetHeadToCheckpointChainRequest) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.GetHeadToCheckpointChainRequest == nil {
		return &bcmpb.Event_GetHeadToCheckpointChainRequest{}
	}
	return &bcmpb.Event_GetHeadToCheckpointChainRequest{GetHeadToCheckpointChainRequest: (w.GetHeadToCheckpointChainRequest).Pb()}
}

func (*Event_GetHeadToCheckpointChainRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_GetHeadToCheckpointChainRequest]()}
}

type Event_GetHeadToCheckpointChainResponse struct {
	GetHeadToCheckpointChainResponse *GetHeadToCheckpointChainResponse
}

func (*Event_GetHeadToCheckpointChainResponse) isEvent_Type() {}

func (w *Event_GetHeadToCheckpointChainResponse) Unwrap() *GetHeadToCheckpointChainResponse {
	return w.GetHeadToCheckpointChainResponse
}

func (w *Event_GetHeadToCheckpointChainResponse) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.GetHeadToCheckpointChainResponse == nil {
		return &bcmpb.Event_GetHeadToCheckpointChainResponse{}
	}
	return &bcmpb.Event_GetHeadToCheckpointChainResponse{GetHeadToCheckpointChainResponse: (w.GetHeadToCheckpointChainResponse).Pb()}
}

func (*Event_GetHeadToCheckpointChainResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_GetHeadToCheckpointChainResponse]()}
}

type Event_RegisterCheckpoint struct {
	RegisterCheckpoint *RegisterCheckpoint
}

func (*Event_RegisterCheckpoint) isEvent_Type() {}

func (w *Event_RegisterCheckpoint) Unwrap() *RegisterCheckpoint {
	return w.RegisterCheckpoint
}

func (w *Event_RegisterCheckpoint) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.RegisterCheckpoint == nil {
		return &bcmpb.Event_RegisterCheckpoint{}
	}
	return &bcmpb.Event_RegisterCheckpoint{RegisterCheckpoint: (w.RegisterCheckpoint).Pb()}
}

func (*Event_RegisterCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_RegisterCheckpoint]()}
}

type Event_InitBlockchain struct {
	InitBlockchain *InitBlockchain
}

func (*Event_InitBlockchain) isEvent_Type() {}

func (w *Event_InitBlockchain) Unwrap() *InitBlockchain {
	return w.InitBlockchain
}

func (w *Event_InitBlockchain) Pb() bcmpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.InitBlockchain == nil {
		return &bcmpb.Event_InitBlockchain{}
	}
	return &bcmpb.Event_InitBlockchain{InitBlockchain: (w.InitBlockchain).Pb()}
}

func (*Event_InitBlockchain) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event_InitBlockchain]()}
}

func EventFromPb(pb *bcmpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *bcmpb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.Event{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.Event]()}
}

type NewBlock struct {
	Block *types.Block
}

func NewBlockFromPb(pb *bcmpb.NewBlock) *NewBlock {
	if pb == nil {
		return nil
	}
	return &NewBlock{
		Block: types.BlockFromPb(pb.Block),
	}
}

func (m *NewBlock) Pb() *bcmpb.NewBlock {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.NewBlock{}
	{
		if m.Block != nil {
			pbMessage.Block = (m.Block).Pb()
		}
	}

	return pbMessage
}

func (*NewBlock) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.NewBlock]()}
}

type NewChain struct {
	Blocks []*types.Block
}

func NewChainFromPb(pb *bcmpb.NewChain) *NewChain {
	if pb == nil {
		return nil
	}
	return &NewChain{
		Blocks: types1.ConvertSlice(pb.Blocks, func(t *blockchainpb.Block) *types.Block {
			return types.BlockFromPb(t)
		}),
	}
}

func (m *NewChain) Pb() *bcmpb.NewChain {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.NewChain{}
	{
		pbMessage.Blocks = types1.ConvertSlice(m.Blocks, func(t *types.Block) *blockchainpb.Block {
			return (t).Pb()
		})
	}

	return pbMessage
}

func (*NewChain) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.NewChain]()}
}

type GetBlockRequest struct {
	RequestId    string
	SourceModule types2.ModuleID
	BlockId      uint64
}

func GetBlockRequestFromPb(pb *bcmpb.GetBlockRequest) *GetBlockRequest {
	if pb == nil {
		return nil
	}
	return &GetBlockRequest{
		RequestId:    pb.RequestId,
		SourceModule: (types2.ModuleID)(pb.SourceModule),
		BlockId:      pb.BlockId,
	}
}

func (m *GetBlockRequest) Pb() *bcmpb.GetBlockRequest {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.GetBlockRequest{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.SourceModule = (string)(m.SourceModule)
		pbMessage.BlockId = m.BlockId
	}

	return pbMessage
}

func (*GetBlockRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.GetBlockRequest]()}
}

type GetBlockResponse struct {
	RequestId string
	Found     bool
	Block     *types.Block
}

func GetBlockResponseFromPb(pb *bcmpb.GetBlockResponse) *GetBlockResponse {
	if pb == nil {
		return nil
	}
	return &GetBlockResponse{
		RequestId: pb.RequestId,
		Found:     pb.Found,
		Block:     types.BlockFromPb(pb.Block),
	}
}

func (m *GetBlockResponse) Pb() *bcmpb.GetBlockResponse {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.GetBlockResponse{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.Found = m.Found
		if m.Block != nil {
			pbMessage.Block = (m.Block).Pb()
		}
	}

	return pbMessage
}

func (*GetBlockResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.GetBlockResponse]()}
}

type GetChainRequest struct {
	RequestId      string
	SourceModule   types2.ModuleID
	EndBlockId     uint64
	SourceBlockIds []uint64
}

func GetChainRequestFromPb(pb *bcmpb.GetChainRequest) *GetChainRequest {
	if pb == nil {
		return nil
	}
	return &GetChainRequest{
		RequestId:      pb.RequestId,
		SourceModule:   (types2.ModuleID)(pb.SourceModule),
		EndBlockId:     pb.EndBlockId,
		SourceBlockIds: pb.SourceBlockIds,
	}
}

func (m *GetChainRequest) Pb() *bcmpb.GetChainRequest {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.GetChainRequest{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.SourceModule = (string)(m.SourceModule)
		pbMessage.EndBlockId = m.EndBlockId
		pbMessage.SourceBlockIds = m.SourceBlockIds
	}

	return pbMessage
}

func (*GetChainRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.GetChainRequest]()}
}

type GetChainResponse struct {
	RequestId string
	Success   bool
	Chain     []*types.Block
}

func GetChainResponseFromPb(pb *bcmpb.GetChainResponse) *GetChainResponse {
	if pb == nil {
		return nil
	}
	return &GetChainResponse{
		RequestId: pb.RequestId,
		Success:   pb.Success,
		Chain: types1.ConvertSlice(pb.Chain, func(t *blockchainpb.Block) *types.Block {
			return types.BlockFromPb(t)
		}),
	}
}

func (m *GetChainResponse) Pb() *bcmpb.GetChainResponse {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.GetChainResponse{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.Success = m.Success
		pbMessage.Chain = types1.ConvertSlice(m.Chain, func(t *types.Block) *blockchainpb.Block {
			return (t).Pb()
		})
	}

	return pbMessage
}

func (*GetChainResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.GetChainResponse]()}
}

type GetHeadToCheckpointChainRequest struct {
	RequestId    string
	SourceModule types2.ModuleID
}

func GetHeadToCheckpointChainRequestFromPb(pb *bcmpb.GetHeadToCheckpointChainRequest) *GetHeadToCheckpointChainRequest {
	if pb == nil {
		return nil
	}
	return &GetHeadToCheckpointChainRequest{
		RequestId:    pb.RequestId,
		SourceModule: (types2.ModuleID)(pb.SourceModule),
	}
}

func (m *GetHeadToCheckpointChainRequest) Pb() *bcmpb.GetHeadToCheckpointChainRequest {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.GetHeadToCheckpointChainRequest{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.SourceModule = (string)(m.SourceModule)
	}

	return pbMessage
}

func (*GetHeadToCheckpointChainRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.GetHeadToCheckpointChainRequest]()}
}

type GetHeadToCheckpointChainResponse struct {
	RequestId       string
	Chain           []*types.Block
	CheckpointState *types3.State
}

func GetHeadToCheckpointChainResponseFromPb(pb *bcmpb.GetHeadToCheckpointChainResponse) *GetHeadToCheckpointChainResponse {
	if pb == nil {
		return nil
	}
	return &GetHeadToCheckpointChainResponse{
		RequestId: pb.RequestId,
		Chain: types1.ConvertSlice(pb.Chain, func(t *blockchainpb.Block) *types.Block {
			return types.BlockFromPb(t)
		}),
		CheckpointState: types3.StateFromPb(pb.CheckpointState),
	}
}

func (m *GetHeadToCheckpointChainResponse) Pb() *bcmpb.GetHeadToCheckpointChainResponse {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.GetHeadToCheckpointChainResponse{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.Chain = types1.ConvertSlice(m.Chain, func(t *types.Block) *blockchainpb.Block {
			return (t).Pb()
		})
		if m.CheckpointState != nil {
			pbMessage.CheckpointState = (m.CheckpointState).Pb()
		}
	}

	return pbMessage
}

func (*GetHeadToCheckpointChainResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.GetHeadToCheckpointChainResponse]()}
}

type RegisterCheckpoint struct {
	BlockId uint64
	State   *types3.State
}

func RegisterCheckpointFromPb(pb *bcmpb.RegisterCheckpoint) *RegisterCheckpoint {
	if pb == nil {
		return nil
	}
	return &RegisterCheckpoint{
		BlockId: pb.BlockId,
		State:   types3.StateFromPb(pb.State),
	}
}

func (m *RegisterCheckpoint) Pb() *bcmpb.RegisterCheckpoint {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.RegisterCheckpoint{}
	{
		pbMessage.BlockId = m.BlockId
		if m.State != nil {
			pbMessage.State = (m.State).Pb()
		}
	}

	return pbMessage
}

func (*RegisterCheckpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.RegisterCheckpoint]()}
}

type InitBlockchain struct {
	InitialState *types3.State
}

func InitBlockchainFromPb(pb *bcmpb.InitBlockchain) *InitBlockchain {
	if pb == nil {
		return nil
	}
	return &InitBlockchain{
		InitialState: types3.StateFromPb(pb.InitialState),
	}
}

func (m *InitBlockchain) Pb() *bcmpb.InitBlockchain {
	if m == nil {
		return nil
	}
	pbMessage := &bcmpb.InitBlockchain{}
	{
		if m.InitialState != nil {
			pbMessage.InitialState = (m.InitialState).Pb()
		}
	}

	return pbMessage
}

func (*InitBlockchain) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*bcmpb.InitBlockchain]()}
}
