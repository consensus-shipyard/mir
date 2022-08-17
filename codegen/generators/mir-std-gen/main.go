package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/generators/mir-std-gen/generator"
)

func main() {
	gen := kingpin.New("mir-std-gen", "Runs a set of standard Mir code generators. "+
		"Namely, types-gen, events-gen, dsl-gen, and net-gen. "+
		"The generators are bundled together in one binary for better performance. "+
		"See the help messages and the documentation for the corresponding tools for details.")
	inputPkgPath := gen.Arg("package", "The full package path at which the protoc-generated types can be imported. "+
		"It should be possible to import the package from the current working directory.").String()

	if _, err := gen.Parse(os.Args[1:]); err != nil {
		gen.FatalUsage("could not parse arguments: %v\n", err)
	}

	err := codegen.RunGenerator[generator.CombinedGenerator](*inputPkgPath)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(3)
	}
}
