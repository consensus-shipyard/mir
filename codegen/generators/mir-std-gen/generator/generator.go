package generator

import (
	"reflect"

	dslgenerator "github.com/filecoin-project/mir/codegen/generators/dsl-gen/generator"
	eventsgenerator "github.com/filecoin-project/mir/codegen/generators/events-gen/generator"
	netgenerator "github.com/filecoin-project/mir/codegen/generators/net-gen/generator"
	typesgenerator "github.com/filecoin-project/mir/codegen/generators/types-gen/generator"
)

// CombinedGenerator combines the functionality of several standard generators.
type CombinedGenerator struct{}

func (CombinedGenerator) Run(structTypes []reflect.Type) error {
	err := typesgenerator.TypeGenerator{}.Run(structTypes)
	if err != nil {
		return err
	}

	err = eventsgenerator.EventsGenerator{}.Run(structTypes)
	if err != nil {
		return err
	}

	err = netgenerator.NetMsgGenerator{}.Run(structTypes)
	if err != nil {
		return err
	}

	err = dslgenerator.DslGenerator{}.Run(structTypes)
	if err != nil {
		return err
	}

	return nil
}
