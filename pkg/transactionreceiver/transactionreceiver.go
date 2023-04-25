/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package transactionreceiver

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/events"
	"github.com/filecoin-project/mir/pkg/logging"
	mempoolpbevents "github.com/filecoin-project/mir/pkg/pb/mempoolpb/events"
	"github.com/filecoin-project/mir/pkg/pb/trantorpb"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

type TransactionReceiver struct {
	UnimplementedTransactionReceiverServer

	// The Node to which to submit the received transactions.
	node *mir.Node

	// The ID of the module to which to submit the received transactions.
	moduleID t.ModuleID

	// The gRPC server used by this networking module.
	grpcServer *grpc.Server

	// Wait group that is notified when the grpcServer stops.
	// Waiting on this WaitGroup makes sure that the server exit status has been recorded correctly.
	grpcServerWg sync.WaitGroup

	// Error returned from the grpcServer.Serve() call (see Start() method).
	grpcServerError error

	// Logger use for all logging events of this TransactionReceiver
	logger logging.Logger
}

// NewTransactionReceiver returns a new initialized transaction receiver.
// The returned TransactionReceiver is not yet running (able to receive transactions).
// This needs to be done explicitly by calling the Start() method.
// For the transactions to be processed by passed Node, the Node must also be running.
func NewTransactionReceiver(node *mir.Node, moduleID t.ModuleID, logger logging.Logger) *TransactionReceiver {
	// If no logger was given, only write errors to the console.
	if logger == nil {
		logger = logging.ConsoleErrorLogger
	}

	return &TransactionReceiver{
		node:     node,
		moduleID: moduleID,
		logger:   logger,
	}
}

// Listen implements the gRPC Listen service (multi-request-single-response).
// It receives messages from the gRPC client running on the Mir client
// and submits them to the Node associated with this TransactionReceiver.
// This function is called by the gRPC system on every new connection
// from a Mir client's gRPC client.
func (rr *TransactionReceiver) Listen(srv TransactionReceiver_ListenServer) error {

	// Print address of incoming connection.
	p, ok := peer.FromContext(srv.Context())
	if ok {
		rr.logger.Log(logging.LevelDebug, fmt.Sprintf("Incoming connection from %s", p.Addr.String()))
	} else {
		return fmt.Errorf("failed to get grpc peer info from context")
	}

	// Declare loop variables outside, since err is checked also after the loop finishes.
	var err error
	var req *trantorpb.Transaction

	// For each received transaction
	for req, err = srv.Recv(); err == nil; req, err = srv.Recv() {

		rr.logger.Log(logging.LevelInfo, "Received transaction", "clId", req.ClientId, "reqNo", req.TxNo)

		// Submit the transaction to the Node.
		if srErr := rr.node.InjectEvents(srv.Context(), events.ListOf(mempoolpbevents.NewRequests(
			rr.moduleID,
			[]*trantorpbtypes.Transaction{trantorpbtypes.TransactionFromPb(req)},
		).Pb())); srErr != nil {

			// If submitting fails, stop receiving further transaction (and close connection).
			rr.logger.Log(logging.LevelError, fmt.Sprintf("Could not submit transaction (%v-%d): %v. Closing connection.",
				req.ClientId, req.TxNo, srErr))
			break
		}
	}

	// If the connection was terminated by the gRPC client, print the reason.
	// (This line could also be reached by breaking out of the above loop on transaction submission error.)
	if err != nil {
		rr.logger.Log(logging.LevelWarn, fmt.Sprintf("Connection terminated: %s (%v)", p.Addr.String(), err))
	}

	// Send gRPC response message and close connection.
	return srv.SendAndClose(&ByeBye{})
}

// Start starts the TransactionReceiver by initializing and starting the internal gRPC server,
// listening on the passed port.
// Before ths method is called, no client connections are accepted.
func (rr *TransactionReceiver) Start(port int) error {

	rr.logger.Log(logging.LevelInfo, fmt.Sprintf("Listening for transaction connections on port %d", port))

	// Create a gRPC server and assign it the logic of this TransactionReceiver.
	rr.grpcServer = grpc.NewServer()
	RegisterTransactionReceiverServer(rr.grpcServer, rr)

	// Start listening on the network
	conn, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return fmt.Errorf("failed to listen for connections on port %d: %w", port, err)
	}

	// Start the gRPC server in a separate goroutine.
	// When the server stops, it will write its exit error into gt.grpcServerError.
	rr.grpcServerWg.Add(1)
	go func() {
		rr.grpcServerError = rr.grpcServer.Serve(conn)
		rr.grpcServerWg.Done()
	}()

	// If we got all the way here, no error occurred.
	return nil
}

// Stop stops the own gRPC server (preventing further incoming connections).
// After Stop() returns, the error returned by the gRPC server's Serve() call
// can be obtained through the ServerError() method.
func (rr *TransactionReceiver) Stop() {

	rr.logger.Log(logging.LevelDebug, "Stopping transaction receiver.")

	// Stop own gRPC server and wait for its exit status to be recorded.
	rr.grpcServer.Stop()
	rr.grpcServerWg.Wait()

	rr.logger.Log(logging.LevelDebug, "Transaction receiver stopped.")
}

// ServerError returns the error returned by the gRPC server's Serve() call.
// ServerError() must not be called before the TransactionReceiver is stopped and its Stop() method has returned.
func (rr *TransactionReceiver) ServerError() error {
	return rr.grpcServerError
}
