package requestpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types1 "github.com/filecoin-project/mir/codegen/model/types"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
	types "github.com/filecoin-project/mir/pkg/types"
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

type HashedRequest struct {
	Req    *Request
	Digest []uint8
}

func HashedRequestFromPb(pb *requestpb.HashedRequest) *HashedRequest {
	return &HashedRequest{
		Req:    RequestFromPb(pb.Req),
		Digest: pb.Digest,
	}
}

func (m *HashedRequest) Pb() *requestpb.HashedRequest {
	return &requestpb.HashedRequest{
		Req:    (m.Req).Pb(),
		Digest: m.Digest,
	}
}

func (*HashedRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*requestpb.HashedRequest]()}
}

type Batch struct {
	Requests []*HashedRequest
}

func BatchFromPb(pb *requestpb.Batch) *Batch {
	return &Batch{
		Requests: types1.ConvertSlice(pb.Requests, func(t *requestpb.HashedRequest) *HashedRequest {
			return HashedRequestFromPb(t)
		}),
	}
}

func (m *Batch) Pb() *requestpb.Batch {
	return &requestpb.Batch{
		Requests: types1.ConvertSlice(m.Requests, func(t *HashedRequest) *requestpb.HashedRequest {
			return (t).Pb()
		}),
	}
}

func (*Batch) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*requestpb.Batch]()}
}
