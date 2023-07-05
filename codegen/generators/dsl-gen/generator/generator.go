package generator

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"

	eventsgenerator "github.com/filecoin-project/mir/codegen/generators/events-gen/generator"
	netgenerator "github.com/filecoin-project/mir/codegen/generators/net-gen/generator"
	"github.com/filecoin-project/mir/codegen/model/events"
	"github.com/filecoin-project/mir/codegen/model/messages"
)

type DslGenerator struct{}

func (DslGenerator) Run(pbGoStructTypes []reflect.Type) error {
	eventHierarchyRoots, err := eventsgenerator.GetEventHierarchyRootMessages(pbGoStructTypes)
	if err != nil {
		return err
	}

	for _, eventHierarchyRoot := range eventHierarchyRoots {
		eventParser := events.DefaultParser()

		rootNode, err := eventParser.ParseEventHierarchy(eventHierarchyRoot)
		if err != nil {
			return err
		}

		err = GenerateDslFunctionsForEvents(rootNode)
		if err != nil {
			return fmt.Errorf("error generating dsl functions for events: %w", err)
		}
	}

	netMessageHierarchyRoots, err := netgenerator.GetNetMsgHierarchyRootMessages(pbGoStructTypes)
	if err != nil {
		return err
	}

	for _, netMessageHierarchyRoot := range netMessageHierarchyRoots {
		netMessageParser := messages.DefaultParser()

		rootNode, err := netMessageParser.ParseNetMessageHierarchy(netMessageHierarchyRoot)
		if err != nil {
			return err
		}

		err = GenerateDslFunctionsForNetMessages(rootNode)
		if err != nil {
			return fmt.Errorf("error generating dsl functions for net messages: %w", err)
		}
	}

	return nil
}

func DslPackagePath(sourcePackagePath string) string {
	return sourcePackagePath + "/dsl"
}

func DslPackageName(sourcePackagePath string) string {
	return sourcePackagePath[strings.LastIndex(sourcePackagePath, "/")+1:] + "dsl"
}

func DslOutputDir(sourceDir string) string {
	return path.Join(sourceDir, "dsl")
}

var (
	// Note: using reflection to determine this package path would cause a build dependency cycle.
	dslPackagePath = "github.com/filecoin-project/mir/pkg/dsl"

	dslModule    jen.Code = jen.Qual(dslPackagePath, "Module")
	dslContextID jen.Code = jen.Qual(dslPackagePath, "ContextID")
	dslMirOrigin jen.Code = jen.Qual(dslPackagePath, "MirOrigin")
	dslEmitEvent jen.Code = jen.Qual(dslPackagePath, "EmitEvent")
	dslUponEvent jen.Code = jen.Qual(dslPackagePath, "UponEvent")
)
