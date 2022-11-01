package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/events"
	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

func GenerateDslFunctionsForEvents(eventRoot *events.EventNode) error {
	err := generateDslFunctionsForEmittingEvents(eventRoot)
	if err != nil {
		return err
	}

	return generateDslFunctionForHandlingEvents(eventRoot)
}

func generateDslFunctionsForEmittingEvents(eventRoot *events.EventNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	err := generateRecursivelyDslFunctionsForEmittingEvents(
		/*eventNode*/ eventRoot,
		jenFileBySourcePackagePath,
	)
	if err != nil {
		return err
	}

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, DslOutputDir, "emit.mir.go")
}

func generateRecursivelyDslFunctionsForEmittingEvents(
	eventNode *events.EventNode,
	jenFileBySourcePackagePath map[string]*jen.File,
) error {

	// If this is an internal node in the hierarchy, recursively call the function for subtypes and return.
	if eventNode.IsEventClass() {
		for _, child := range eventNode.Children() {
			err := generateRecursivelyDslFunctionsForEmittingEvents(child, jenFileBySourcePackagePath)
			if err != nil {
				return err
			}
		}

		return nil
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

	if eventNode.OriginRequest() != nil {
		return generateDslFunctionForEmittingRequestWithOrigin(eventNode, jenFile)
	}

	generateDslFunctionForEmittingSimpleEvent(eventNode, jenFile)
	return nil
}

func generateDslFunctionForEmittingRequestWithOrigin(eventNode *events.EventNode, jenFile *jen.File) error {

	// Example:
	//		func RequestCert[C any](m dsl.Module, dest t.ModuleID, context *C) {
	//			contextID := m.DslHandle().StoreContext(context)
	//
	//			origin := &apb.RequestCertOrigin{
	//				Module: m.ModuleID().Pb(),
	//				Type: &apb.RequestCertOrigin_Dsl{
	//					Dsl: dsl.Origin(contextID),
	//				},
	//			}
	//
	//			dsl.EmitEvent(m, aevents.RequestCert(dest, origin))
	//		}

	origin := eventNode.OriginRequest()

	dslOriginOption, ok := GetDslOriginOption(origin)
	if !ok {
		return fmt.Errorf("the 'Type' oneof of the Origin message %v in package %v must have an option named 'Dsl'",
			origin.Message.Name(), origin.Message.PbPkgPath())
	}

	originParam, _ := eventNode.ThisNodeConstructorParameters().FindParamByField(origin.Field)
	constructorParamsWithoutOrigin, _ := eventNode.AllConstructorParameters().RemoveParam(originParam.FunctionParam)

	funcParams := []jen.Code{jen.Id("m").Add(dslModule)}
	funcParams = append(funcParams, constructorParamsWithoutOrigin.MirCode()...)
	funcParams = append(funcParams, jen.Id("context").Op("*").Id("C"))

	jenFile.Func().Id(eventNode.Name()).Types(jen.Id("C").Any()).Params(funcParams...).Block(
		jen.Id("contextID").Op(":=").Id("m").Dot("DslHandle").Call().Dot("StoreContext").Call(jen.Id("context")),
		jen.Line(), // empty line
		jen.Id(originParam.Name()).Op(":=").Add(origin.Message.NewMirType()).Values(jen.Dict{
			jen.Id("Module"): jen.Id("m").Dot("ModuleID").Call(),
			jen.Id("Type"): dslOriginOption.NewMirWrapperType().Values(jen.Dict{
				jen.Id("Dsl"): jen.Add(dslMirOrigin).Call(jen.Id("contextID")),
			}),
		}),
		jen.Line(), // empty line
		jen.Add(dslEmitMirEvent).Call(
			jen.Id("m"),
			eventNode.Constructor().Call(eventNode.AllConstructorParameters().IDs()...),
		),
	).Line()

	return nil
}

func generateDslFunctionForEmittingSimpleEvent(eventNode *events.EventNode, jenFile *jen.File) {
	funcParams := []jen.Code{jen.Id("m").Add(dslModule)}
	funcParams = append(funcParams, eventNode.AllConstructorParameters().MirCode()...)

	jenFile.Func().Id(eventNode.Name()).Params(funcParams...).Block(
		jen.Add(dslEmitMirEvent).Call(
			jen.Id("m"),
			eventNode.Constructor().Call(eventNode.AllConstructorParameters().IDs()...),
		),
	).Line()
}

func generateDslFunctionForHandlingEvents(eventRoot *events.EventNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	for _, child := range eventRoot.Children() {
		err := generateRecursivelyDslFunctionsForHandlingEvents(
			/*eventNode*/ child,
			/*uponEvent*/ dslUponMirEvent,
			jenFileBySourcePackagePath,
		)

		if err != nil {
			return err
		}
	}

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, DslOutputDir, "upon.mir.go")
}

