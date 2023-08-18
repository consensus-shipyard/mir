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

	"github.com/filecoin-project/mir/cmd/bench/localtxgenerator"
	"github.com/filecoin-project/mir/pkg/rendezvous"
	"github.com/filecoin-project/mir/pkg/trantor/appmodule"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	es "github.com/go-errors/errors"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/spf13/cobra"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/cmd/bench/stats"
	"github.com/filecoin-project/mir/pkg/deploytest"
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
			if err := runNode(); !es.Is(err, mir.ErrStopped) {
				return err
			}
			return nil
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

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	// Load configuration parameters
	var params BenchParams
	if err := loadFromFile(configFileName, &params); err != nil {
		return es.Errorf("could not load parameters from file '%s': %w", configFileName, err)
	}

	// Parse own ID.
	ownNumericID, err := strconv.Atoi(id)
	if err != nil {
		return es.Errorf("unable to convert node ID: %w", err)
	}

	// Check if own id is in the membership
	initialMembership := params.Trantor.Iss.InitialMembership
	if _, ok := initialMembership.Nodes[t.NodeID(id)]; !ok {
		return es.Errorf("own ID (%v) not found in membership (%v)", id, maputil.GetKeys(initialMembership.Nodes))
	}
	ownID := t.NodeID(id)

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
	transport := libp2p2.NewTransport(params.Trantor.Net, ownID, h, logger)

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
	genesisCheckpoint, err := trantor.GenesisCheckpoint([]byte{}, params.Trantor)
	if err != nil {
		return es.Errorf("could not create genesis checkpoint: %w", err)
	}

	// Create a local transaction generator.
	// It has, at the same time, the interface of a trantor App,
	// so it knows when transactions are delivered and can submit new ones accordingly.
	// If the client ID is not specified, use the local node's ID
	if params.TxGen.ClientID == "" {
		params.TxGen.ClientID = tt.ClientID(ownID)
	}
	txGen := localtxgenerator.New(localtxgenerator.DefaultModuleConfig(), params.TxGen)

	// Create a Trantor instance.
	trantorInstance, err := trantor.New(
		ownID,
		transport,
		genesisCheckpoint,
		localCrypto,
		appmodule.AppLogicFromStatic(txGen, initialMembership), // The transaction generator is also a static app.
		params.Trantor,
		logger,
	)
	if err != nil {
		return es.Errorf("could not create bench app: %w", err)
	}

	// Add transaction generator module to the setup.
	trantorInstance.WithModule("localtxgen", txGen)

	// Create trackers for gathering statistics about the performance.
	liveStats := stats.NewLiveStats()
	txGen.TrackStats(liveStats)

	// Instantiate the Mir Node.
	nodeConfig := mir.DefaultNodeConfig().WithLogger(logger)
	nodeConfig.Stats.Period = time.Second
	node, err := mir.NewNode(t.NodeID(id), nodeConfig, trantorInstance.Modules(), nil)
	if err != nil {
		return es.Errorf("could not create node: %w", err)
	}

	if err := trantorInstance.Start(); err != nil {
		return es.Errorf("could not start bench app: %w", err)
	}

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
	liveStats.WriteCSVHeader(statCSV)
	liveStatsCtx, stopLiveStats := context.WithCancel(ctx)

	go func() {
		timestamp := time.Now()
		ticker := time.NewTicker(statPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-liveStatsCtx.Done():
				return
			case ts := <-ticker.C:
				d := ts.Sub(timestamp)
				liveStats.WriteCSVRecord(statCSV, d)
				statCSV.Flush()
				timestamp = ts
				liveStats.Reset()
			}
		}
	}()

	// Stop outputting real-time stats and submitting transactions,
	// wait until everything is delivered, and stop node.
	shutDown := func() {

		stopLiveStats()
		txGen.Stop()
		txGenCtx, stopTxGen := context.WithTimeout(ctx, syncLimit)
		if err := txGen.Wait(txGenCtx); err != nil {
			logger.Log(logging.LevelError, "Not all submitted transactions delivered.", "error", err)
		} else {
			logger.Log(logging.LevelInfo, "Successfully delivered all submitted transactions.", "error", err)
		}
		stopTxGen()

		// Wait for other nodes to deliver their transactions.
		if deliverSyncFileName != "" {
			syncerCtx, stopWaiting := context.WithTimeout(ctx, syncLimit)
			err := rendezvous.NewFileSyncer(deliverSyncFileName, syncPollInterval).Ready(syncerCtx)
			stopWaiting()
			if err != nil {
				logger.Log(logging.LevelError, "Aborting waiting for other nodes transaction delivery.", "error", err)
			} else {
				logger.Log(logging.LevelInfo, "All nodes successfully delivered all transactions they submitted.", "error", err)
			}
		}

		// Stop Mir node and Trantor instance.
		logger.Log(logging.LevelInfo, "Stopping Mir node.")
		node.Stop()
		logger.Log(logging.LevelInfo, "Mir node stopped.")
		logger.Log(logging.LevelInfo, "Stopping Trantor.")
		trantorInstance.Stop()
		logger.Log(logging.LevelInfo, "Trantor stopped.")

	}

	done := make(chan struct{})
	if params.Duration > 0 {
		go func() {
			// Wait until the end of the benchmark and shut down the node.
			select {
			case <-ctx.Done():
			case <-time.After(params.Duration):
			}
			shutDown()
			close(done)
		}()
	} else {
		// TODO: This is not right. Only have this branch to quit on node error.
		//   Set up signal handlers so that the nodes stops and cleans up after itself upon SIGINT and / or SIGTERM.
		close(done)
	}

	// Start generating the load and measuring performance.
	txGen.Start()

	nodeError := node.Run(ctx)
	<-done
	return nodeError
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
