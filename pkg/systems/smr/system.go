package smr

import (
	"context"

	"github.com/libp2p/go-libp2p"
	lp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	mircrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	libp2pnet "github.com/filecoin-project/mir/pkg/net/libp2p"
	t "github.com/filecoin-project/mir/pkg/types"
)

// System represents a Mir SMR system.
// It groups and configures the various Mir modules that need to work together to implement state machine replication.
type System struct {
	// modules is the set of Mir modules that make up the system.
	modules modules.Modules

	// transport is the network transport module.
	// We keep an additional reference to it so that we can start and stop it and connect to other nodes
	// (at startup and after reconfiguration).
	transport net.Transport

	// initialMembership is the initial membership of the system specified at creation of the system.
	initialMembership map[t.NodeID]t.NodeAddress
}

// Modules returns the Mir modules that make up the system.
// The return value of Modules is to be used as an argument to mir.NewNode.
func (sys *System) Modules() modules.Modules {
	return sys.modules
}

// Start starts the operation of the modules of the SMR system.
// It starts the network transport and connects to the initial members of the system.
func (sys *System) Start(ctx context.Context) error {
	if err := sys.transport.Start(); err != nil {
		return errors.Wrap(err, "could not start network transport")
	}
	sys.transport.Connect(ctx, sys.initialMembership)
	return nil
}

// Stop stops the operation of the modules of the SMR system.
// Currently, it only stops the network transport, as no other modules need to be stopped.
func (sys *System) Stop() {
	sys.transport.Stop()
}

// New creates a new SMR system.
// It instantiates the various Mir modules that make up the system and configures them to work together.
// The returned system's Start method must be called before the system can be used.
// The returned system's Stop method should be called when the system is no longer needed.
// The returned system's Modules method can be used to obtain the Mir modules to be passed to mir.NewNode.
func New(
	// The ID of this node.
	ownID t.NodeID,

	// The libp2p private key to use for communicating over libp2p.
	ownKey lp2pcrypto.PrivKey,

	// The initial membership of the system, containing the addresses of all participating nodes (including the own).
	initialMembership map[t.NodeID]t.NodeAddress,

	// Implementation of the cryptographic primitives to be used for signing and verifying protocol messages.
	crypto mircrypto.Crypto,

	// The replicated application logic.
	// This is what the user of the SMR system is expected to implement.
	// If the system needs to support reconfiguration,
	// the user is expected to implement the AppLogic interface directly.
	// For a static application, the user can implement the StaticAppLogic interface instead and transform it into to AppLogic
	// using AppLogicFromStatic.
	app AppLogic,

	// The logger to which the system will pass all its log messages.
	logger logging.Logger,
) (*System, error) {

	// Initialize the libp2p transport subsystem.
	var transport *libp2pnet.Transport
	libp2pPeerID, err := peer.AddrInfoFromP2pAddr(initialMembership[ownID])
	if err != nil {
		return nil, errors.Wrap(err, "failed to get own libp2p addr info")
	}
	h, err := libp2p.New(libp2p.Identity(ownKey), libp2p.DefaultTransports, libp2p.ListenAddrs(libp2pPeerID.Addrs[0]))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create libp2p host")
	}
	transport, err = libp2pnet.NewTransport(h, ownID, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create libp2p transport")
	}

	// Obtain initial app snapshot.
	initialSnapshot, err := app.Snapshot()
	if err != nil {
		return nil, errors.Wrap(err, "could not obtain initial app snapshot")
	}

	// Instantiate the ISS ordering protocol with default configuration.
	// We use the ISS' default module configuration (the expected IDs of modules it interacts with)
	// also to configure other modules of the system.
	issModuleConfig := iss.DefaultModuleConfig()
	issParams := iss.DefaultParams(initialMembership)
	issProtocol, err := iss.New(
		ownID,
		issModuleConfig, issParams, iss.InitialStateSnapshot(initialSnapshot, issParams),
		logging.Decorate(logger, "ISS: "),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating ISS protocol module")
	}

	// Use a simple mempool for incoming requests.
	mempool := simplemempool.NewModule(
		&simplemempool.ModuleConfig{
			Self:   "mempool",
			Hasher: issModuleConfig.Hasher,
		},
		&simplemempool.ModuleParams{
			MaxTransactionsInBatch: 10,
		},
	)

	// Use fake batch database that only stores batches in memory and does not persist them to disk.
	batchdb := fakebatchdb.NewModule(
		&fakebatchdb.ModuleConfig{
			Self: "batchdb",
		},
	)

	// Instantiate the availability layer.
	availability := multisigcollector.NewReconfigurableModule(
		&multisigcollector.ModuleConfig{
			Self:    issModuleConfig.Availability,
			Net:     issModuleConfig.Net,
			Crypto:  issModuleConfig.Crypto,
			Mempool: "mempool",
			BatchDB: "batchdb",
		},
		ownID,
		logger,
	)

	// Instantiate the batch fetcher module that transforms availability certificates ordered by ISS
	// into batches of transactions that can be applied to the replicated application.
	batchFetcher := batchfetcher.NewModule(batchfetcher.DefaultModuleConfig())

	// Let the ISS implementation complete the module set by adding default implementations of helper modules
	// that it needs but that have not been specified explicitly.
	modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
		issModuleConfig.App:          batchFetcher,
		issModuleConfig.Crypto:       mircrypto.New(crypto),
		issModuleConfig.Self:         issProtocol,
		issModuleConfig.Net:          transport,
		issModuleConfig.Availability: availability,
		"batchdb":                    batchdb,
		"mempool":                    mempool,
		"app":                        NewAppModule(app, transport, issModuleConfig.Self),
	}, issModuleConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing the Mir modules")
	}

	return &System{
		modules:           modulesWithDefaults,
		transport:         transport,
		initialMembership: initialMembership,
	}, nil
}
