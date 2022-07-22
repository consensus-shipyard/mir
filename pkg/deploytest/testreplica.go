package deploytest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/filecoin-project/mir/pkg/net"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/requestpb"
	"github.com/filecoin-project/mir/pkg/requestreceiver"
	"github.com/filecoin-project/mir/pkg/simplewal"
	"github.com/filecoin-project/mir/pkg/testsim"
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
	Membership []t.NodeID

	// Node's representation within the simulation runtime
	Sim *SimNode

	// Node's process within the simulation runtime
	Proc *testsim.Process

	// Number of simulated requests inserted in the test replica by a hypothetical client.
	NumFakeRequests int

	// ID of the module to which fake requests should be sent.
	FakeRequestsDestModule t.ModuleID
}

// EventLogFile returns the name of the file where the replica's event log is stored.
func (tr *TestReplica) EventLogFile() string {
	return filepath.Join(tr.Dir, "eventlog.gz")
}

// Run initializes all the required modules and starts the test replica.
// The function blocks until the replica stops.
// The replica stops when stopC is closed.
// Run returns the error returned by the run of the underlying Mir node.
func (tr *TestReplica) Run(ctx context.Context) error {

	// Initialize the write-ahead log.
	walPath := filepath.Join(tr.Dir, "wal")
	if err := os.MkdirAll(walPath, 0700); err != nil {
		return fmt.Errorf("error creating WAL directory: %w", err)
	}
	wal, err := simplewal.Open(walPath)
	if err != nil {
		return fmt.Errorf("error opening WAL: %w", err)
	}
	defer wal.Close()

	// Initialize recording of events.
	file, err := os.Create(tr.EventLogFile())
	if err != nil {
		return fmt.Errorf("error creating event log file: %w", err)
	}
	defer file.Close()
	interceptor := eventlog.NewRecorder(tr.ID, file, logging.Decorate(tr.Config.Logger, "Interceptor: "))
	defer func() {
		if err := interceptor.Stop(); err != nil {
			panic(err)
		}
	}()

	// TODO: avoid hacky special cases like this.
	if tr.Modules["wal"] != nil {
		if _, isNull := tr.Modules["wal"].(modules.NullPassive); !isNull {
			return fmt.Errorf("\"wal\" module is already present in replica configuration")
		}
	}
	tr.Modules["wal"] = wal

	modules := tr.Modules
	if tr.Sim != nil {
		modules["timer"] = NewSimTimerModule(tr.Sim)
		modules = tr.Sim.WrapModules(modules)
	}

	// Create the mir node for this replica.
	node, err := mir.NewNode(
		tr.ID,
		tr.Config,
		modules,
		wal,
		interceptor,
	)
	if err != nil {
		return fmt.Errorf("error creating Mir node: %w", err)
	}

	// Create a RequestReceiver for request coming over the network.
	requestReceiver := requestreceiver.NewRequestReceiver(node, "iss", logging.Decorate(tr.Config.Logger, "ReqRec: "))

	// TODO: do not assume that node IDs are integers.
	p, err := strconv.Atoi(tr.ID.Pb())
	if err != nil {
		return fmt.Errorf("error converting node ID %s: %w", tr.ID, err)
	}
	err = requestReceiver.Start(RequestListenPort + p)
	if err != nil {
		return fmt.Errorf("error starting request receiver: %w", err)
	}

	// Initialize WaitGroup for the replica's request submission thread.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start thread submitting requests from a (single) hypothetical client.
	// The client submits a predefined number of requests and then stops.
	go func() {
		if tr.Proc != nil {
			walEvents, err := wal.LoadAll(ctx)
			if err != nil {
				panic(fmt.Errorf("error loading WAL events %w", err))
			}
			tr.Sim.Start(tr.Proc, walEvents)
		}
		tr.submitFakeRequests(ctx, node, tr.FakeRequestsDestModule, &wg)
	}()

	// TODO: avoid hacky special cases like this.
	if transport, ok := tr.Modules["net"]; ok {
		transport := transport.(net.Transport)

		err = transport.Start()
		if err != nil {
			return fmt.Errorf("error starting the network link: %w", err)
		}
		transport.Connect(ctx)
		defer transport.Stop()
	}

	// Run the node until it stops.
	exitErr := node.Run(ctx)
	tr.Config.Logger.Log(logging.LevelDebug, "Node run returned!")

	// Stop the request receiver.
	requestReceiver.Stop()
	if err := requestReceiver.ServerError(); err != nil {
		return fmt.Errorf("request receiver returned server error: %w", err)
	}

	// Wait for the local request submission thread.
	wg.Wait()
	tr.Config.Logger.Log(logging.LevelInfo, "Fake request submission done.")

	return exitErr
}

// Submits n fake requests to node.
// Aborts when stopC is closed.
// Decrements wg when done.
func (tr *TestReplica) submitFakeRequests(ctx context.Context, node *mir.Node, destModule t.ModuleID, wg *sync.WaitGroup) {
	defer wg.Done()

	if tr.Proc != nil {
		defer tr.Proc.Exit()
	}

	// The ID of the fake client is always 0.
	for i := 0; i < tr.NumFakeRequests; i++ {
		select {
		case <-ctx.Done():
			// Stop submitting if shutting down.
			break
		default:
			// Otherwise, submit next request.
			eventList := events.ListOf(events.NewClientRequests(
				destModule,
				[]*requestpb.Request{events.ClientRequest(
					t.NewClientIDFromInt(0),
					t.ReqNo(i),
					[]byte(fmt.Sprintf("Request %d", i)),
				)},
			))

			if err := node.InjectEvents(ctx, eventList); err != nil {

				// TODO (Jason), failing on err causes flakes in the teardown,
				// so just returning for now, we should address later
				break
			}

			if tr.Proc != nil {
				tr.Sim.SendEvents(tr.Proc, eventList)
			}

			// TODO: Add some configurable delay here
		}
	}
}
