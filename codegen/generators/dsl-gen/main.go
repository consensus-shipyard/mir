package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/generators/dsl-gen/generator"
)

func main() {
	gen := kingpin.New("dsl-gen", "Generates helper functions to emit and handle events in DSL modules.")
	inputPkgPath := gen.Arg("package", "The full package path at which the protoc-generated types can be imported. "+
		"It should be possible to import the package from the current working directory.").String()

	if _, err := gen.Parse(os.Args[1:]); err != nil {
		gen.FatalUsage("could not parse arguments: %v\n", err)
	}

	err := codegen.RunGenerator[generator.DslGenerator](*inputPkgPath)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(3)
	}
}
