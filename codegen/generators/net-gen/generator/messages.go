package generator

import (
	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/messages"
)

func generateMessageConstructorsRecursively(
	msgNode *messages.NetMessageNode,
	constructParent func(child jen.Code) *jen.Statement,
	msgRootType jen.Code,
	jenFileBySourcePackagePath map[string]*jen.File,
) {

	// If this is an intermediate node in the hierarchy, recursively call the function for subtypes.
	if msgNode.IsMsgClass() {
		for _, childNode := range msgNode.Children() {
			// constructThis is a function that takes the code to construct a child in the hierarchy
			// and constructs a message.
			constructThis := func(child jen.Code) *jen.Statement {
				return constructParent(
					msgNode.Message().NewMirType().ValuesFunc(func(group *jen.Group) {
						// Initialize fields other than the message type.
						for _, param := range msgNode.ThisNodeConstructorParameters().Slice() {
							group.Line().Id(param.Field().Name).Op(":").Id(param.Name())
						}

						// Initialize the Type field
						group.Line().Id(msgNode.TypeOneof().Name).Op(":").
							Add(childNode.OneofOption().ConstructMirWrapperType(child))

						// Put the closing bracket on a new line.
						group.Line()
					}),
				)
			}

			generateMessageConstructorsRecursively(
				/*msgNode*/ childNode,
				/*constructParent*/ constructThis,
				msgRootType,
				jenFileBySourcePackagePath,
			)
		}

		return
	}

	// If this is a net message (i.e., a leaf in the hierarchy), create the net message constructor.

	// Get a jen file to which the constructor will be added.
	sourcePackage := msgNode.Message().PbPkgPath()
	jenFile, ok := jenFileBySourcePackagePath[sourcePackage]
	if !ok {
		jenFile = jen.NewFilePathName(messages.PackagePath(sourcePackage), messages.PackageName(sourcePackage))
		jenFileBySourcePackagePath[sourcePackage] = jenFile
	}

	// Generate the constructor.
	jenFile.Func().Id(msgNode.Name()).Params(
		msgNode.AllConstructorParameters().MirCode()...,
	).Add(msgRootType).Block(
		jen.Return(constructParent(
			msgNode.Message().NewMirType().ValuesFunc(func(group *jen.Group) {
				for _, param := range msgNode.ThisNodeConstructorParameters().Slice() {
					group.Line().Id(param.Field().Name).Op(":").Id(param.Name())
				}
				group.Line()
			}),
		)),
	).Line()
}

func GenerateMessageConstructors(messageRoot *messages.NetMessageNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	generateMessageConstructorsRecursively(
		/*msgNode*/ messageRoot,
		/*constructParent*/ func(child jen.Code) *jen.Statement { return jen.Add(child) },
		/*msgRootType*/ messageRoot.Message().MirType(),
		jenFileBySourcePackagePath,
	)

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, messages.OutputDir, "msgs.mir.go")
}
