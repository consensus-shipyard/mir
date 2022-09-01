// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/cmd/bench/stats"
	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/membership"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
	mirnet "github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	t "github.com/filecoin-project/mir/pkg/types"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

const (
	ReqReceiverBasePort = 20000
)

var (
	transportType string
	statFileName  string
	statPeriod    time.Duration

	nodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Start a Mir node",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNode()
		},
	}
)

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.Flags().StringVarP(&transportType, "net", "t", "libp2p", "network transport (libp2p|grpc)")
	nodeCmd.Flags().StringVarP(&statFileName, "statFile", "o", "", "output file for statistics")
	nodeCmd.Flags().DurationVar(&statPeriod, "statPeriod", time.Second, "statistic record period")
}

func runNode() error {
	var logger logging.Logger
	if verbose {
		logger = logging.ConsoleDebugLogger
	} else {
		logger = logging.ConsoleWarnLogger
	}

	initialMembership, err := membership.FromFileName(membershipFile)
	if err != nil {
		return fmt.Errorf("could not load membership: %w", err)
	}
	addresses, err := membership.GetIPs(initialMembership)
	if err != nil {
		return fmt.Errorf("could not load node IPs: %w", err)
	}

	ownNumericID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("unable to convert node ID: %w", err)
	} else if ownNumericID < 0 || ownNumericID >= len(initialMembership) {
		return fmt.Errorf("ID must be in [0, %d]", len(initialMembership)-1)
	}
	ownID := t.NodeID(id)

	// Generate addresses and ports for client request receivers.
	// Each node uses different ports for receiving protocol messages and requests.
	// These addresses will be used by the client code to know where to send its requests.
	reqReceiverAddrs := make(map[t.NodeID]string)
	for nodeID, nodeIP := range addresses {
		numericID, err := strconv.Atoi(string(nodeID))
		if err != nil {
			return fmt.Errorf("node IDs must be numeric in the sample app: %w", err)
		}
		reqReceiverAddrs[nodeID] = net.JoinHostPort(nodeIP, fmt.Sprintf("%d", ReqReceiverBasePort+numericID))
	}

	// In the current implementation, all nodes are on the local machine, but listen on different port numbers.
	// Change this or make this configurable to deploy different nodes on different physical machines.
	var transport mirnet.Transport
	var nodeAddrs map[t.NodeID]t.NodeAddress
	switch strings.ToLower(transportType) {
	case "grpc":
		transport, err = grpc.NewTransport(ownID, initialMembership[ownID], logger)
		nodeAddrs = initialMembership
	case "libp2p":
		h := libp2ptools.NewDummyHost(ownNumericID, initialMembership[ownID])
		nodeAddrs, err = membership.DummyMultiAddrs(initialMembership)
		if err != nil {
			return fmt.Errorf("could not generate libp2p multiaddresses: %w", err)
		}
		transport, err = libp2p.NewTransport(h, ownID, logger)
	default:
		return fmt.Errorf("unknown network transport %s", strings.ToLower(transportType))
	}
	if err != nil {
		return fmt.Errorf("failed to get network transport %w", err)
	}

	issConfig := iss.DefaultConfig(initialMembership)
	issProtocol, err := iss.New(
		ownID,
		iss.DefaultModuleConfig(), issConfig, iss.InitialStateSnapshot([]byte{}, issConfig),
		logger,
	)
	if err != nil {
		return fmt.Errorf("could not instantiate ISS protocol module: %w", err)
	}

	// Use a simple mempool for incoming requests.
	mempool := simplemempool.NewModule(
		&simplemempool.ModuleConfig{
			Self:   "mempool",
			Hasher: "hasher",
		},
		&simplemempool.ModuleParams{
			MaxTransactionsInBatch: 10,
		},
	)

	// Use fake batch database.
	batchdb := fakebatchdb.NewModule(
		&fakebatchdb.ModuleConfig{
			Self: "batchdb",
		},
	)

	// Instantiate a reconfigurable availability layer.
	availability := multisigcollector.NewReconfigurableModule(
		&multisigcollector.ModuleConfig{
			Self:    "availability",
			Mempool: "mempool",
			BatchDB: "batchdb",
			Net:     "net",
			Crypto:  "crypto",
		},
		ownID,
		logger,
	)

	// The batch fetcher receives availability certificates from the ordering protocol (ISS),
	// retrieves the corresponding transaction batches, and delivers them to the application.
	batchFetcher := batchfetcher.NewModule(batchfetcher.DefaultModuleConfig())

	cryptoSystem := deploytest.NewLocalCryptoSystem("pseudo", membership.GetIDs(initialMembership), logger)

	stat := stats.NewStats()
	interceptor := stats.NewStatInterceptor(stat, "app")

	nodeModules, err := iss.DefaultModules(modules.Modules{
		"net":          transport,
		"crypto":       cryptoSystem.Module(ownID),
		"iss":          issProtocol,
		"app":          &App{Logger: logger, ProtocolModule: "iss", Membership: nodeAddrs},
		"batchfetcher": batchFetcher,
		"mempool":      mempool,
		"batchdb":      batchdb,
		"availability": availability,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Mir modules: %w", err)
	}

	nodeConfig := &mir.NodeConfig{Logger: logger}
	node, err := mir.NewNode(t.NodeID(id), nodeConfig, nodeModules, nil, interceptor)
	if err != nil {
		return fmt.Errorf("could not create node: %w", err)
	}

	ctx := context.Background()

	reqReceiver := requestreceiver.NewRequestReceiver(node, "mempool", logger)
	if err := reqReceiver.Start(ReqReceiverBasePort + ownNumericID); err != nil {
		return fmt.Errorf("could not start request receiver: %w", err)
	}
	defer reqReceiver.Stop()

	if err := transport.Start(); err != nil {
		return fmt.Errorf("could not start network transport: %w", err)
	}
	transport.Connect(ctx, nodeAddrs)
	defer transport.Stop()

	var statFile *os.File
	if statFileName != "" {
		statFile, err = os.Create(statFileName)
		if err != nil {
			return fmt.Errorf("could not open output file for statistics: %w", err)
		}
	} else {
		statFile = os.Stdout
	}

	statCSV := csv.NewWriter(statFile)
	stat.WriteCSVHeader(statCSV)

	go func() {
		timestamp := time.Now()
		for {
			ticker := time.NewTicker(statPeriod)
			defer ticker.Stop()

			select {
			case <-ctx.Done():
				return
			case ts := <-ticker.C:
				d := ts.Sub(timestamp)
				stat.WriteCSVRecord(statCSV, d)
				statCSV.Flush()
				timestamp = ts
				stat.Reset()
			}
		}
	}()

	defer node.Stop()
	return node.Run(ctx)
}
