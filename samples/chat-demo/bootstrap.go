package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	t "github.com/filecoin-project/mir/pkg/types"
)

func loadMembership(f *os.File) (map[t.NodeID]t.NodeAddress, error) {

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

func membershipIPs(membership map[t.NodeID]t.NodeAddress) (map[t.NodeID]string, error) {
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