func generateRecursivelyDslFunctionsForHandlingEvents(
	eventNode *events.EventNode,
	uponEvent jen.Code,
	jenFileBySourcePackagePath map[string]*jen.File,
) error {

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
		generateDslFunctionForHandlingEventClass(eventNode, uponEvent, jenFile)

		// Recursively invoke the generation function for the children in the hierarchy.
		uponChildEvent := jen.Qual(DslPackagePath(eventNode.Message().PbPkgPath()), "Upon"+eventNode.Name())
		for _, child := range eventNode.Children() {
			err := generateRecursivelyDslFunctionsForHandlingEvents(
				/*eventNode*/ child,
				/*uponEvent*/ uponChildEvent,
				jenFileBySourcePackagePath,
			)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if eventNode.OriginResponse() != nil {
		return generateDslFunctionForHandlingResponseWithOrigin(eventNode, uponEvent, jenFile)
	}

	generateDslFunctionForHandlingSimpleEvent(eventNode, uponEvent, jenFile)
	return nil
}

func generateDslFunctionForHandlingEventClass(eventNode *events.EventNode, uponEvent jen.Code, jenFile *jen.File) {
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
}

func generateDslFunctionForHandlingResponseWithOrigin(
	eventNode *events.EventNode,
	uponEvent jen.Code,
	jenFile *jen.File,
) error {

	// Example:
	//		func UponSignResult[C any](m Module, handler func(signature []byte, context *C) error) {
	//			UponEvent[*eventpb.Event_SignResult](m, func(ev *eventpb.SignResult) error {
	//				originWrapper, ok := ev.Origin.Type.(*eventpb.SignOrigin_Dsl)
	//				if !ok {
	//					return nil
	//				}
	//
	//				contextRaw := m.DslHandle().RecoverAndCleanupContext(ContextID(originWrapper.Dsl.ContextID))
	//				context, ok := contextRaw.(*C)
	//				if !ok {
	//					return nil
	//				}
	//
	//				return handler(ev.Signature, context)
	//			})
	//		}

	origin := eventNode.OriginResponse()

	dslOriginOption, ok := GetDslOriginOption(origin)
	if !ok {
		return fmt.Errorf("the 'Type' oneof of the Origin message %v in package %v must have an option named 'Dsl'",
			origin.Message.Name(), origin.Message.PbPkgPath())
	}

	// TODO: consider if we need to propagate some parameters from the parent.
	originParam, _ := eventNode.ThisNodeConstructorParameters().FindParamByField(origin.Field)
	constructorParamsWithoutOrigin, _ := eventNode.ThisNodeConstructorParameters().RemoveParam(originParam)

	handlerParams := append(constructorParamsWithoutOrigin.MirCode(), jen.Id("context").Op("*").Id("C"))

	jenFile.Func().Id("Upon"+eventNode.Name()).Types(jen.Id("C").Any()).Params(
		jen.Id("m").Add(dslModule),
		jen.Id("handler").Func().Params(handlerParams...).Id("error"),
	).Block(
		jen.Add(uponEvent).Types(eventNode.OneofOption().MirWrapperType()).Params(
			jen.Id("m"),
			jen.Func().Params(jen.Id("ev").Add(eventNode.Message().MirType())).Id("error").Block(
				jen.List(jen.Id("originWrapper"), jen.Id("ok")).Op(":=").Id("ev").Dot(origin.Field.Name).Dot("Type").
					Op(".").Parens(dslOriginOption.MirWrapperType()),
				jen.If(jen.Op("!").Id("ok")).Block(
					jen.Return(jen.Id("nil")),
				),
				jen.Line(), // empty line

				jen.Id("contextRaw").Op(":=").Id("m").Dot("DslHandle").Params().Dot("RecoverAndCleanupContext").
					Params(jen.Add(dslContextID).Parens(jen.Id("originWrapper").Dot("Dsl").Dot("ContextID"))),
				jen.Id("context").Op(",").Id("ok").Op(":=").Id("contextRaw").Op(".").Parens(jen.Id("*C")),
				jen.If(jen.Op("!").Id("ok")).Block(
					jen.Return(jen.Id("nil")),
				),
				jen.Line(), // empty line

				jen.Return(jen.Id("handler").ParamsFunc(func(group *jen.Group) {
					for _, param := range constructorParamsWithoutOrigin.Slice() {
						group.Id("ev").Dot(param.Field().Name)
					}
					group.Id("context")
				})),
			),
		),
	).Line()

	return nil
}

func generateDslFunctionForHandlingSimpleEvent(eventNode *events.EventNode, uponEvent jen.Code, jenFile *jen.File) {
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

func GetDslOriginOption(origin *events.Origin) (*types.OneofOption, bool) {
	return sliceutil.Any(
		sliceutil.Filter(origin.TypeOneof.Options, func(i int, opt *types.OneofOption) bool {
			return opt.Name() == "Dsl"
		}),
	)
}
