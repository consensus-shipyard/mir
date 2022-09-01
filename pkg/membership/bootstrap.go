package membership

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

func FromFileName(fileName string) (map[t.NodeID]t.NodeAddress, error) {

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

func FromFile(f *os.File) (map[t.NodeID]t.NodeAddress, error) {

	membership := make(map[t.NodeID]t.NodeAddress)

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			tokens := strings.Fields(scanner.Text())
			var err error
			membership[t.NodeID(tokens[0])], err = multiaddr.NewMultiaddr(tokens[1])
			if err != nil {
				return nil, err
			}
		}
	}

	return membership, nil
}

func GetIPs(membership map[t.NodeID]t.NodeAddress) (map[t.NodeID]string, error) {
	ips := make(map[t.NodeID]string)
	for nodeID, multiAddr := range membership {
		_, addrPort, err := manet.DialArgs(multiAddr)
		if err != nil {
			return nil, err
		}
		ips[nodeID] = strings.Split(addrPort, ":")[0]
	}
	return ips, nil
}

func GetIDs(membership map[t.NodeID]t.NodeAddress) []t.NodeID {
	ids := make([]t.NodeID, 0, len(membership))
	for nodeID := range membership {
		ids = append(ids, nodeID)
	}
	return ids
}

// DummyMultiAddrs returns a set of libp2p multiaddresses based on the given membership,
// generating host keys deterministically based on node IDs.
// The node IDs must be convertable to numbers, otherwise DummyMultiAddrs returns an error.
func DummyMultiAddrs(membership map[t.NodeID]t.NodeAddress) (map[t.NodeID]t.NodeAddress, error) {
	nodeAddrs := make(map[t.NodeID]t.NodeAddress)

	for nodeID, nodeAddr := range membership {
		numericID, err := strconv.Atoi(string(nodeID))
		if err != nil {
			return nil, fmt.Errorf("node IDs must be numeric in the sample app: %w", err)
		}
		nodeAddrs[nodeID] = t.NodeAddress(libp2ptools.NewDummyMultiaddr(numericID, nodeAddr))
	}

	return nodeAddrs, nil
}
