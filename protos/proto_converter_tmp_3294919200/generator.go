
package main

import (
	"log"
	"os"
	"reflect"

	generator_ "github.com/filecoin-project/mir/codegen/generators/mir-std-gen/generator"
	pkg_ "github.com/filecoin-project/mir/pkg/pb/hasherpb"
)

func main() {
	var generator generator_.CombinedGenerator
	err := generator.Run(
		[]reflect.Type{
			
			reflect.TypeOf((*pkg_.Event)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Event_Request)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Event_Result)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Event_RequestOne)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Event_ResultOne)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Request)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.Result)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.RequestOne)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.ResultOne)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.HashOrigin)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.HashOrigin_ContextStore)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.HashOrigin_Dsl)(nil)).Elem(),
			
			reflect.TypeOf((*pkg_.HashData)(nil)).Elem(),
			
		})

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}
