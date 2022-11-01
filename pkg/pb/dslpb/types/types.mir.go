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
	return &Origin{
		ContextID: pb.ContextID,
	}
}

func (m *Origin) Pb() *dslpb.Origin {
	return &dslpb.Origin{
		ContextID: m.ContextID,
	}
}

func (*Origin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*dslpb.Origin]()}
}
