package generator

import (
	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/events"
)

func generateEventConstructorsRecursively(
	eventNode *events.EventNode,
	constructParent func(child jen.Code) *jen.Statement,
	eventRootType jen.Code,
	jenFileBySourcePackagePath map[string]*jen.File,
) {

	// If this is an intermediate node in the hierarchy, recursively call the function for subtypes.
	if eventNode.IsEventClass() {
		for _, childNode := range eventNode.Children() {
			// constructThisFromChild is a function that takes the code to construct a child in the hierarchy
			// and constructs an event of type eventNode.
			constructThisFromChild := func(child jen.Code) *jen.Statement {
				return constructParent(
					eventNode.Message().NewMirType().ValuesFunc(func(group *jen.Group) {
						// Initialize fields other than the event type.
						for _, param := range eventNode.ThisNodeConstructorParameters().Slice() {
							group.Line().Id(param.Field().Name).Op(":").Id(param.Name())
						}

						// Initialize the Type field
						group.Line().Id(eventNode.TypeOneof().Name).Op(":").
							Add(childNode.OneofOption().ConstructMirWrapperType(child))

						// Put the closing bracket on a new line.
						group.Line()
					}),
				)
			}

			generateEventConstructorsRecursively(
				/*eventNode*/ childNode,
				/*constructParent*/ constructThisFromChild,
				eventRootType,
				jenFileBySourcePackagePath,
			)
		}

		return
	}

	// If this is an event (i.e., a leaf in the hierarchy), create the event constructor.

	// get a jen file to which the code will be added.
	sourcePackage := eventNode.Message().PbPkgPath()
	jenFile, ok := jenFileBySourcePackagePath[sourcePackage]
	if !ok {
		jenFile = jen.NewFilePathName(events.PackagePath(sourcePackage), events.PackageName(sourcePackage))
		jenFileBySourcePackagePath[sourcePackage] = jenFile
	}

	// Generate the constructor.
	jenFile.Func().Id(eventNode.Name()).Params(
		eventNode.AllConstructorParameters().MirCode()...,
	).Add(eventRootType).Block(
		jen.Return(constructParent(
			eventNode.Message().NewMirType().ValuesFunc(func(group *jen.Group) {
				for _, param := range eventNode.ThisNodeConstructorParameters().Slice() {
					group.Line().Id(param.Field().Name).Op(":").Id(param.Name())
				}
				group.Line()
			}),
		)),
	).Line()
}

func GenerateEventConstructors(eventRoot *events.EventNode) error {
	jenFileBySourcePackagePath := make(map[string]*jen.File)

	generateEventConstructorsRecursively(
		/*eventNode*/ eventRoot,
		/*constructParent*/ func(child jen.Code) *jen.Statement { return jen.Add(child) },
		/*eventRootType*/ eventRoot.Message().MirType(),
		jenFileBySourcePackagePath,
	)

	return codegen.RenderJenFiles(jenFileBySourcePackagePath, events.OutputDir, "events.mir.go")
}
