package deploytest

import (
	"bytes"
	"context"
	"crypto"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/clients"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/grpctransport"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/pb/statuspb"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	"github.com/filecoin-project/mir/pkg/serializing"
	"github.com/filecoin-project/mir/pkg/simplewal"
	t "github.com/filecoin-project/mir/pkg/types"
)

var (
	FakeClientHasher modules.Hasher = crypto.SHA256
)

// TestReplica represents one replica (that uses one instance of the mir.Node) in the test system.
type TestReplica struct {

	// ID of the replica as seen by the protocol.
	Id t.NodeID

	// Dummy test application the replica is running.
	App *FakeApp

	// Request store
	ReqStore modules.RequestStore

	// Name of the directory where the persisted state of this TestReplica will be stored,
	// along with the logs produced by running the replica.
	Dir string

	// Configuration of the node corresponding to this replica.
	Config *mir.NodeConfig

	// List of replica IDs constituting the (static) membership.
	Membership []t.NodeID

	// List of IDs of all clients the replica will accept requests from.
	ClientIDs []t.ClientID

	// Network transport subsystem.
	Net modules.Net

	// Number of simulated requests inserted in the test replica by a hypothetical client.
	NumFakeRequests int

	// Configuration of the ISS protocol, if used. If set to nil, the default ISS configuration is assumed.
	ISSConfig *iss.Config
}

// EventLogFile returns the name of the file where the replica's event log is stored.
func (tr *TestReplica) EventLogFile() string {
	return filepath.Join(tr.Dir, "eventlog.gz")
}

// Run initializes all the required modules and starts the test replica.
// The function blocks until the replica stops.
// The replica stops when stopC is closed.
// Run returns, in this order
//   - The final status of the replica
//   - The error that made the node terminate
//   - The error that occurred while obtaining the final node status
func (tr *TestReplica) Run(ctx context.Context) NodeStatus {

	// // Initialize the request store.
	// reqStorePath := filepath.Join(tr.TmpDir, "reqstore")
	// err := os.MkdirAll(reqStorePath, 0700)
	// Expect(err).NotTo(HaveOccurred())
	// reqStore, err := reqstore.Open(reqStorePath)
	// Expect(err).NotTo(HaveOccurred())
	// defer reqStore.Close()

	// Initialize the write-ahead log.
	walPath := filepath.Join(tr.Dir, "wal")
	err := os.MkdirAll(walPath, 0700)
	Expect(err).NotTo(HaveOccurred())
	wal, err := simplewal.Open(walPath)
	Expect(err).NotTo(HaveOccurred())
	defer wal.Close()

	// Initialize recording of events.
	file, err := os.Create(tr.EventLogFile())
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()
	interceptor := eventlog.NewRecorder(tr.Id, file, logging.Decorate(tr.Config.Logger, "Interceptor: "))
	defer func() {
		err := interceptor.Stop()
		Expect(err).NotTo(HaveOccurred())
	}()

	// If no ISS Protocol configuration has been specified, use the default one.
	if tr.ISSConfig == nil {
		tr.ISSConfig = iss.DefaultConfig(tr.Membership)
	}

	issProtocol, err := iss.New(tr.Id, tr.ISSConfig, logging.Decorate(tr.Config.Logger, "ISS: "))
	Expect(err).NotTo(HaveOccurred())

	cryptoModule, err := mirCrypto.NodePseudo(tr.Membership, tr.ClientIDs, tr.Id, mirCrypto.DefaultPseudoSeed)
	Expect(err).NotTo(HaveOccurred())

	// Create the mir node for this replica.
	node, err := mir.NewNode(
		tr.Id,
		tr.Config,
		&modules.Modules{
			Net:           tr.Net,
			App:           tr.App,
			RequestStore:  tr.ReqStore,
			WAL:           wal,
			ClientTracker: clients.SigningTracker(logging.Decorate(tr.Config.Logger, "CT: ")),
			// Protocol:    ordering.NewDummyProtocol(tr.Config.Logger, tr.Membership, tr.Id),
			Protocol:    issProtocol,
			Interceptor: interceptor,
			// // Use dummy crypto module that only produces signatures
			// // consisting of a single zero byte and treats those signatures as valid.
			// Crypto: &mirCrypto.DummyCrypto{DummySig: []byte{0}},
			Crypto: cryptoModule,
		},
	)
	Expect(err).NotTo(HaveOccurred())

	// Create a RequestReceiver for request coming over the network.
	requestReceiver := requestreceiver.NewRequestReceiver(node, logging.Decorate(tr.Config.Logger, "ReqRec: "))
	p, err := strconv.Atoi(tr.Id.Pb())
	if err != nil {
		panic(fmt.Errorf("could not convert node ID %s: %w", tr.Id, err))
	}
	err = requestReceiver.Start(RequestListenPort + p)
	Expect(err).NotTo(HaveOccurred())

	// Initialize WaitGroup for the replica's request submission thread.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start thread submitting requests from a (single) hypothetical client.
	// The client submits a predefined number of requests and then stops.
	go tr.submitFakeRequests(ctx, node, &wg)

	// ATTENTION! This is hacky!
	// If the test replica used the GRPC transport, initialize the Net module.
	switch transport := tr.Net.(type) {
	case *grpctransport.GrpcTransport:
		err := transport.Start()
		Expect(err).NotTo(HaveOccurred())
		transport.Connect(ctx)
	}

	// Run the node until it stops and obtain the node's final status.
	exitErr := node.Run(ctx)
	tr.Config.Logger.Log(logging.LevelDebug, "Node run returned!")

	finalStatus, statusErr := node.Status(context.Background())

	// Stop the request receiver.
	requestReceiver.Stop()
	Expect(requestReceiver.ServerError()).NotTo(HaveOccurred())

	// Wait for the local request submission thread.
	wg.Wait()
	tr.Config.Logger.Log(logging.LevelInfo, "Fake request submission done.")

	// ATTENTION! This is hacky!
	// If the test replica used the GRPC transport, stop the Net module.
	switch transport := tr.Net.(type) {
	case *grpctransport.GrpcTransport:
		tr.Config.Logger.Log(logging.LevelDebug, "Stopping gRPC transport.")
		transport.Stop()
		tr.Config.Logger.Log(logging.LevelDebug, "gRPC transport stopped.")
	}

	// Return the final node status.
	return NodeStatus{
		Status:    finalStatus,
		StatusErr: statusErr,
		ExitErr:   exitErr,
	}
}

