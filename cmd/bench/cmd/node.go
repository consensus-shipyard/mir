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

	"github.com/filecoin-project/mir/pkg/localtxgenerator"
	"github.com/filecoin-project/mir/pkg/rendezvous"
	"github.com/filecoin-project/mir/pkg/trantor/appmodule"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
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
	"github.com/filecoin-project/mir/pkg/trantor"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/libp2p"
)

const (
	syncLimit        = 30 * time.Second
	syncPollInterval = 100 * time.Millisecond
)

var (
	configFileName      string
	liveStatFileName    string
	clientStatFileName  string
	statPeriod          time.Duration
	readySyncFileName   string
	deliverSyncFileName string

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

	// Required arguments
	nodeCmd.Flags().StringVarP(&configFileName, "config-file", "c", "", "configuration file")
	_ = nodeCmd.MarkFlagRequired("config-file")
	nodeCmd.PersistentFlags().StringVarP(&id, "id", "i", "", "node ID")
	_ = nodeCmd.MarkPersistentFlagRequired("id")

	// Optional arguments
	nodeCmd.Flags().DurationVar(&statPeriod, "stat-period", time.Second, "statistic record period")
	nodeCmd.Flags().StringVar(&clientStatFileName, "client-stat-file", "bench-output.json", "statistics output file")
	nodeCmd.Flags().StringVar(&liveStatFileName, "live-stat-file", "", "output file for live statistics, default is standard output")

	// Sync files
	nodeCmd.Flags().StringVar(&readySyncFileName, "ready-sync-file", "", "file to use for initial synchronization when ready to start the benchmark")
	nodeCmd.Flags().StringVar(&deliverSyncFileName, "deliver-sync-file", "", "file to use for synchronization when waiting to deliver all transactions")
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
	trantorParams := trantor.DefaultParams(initialMembership)
	trantorParams.Mempool.MaxTransactionsInBatch = 1024
	trantorParams.Iss.MaxProposeDelay = 0 * time.Millisecond
	trantorParams.Iss.PBFTViewChangeSNTimeout = 8 * time.Second
	trantorParams.Iss.PBFTViewChangeSegmentTimeout = time.Duration(trantorParams.Iss.SegmentLength) * trantorParams.Iss.PBFTViewChangeSNTimeout

	// Assemble listening address.
	// In this benchmark code, we always listen on the address 0.0.0.0.
	portStr, err := getPortStr(initialMembership.Nodes[ownID].Addr)
	if err != nil {
		return es.Errorf("could not parse port from own address: %w", err)
	}
	addrStr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", portStr)
	listenAddr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return es.Errorf("could not create listen address: %w", err)
	}

	// Create libp2p host
	h, err := libp2p.NewDummyHostWithPrivKey(
		t.NodeAddress(libp2p.NewDummyMultiaddr(ownNumericID, listenAddr)),
		libp2p.NewDummyHostKey(ownNumericID),
	)
	if err != nil {
		return es.Errorf("failed to create libp2p host: %w", err)
	}

	// Initialize the libp2p transport subsystem.
	transport := libp2p2.NewTransport(trantorParams.Net, ownID, h, logger)

	// Instantiate the crypto module.
	localCryptoSystem, err := deploytest.NewLocalCryptoSystem("pseudo", membership.GetIDs(initialMembership), logger)
	if err != nil {
		return es.Errorf("could not create a local crypto system: %w", err)
	}
	localCrypto, err := localCryptoSystem.Crypto(ownID)
	if err != nil {
		return es.Errorf("could not create a local crypto module: %w", err)
	}

	// Generate the initial checkpoint.
	genesisCheckpoint, err := trantor.GenesisCheckpoint([]byte{}, trantorParams)
	if err != nil {
		return es.Errorf("could not create genesis checkpoint: %w", err)
	}

	// Create a local transaction generator. It has, at the same time, the interface of a trantor App,
	// So it knows when transactions are delivered and can submit new ones accordingly.
	txGenParams := localtxgenerator.DefaultModuleParams(tt.ClientID(ownID))
	// Use this line instead of the above one to simulate submitting transactions to all nodes.
	//txGenParams := localtxgenerator.DefaultModuleParams(t.NewClientIDFromInt(0))
	txGenParams.BufSize = 100
	txGenParams.Tps = 100
	txGen := localtxgenerator.New(
		localtxgenerator.DefaultModuleConfig(),
		txGenParams,
	)

	// Create a Trantor instance.
	trantorInstance, err := trantor.New(
		ownID,
		transport,
		genesisCheckpoint,
		localCrypto,
		appmodule.AppLogicFromStatic(txGen, initialMembership), // The transaction generator is also a static app.
		trantorParams,
		logger,
	)
	if err != nil {
		return es.Errorf("could not create bench app: %w", err)
	}

	// Add transaction generator module to the setup.
	trantorInstance.WithModule("localtxgen", txGen)

	// Create recorder for gathering statistics about the performance.
	stat := stats.NewStats()
	statsRecorder := stats.NewStatInterceptor(stat, "app")

	// Assemble the main event interceptor.
	// We could use the statsRecorder directly, but this construction makes it convenient to add more when needed.
	interceptor := eventlog.MultiInterceptor(
		statsRecorder,
	)

	// Instantiate the Mir Node.
	nodeConfig := mir.DefaultNodeConfig().WithLogger(logger)
	nodeConfig.Stats.Period = time.Second
	node, err := mir.NewNode(t.NodeID(id), nodeConfig, trantorInstance.Modules(), interceptor)
	if err != nil {
		return es.Errorf("could not create node: %w", err)
	}

	if err := trantorInstance.Start(); err != nil {
		return es.Errorf("could not start bench app: %w", err)
	}
	defer trantorInstance.Stop()

	if err := transport.WaitFor(len(initialMembership.Nodes)); err != nil {
		return es.Errorf("failed waiting for network connections: %w", err)
	}

	// Synchronize with other nodes if necessary.
	// If invoked, this code blocks until all the nodes have connected to each other.
	// (The file created by Ready must be deleted by some external code (or manually) after all nodes have created it.)
	if readySyncFileName != "" {
		syncCtx, cancelFunc := context.WithTimeout(ctx, syncLimit)
		err = rendezvous.NewFileSyncer(readySyncFileName, syncPollInterval).Ready(syncCtx)
		cancelFunc()
		if err != nil {
			return fmt.Errorf("error synchronizing nodes: %w", err)
		}
	}

	// Start generating the load.
	txGen.Start()
	defer txGen.Stop()

	// Output the statistics.
	var statFile *os.File
	if liveStatFileName != "" {
		statFile, err = os.Create(liveStatFileName)
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
