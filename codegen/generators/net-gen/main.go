package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/codegen"
	"github.com/filecoin-project/mir/codegen/generators/net-gen/generator"
)

func main() {
	gen := kingpin.New("net-gen", "Generates constructor functions for net messages.")
	inputPkgPath := gen.Arg("package", "The full package path at which the protoc-generated types can be imported. "+
		"It should be possible to import the package from the current working directory.").String()

	if _, err := gen.Parse(os.Args[1:]); err != nil {
		gen.FatalUsage("could not parse arguments: %v\n", err)
	}

	err := codegen.RunGenerator[generator.NetMsgGenerator](*inputPkgPath)
	if err != nil {
		log.Println("Error:", err)
		os.Exit(3)
	}
}
