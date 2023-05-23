package membership

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	es "github.com/go-errors/errors"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
	"github.com/filecoin-project/mir/pkg/util/maputil"
)

func FromFileName(fileName string) (*trantorpbtypes.Membership, error) {

	// Open file.
	f, err := os.Open(fileName)
	if err != nil {
		return nil, es.Errorf("could not open file: %w", err)
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

// FromFile reads a membership for a file. It expects a text file containing valid JSON with the following format.
// (The strange field names are a result of trying to be compatible with the format used by Lotus/Eudico.)
// TODO: Use a better format for the membership, maybe even if it's incompatible with Eudico.
func FromFile(f *os.File) (*trantorpbtypes.Membership, error) {
	// Sample input:
	// {
	//     "configuration_number": 0,
	//     "validators": [
	//         {
	//             "addr": "t1dgw4345grpw53zdhu75dc6jj4qhrh4zoyrtq6di",
	//             "net_addr": "/ip4/172.31.39.78/tcp/43077/p2p/12D3KooWNzTunrQtcoo4SLWNdQ4EdFWSZtah6mgU44Q5XWM61aan",
	//             "weight": "1"
	//         },
	//         {
	//             "addr": "t1a5gxsoogaofa5nzfdh66l6uynx4m6m4fiqvcx6y",
	//             "net_addr": "/ip4/172.31.33.169/tcp/38257/p2p/12D3KooWABvxn3CHjz9r5TYGXGDqm8549VEuAyFpbkH8xWkNLSmr",
	//             "weight": "1"
	//         },
	//         {
	//             "addr": "t1q4j6esoqvfckm7zgqfjynuytjanbhirnbwfrsty",
	//             "net_addr": "/ip4/172.31.42.15/tcp/44407/p2p/12D3KooWGdQGu1utYP6KD1Cq4iXTLV6hbZa8yQN34zwuHNP5YbCi",
	//             "weight": "1"
	//         },
	//         {
	//             "addr": "t16biatgyushsfcidabfy2lm5wo22ppe6r7ddir6y",
	//             "net_addr": "/ip4/172.31.47.117/tcp/34355/p2p/12D3KooWEtfTyoWW7pFLsErAb6jPiQQCC3y3junHtLn9jYnFHei8",
	//             "weight": "1"
	//         }
	//     ]
	// }

	// Define data structure to load from JSON.
	data := struct {
		// The configuration_number field (if present) is ignored.
		Identities []struct {
			ID     string `json:"addr"`
			Addr   string `json:"net_addr"`
			Weight uint64 `json:"weight,string"`
		} `json:"validators"`
	}{}

	// Decode the JSON data.
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	// Construct membership from dedoced data.
	membership := &trantorpbtypes.Membership{make(map[t.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet
	for _, identity := range data.Identities {
		membership.Nodes[t.NodeID(identity.ID)] = &trantorpbtypes.NodeIdentity{ // nolint:govet
			t.NodeID(identity.ID),
			identity.Addr,
			nil,
			tt.VoteWeight(identity.Weight),
		}
	}

	return membership, nil
}

func GetIPs(membership *trantorpbtypes.Membership) (map[t.NodeID]string, error) {
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

func GetIDs(membership *trantorpbtypes.Membership) []t.NodeID {
	return maputil.GetSortedKeys(membership.Nodes)
}

// DummyMultiAddrs augments a membership by deterministically generated host keys based on node IDs.
// The node IDs must be convertable to numbers, otherwise DummyMultiAddrs returns an error.
func DummyMultiAddrs(membershipIn *trantorpbtypes.Membership) (*trantorpbtypes.Membership, error) {
	membershipOut := &trantorpbtypes.Membership{make(map[t.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet

	for nodeID, identity := range membershipIn.Nodes {
		numericID, err := strconv.Atoi(string(nodeID))
		if err != nil {
			return nil, es.Errorf("node IDs must be numeric in the sample app: %w", err)
		}

		newAddr, err := multiaddr.NewMultiaddr(identity.Addr)
		if err != nil {
			return nil, err
		}

		membershipOut.Nodes[nodeID] = &trantorpbtypes.NodeIdentity{ // nolint:govet
			identity.Id,
			libp2ptools.NewDummyMultiaddr(numericID, newAddr).String(),
			nil,
			1,
		}
	}

	return membershipOut, nil
}
