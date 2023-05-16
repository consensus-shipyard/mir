package dslpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	dslpb "github.com/filecoin-project/mir/pkg/pb/dslpb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Origin struct {
	ContextID uint64
}

func OriginFromPb(pb *dslpb.Origin) *Origin {
	if pb == nil {
		return nil
	}
	return &Origin{
		ContextID: pb.ContextID,
	}
}

func (m *Origin) Pb() *dslpb.Origin {
	if m == nil {
		return nil
	}
	pbMessage := &dslpb.Origin{}
	{
		pbMessage.ContextID = m.ContextID
	}

	return pbMessage
}

func (*Origin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*dslpb.Origin]()}
}
