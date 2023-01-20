package emptycert

import (
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool/emptybatchid"
	mscpbtypes "github.com/filecoin-project/mir/pkg/pb/availabilitypb/mscpb/types"
)

func EmptyCert() *mscpbtypes.Cert {
	return &mscpbtypes.Cert{
		BatchId: emptybatchid.EmptyBatchID(),
	}
}

func IsEmpty(cert *mscpbtypes.Cert) bool {
	return emptybatchid.IsEmptyBatchID(cert.BatchId)
}
