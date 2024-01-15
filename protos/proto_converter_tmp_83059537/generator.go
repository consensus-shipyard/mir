
package main

import (
	"log"
	"os"
	"reflect"

	generator_ "github.com/filecoin-project/mir/codegen/generators/mir-std-gen/generator"
	pkg_ "github.com/filecoin-project/mir/pkg/pb/pingpongpb"
)

func main() {
	var generator generator_.CombinedGenerator
	err := generator.Run(
		[]reflect.Type{
			
			reflect.TypeOf((*pkg_.Event)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Event_PingTime)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.PingTime)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Message)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Message_Ping)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Message_Pong)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Ping)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Pong)(nil)).Elem(),
			
		})

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}
