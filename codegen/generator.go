package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"text/template"

	"golang.org/x/exp/slices"

	"github.com/filecoin-project/mir/codegen/util/astutil"
	"github.com/filecoin-project/mir/codegen/util/buildutil"
	"github.com/filecoin-project/mir/pkg/util/reflectutil"
)

// Generator receives a list of all struct types exported by a package as input and produces the generated code.
// It is assumed to use reflection to inspect the input types.
// All the types are from the same package and, thus, from the same folder.
// To obtain the source folder, buildutil.GetSourceDirForPackage can be used.
type Generator interface {
	Run(structTypes []reflect.Type) error
}

// RunGenerator runs a generator on the exported struct types of the package corresponding to inputPkgPath.
// The zero value of the provided GeneratorType is used. GeneratorType cannot be an interface, it must be exported
// (i.e., the name of the type should start from a capital letter), and it cannot be in "main" or "internal" package.
func RunGenerator[GeneratorType Generator](inputPkgPath string) error {
	genType := reflectutil.TypeOf[GeneratorType]()
	if genType.Kind() == reflect.Interface {
		return fmt.Errorf("type %v is an interface. A concrete type was expected", genType)
	}

	generatorPkgPath := genType.PkgPath()
	generatorName := genType.Name()

	if !ast.IsExported(generatorName) {
		log.Println(generatorName)
		log.Println(ast.IsExported(generatorName))
		return fmt.Errorf("generator %v.%v is not exported. Consider renaming it to %v",
			generatorPkgPath, generatorName, astutil.ToExported(generatorName))
	}

	if slices.Contains(strings.Split(generatorPkgPath, "/"), "internal") {
		return fmt.Errorf("generator %v.%v is in an 'internal' package", generatorPkgPath, generatorName)
	}

	if generatorPkgPath == "main" {
		return fmt.Errorf("generator %v.%v is in a 'main' package", generatorPkgPath, generatorName)
	}

	// Get the list of all structs (i.e., potential message types) in the package.
	structNames, err := getAllTypesInAPackage(inputPkgPath)
	if err != nil {
		return fmt.Errorf("failed to extract struct types from package %v: %w", inputPkgPath, err)
	}

	// Create the generator program code.
	generatorProgram, err := writeGeneratorProgram(generatorPkgPath, generatorName, inputPkgPath, structNames)
	if err != nil {
		return err
	}

	// Run the program in the current working directory.
	// TODO: consider running it from a different directory or trying several options until one of them succeeds
	//       (this is the approach gomock takes).
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get the working directory: %w", err)
	}

	return runInDir(generatorProgram, wd)
}

type generatorInputs struct {
	GeneratorPath string
	GeneratorName string
	ImportPath    string
	StructNames   []string
}

var generatorTemplate = template.Must(template.New("generator").Parse(`
package main

import (
	"log"
	"os"
	"reflect"

	generator_ "{{.GeneratorPath}}"
	pkg_ "{{.ImportPath}}"
)

func main() {
	var generator generator_.{{.GeneratorName}}
	err := generator.Run(
		[]reflect.Type{
			{{range .StructNames}}
			reflect.TypeOf((*pkg_.{{.}})(nil)).Elem(),
			{{end}}
		})

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}
`))

func writeGeneratorProgram(generatorPkgPath, generatorName, inputPkgPath string, structNames []string) ([]byte, error) {
	var program bytes.Buffer

	err := generatorTemplate.Execute(
		&program,
		&generatorInputs{generatorPkgPath, generatorName, inputPkgPath, structNames},
	)
	if err != nil {
		return nil, err
	}

	return program.Bytes(), nil
}

// Run the program in the given directory.
func runInDir(program []byte, dir string) error {
	// Create a temporary folder for the generator program.
	tmpDir, err := os.MkdirTemp(dir, "proto_converter_tmp_")
	if err != nil {
		return fmt.Errorf("error creating a temporary directory in %v: %w", dir, err)
	}

	// Remove the temporary folder when finished.
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Printf("failed to remove the temporary directory %v: %v", tmpDir, err)
		}
	}()

	progSource := "generator.go"
	progBinary := "generator.bin"
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		// Windows won't execute a program unless it has a ".exe" suffix.
		progBinary += ".exe"
	}

	if err := os.WriteFile(filepath.Join(tmpDir, progSource), program, 0600); err != nil {
		return err
	}

	cmdArgs := []string{"build", "-o", progBinary, progSource}

	// Build the program.
	buildCmd := exec.Command("go", cmdArgs...)
	buildCmd.Dir = tmpDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return err
	}

	// Run the program.
	generatorCmd := exec.Command(path.Join(tmpDir, progBinary)) //nolint:gosec
	generatorCmd.Stdout = os.Stdout
	generatorCmd.Stderr = os.Stderr
	return generatorCmd.Run()
}

func getAllTypesInAPackage(sourcePkgPath string) ([]string, error) {
	// Get the source dir of the package.
	sourceDir, err := buildutil.GetSourceDirForPackage(sourcePkgPath)
	if err != nil {
		return nil, fmt.Errorf("could not obtain the source dir for package %v: %w", sourcePkgPath, err)
	}

	// Parse the source dir.
	fset := token.FileSet{}
	pkgs, err := parser.ParseDir(&fset, sourceDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("could not parse the package sources: %w", err)
	}

	// Get the ast of the target package.
	pkgName := sourcePkgPath[strings.LastIndex(sourcePkgPath, "/")+1:]
	pkgAst, ok := pkgs[pkgName]
	if !ok {
		return nil, fmt.Errorf("did not find package %v in %v", pkgName, sourceDir)
	}

	var res []string
	for _, file := range pkgAst.Files {
		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range decl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if typeSpec.Assign != token.NoPos {
					// Ignore type aliases.
					continue
				}

				if !ast.IsExported(typeSpec.Name.Name) {
					// Skip unexported types.
					continue
				}

				if typeSpec.TypeParams != nil {
					// Skip generic types.
					continue
				}

				res = append(res, typeSpec.Name.Name)
			}
		}
	}

	return res, nil
}
