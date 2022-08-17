package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/generators/types-gen/generator"
)

func main() {
	gen := kingpin.New("types-gen", "Generates types to be used in the Mir framework from the protoc-generated types. "+
		"It also augments the protoc-generated structs for oneofs with some auxiliary methods and interfaces.")
	inputPkgPath := gen.Arg("package", "The full package path at which the protoc-generated types can be imported. "+
		"It should be possible to import the package from the current working directory.").String()

	if _, err := gen.Parse(os.Args[1:]); err != nil {
		gen.FatalUsage("could not parse arguments: %v\n", err)
	}

	err := codegen.RunGenerator[generator.TypeGenerator](*inputPkgPath)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(3)
	}
}
