/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deploytest

import (
	"context"
	"crypto"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/dummyclient"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (
	// BaseListenPort defines the starting port number on which test replicas will be listening
	// in case the test is being run with the "grpc" or "libp2p" setting for networking.
	// A node with numeric ID id will listen on port (BaseListenPort + id)
	BaseListenPort = 10000

	// RequestListenPort is the port number on which nodes' RequestReceivers listen for incoming requests.
	RequestListenPort = 20000
)

// TestConfig contains the parameters of the deployment to be tested.
type TestConfig struct {
	// Optional information about the test.
	Info string

	// Number of replicas in the tested deployment.
	NumReplicas int

	// Number of clients in the tested deployment.
	NumClients int

	// Type of networking to use.
	// Current possible values: "fake", "grpc", "libp2p"
	Transport string

	// The number of requests each client submits during the execution of the deployment.
	NumFakeRequests int

	// The number of requests sent over the network (by a single DummyClient)
	NumNetRequests int

	// Directory where all the test-related files will be stored.
	// If empty, an OS-default temporary directory will be used.
	Directory string

	// Duration after which the test deployment will be asked to shut down.
	Duration time.Duration

	// A set of replicas that are slow in proposing batches that is likely to trigger view change.
	SlowProposeReplicas map[int]bool

	// Logger to use for producing diagnostic messages.
	Logger logging.Logger
}

// The Deployment represents a list of replicas interconnected by a simulated network transport.
type Deployment struct {
	TestConfig *TestConfig

	// The replicas of the deployment.
	TestReplicas []*TestReplica

	// Dummy clients to submit requests to replicas over the (local loopback) network.
	Clients []*dummyclient.DummyClient
}

// NewDeployment returns a Deployment initialized according to the passed configuration.
func NewDeployment(conf *TestConfig) (*Deployment, error) {
	if conf == nil {
		return nil, fmt.Errorf("test config is nil")
	}

	// Use a common logger for all clients and replicas.
	var logger logging.Logger
	if conf.Logger != nil {
		logger = logging.Synchronize(conf.Logger)
	} else {
		logger = logging.Synchronize(logging.ConsoleDebugLogger)
	}

	// Create a dummy static membership with replica IDs from 0 to len(replicas) - 1
	nodeIDs := make([]t.NodeID, conf.NumReplicas)
	for i := 0; i < len(nodeIDs); i++ {
		nodeIDs[i] = t.NewNodeIDFromInt(i)
	}

	// Create a simulated network transport to route messages between replicas.
	var transportLayer LocalTransportLayer
	switch conf.Transport {
	case "fake":
		transportLayer = NewFakeTransport(nodeIDs)
	case "grpc":
		transportLayer = NewLocalGrpcTransport(nodeIDs, logger)
	case "libp2p":
		transportLayer = NewLocalLibp2pTransport(nodeIDs, logger)
	}

	// Create all TestReplicas for this deployment.
	replicas := make([]*TestReplica, conf.NumReplicas)
	for i := range replicas {
		nodeID := t.NewNodeIDFromInt(i)

		// Configure the test replica's node.
		config := &mir.NodeConfig{
			Logger: logging.Decorate(logger, fmt.Sprintf("Node %d: ", i)),
		}

		// ISS configuration
		issConfig := iss.DefaultConfig(nodeIDs)
		if conf.SlowProposeReplicas[i] {
			// Increase MaxProposeDelay such that it is likely to trigger view change by the batch timeout.
			// Since a sensible value for the segment timeout needs to be stricter than the batch timeout,
			// in the worst case, it will trigger view change by the segment timeout.
			issConfig.MaxProposeDelay = issConfig.PBFTViewChangeBatchTimeout
		}

		transport, err := transportLayer.Link(nodeID)
		if err != nil {
			return nil, err
		}

		// Create instance of TestReplica.
		replicas[i] = &TestReplica{
			ID:              nodeID,
			Config:          config,
			NodeIDs:         nodeIDs,
			Nodes:           transportLayer.Nodes(),
			Dir:             filepath.Join(conf.Directory, fmt.Sprintf("node%d", i)),
			App:             &FakeApp{},
			Transport:       transport,
			NumFakeRequests: conf.NumFakeRequests,
			ISSConfig:       issConfig,
		}
	}

	// Create dummy clients.
	netClients := make([]*dummyclient.DummyClient, 0)
	for i := 1; i <= conf.NumClients; i++ {
		// The loop counter i is used as client ID.
		// We start counting at 1 (and not 0), since client ID 0 is reserved
		// for the "fake" requests submitted directly by the TestReplicas.

		// Create new DummyClient
		netClients = append(netClients, dummyclient.NewDummyClient(
			t.NewClientIDFromInt(i),
			crypto.SHA256,
			logger,
		))
	}

	return &Deployment{
		TestConfig:   conf,
		TestReplicas: replicas,
		Clients:      netClients,
	}, nil
}

// Run launches the test deployment.
// It starts all test replicas, the dummy client, and the fake message transport subsystem,
// waits until the replicas stop, and returns the final statuses of all the replicas.
func (d *Deployment) Run(ctx context.Context) (nodeErrors []error, heapObjects int64, heapAlloc int64) {
	fmt.Println(">>>>>>>>>>> ", d.TestConfig.Info)

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
	for i, testReplica := range d.TestReplicas {

		// Start the replica in a separate goroutine.
		go func(i int, testReplica *TestReplica) {
			defer nodeWg.Done()

			testReplica.Config.Logger.Log(logging.LevelDebug, "running")
			nodeErrors[i] = testReplica.Run(ctx2)
			if err := nodeErrors[i]; err != nil {
				testReplica.Config.Logger.Log(logging.LevelError, "exit with error", "err", err)
			} else {
				testReplica.Config.Logger.Log(logging.LevelDebug, "exit")
			}
		}(i, testReplica)
	}

	// Connect the deployment's DummyClients to all replicas and have them submit their requests in separate goroutines.
	// Each dummy client connects to the replicas, submits the prescribed number of requests and disconnects.
	clientWg.Add(len(d.Clients))
	for _, client := range d.Clients {
		go func(c *dummyclient.DummyClient) {
			defer clientWg.Done()

			c.Connect(ctx2, d.localRequestReceiverAddrs())
			submitDummyRequests(ctx2, c, d.TestConfig.NumNetRequests)
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
	return
}

// localRequestReceiverAddrs computes network addresses and ports for the RequestReceivers at all replicas and returns
// an address map.
// It is assumed that node ID strings must be parseable to decimal numbers.
// Each test replica is on the local machine - 127.0.0.1
func (d *Deployment) localRequestReceiverAddrs() map[t.NodeID]string {

	addrs := make(map[t.NodeID]string, len(d.TestReplicas))
	for i, tr := range d.TestReplicas {
		addrs[tr.ID] = fmt.Sprintf("127.0.0.1:%d", RequestListenPort+i)
	}

	return addrs
}

// submitDummyRequests submits n dummy requests using client.
// It returns when all requests have been submitted or when ctx is done.
func submitDummyRequests(ctx context.Context, client *dummyclient.DummyClient, n int) {
	for i := 0; i < n; i++ {
		// For each request to be submitted

		select {
		case <-ctx.Done():
			// Return immediately if context finished.
			return
		default:
			// Submit the request and check for error.
			if err := client.SubmitRequest([]byte(fmt.Sprintf("Request %d", i))); err != nil {
				panic(err)
			}
		}
	}
}