// NodeStatus represents the final status of a test replica.
type NodeStatus struct {

	// Status as returned by mir.Node.Status()
	Status *statuspb.NodeStatus

	// Potential error returned by mir.Node.Status() in case of obtaining of the status failed.
	StatusErr error

	// Reason the node terminated, as returned by mir.Node.Run()
	ExitErr error
}

// Submits n fake requests to node.
// Aborts when stopC is closed.
// Decrements wg when done.
func (tr *TestReplica) submitFakeRequests(ctx context.Context, node *mir.Node, wg *sync.WaitGroup) {
	defer GinkgoRecover()
	defer wg.Done()

	// Instantiate a Crypto module for signing the requests.
	// The ID of the fake client is always 0.
	cryptoModule, err := mirCrypto.ClientPseudo(tr.Membership, tr.ClientIDs, t.NewClientIDFromInt(0), mirCrypto.DefaultPseudoSeed)
	Expect(err).NotTo(HaveOccurred())

	for i := 0; i < tr.NumFakeRequests; i++ {
		select {
		case <-ctx.Done():
			// Stop submitting if shutting down.
			break
		default:
			// Otherwise, submit next request.

			// Create new request message. This is only necessary for proper signing
			// and the message will be "taken apart" just a few lines later, when submitting it to the Node.
			reqMsg := &requestpb.Request{
				ClientId: string(t.NewClientIDFromInt(0)),
				ReqNo:    t.ReqNo(i).Pb(),
				Data:     []byte(fmt.Sprintf("Request %d", i)),
			}

			// Sign (the hash of) the request, adding the signature to the request message.
			h := FakeClientHasher.New()
			h.Write(bytes.Join(serializing.RequestForHash(reqMsg), nil))
			reqMsg.Authenticator, err = cryptoModule.Sign([][]byte{h.Sum(nil)})
			Expect(err).NotTo(HaveOccurred())

			// Submit the signed request to the local Node.
			if err := node.SubmitRequest(
				context.Background(),
				t.ClientID(reqMsg.ClientId),
				t.ReqNo(reqMsg.ReqNo),
				reqMsg.Data,
				reqMsg.Authenticator,
				// []byte{0}, // Fake signature. Relies on the Nodes using DummyCrypto{DummySig: []byte{0}}
			); err != nil {

				// TODO (Jason), failing on err causes flakes in the teardown,
				// so just returning for now, we should address later
				break
			}
			// TODO: Add some configurable delay here
		}
	}
}
