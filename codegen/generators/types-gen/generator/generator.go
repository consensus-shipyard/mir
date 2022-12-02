package generator

import (
	"fmt"
	"reflect"

	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/buildutil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type TypeGenerator struct{}

// Run runs the TypeGenerator on the given set of struct types.
// For simplicity, all struct types are assumed to be in the same go package.
func (TypeGenerator) Run(pbGoStructTypes []reflect.Type) error {
	if len(pbGoStructTypes) == 0 {
		// Nothing to do.
		return nil
	}

	// Determine the input package.
	inputPackagePath := pbGoStructTypes[0].PkgPath()

	// Check that all structs are in the same package.
	for _, tp := range pbGoStructTypes {
		if tp.PkgPath() != inputPackagePath {
			return fmt.Errorf("received structs from different packages as input: %v and %v",
				inputPackagePath, tp.PkgPath())
		}
	}

	// Get the directory with input sources.
	inputDir, err := buildutil.GetSourceDirForPackage(inputPackagePath)
	if err != nil {
		return err
	}

	parser := types.DefaultParser()

	// For convenience, the parser operates on pointers to struct types and not struct types themselves.
	// The reason for this is that protobuf messages are always used as pointers in Go code.
	ptrTypes := sliceutil.Transform(pbGoStructTypes, func(_ int, tp reflect.Type) reflect.Type {
		return reflect.PointerTo(tp)
	})

	msgs, err := parser.ParseMessages(ptrTypes)
	if err != nil {
		return err
	}

	err = GenerateOneofInterfaces(inputDir, inputPackagePath, msgs, parser)
	if err != nil {
		return err
	}

	err = GenerateMirTypes(inputDir, inputPackagePath, msgs, parser)
	if err != nil {
		return err
	}

	return nil
}
