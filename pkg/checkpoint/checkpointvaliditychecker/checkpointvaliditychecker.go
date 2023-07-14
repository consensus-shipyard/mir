package checkpointvaliditychecker

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	cvcpbdsl "github.com/filecoin-project/mir/pkg/pb/checkpointpb/checkpointvaliditycheckerpb/dsl"
	cvcpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/checkpointvaliditycheckerpb/types"
	checkpointpbtypes "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// ModuleConfig sets the module ids.
type ModuleConfig struct {
	Self t.ModuleID
}

// NewModule returns a passive module for the CheckpointValidityChecker module.
func NewModule(mc ModuleConfig, cvc CVC) modules.PassiveModule {
	m := dsl.NewModule(mc.Self)

	cvcpbdsl.UponValidateCheckpoint(m, func(checkpoint *checkpointpbtypes.StableCheckpoint, epochNr types.EpochNr, memberships []*trantorpbtypes.Membership, origin *cvcpbtypes.ValidateChkpOrigin) error {
		err := cvc.Verify(checkpoint, epochNr, memberships)
		cvcpbdsl.CheckpointValidated(m, origin.Module, err, origin)
		return nil
	})

	return m
}
