/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"context"
	"crypto"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/dummyclient"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/testsim"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
	"github.com/filecoin-project/mir/pkg/util/errstack"
)

const (
	// BaseListenPort defines the starting port number on which test replicas will be listening
	// in case the test is being run with the "grpc" or "libp2p" setting for networking.
	// A node with numeric ID id will listen on port (BaseListenPort + id)
	BaseListenPort = 10000

	// TXListenPort is the port number on which nodes' TransactionReceivers listen for incoming transactions.
	TXListenPort = 20000
)

// TestConfig contains the parameters of the deployment to be tested.
type TestConfig struct {
	// Optional information about the test.
	Info string

	// The test simulation is only used if the deployment is configured to use "sim" transport.
	Simulation *Simulation

	// IDs of nodes in this test.
	NodeIDs []t.NodeID

	// List of nodes.
	Membership *trantorpbtypes.Membership

	// The modules that will be run by each replica.
	NodeModules map[t.NodeID]modules.Modules

	// Number of clients in the tested deployment.
	NumClients int

	// The number of transactions each client submits during the execution of the deployment.
	NumFakeTXs int

	// The number of transactions sent over the network (by a single DummyClient)
	NumNetTXs int

	// The target module for the clients' transactions.
	FakeTXDestModule t.ModuleID

	// Directory where all the test-related files will be stored.
	// If empty, an OS-default temporary directory will be used.
	Directory string

	// Logger to use for producing diagnostic messages.
	Logger logging.Logger

	// TransportLayer to work with network transport.
	TransportLayer LocalTransportLayer

	// Fake applications of all the test replicas. Required for checking results.
	FakeApps map[t.NodeID]*FakeApp
}

// The Deployment represents a list of replicas interconnected by a simulated network transport.
type Deployment struct {
	TestConfig *TestConfig

	// The test simulation is only used if the deployment is configured to use "sim" transport.
	Simulation *Simulation

	// The replicas of the deployment.
	TestReplicas []*TestReplica

	// Dummy clients to submit transactions to replicas over the (local loopback) network.
	Clients []*dummyclient.DummyClient
}

// NewDeployment returns a Deployment initialized according to the passed configuration.
func NewDeployment(conf *TestConfig) (*Deployment, error) {
	if conf == nil {
		return nil, es.Errorf("test config is nil")
	}

	if conf.Logger == nil {
		return nil, es.Errorf("logger is nil")
	}

	// The logger will be used concurrently by all clients and nodes, so it has to be thread-safe.
	conf.Logger = logging.Synchronize(conf.Logger)

	// Create all TestReplicas for this deployment.
	replicas := make([]*TestReplica, len(conf.NodeIDs))
	for i, nodeID := range conf.NodeIDs {

		// Configure the test replica's node.
		config := mir.DefaultNodeConfig().WithLogger(logging.Decorate(conf.Logger, fmt.Sprintf("Node %d: ", i)))
		config.StatsLogInterval = 5 * time.Second

		// Create instance of TestReplica.
		replicas[i] = &TestReplica{
			ID:               nodeID,
			Config:           config,
			NodeIDs:          conf.NodeIDs,
			Membership:       conf.Membership,
			Dir:              filepath.Join(conf.Directory, fmt.Sprintf("node%d", i)),
			NumFakeTXs:       conf.NumFakeTXs,
			Modules:          conf.NodeModules[nodeID],
			FakeTXDestModule: conf.FakeTXDestModule,
		}

		if conf.Simulation != nil {
			replicas[i].Sim = conf.Simulation.Node(nodeID)
			replicas[i].Proc = conf.Simulation.Spawn()
		}
	}

	// Create dummy clients.
	netClients := make([]*dummyclient.DummyClient, 0)
	for i := 1; i <= conf.NumClients; i++ {
		// The loop counter i is used as client ID.
		// We start counting at 1 (and not 0), since client ID 0 is reserved
		// for the "fake" transactions submitted directly by the TestReplicas.

		// Create new DummyClient
		netClients = append(netClients, dummyclient.NewDummyClient(
			tt.NewClientIDFromInt(i),
			crypto.SHA256,
			conf.Logger,
		))
	}

	return &Deployment{
		TestConfig:   conf,
		Simulation:   conf.Simulation,
		TestReplicas: replicas,
		Clients:      netClients,
	}, nil
}

