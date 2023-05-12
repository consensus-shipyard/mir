package contextstorepbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	contextstorepb "github.com/filecoin-project/mir/pkg/pb/contextstorepb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Origin struct {
	ItemID uint64
}

func OriginFromPb(pb *contextstorepb.Origin) *Origin {
	if pb == nil {
		return nil
	}
	return &Origin{
		ItemID: pb.ItemID,
	}
}

func (m *Origin) Pb() *contextstorepb.Origin {
	if m == nil {
		return nil
	}
	return &contextstorepb.Origin{
		ItemID: m.ItemID,
	}
}

func (*Origin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*contextstorepb.Origin]()}
}
