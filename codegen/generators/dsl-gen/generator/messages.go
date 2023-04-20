package generator

import (
	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen/model/messages"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/jenutil"
	"github.com/filecoin-project/mir/codegen/util/params"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/reflectutil"
)

var tNodeIDCode jen.Code = jenutil.QualFromType(reflectutil.TypeOf[t.NodeID]())

func GenerateDslFunctionsForNetMessages(hierarchyRoot *messages.NetMessageNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	uponMessageReceived := jen.Qual(
		"github.com/filecoin-project/mir/pkg/pb/transportpb/dsl", "UponMessageReceived")

	generateDslFunctionsForHandlingNetMessagesRecursively(
		/*hierarchyNode*/ hierarchyRoot,
		uponMessageReceived,
		jenFileBySourcePackagePath,
	)

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, DslOutputDir, "messages.mir.go")
}

func generateDslFunctionsForHandlingNetMessagesRecursively(
	hierarchyNode *messages.NetMessageNode,
	uponMessageReceived jen.Code,
	jenFileBySourcePackagePath map[string]*jen.File,
) {

	// Get a jen file to which the code will be added.
	sourcePackage := hierarchyNode.Message().PbPkgPath()
	jenFile, ok := jenFileBySourcePackagePath[sourcePackage]
	if !ok {
		jenFile = jen.NewFilePathName(DslPackagePath(sourcePackage), DslPackageName(sourcePackage))
		jenFileBySourcePackagePath[sourcePackage] = jenFile

		jenFile.Comment("Module-specific dsl functions for processing net messages.")
		jenFile.Line()
	}

	// Check if this is the root of the hierarchy.
	if hierarchyNode.IsRoot() {

		jenFile.Func().Id("Upon"+hierarchyNode.Name()+"Received").Types(
			jen.Id("W").Add(hierarchyNode.TypeOneof().MirWrapperInterface()).Types(jen.Id("M")),
			jen.Id("M").Any(),
		).Params(
			jen.Id("m").Add(dslModule),
			jen.Id("handler").Func().Params(jen.Id("from").Add(tNodeIDCode), jen.Id("msg").Op("*").Id("M")).Id("error"),
		).Block(
			jen.Add(uponMessageReceived).Params(
				jen.Id("m"),
				jen.Func().Params(
					jen.Id("from").Add(tNodeIDCode),
					jen.Id("msg").Add(hierarchyNode.Message().MirType()),
				).Id("error").Block(
					jen.List(jen.Id("w"), jen.Id("ok")).Op(":=").
						Id("msg").Dot(hierarchyNode.TypeOneof().Name).Op(".").Parens(jen.Add(jen.Id("W"))),
					jen.If(jen.Op("!").Id("ok")).Block(
						jen.Return(jen.Id("nil")),
					),
					jen.Line(), // empty line
					jen.Return(jen.Id("handler").Params(jen.Id("from"), jen.Id("w").Dot("Unwrap").Params())),
				),
			),
		).Line()

		uponChildMessageReceived := jen.Qual(
			DslPackagePath(hierarchyNode.Message().PbPkgPath()), "Upon"+hierarchyNode.Name()+"Received")

		// Recursively invoke the function for the children in the hierarchy.
		for _, child := range hierarchyNode.Children() {
			generateDslFunctionsForHandlingNetMessagesRecursively(
				/*hierarchyNode*/ child,
				/*uponMessageReceived*/ uponChildMessageReceived,
				jenFileBySourcePackagePath,
			)
		}

		return
	}

	// Check if this is an internal node in the hierarchy.
	if hierarchyNode.IsMsgClass() {

		// Generate function for handling the message class.
		jenFile.Func().Id("Upon"+hierarchyNode.Name()+"Received").Types(
			jen.Id("W").Add(hierarchyNode.TypeOneof().MirWrapperInterface()).Types(jen.Id("M")),
			jen.Id("M").Any(),
		).Params(
			jen.Id("m").Add(dslModule),
			// TODO: consider if we need to propagate some parameters from the parent.
			jen.Id("handler").Func().Params(jen.Id("from").Add(tNodeIDCode), jen.Id("msg").Op("*").Id("M")).Id("error"),
		).Block(
			jen.Add(uponMessageReceived).Types(hierarchyNode.OneofOption().MirWrapperType()).Params(
				jen.Id("m"),
				jen.Func().Params(
					jen.Id("from").Add(tNodeIDCode),
					jen.Id("msg").Add(hierarchyNode.Message().MirType()),
				).Id("error").Block(
					jen.List(jen.Id("w"), jen.Id("ok")).Op(":=").
						Id("msg").Dot(hierarchyNode.Parent().TypeOneof().Name).Op(".").Parens(jen.Add(jen.Id("W"))),
					jen.If(jen.Op("!").Id("ok")).Block(
						jen.Return(jen.Id("nil")),
					),
					jen.Line(), // empty line
					jen.Return(jen.Id("handler").Params(jen.Id("from"), jen.Id("w").Dot("Unwrap").Params())),
				),
			),
		).Line()

		uponChildMessageReceived := jen.Qual(
			DslPackagePath(hierarchyNode.Message().PbPkgPath()), "Upon"+hierarchyNode.Name()+"Received")

		// Recursively invoke the function for the children in the hierarchy.
		for _, child := range hierarchyNode.Children() {
			generateDslFunctionsForHandlingNetMessagesRecursively(
				/*hierarchyNode*/ child,
				/*uponMessageReceived*/ uponChildMessageReceived,
				jenFileBySourcePackagePath,
			)
		}
		return
	}

	// Generate the function for handling the net message.
	// TODO: consider if we need to propagate some parameters from the parent.

	nodeIDType := types.Castable{
		PbType_:  jen.Id("string"),
		MirType_: tNodeIDCode,
	}

	nonConstructorHandlerParams := params.FunctionParamList{}.Append("from", nodeIDType)
	constructorHandlerParams := hierarchyNode.ThisNodeConstructorParameters().Adapt(nonConstructorHandlerParams)
	allHandlerParameters := nonConstructorHandlerParams.AppendAll(constructorHandlerParams.FunctionParamList())

	jenFile.Func().Id("Upon"+hierarchyNode.Name()+"Received").Params(
		jen.Id("m").Add(dslModule),
		jen.Id("handler").Func().Params(allHandlerParameters.MirCode()...).Id("error"),
	).Block(
		jen.Add(uponMessageReceived).Types(hierarchyNode.OneofOption().MirWrapperType()).Params(
			jen.Id("m"),
			jen.Func().Params(jen.Id("from").Add(tNodeIDCode), jen.Id("msg").Add(hierarchyNode.Message().MirType())).Id("error").Block(
				jen.Return(jen.Id("handler").ParamsFunc(func(group *jen.Group) {
					group.Id("from")
					for _, param := range constructorHandlerParams.Slice() {
						group.Id("msg").Dot(param.Field().Name)
					}
				})),
			),
		),
	).Line()
}
