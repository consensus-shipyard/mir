package trantor

import (
	"crypto"

	"github.com/filecoin-project/mir/pkg/timer"
	"github.com/filecoin-project/mir/pkg/trantor/appmodule"

	"github.com/pkg/errors"

	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	mircrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/orderers"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	t "github.com/filecoin-project/mir/pkg/types"
)

// System represents a Trantor system.
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
	initialMemberships []*trantorpbtypes.Membership
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

	// Network transport system to be used by Trantor to send and receive messages.
	transport net.Transport,

	// Initial checkpoint of the application state and configuration.
	// The SMR system will continue operating from this checkpoint.
	startingCheckpoint *checkpoint.StableCheckpoint,

	// Implementation of the cryptographic primitives to be used for signing and verifying protocol messages.
	cryptoImpl mircrypto.Crypto,

	// The replicated application logic.
	// This is what the user of the SMR system is expected to implement.
	// If the system needs to support reconfiguration,
	// the user is expected to implement the AppLogic interface directly.
	// For a static application, the user can implement the StaticAppLogic interface instead and transform it into to AppLogic
	// using AppLogicFromStatic.
	app appmodule.AppLogic,

	// Parameters of the SMR system, like batch size or batch timeout.
	params Params,

	// The logger to which the system will pass all its log messages.
	logger logging.Logger,
) (*System, error) {

	// Hash function to be used by all modules of the system.
	hashImpl := crypto.SHA256

	moduleConfig := DefaultModuleConfig()
	trantorModules := make(map[t.ModuleID]modules.Module)

	// Instantiate the ISS ordering protocol.
	issModuleConfig := iss.ModuleConfig{
		Self:         moduleConfig.ISS,
		App:          moduleConfig.BatchFetcher,
		Availability: moduleConfig.Availability,
		Checkpoint:   moduleConfig.Checkpointing,
		Net:          moduleConfig.Net,
		Ordering:     moduleConfig.Ordering,
		Timer:        moduleConfig.Timer,
	}
	issProtocol, err := iss.New(
		ownID,
		issModuleConfig,
		params.Iss,
		startingCheckpoint,
		hashImpl,
		cryptoImpl,
		logging.Decorate(logger, "ISS: "),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating ISS protocol module")
	}
	trantorModules[moduleConfig.ISS] = issProtocol

	// Factory module with instances of the checkpointing protocol.
	chkpModuleConfig := checkpoint.ModuleConfig{
		Self:   moduleConfig.Checkpointing,
		App:    moduleConfig.BatchFetcher,
		Crypto: moduleConfig.Crypto,
		Hasher: moduleConfig.Hasher,
		Net:    moduleConfig.Net,
		Ord:    moduleConfig.ISS,
	}
	trantorModules[moduleConfig.Checkpointing] = checkpoint.Factory(
		chkpModuleConfig,
		ownID,
		logging.Decorate(logger, "CHKP: "),
	)

	// Ordering module, dynamically creating instances of the PBFT protocol as segments are created by ISS.
	orderingModuleConfig := orderers.ModuleConfig{
		Self:   moduleConfig.Ordering,
		App:    moduleConfig.BatchFetcher,
		Ava:    "", // Ava not initialized yet. It will be set at sub-module instantiation within the factory.
		Crypto: moduleConfig.Crypto,
		Hasher: moduleConfig.Hasher,
		Net:    moduleConfig.Net,
		Ord:    moduleConfig.ISS,
		Timer:  moduleConfig.Timer,
	}
	trantorModules[moduleConfig.Ordering] = orderers.Factory(
		orderingModuleConfig,
		params.Iss,
		ownID,
		hashImpl,
		cryptoImpl,
		logging.Decorate(logger, "PBFT: "),
	)

	// Use a simple mempool for incoming transactions.
	mempoolModuleConfig := simplemempool.ModuleConfig{
		Self:   moduleConfig.Mempool,
		Hasher: moduleConfig.Hasher,
	}
	trantorModules[moduleConfig.Mempool] = simplemempool.NewModule(
		mempoolModuleConfig,
		params.Mempool,
	)

	// Use fake batch database that only stores batches in memory and does not persist them to disk.
	batchDbModuleConfig := fakebatchdb.ModuleConfig{
		Self: moduleConfig.BatchDB,
	}
	trantorModules[moduleConfig.BatchDB] = fakebatchdb.NewModule(batchDbModuleConfig)

	// Instantiate the availability component.
	availModuleConfig := multisigcollector.ModuleConfig{
		Self:    moduleConfig.Availability,
		Mempool: moduleConfig.Mempool,
		BatchDB: moduleConfig.BatchDB,
		Net:     moduleConfig.Net,
		Crypto:  moduleConfig.Crypto,
	}
	trantorModules[moduleConfig.Availability] = multisigcollector.NewReconfigurableModule(availModuleConfig, logger)

	// Instantiate the batch fetcher module that transforms availability certificates ordered by ISS
	// into batches of transactions that can be applied to the replicated application.
	// It acts as a proxy between the application module and the rest of the system.
	bfModuleConfig := batchfetcher.ModuleConfig{
		Self:         moduleConfig.BatchFetcher,
		Availability: moduleConfig.Availability,
		Checkpoint:   moduleConfig.Checkpointing,
		Destination:  moduleConfig.App,
	}
	trantorModules[moduleConfig.BatchFetcher] = batchfetcher.NewModule(
		bfModuleConfig,
		startingCheckpoint.Epoch(),
		startingCheckpoint.ClientProgress(logger),
		logger,
	)

	trantorModules[moduleConfig.Net] = transport
	trantorModules[moduleConfig.App] = appmodule.NewAppModule(app, transport, issModuleConfig.Self)
	trantorModules[moduleConfig.Hasher] = mircrypto.NewHasher(hashImpl)
	trantorModules[moduleConfig.Crypto] = mircrypto.New(cryptoImpl)
	trantorModules[moduleConfig.Timer] = timer.New()
	trantorModules[moduleConfig.Null] = modules.NullPassive{}

	return &System{
		modules:            trantorModules,
		transport:          transport,
		initialMemberships: startingCheckpoint.Memberships(),
	}, nil
}

// GenesisCheckpoint returns an initial stable checkpoint used for bootstrapping.
// It is a special checkpoint for epoch 0, corresponding to the state of the application
// (the serialization of which is passed as the initialAppState parameter) before applying any transactions.
// The associated certificate is empty (and should still be considered valid, as a special case).
func GenesisCheckpoint(initialAppState []byte, params Params) (*checkpoint.StableCheckpoint, error) {
	stateSnapshotpb, err := iss.InitialStateSnapshot(initialAppState, params.Iss)
	if err != nil {
		return nil, err
	}
	return checkpoint.Genesis(stateSnapshotpb), nil
}
