package membership

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	commonpbtypes "github.com/filecoin-project/mir/pkg/pb/commonpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func FromFileName(fileName string) (*commonpbtypes.Membership, error) {

	// Open file.
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	// Schedule closing file.
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Could not close membership file: %s\n", fileName)
		}
	}()

	// Read membership from file.
	return FromFile(f)
}

func FromFile(f *os.File) (*commonpbtypes.Membership, error) {
	// TODO: Make this function parse the standard JSON membership format.

	membership := &commonpbtypes.Membership{make(map[t.NodeID]*commonpbtypes.NodeIdentity)} // nolint:govet

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			tokens := strings.Fields(scanner.Text())
			id := t.NodeID(tokens[0])

			// Converting the parsed string to (and later back from) multiaddress
			// serves as a check that the address is correct.
			addr, err := multiaddr.NewMultiaddr(tokens[1])
			if err != nil {
				return nil, err
			}

			membership.Nodes[id] = &commonpbtypes.NodeIdentity{id, addr.String(), nil, 0} // nolint:govet
		}
	}

	return membership, nil
}

func GetIPs(membership *commonpbtypes.Membership) (map[t.NodeID]string, error) {
	ips := make(map[t.NodeID]string)
	for nodeID, identity := range membership.Nodes {

		address, err := multiaddr.NewMultiaddr(identity.Addr)
		if err != nil {
			return nil, err
		}

		_, addrPort, err := manet.DialArgs(address)
		if err != nil {
			return nil, err
		}
		ips[nodeID] = strings.Split(addrPort, ":")[0]
	}
	return ips, nil
}

func GetIDs(membership *commonpbtypes.Membership) []t.NodeID {
	return maputil.GetSortedKeys(membership.Nodes)
}

// DummyMultiAddrs augments a membership by deterministically generated host keys based on node IDs.
// The node IDs must be convertable to numbers, otherwise DummyMultiAddrs returns an error.
func DummyMultiAddrs(membershipIn *commonpbtypes.Membership) (*commonpbtypes.Membership, error) {
	membershipOut := &commonpbtypes.Membership{make(map[t.NodeID]*commonpbtypes.NodeIdentity)} // nolint:govet

	for nodeID, identity := range membershipIn.Nodes {
		numericID, err := strconv.Atoi(string(nodeID))
		if err != nil {
			return nil, fmt.Errorf("node IDs must be numeric in the sample app: %w", err)
		}

		newAddr, err := multiaddr.NewMultiaddr(identity.Addr)
		if err != nil {
			return nil, err
		}

		membershipOut.Nodes[nodeID] = &commonpbtypes.NodeIdentity{ // nolint:govet
			identity.Id,
			libp2ptools.NewDummyMultiaddr(numericID, newAddr).String(),
			nil,
			0,
		}
	}

	return membershipOut, nil
}
