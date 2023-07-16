package main

import (
	"log"
	"os"
	"reflect"

	generator_ "github.com/filecoin-project/mir/codegen/generators/mir-std-gen/generator"
	pkg_ "github.com/filecoin-project/mir/pkg/pb/bcbpb"
)

func main() {
	var generator generator_.CombinedGenerator
	err := generator.Run(
		[]reflect.Type{

			reflect.TypeOf((*pkg_.Event)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Request)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Deliver)(nil)).Elem(),

			reflect.TypeOf((*pkg_.BroadcastRequest)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Deliver)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Message)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Message_StartMessage)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Message_EchoMessage)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Message_FinalMessage)(nil)).Elem(),

			reflect.TypeOf((*pkg_.StartMessage)(nil)).Elem(),

			reflect.TypeOf((*pkg_.EchoMessage)(nil)).Elem(),

			reflect.TypeOf((*pkg_.FinalMessage)(nil)).Elem(),
		})

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}
