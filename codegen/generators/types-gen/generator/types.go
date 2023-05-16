package generator

import (
	"reflect"

	"github.com/dave/jennifer/jen"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/mirreflect"
	"github.com/filecoin-project/mir/codegen/model/types"
	"github.com/filecoin-project/mir/codegen/util/jenutil"
	"github.com/filecoin-project/mir/pkg/util/reflectutil"
)

var (
	mirreflectType          jen.Code = jenutil.QualFromType(reflectutil.TypeOf[mirreflect.Type]())
	mirreflectTypeImpl      jen.Code = jenutil.QualFromType(reflectutil.TypeOf[mirreflect.TypeImpl]())
	mirreflectGeneratedType jen.Code = jenutil.QualFromType(reflectutil.TypeOf[mirreflect.GeneratedType]())

	reflectutilTypeOf jen.Code = jen.Qual("github.com/filecoin-project/mir/pkg/util/reflectutil", "TypeOf")
)

func generateMirType(g *jen.File, msg *types.Message, parser *types.Parser) error {
	if !msg.ShouldGenerateMirType() {
		// Ignore non-annotated messages.
		return nil
	}

	fields, err := parser.ParseFields(msg)
	if err != nil {
		return err
	}

	// Generate the struct.
	g.Type().Id(msg.Name()).StructFunc(func(group *jen.Group) {
		for _, field := range fields {
			group.Id(field.Name).Add(field.Type.MirType())
		}
	}).Line()

	// Generate the code for oneof fields.
	for _, field := range fields {
		if oneof, ok := field.Type.(*types.Oneof); ok {
			// Generate the interfaces.
			g.Type().Id(oneof.MirInterfaceName()).Interface(
				mirreflectGeneratedType,
				jen.Id(oneof.MirMethodName()).Params(),
				jen.Id("Pb").Params().Add(oneof.PbType()),
			).Line()

			g.Type().Id(oneof.MirWrapperInterfaceName()).Types(jen.Id("T").Any()).Interface(
				jen.Id(oneof.MirInterfaceName()),
				jen.Id("Unwrap").Params().Op("*").Id("T"),
			).Line()

			// Generate the FromPb function.
			g.Func().Id(oneof.MirInterfaceName()+"FromPb").Params(jen.Id("pb").Add(oneof.PbType())).Add(oneof.MirType()).Block(
				jen.If(jen.Id("pb").Op("==").Nil()).Block(
					jen.Return(jen.Nil()),
				),
				jen.Switch(jen.Id("pb").Op(":=").Id("pb").Dot("(type)")).BlockFunc(func(group *jen.Group) {
					for _, opt := range oneof.Options {
						group.Case(opt.PbWrapperType()).Block(
							jen.Return(jen.Add(opt.NewMirWrapperType()).Values(
								jen.Id(opt.Field.Name).Op(":").Add(opt.Field.Type.ToMir(jen.Id("pb").Dot(opt.Field.Name))),
							)),
						)
					}
				}),
				jen.Return(jen.Nil()),
			).Line()

			// Generate the wrappers.
			for _, opt := range oneof.Options {
				g.Type().Id(opt.WrapperName).Struct(
					jen.Id(opt.Field.Name).Add(opt.Field.Type.MirType()),
				).Line()

				g.Func().Params(opt.MirWrapperType()).Id(oneof.MirMethodName()).Params().Block().Line()

				// Generate the Unwrap method.
				g.Func().Params(jen.Id("w").Add(opt.MirWrapperType())).Id("Unwrap").Params().Add(opt.Field.Type.MirType()).Block(
					jen.Return(jen.Id("w").Dot(opt.Field.Name)),
				).Line()

				// Generate the Pb method.
				g.Func().Params(jen.Id("w").Add(opt.MirWrapperType())).Id("Pb").Params().Add(oneof.PbType()).Block(
					jen.If(jen.Id("w").Op("==").Nil()).Block(
						jen.Return(jen.Nil()),
					),
					jen.If(jen.Id("w").Dot(opt.Field.Name).Op("==").Nil()).Block(
						jen.Return(jen.Add(opt.NewPbWrapperType().Values(jen.Empty()))),
					),
					jen.Return(jen.Add(opt.NewPbWrapperType()).Values(
						jen.Id(opt.Field.Name).Op(":").Add(opt.Field.Type.ToPb(jen.Id("w").Dot(opt.Field.Name))),
					)),
				).Line()

				// Generate the MirReflect method.
				g.Func().Params(jen.Add(opt.MirWrapperType())).Id("MirReflect").Params().Add(mirreflectType).Block(
					jen.Return(jen.Add(mirreflectTypeImpl).Values(
						jen.Id("PbType_").Op(":").Add(reflectutilTypeOf).Types(opt.PbWrapperType()).Params(),
					)),
				).Line()
			}
		}
	}

	// Generate [Name]FromPb function.
	// NB: it would be nicer to generate .ToMir() methods for pb types, but this would cause a cyclic dependency.
	g.Func().Id(msg.Name()+"FromPb").Params(jen.Id("pb").Add(msg.PbType())).Add(msg.MirType()).Block(
		jen.If(jen.Id("pb").Op("==").Nil()).Block(
			jen.Return(jen.Nil()),
		),
		jen.Return().Add(msg.NewMirType()).ValuesFunc(func(group *jen.Group) {
			for _, field := range fields {
				group.Line().Id(field.Name).Op(":").Add(field.Type.ToMir(jen.Id("pb").Dot(field.Name)))
			}
			group.Line()
		}),
	).Line()

	// Generate the Pb method.
	g.Func().Params(jen.Id("m").Add(msg.MirType())).Id("Pb").Params().Add(msg.PbType()).Block(
		jen.If(jen.Id("m").Op("==").Nil()).Block(
			jen.Return(jen.Nil()),
		),
		jen.Id("pbMessage").Op(":=").Add(msg.NewPbType()).Values(jen.Empty()),

		// Iterate over struct fields and set values for non-nil fields.
		jen.BlockFunc(func(g *jen.Group) {
			for _, field := range fields {
				if !canBeNil(field) {
					g.Add(jen.Id("pbMessage").Dot(field.Name).Op("=").Add(field.Type.ToPb(jen.Id("m").Dot(field.Name))))
				} else {
					g.If(jen.Id("m").Dot(field.Name).Op("!=").Nil()).Block(
						jen.Id("pbMessage").Dot(field.Name).Op("=").Add(field.Type.ToPb(jen.Id("m").Dot(field.Name))),
					)
				}
			}
		}),
		jen.Empty(),
		jen.Return(jen.Id("pbMessage")),
	).Line()

	// Generate the MirReflect method.

	// Generate the MirReflect method.
	g.Func().Params(jen.Add(msg.MirType())).Id("MirReflect").Params().Add(mirreflectType).Block(
		jen.Return(jen.Add(mirreflectTypeImpl).Values(
			jen.Id("PbType_").Op(":").Add(reflectutilTypeOf).Types(msg.PbType()).Params(),
		)),
	).Line()

	return nil
}

func GenerateMirTypes(inputDir, sourcePackagePath string, msgs []*types.Message, parser *types.Parser) (err error) {
	jenFile := jen.NewFilePathName(
		types.PackagePath(sourcePackagePath),
		types.PackageName(sourcePackagePath),
	)

	// Generate Mir types for messages.
	for _, msg := range msgs {
		err := generateMirType(jenFile, msg, parser)
		if err != nil {
			return err
		}
	}

	return codegen.RenderJenFile(jenFile, types.OutputDir(inputDir), "types.mir.go")
}

// canBeNil returns true if the field can be nil.
// at the moment, Maps and Slices cannot be nil by this check, as we use an internal conversion to our types.Slice and types.Map
// the check for Maps and Slices is in their respective converSlice and convertMap functions
func canBeNil(field *types.Field) bool {
	switch reflect.TypeOf(field.Type).Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.UnsafePointer:
		return true
	}
	return false
}
