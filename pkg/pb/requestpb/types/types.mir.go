package requestpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/trantor/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Request struct {
	ClientId types.ClientID
	ReqNo    types.ReqNo
	Type     uint64
	Data     []uint8
}

func RequestFromPb(pb *requestpb.Request) *Request {
	return &Request{
		ClientId: (types.ClientID)(pb.ClientId),
		ReqNo:    (types.ReqNo)(pb.ReqNo),
		Type:     pb.Type,
		Data:     pb.Data,
	}
}

func (m *Request) Pb() *requestpb.Request {
	return &requestpb.Request{
		ClientId: (string)(m.ClientId),
		ReqNo:    (uint64)(m.ReqNo),
		Type:     m.Type,
		Data:     m.Data,
	}
}

func (*Request) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*requestpb.Request]()}
}
