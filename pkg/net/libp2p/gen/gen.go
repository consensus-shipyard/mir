package main

import (
	gen "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/mir/pkg/net/libp2p"
)

func main() {
	if err := gen.WriteTupleEncodersToFile("./cbor_gen.go", "libp2p",
		libp2p.TransportMessage{},
	); err != nil {
		panic(err)
	}
}
