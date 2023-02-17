package checkpointpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type HashOrigin struct{}

func HashOriginFromPb(pb *checkpointpb.HashOrigin) *HashOrigin {
	return &HashOrigin{}
}

func (m *HashOrigin) Pb() *checkpointpb.HashOrigin {
	return &checkpointpb.HashOrigin{}
}

func (*HashOrigin) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*checkpointpb.HashOrigin]()}
}
