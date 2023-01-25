package deploytest

import (
	"context"
	"fmt"
	gonet "net"
	"path/filepath"
	"sync"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	mempoolpbevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/testsim"
	"github.com/filecoin-project/mir/pkg/transactionreceiver"
	tt "github.com/filecoin-project/mir/pkg/trantor/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// TestReplica represents one replica (that uses one instance of the mir.Node) in the test system.
type TestReplica struct {

	// ID of the replica as seen by the protocol.
	ID t.NodeID

	// The modules that the node will run.
	Modules modules.Modules

	// Name of the directory where the persisted state of this TestReplica will be stored,
	// along with the logs produced by running the replica.
	Dir string

	// Configuration of the node corresponding to this replica.
	Config *mir.NodeConfig

	// List of replica IDs constituting the (static) membership.
	NodeIDs []t.NodeID

	// List of replicas.
	Membership *trantorpbtypes.Membership

	// Node's representation within the simulation runtime
	Sim *SimNode

	// Node's process within the simulation runtime
	Proc *testsim.Process

	// Number of simulated transactions inserted in the test replica by a hypothetical client.
	NumFakeTXs int

	// ID of the module to which fake transactions should be sent.
	FakeTXDestModule t.ModuleID
}

// EventLogFile returns the name of the file where the replica's event log is stored.
func (tr *TestReplica) EventLogFile() string {
	return filepath.Join(tr.Dir, "eventlog0.gz")
}

// Run initializes all the required modules and starts the test replica.
// The function blocks until the replica stops.
// The replica stops when stopC is closed.
// Run returns the error returned by the run of the underlying Mir node.
func (tr *TestReplica) Run(ctx context.Context, txReceiverListener gonet.Listener) error {

	// Initialize recording of events.
	interceptor, err := eventlog.NewRecorder(
		tr.ID,
		tr.Dir,
		logging.Decorate(tr.Config.Logger, "Interceptor: "),
		//eventlog.SyncWriteOpt(),
	)

	if err != nil {
		return es.Errorf("error creating interceptor: %w", err)
	}
	defer func() {
		if err := interceptor.Stop(); err != nil {
			panic(err)
		}
	}()

	mod := tr.Modules
	if tr.Sim != nil {
		mod["timer"] = NewSimTimerModule(tr.Sim)
		mod = tr.Sim.WrapModules(mod)
	}

	// Create the mir node for this replica.
	node, err := mir.NewNode(
		tr.ID,
		tr.Config,
		mod,
		interceptor,
	)
	if err != nil {
		return es.Errorf("error creating Mir node: %w", err)
	}

	// Create a Transactionreceiver for transactions coming over the network.
	txReceiver := transactionreceiver.NewTransactionReceiver(node, tr.FakeTXDestModule, logging.Decorate(tr.Config.Logger, "TxRec: "))
	txReceiver.Start(txReceiverListener)

	// Initialize WaitGroup for the replica's transaction submission thread.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start thread submitting transactions from a (single) hypothetical client.
	// The client submits a predefined number of transactions and then stops.
	go func() {
		if tr.Proc != nil {
			tr.Sim.Start(tr.Proc)
		}
		tr.submitFakeTransactions(ctx, node, tr.FakeTXDestModule, &wg)
	}()

	// TODO: avoid hacky special cases like this.
	if transport, ok := tr.Modules["net"]; ok {
		transport, ok := transport.(net.Transport)
		if !ok {
			transport = tr.Modules["truenet"].(net.Transport)
		}

		err = transport.Start()
		if err != nil {
			return es.Errorf("error starting the network link: %w", err)
		}
		transport.Connect(tr.Membership)
		defer transport.Stop()
	}

	// Run the node until it stops.
	exitErr := node.Run(ctx)
	tr.Config.Logger.Log(logging.LevelDebug, "Node run returned!")

	// Stop the transaction receiver.
	txReceiver.Stop()
	if err := txReceiver.ServerError(); err != nil {
		return es.Errorf("transaction receiver returned server error: %w", err)
	}

	// Wait for the local transaction submission thread.
	wg.Wait()
	tr.Config.Logger.Log(logging.LevelInfo, "Fake transaction submission done.")

	return exitErr
}

// Submits n fake transactions to node.
// Aborts when stopC is closed.
// Decrements wg when done.
func (tr *TestReplica) submitFakeTransactions(ctx context.Context, node *mir.Node, destModule t.ModuleID, wg *sync.WaitGroup) {
	defer wg.Done()

	if tr.Proc != nil {
		defer tr.Proc.Exit()
	}

	// The ID of the fake client is always 0.
	for i := 0; i < tr.NumFakeTXs; i++ {
		select {
		case <-ctx.Done():
			// Stop submitting if shutting down.
			break
		default:
			// Otherwise, submit next transaction.
			eventList := events.ListOf(mempoolpbevents.NewTransactions(
				destModule,
				[]*trantorpbtypes.Transaction{{
					ClientId: tt.NewClientIDFromInt(0),
					TxNo:     tt.TxNo(i),
					Data:     []byte(fmt.Sprintf("Transaction %d", i)),
				}},
			).Pb())

			if err := node.InjectEvents(ctx, eventList); err != nil {
				tr.Config.Logger.Log(logging.LevelError, "failed to inject events", "err", err)
				break
			}

			if tr.Proc != nil {
				tr.Sim.SendEvents(tr.Proc, eventList)
			}

			// TODO: Add some configurable delay here
		}
	}
}
