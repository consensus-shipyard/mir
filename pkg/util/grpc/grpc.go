package grpc

import (
	"fmt"

	"github.com/multiformats/go-multiaddr"
)

// NewDummyMultiaddr generates a gRPC node multiaddress.
func NewDummyMultiaddr(port int) multiaddr.Multiaddr {
	a, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err != nil {
		panic(err)
	}
	return a
}
