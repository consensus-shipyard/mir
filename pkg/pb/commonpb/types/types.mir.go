package commonpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	commonpb "github.com/filecoin-project/mir/pkg/pb/commonpb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type HashData struct {
	Data [][]uint8
}

func HashDataFromPb(pb *commonpb.HashData) *HashData {
	return &HashData{
		Data: pb.Data,
	}
}

func (m *HashData) Pb() *commonpb.HashData {
	return &commonpb.HashData{
		Data: m.Data,
	}
}

func (*HashData) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.HashData]()}
}

type ClientProgress struct {
	Progress map[string]*commonpb.DeliveredReqs
}

func ClientProgressFromPb(pb *commonpb.ClientProgress) *ClientProgress {
	return &ClientProgress{
		Progress: pb.Progress,
	}
}

func (m *ClientProgress) Pb() *commonpb.ClientProgress {
	return &commonpb.ClientProgress{
		Progress: m.Progress,
	}
}

func (*ClientProgress) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*commonpb.ClientProgress]()}
}
