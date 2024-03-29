package stdtypes

import (
	"github.com/multiformats/go-multiaddr"
)

// NodeAddress represents the address of a node.
type NodeAddress multiaddr.Multiaddr

func NodeAddressFromString(addrString string) (NodeAddress, error) {
	return multiaddr.NewMultiaddr(addrString)
}
