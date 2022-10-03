package generator

import (
	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/events"
)

func GenerateDslFunctionsForEvents(eventRoot *events.EventNode) error {
	err := generateDslFunctionsForEmittingEvents(eventRoot)
	if err != nil {
		return err
	}

	return generateDslFunctionForHandlingEvents(eventRoot)
}

func generateDslFunctionsForEmittingEventsRecursively(
	eventNode *events.EventNode,
	jenFileBySourcePackagePath map[string]*jen.File,
) {

	// If this is an internal node in the hierarchy, recursively call the function for subtypes.
	if eventNode.IsEventClass() {
		for _, child := range eventNode.Children() {
			generateDslFunctionsForEmittingEventsRecursively(child, jenFileBySourcePackagePath)
		}
		return
	}

	// get a jen file to which the code will be added.
	sourcePackage := eventNode.Message().PbPkgPath()
	jenFile, ok := jenFileBySourcePackagePath[sourcePackage]
	if !ok {
		jenFile = jen.NewFilePathName(DslPackagePath(sourcePackage), DslPackageName(sourcePackage))
		jenFileBySourcePackagePath[sourcePackage] = jenFile

		jenFile.Comment("Module-specific dsl functions for emitting events.")
		jenFile.Line()
	}

	// Generate the function for emitting the event
	funcParams := append(
		[]jen.Code{jen.Id("m").Add(dslModule)},
		eventNode.AllConstructorParameters().MirCode()...,
	)

	jenFile.Func().Id(eventNode.Name()).Params(funcParams...).Block(
		jen.Add(dslEmitMirEvent).Params(
			jen.Id("m"),
			eventNode.Constructor().Params(eventNode.AllConstructorParameters().IDs()...),
		),
	).Line()
}

func generateDslFunctionsForEmittingEvents(eventRoot *events.EventNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	generateDslFunctionsForEmittingEventsRecursively(
		/*eventNode*/ eventRoot,
		jenFileBySourcePackagePath,
	)

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, DslOutputDir, "emit.mir.go")
}

func generateDslFunctionsForHandlingEventsRecursively(
	eventNode *events.EventNode,
	uponEvent jen.Code,
	jenFileBySourcePackagePath map[string]*jen.File,
) {

	// get a jen file to which the code will be added.
	sourcePackage := eventNode.Message().PbPkgPath()
	jenFile, ok := jenFileBySourcePackagePath[sourcePackage]
	if !ok {
		jenFile = jen.NewFilePathName(DslPackagePath(sourcePackage), DslPackageName(sourcePackage))
		jenFileBySourcePackagePath[sourcePackage] = jenFile

		jenFile.Comment("Module-specific dsl functions for processing events.")
		jenFile.Line()
	}

	// Check if this is an internal node in the hierarchy.
	if eventNode.IsEventClass() {

		// Generate function for handling the event class.
		jenFile.Func().Id("Upon"+eventNode.Name()).Types(
			jen.Id("W").Add(eventNode.TypeOneof().MirWrapperInterface()).Types(jen.Id("Ev")),
			jen.Id("Ev").Any(),
		).Params(
			jen.Id("m").Add(dslModule),
			// TODO: consider if we need to propagate some parameters from the parent.
			jen.Id("handler").Func().Params(jen.Id("ev").Op("*").Id("Ev")).Id("error"),
		).Block(
			jen.Add(uponEvent).Types(eventNode.OneofOption().MirWrapperType()).Params(
				jen.Id("m"),
				jen.Func().Params(jen.Id("ev").Add(eventNode.Message().MirType())).Id("error").Block(
					jen.List(jen.Id("w"), jen.Id("ok")).Op(":=").
						Id("ev").Dot(eventNode.Parent().TypeOneof().Name).Op(".").Parens(jen.Add(jen.Id("W"))),
					jen.If(jen.Op("!").Id("ok")).Block(
						jen.Return(jen.Id("nil")),
					),
					jen.Line(), // empty line
					jen.Return(jen.Id("handler").Params(jen.Id("w").Dot("Unwrap").Params())),
				),
			),
		).Line()

		uponChildEvent := jen.Qual(DslPackagePath(eventNode.Message().PbPkgPath()), "Upon"+eventNode.Name())

		// Recursively invoke the function for the children in the hierarchy.
		for _, child := range eventNode.Children() {
			generateDslFunctionsForHandlingEventsRecursively(
				/*eventNode*/ child,
				/*uponEvent*/ uponChildEvent,
				jenFileBySourcePackagePath,
			)
		}
		return
	}

	// Generate the function for handling the event.
	// TODO: consider if we need to propagate some parameters from the parent.
	handlerParameters := eventNode.ThisNodeConstructorParameters()

	jenFile.Func().Id("Upon"+eventNode.Name()).Params(
		jen.Id("m").Add(dslModule),
		jen.Id("handler").Func().Params(handlerParameters.MirCode()...).Id("error"),
	).Block(
		jen.Add(uponEvent).Types(eventNode.OneofOption().MirWrapperType()).Params(
			jen.Id("m"),
			jen.Func().Params(jen.Id("ev").Add(eventNode.Message().MirType())).Id("error").Block(
				jen.Return(jen.Id("handler").ParamsFunc(func(group *jen.Group) {
					for _, param := range handlerParameters.Slice() {
						group.Id("ev").Dot(param.Field().Name)
					}
				})),
			),
		),
	).Line()
}

func generateDslFunctionForHandlingEvents(eventRoot *events.EventNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	for _, child := range eventRoot.Children() {
		generateDslFunctionsForHandlingEventsRecursively(
			/*eventNode*/ child,
			/*uponEvent*/ dslUponMirEvent,
			jenFileBySourcePackagePath,
		)
	}

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, DslOutputDir, "upon.mir.go")
}
