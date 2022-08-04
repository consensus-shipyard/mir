// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/filecoin-project/mir"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	t "github.com/filecoin-project/mir/pkg/types"
	grpctools "github.com/filecoin-project/mir/pkg/util/grpc"
	libp2ptools "github.com/filecoin-project/mir/pkg/util/libp2p"
)

const (
	NodeBasePort        = 10000
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

	ownID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("unable to convert node ID: %w", err)
	} else if ownID < 0 || ownID >= nrNodes {
		return fmt.Errorf("ID must be in [0, %d]", nrNodes-1)
	}

	nodeIDs := make([]t.NodeID, nrNodes)
	for i := range nodeIDs {
		nodeIDs[i] = t.NewNodeIDFromInt(i)
	}

	reqReceiverAddrs := make(map[t.NodeID]string)
	for i := range nodeIDs {
		reqReceiverAddrs[t.NewNodeIDFromInt(i)] = fmt.Sprintf("127.0.0.1:%d", ReqReceiverBasePort+i)
	}

	var transport net.Transport
	nodeAddrs := make(map[t.NodeID]t.NodeAddress)
	switch strings.ToLower(transportType) {
	case "grpc":
		for i := range nodeIDs {
			nodeAddrs[t.NewNodeIDFromInt(i)] = t.NodeAddress(grpctools.NewDummyMultiaddr(i + NodeBasePort))
		}
		transport, err = grpc.NewTransport(t.NodeID(id), nodeAddrs[t.NodeID(id)], logger)
	case "libp2p":
		h := libp2ptools.NewDummyHost(ownID, NodeBasePort)
		for i := range nodeIDs {
			nodeAddrs[t.NewNodeIDFromInt(i)] = t.NodeAddress(libp2ptools.NewDummyMultiaddr(i, NodeBasePort))
		}
		transport, err = libp2p.NewTransport(h, t.NodeID(id), logger)
	default:
		return fmt.Errorf("unknown network transport %s", strings.ToLower(transportType))
	}
	if err != nil {
		return fmt.Errorf("failed to get network transport %w", err)
	}

	issConfig := iss.DefaultConfig(nodeIDs)
	issProtocol, err := iss.New(t.NodeID(id), issConfig, logger)
	if err != nil {
		return fmt.Errorf("could not instantiate ISS protocol module: %w", err)
	}

	// Use dummy crypto module that only produces signatures
	// consisting of a single zero byte and treats those signatures as valid.
	// TODO: Adjust once a default crypto implementation is provided by Mir.
	crypto := mirCrypto.New(&mirCrypto.DummyCrypto{DummySig: []byte{0}})

	stats := NewStats()
	interceptor := NewStatInterceptor(stats)

	nodeModules, err := iss.DefaultModules(modules.Modules{
		"net":    transport,
		"crypto": crypto,
		"iss":    issProtocol,
		"app":    &App{Logger: logger},
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

	reqReceiver := requestreceiver.NewRequestReceiver(node, "iss", logger)
	if err := reqReceiver.Start(ReqReceiverBasePort + ownID); err != nil {
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

	statsCSV := csv.NewWriter(statFile)
	stats.WriteCSVHeader(statsCSV)

	go func() {
		timestamp := time.Now()
		for {
			ticker := time.NewTicker(statPeriod)
			defer ticker.Stop()

			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				d := t.Sub(timestamp)
				stats.WriteCSVRecord(statsCSV, d)
				statsCSV.Flush()
				timestamp = t
				stats.Reset()
			}
		}
	}()

	defer node.Stop()
	return node.Run(ctx)
}
