package factorymodule

import (
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/factorymodulepb"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (
	DefaultMsgBufSize = 1024 * 1024 // 1 MB
)

type ModuleGenerator func(id t.ModuleID, params *factorymodulepb.GeneratorParams) (modules.PassiveModule, error)

type ModuleParams struct {
	Generator  ModuleGenerator
	MsgBufSize int
}

func DefaultParams(generator ModuleGenerator) ModuleParams {
	return ModuleParams{
		Generator:  generator,
		MsgBufSize: DefaultMsgBufSize,
	}
}