// Run launches the test deployment.
// It starts all test replicas, the dummy client, and the fake message transport subsystem,
// waits until the replicas stop, and returns the final statuses of all the replicas.
func (d *Deployment) Run(ctx context.Context) (nodeErrors []error, heapObjects int64, heapAlloc int64) {

	// Initialize helper variables.
	nodeErrors = make([]error, len(d.TestReplicas))
	var nodeWg sync.WaitGroup
	var clientWg sync.WaitGroup

	ctx2, cancel := context.WithCancel(context.Background())

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	go func() {
		<-ctx.Done()
		runtime.GC()
		runtime.ReadMemStats(&m2)
		heapObjects = int64(m2.HeapObjects - m1.HeapObjects)
		heapAlloc = int64(m2.HeapAlloc - m1.HeapAlloc)
		cancel()
	}()

	// Start the Mir nodes.
	nodeWg.Add(len(d.TestReplicas))
	trAddrs := make(map[t.NodeID]string, len(d.TestReplicas))
	for i, testReplica := range d.TestReplicas {
		i, testReplica := i, testReplica

		trListener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			nodeErrors[i] = err
			return nodeErrors, 0, 0
		}
		trAddrs[testReplica.ID] = fmt.Sprintf("127.0.0.1:%v", trListener.Addr().(*net.TCPAddr).Port)

		// Start the replica in a separate goroutine.
		start := make(chan struct{})
		go func() {
			defer nodeWg.Done()

			<-start
			testReplica.Config.Logger.Log(logging.LevelDebug, "running")
			nodeErrors[i] = testReplica.Run(ctx2, trListener)
			if err := nodeErrors[i]; err != nil {
				testReplica.Config.Logger.Log(logging.LevelError, "exit with error", "err", errstack.ToString(err))
			} else {
				testReplica.Config.Logger.Log(logging.LevelDebug, "exit")
			}
		}()

		if d.Simulation != nil {
			// TODO: Make replica start delay configurable?
			delay := testsim.RandDuration(d.Simulation.Rand, 1, time.Microsecond)
			go func() {
				testReplica.Proc.Delay(delay)
				close(start)
			}()
		} else {
			close(start)
		}
	}

	// Connect the deployment's DummyClients to all replicas and have them submit their transactions in separate goroutines.
	// Each dummy client connects to the replicas, submits the prescribed number of transactions and disconnects.
	clientWg.Add(len(d.Clients))
	for _, client := range d.Clients {
		go func(c *dummyclient.DummyClient) {
			defer clientWg.Done()

			c.Connect(ctx2, trAddrs)
			submitDummyTransactions(ctx2, c, d.TestConfig.NumNetTXs)
			c.Disconnect()
		}(client)
	}

	// Wait for one of:
	// - The garbage collection is done.
	//   Normally this happens before the nodes finish, as they only receive the stop signal when the GC finished.
	// - The nodes finished before the GC.
	//   This happens on error and prevents the system from waiting until normal scheduled shutdown.
	nodesDone := make(chan struct{})
	go func() {
		nodeWg.Wait()
		close(nodesDone)
	}()
	select {
	case <-ctx2.Done():
	case <-nodesDone:

	}

	// Wait for all replicas and clients to terminate
	nodeWg.Wait()
	clientWg.Wait()

	fmt.Printf("All go routines shut down\n")
	return nodeErrors, heapObjects, heapAlloc
}

func (d *Deployment) EventLogFiles() map[t.NodeID]string {
	logFiles := make(map[t.NodeID]string)
	for _, r := range d.TestReplicas {
		logFiles[r.ID] = r.EventLogFile()
	}
	return logFiles
}

// submitDummyTransactions submits n dummy transactions using client.
// It returns when all transactions have been submitted or when ctx is done.
func submitDummyTransactions(ctx context.Context, client *dummyclient.DummyClient, n int) {
	for i := 0; i < n; i++ {
		// For each transaction to be submitted

		select {
		case <-ctx.Done():
			// Return immediately if context finished.
			return
		default:
			// Submit the transaction and check for error.
			if err := client.SubmitTransaction([]byte(fmt.Sprintf("Transaction %d", i))); err != nil {
				panic(err)
			}
		}
	}
}
