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
	"time"

	es "github.com/go-errors/errors"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/spf13/cobra"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/cmd/bench/stats"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/membership"
	libp2p2 "github.com/filecoin-project/mir/pkg/net/libp2p"
	"github.com/filecoin-project/mir/pkg/pb/batchfetcherpb"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/mempoolpb"
	"github.com/filecoin-project/mir/pkg/transactionreceiver"
	"github.com/filecoin-project/mir/pkg/trantor"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
)

const (
	TxReceiverBasePort = 20000
)

var (
	statFileName string
	statPeriod   time.Duration

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

	ctx := context.Background()

	// Load system membership.
	nodeAddrs, err := membership.FromFileName(membershipFile)
	if err != nil {
		return es.Errorf("could not load membership: %w", err)
	}
	initialMembership, err := membership.DummyMultiAddrs(nodeAddrs)
	if err != nil {
		return es.Errorf("could not create dummy multiaddrs: %w", err)
	}

	// Parse own ID.
	ownNumericID, err := strconv.Atoi(id)
	if err != nil {
		return es.Errorf("unable to convert node ID: %w", err)
	} else if ownNumericID < 0 || ownNumericID >= len(initialMembership.Nodes) {
		return es.Errorf("ID must be in [0, %d]", len(initialMembership.Nodes)-1)
	}
	ownID := t.NodeID(id)

	// Set Trantor parameters.
	smrParams := trantor.DefaultParams(initialMembership)
	smrParams.Mempool.MaxTransactionsInBatch = 1024

	// Assemble listening address.
	// In this benchmark code, we always listen on tha address 0.0.0.0.
	portStr, err := getPortStr(initialMembership.Nodes[ownID].Addr)
	if err != nil {
		return es.Errorf("could not parse port from own address: %w", err)
	}
	addrStr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", portStr)
	listenAddr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return es.Errorf("could not create listen address: %w", err)
	}
	h, err := libp2p.NewDummyHostWithPrivKey(
		t.NodeAddress(libp2p.NewDummyMultiaddr(ownNumericID, listenAddr)),
		libp2p.NewDummyHostKey(ownNumericID),
	)
	if err != nil {
		return es.Errorf("failed to create libp2p host: %w", err)
	}

	// Initialize the libp2p transport subsystem.
	transport := libp2p2.NewTransport(smrParams.Net, ownID, h, logger)

	localCryptoSystem, err := deploytest.NewLocalCryptoSystem("pseudo", membership.GetIDs(initialMembership), logger)
	if err != nil {
		return es.Errorf("could not create a local crypto system: %w", err)
	}
	localCrypto, err := localCryptoSystem.Crypto(ownID)
	if err != nil {
		return es.Errorf("could not create a local crypto module: %w", err)
	}
	genesisCheckpoint, err := trantor.GenesisCheckpoint([]byte{}, smrParams)
	if err != nil {
		return es.Errorf("could not create genesis checkpoint: %w", err)
	}
	benchApp, err := trantor.New(
		ownID,
		transport,
		genesisCheckpoint,
		localCrypto,
		&App{Logger: logger, Membership: initialMembership},
		smrParams,
		logger,
	)
	if err != nil {
		return es.Errorf("could not create bench app: %w", err)
	}

	recorder, err := eventlog.NewRecorder(
		ownID,
		"bench-output",
		logging.Decorate(logger, "EVTLOG: "),
		eventlog.EventFilterOpt(func(e *eventpb.Event) bool {
			switch e := e.Type.(type) {
			case *eventpb.Event_Mempool:
				switch e.Mempool.Type.(type) {
				case *mempoolpb.Event_NewTransactions:
					return true
				}
			case *eventpb.Event_BatchFetcher:
				switch e.BatchFetcher.Type.(type) {
				case *batchfetcherpb.Event_NewOrderedBatch:
					return true
				}
			}
			return false
		}),
	)
	if err != nil {
		return es.Errorf("cannot create event recorder: %w", err)
	}
	stat := stats.NewStats()
	interceptor := eventlog.MultiInterceptor(
		stats.NewStatInterceptor(stat, "app"),
		recorder,
	)

	nodeConfig := mir.DefaultNodeConfig().WithLogger(logger)
	node, err := mir.NewNode(t.NodeID(id), nodeConfig, benchApp.Modules(), interceptor)
	if err != nil {
		return es.Errorf("could not create node: %w", err)
	}

	txReceiver := transactionreceiver.NewTransactionReceiver(node, "mempool", logger)
	if err := txReceiver.Start(TxReceiverBasePort + ownNumericID); err != nil {
		return es.Errorf("could not start transaction receiver: %w", err)
	}
	defer txReceiver.Stop()

	if err := benchApp.Start(); err != nil {
		return es.Errorf("could not start bench app: %w", err)
	}
	defer benchApp.Stop()

	var statFile *os.File
	if statFileName != "" {
		statFile, err = os.Create(statFileName)
		if err != nil {
			return es.Errorf("could not open output file for statistics: %w", err)
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

func getPortStr(addressStr string) (string, error) {
	address, err := multiaddr.NewMultiaddr(addressStr)
	if err != nil {
		return "", err
	}

	_, addrStr, err := manet.DialArgs(address)
	if err != nil {
		return "", err
	}

	_, portStr, err := net.SplitHostPort(addrStr)
	if err != nil {
		return "", err
	}

	return portStr, nil
}
