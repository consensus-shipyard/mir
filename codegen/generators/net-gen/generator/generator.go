package generator

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/messages"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type NetMsgGenerator struct{}

func (NetMsgGenerator) Run(pbGoStructTypes []reflect.Type) error {
	rootMessages, err := GetNetMsgHierarchyRootMessages(pbGoStructTypes)
	if err != nil {
		return err
	}

	for _, rootMessage := range rootMessages {
		parser := messages.DefaultParser()

		root, err := parser.ParseNetMessageHierarchy(rootMessage)
		if err != nil {
			return err
		}

		err = GenerateMessageConstructors(root)
		if err != nil {
			return fmt.Errorf("error generating net message constructors: %w", err)
		}
	}

	return nil
}

func GetNetMsgHierarchyRootMessages(pbGoStructTypes []reflect.Type) ([]*types.Message, error) {
	typesParser := types.DefaultParser()

	// For convenience, the parser operates on pointer to struct types and not struct types themselves.
	// The reason for this is that protobuf messages are always used as pointers in Go code.
	ptrTypes := sliceutil.Transform(pbGoStructTypes, func(_ int, tp reflect.Type) reflect.Type {
		return reflect.PointerTo(tp)
	})

	msgs, err := typesParser.ParseMessages(ptrTypes)
	if err != nil {
		return nil, err
	}

	// Look for the root of the hierarchy.
	return sliceutil.Filter(msgs, func(_ int, msg *types.Message) bool {
		return codegen.IsNetMessageRoot(msg.ProtoDesc())
	}), nil
}
