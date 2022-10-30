package smr

import (
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/net"

	"github.com/filecoin-project/mir/pkg/checkpoint"

	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	mircrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
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

	// initialMemberships is a slice of initial memberships of the system specified at creation of the system.
	// They correspond to the starting epoch of the system and configOffset subsequent epochs.
	initialMemberships []map[t.NodeID]t.NodeAddress
}

// Modules returns the Mir modules that make up the system.
// The return value of Modules is to be used as an argument to mir.NewNode.
func (sys *System) Modules() modules.Modules {
	return sys.modules
}

// WithModule associates the given module ID within the SMR system with the given module.
// If a module with the given ID already exists, it is replaced.
// WithModule returns the SMR system itself (not a copy of it), so calls can be chained.
func (sys *System) WithModule(moduleID t.ModuleID, module modules.Module) *System {
	sys.modules[moduleID] = module
	return sys
}

// Start starts the operation of the modules of the SMR system.
// It starts the network transport and connects to the initial members of the system.
func (sys *System) Start() error {
	if err := sys.transport.Start(); err != nil {
		return errors.Wrap(err, "could not start network transport")
	}
	for _, membership := range sys.initialMemberships {
		sys.transport.Connect(membership)
	}
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

	// libp2p host to be used for the network transport module.
	h host.Host,

	// Initial checkpoint of the application state and configuration.
	// The SMR system will continue operating from this checkpoint.
	startingCheckpoint *checkpoint.StableCheckpoint,

	// Implementation of the cryptographic primitives to be used for signing and verifying protocol messages.
	crypto mircrypto.Crypto,

	// The replicated application logic.
	// This is what the user of the SMR system is expected to implement.
	// If the system needs to support reconfiguration,
	// the user is expected to implement the AppLogic interface directly.
	// For a static application, the user can implement the StaticAppLogic interface instead and transform it into to AppLogic
	// using AppLogicFromStatic.
	app AppLogic,

	// Parameters of the SMR system, like batch size or batch timeout.
	params Params,

	// The logger to which the system will pass all its log messages.
	logger logging.Logger,
) (*System, error) {

	// Initialize the libp2p transport subsystem.
	// TODO: Re-enable this check!
	// addrIn := false
	// for _, addr := range h.Addrs() {
	//	// sanity-check to see if the host is configured with the
	//	// right multiaddr.
	//	if addr.Equal(initialMembership[ownID]) {
	//		addrIn = true
	//		break
	//	}
	// }
	// if !addrIn {
	//	return nil, errors.New("libp2p host provided as input not listening to multiaddr specified for node")
	// }
	transport, err := libp2pnet.NewTransport(params.Net, h, ownID, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create libp2p transport")
	}

	// Instantiate the ISS ordering protocol with default configuration.
	// We use the ISS' default module configuration (the expected IDs of modules it interacts with)
	// also to configure other modules of the system.
	issModuleConfig := iss.DefaultModuleConfig()
	issProtocol, err := iss.New(
		ownID,
		issModuleConfig,
		params.Iss,
		startingCheckpoint,
		logging.Decorate(logger, "ISS: "),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating ISS protocol module")
	}

	// Factory module with instances of the checkpointing protocol.
	checkpointing := checkpoint.Factory(checkpoint.DefaultModuleConfig(), ownID, logging.Decorate(logger, "CHKP: "))

	// Use a simple mempool for incoming requests.
	mempool := simplemempool.NewModule(
		&simplemempool.ModuleConfig{
			Self:   "mempool",
			Hasher: issModuleConfig.Hasher,
		},
		params.Mempool,
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
	batchFetcher := batchfetcher.NewModule(
		batchfetcher.DefaultModuleConfig(),
		startingCheckpoint.Epoch(),
		startingCheckpoint.ClientProgress(logger),
	)

	// Let the ISS implementation complete the module set by adding default implementations of helper modules
	// that it needs but that have not been specified explicitly.
	modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
		issModuleConfig.App:          batchFetcher,
		issModuleConfig.Crypto:       mircrypto.New(crypto),
		issModuleConfig.Self:         issProtocol,
		issModuleConfig.Net:          transport,
		issModuleConfig.Availability: availability,
		issModuleConfig.Checkpoint:   checkpointing,
		"batchdb":                    batchdb,
		"mempool":                    mempool,
		"app":                        NewAppModule(app, transport, issModuleConfig.Self),
	}, issModuleConfig)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing the Mir modules")
	}

	return &System{
		modules:            modulesWithDefaults,
		transport:          transport,
		initialMemberships: startingCheckpoint.Memberships(),
	}, nil
}

// GenesisCheckpoint returns an initial stable checkpoint used for bootstrapping.
// It is a special checkpoint for epoch 0, corresponding to the state of the application
// (the serialization of which is passed as the initialAppState parameter) before applying any transactions.
// The associated certificate is empty (and should still be considered valid, as a special case).
func GenesisCheckpoint(initialAppState []byte, params Params) *checkpoint.StableCheckpoint {
	return checkpoint.Genesis(iss.InitialStateSnapshot(initialAppState, params.Iss))
}
