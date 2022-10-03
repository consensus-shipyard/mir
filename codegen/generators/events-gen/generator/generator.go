package generator

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/events"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type EventsGenerator struct{}

func (EventsGenerator) Run(pbGoStructTypes []reflect.Type) error {
	rootMessages, err := GetEventHierarchyRootMessages(pbGoStructTypes)
	if err != nil {
		return err
	}

	for _, eventRootMessage := range rootMessages {
		parser := events.DefaultParser()

		root, err := parser.ParseEventHierarchy(eventRootMessage)
		if err != nil {
			return err
		}

		err = GenerateEventConstructors(root)
		if err != nil {
			return fmt.Errorf("error generating event constructors: %w", err)
		}
	}

	return nil
}

func GetEventHierarchyRootMessages(pbGoStructTypes []reflect.Type) ([]*types.Message, error) {
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
		return codegen.IsMirEventRoot(msg.ProtoDesc())
	}), nil
}
