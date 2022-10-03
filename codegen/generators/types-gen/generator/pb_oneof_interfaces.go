package generator

import (
	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/model/types"
)

func generateOneofInterface(g *jen.File, oneof *types.Oneof) error {
	// Generate [Msg]_[Oneof] interface.
	g.Type().Id(oneof.PbExportedInterfaceName()).Op("=").Id(oneof.PbNativeInterfaceName()).Line()

	// Generate [Msg]_[Oneof]Wrapper interface.
	g.Type().Id(oneof.PbExportedInterfaceName()+"Wrapper").Types(jen.Id("T").Any()).Interface(
		jen.Id(oneof.PbExportedInterfaceName()),
		jen.Id("Unwrap").Params().Op("*").Id("T"),
	).Line()

	// Generate Unwrap() method implementations
	for _, opt := range oneof.Options {
		g.Func().Params(jen.Id("w").Add(opt.PbWrapperType())).Id("Unwrap").Params().Add(opt.Field.Type.PbType()).Block(
			jen.Return(jen.Id("w").Dot(opt.Field.Name)),
		).Line()
	}

	return nil
}

// GenerateOneofInterfaces generates exported interfaces of the form "[Msg]_[Oneof]" "[Msg]_[Oneof]Wrapper" for all
// oneofs in the given messages, where [Msg] is the name of the message and [Oneof] is the name of the oneof.
func GenerateOneofInterfaces(
	inputDir, inputPackagePath string,
	msgs []*types.Message,
	parser *types.Parser,
) (err error) {

	jenFile := jen.NewFilePath(inputPackagePath)

	for _, msg := range msgs {
		if !msg.ShouldGenerateMirType() {
			continue
		}

		fields, err := parser.ParseFields(msg)
		if err != nil {
			return err
		}

		for _, field := range fields {
			if oneof, ok := field.Type.(*types.Oneof); ok {
				err := generateOneofInterface(jenFile, oneof)
				if err != nil {
					return err
				}
			}
		}
	}

	return codegen.RenderJenFile(jenFile, inputDir, "oneof_interfaces.mir.go")
}
