package chkpvalidator

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	cvpbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/chkpvalidatorpb/dsl"
	cvpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/chkpvalidatorpb/types"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids.
type ModuleConfig struct {
	Self t.ModuleID
}

// NewModule returns a (passive) ChkpValidator module.
func NewModule(mc ModuleConfig, cv ChkpValidator) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	cvpbdsl.UponValidateCheckpoint(m, func(checkpoint *checkpointpbtypes.StableCheckpoint, epochNr types.EpochNr, memberships []*trantorpbtypes.Membership, origin *cvpbtypes.ValidateChkpOrigin) error {
		err := cv.Verify(checkpoint, epochNr, memberships)
		cvpbdsl.CheckpointValidated(m, origin.Module, err, origin)
		return nil
	})

	return m
}
