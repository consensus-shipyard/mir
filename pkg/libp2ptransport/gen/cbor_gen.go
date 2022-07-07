package main

import (
	gen "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/mir/pkg/libp2ptransport"
)

func main() {
	if err := gen.WriteTupleEncodersToFile("./cbor_gen.go", "libp2ptransport",
		libp2ptransport.TransportMessage{},

	); err != nil {
		panic(err)
	}
}

