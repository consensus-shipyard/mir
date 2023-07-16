package main

import (
	"log"
	"os"
	"reflect"

	generator_ "github.com/filecoin-project/mir/codegen/generators/mir-std-gen/generator"
	pkg_ "github.com/filecoin-project/mir/pkg/pb/eventpb"
)

func main() {
	var generator generator_.CombinedGenerator
	err := generator.Run(
		[]reflect.Type{

			reflect.TypeOf((*pkg_.Event)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Init)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Timer)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Hasher)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Bcb)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Mempool)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Availability)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_BatchDb)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_BatchFetcher)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_ThreshCrypto)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Checkpoint)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Factory)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Iss)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Orderer)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Crypto)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_App)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_Transport)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_PingPong)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_TestingString)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Event_TestingUint)(nil)).Elem(),

			reflect.TypeOf((*pkg_.Init)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerEvent)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerEvent_Delay)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerEvent_Repeat)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerEvent_GarbageCollect)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerDelay)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerRepeat)(nil)).Elem(),

			reflect.TypeOf((*pkg_.TimerGarbageCollect)(nil)).Elem(),
		})

	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(2)
	}
}
