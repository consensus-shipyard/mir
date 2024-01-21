package testing

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/filecoin-project/mir/pkg/eventmangler"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"

	es "github.com/go-errors/errors"
	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/libp2p"
	"github.com/filecoin-project/mir/pkg/testsim"
	"github.com/filecoin-project/mir/pkg/trantor"
	"github.com/filecoin-project/mir/pkg/trantor/appmodule"
)

const (
	failedTestDir    = "failed-test-data"
	simTransportName = "sim"
)

func TestIntegration(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/filecoin-project/mir/pkg/deploytest.newSimModule"),
		goleak.IgnoreTopFunction("github.com/filecoin-project/mir/pkg/testsim.(*Chan).recv"),

		// Problems with this started occurring after an update to a new version of the quic implementation.
		// Assuming it has nothing to do with Mir or Trantor.
		goleak.IgnoreTopFunction("github.com/libp2p/go-libp2p/p2p/transport/quicreuse.(*reuse).gc"),

		// If an observable is not exhausted when checking an event trace...
		goleak.IgnoreTopFunction("github.com/reactivex/rxgo/v2.Item.SendContext"),
	)
	t.Run("ISS", testIntegrationWithISS)
}

func BenchmarkIntegration(b *testing.B) {
	b.Run("ISS", benchmarkIntegrationWithISS)
}

type TestConfig struct {
	Info                string
	RandomSeed          int64
	NodeIDsWeight       map[stdtypes.NodeID]types.VoteWeight
	NumClients          int
	Transport           string
	NumFakeTXs          int
	NumNetTXs           int
	Duration            time.Duration
	ErrorExpected       *es.Error // whether to expect an error from the test
	Directory           string
	SlowProposeReplicas map[int]bool
	CrashedReplicas     map[int]bool
	CheckFunc           func(tb testing.TB, deployment *deploytest.Deployment, conf *TestConfig)
	Logger              logging.Logger
	Skip                bool // If set to true, test is skipped.
}

func testIntegrationWithISS(tt *testing.T) {
	tests := []struct {
		Desc   string // test description
		Config *TestConfig
	}{
		0: {"Do nothing with 1 node",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				Transport:     "fake",
				Duration:      4 * time.Second,
			}},
		1: {"Do nothing with 4 nodes, one of them slow",
			&TestConfig{
				NodeIDsWeight:       deploytest.NewNodeIDsDefaultWeights(4),
				Transport:           "fake",
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		2: {"Submit 10 fake transactions with 1 node",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				Transport:     "fake",
				NumFakeTXs:    10,
				Directory:     "mirbft-deployment-test",
				Duration:      4 * time.Second,
			}},
		3: {"Submit 100 fake transactions with 4 nodes, one of them slow",
			&TestConfig{
				NodeIDsWeight:       deploytest.NewNodeIDsDefaultWeights(4),
				NumClients:          0,
				Transport:           "fake",
				NumFakeTXs:          100,
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		4: {"Submit 10 fake transactions with 4 nodes and libp2p networking",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(4),
				NumClients:    1,
				Transport:     "libp2p",
				NumFakeTXs:    10,
				Duration:      60 * time.Second,
			}},
		5: {"Submit 10 transactions with 1 node and libp2p networking",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				NumClients:    1,
				Transport:     "libp2p",
				NumNetTXs:     10,
				Duration:      20 * time.Second,
			}},
		6: {"Submit 10 transactions with 4 nodes and libp2p networking",
			&TestConfig{
				Info:          "libp2p 10 transactions and 4 nodes",
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(4),
				NumClients:    1,
				Transport:     "libp2p",
				NumNetTXs:     10,
				Duration:      60 * time.Second,
			}},
		7: {"Do nothing with 1 node in simulation",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				Transport:     simTransportName,
				Duration:      4 * time.Second,
			}},
		8: {"Do nothing with 4 nodes in simulation, one of them slow",
			&TestConfig{
				NodeIDsWeight:       deploytest.NewNodeIDsDefaultWeights(4),
				Transport:           simTransportName,
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		9: {"Submit 10 fake transactions with 1 node in simulation",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				Transport:     simTransportName,
				NumFakeTXs:    10,
				Directory:     "mirbft-deployment-test",
				Duration:      4 * time.Second,
			}},
		10: {"Submit 10 fake transactions with 1 node in simulation, loading WAL",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				NumClients:    1,
				Transport:     simTransportName,
				NumFakeTXs:    10,
				Directory:     "mirbft-deployment-test",
				Duration:      4 * time.Second,
			}},
		11: {"Submit 100 fake transactions with 1 node in simulation",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(1),
				NumClients:    0,
				Transport:     simTransportName,
				NumFakeTXs:    100,
				Duration:      20 * time.Second,
			}},
		12: {"Submit 100 fake transactions with 4 nodes in simulation, one of them slow",
			&TestConfig{
				NodeIDsWeight:       deploytest.NewNodeIDsDefaultWeights(4),
				NumClients:          0,
				Transport:           simTransportName,
				NumFakeTXs:          100,
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		13: {"Submit 100 fake transactions with 4 nodes in simulation, two of them crashed (not messages sent, yes received) and holding the supermajority of stake",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsWeights(4, func(id stdtypes.NodeID) types.VoteWeight {
					numericID, _ := strconv.ParseInt(string(id.Bytes()), 10, 64)
					return types.VoteWeight(fmt.Sprintf("%d0000000000000000000", pow2(int(numericID)))) // ensures last 2 nodes weight is greater than twice the sum of the others'
				}),
				NumClients:      0,
				Transport:       simTransportName,
				NumFakeTXs:      100,
				Duration:        30 * time.Second,
				ErrorExpected:   es.Errorf("no transactions were delivered"),
				CrashedReplicas: map[int]bool{2: true, 3: true},
				CheckFunc: func(tb testing.TB, deployment *deploytest.Deployment, conf *TestConfig) {
					require.Error(tb, conf.ErrorExpected)
					for replica := range conf.NodeIDsWeight {
						app := deployment.TestConfig.FakeApps[replica]
						require.Equal(tb, 0, int(app.TransactionsProcessed))
					}
				},
			}},
		14: {"Submit 100 fake transactions with 4 nodes in simulation, two of them crashed (no messages sent, yes received) but holding the minority of stake",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsWeights(4, func(id stdtypes.NodeID) types.VoteWeight {
					numericID, _ := strconv.ParseInt(string(id.Bytes()), 10, 64)
					return types.VoteWeight(fmt.Sprintf("%d0000000000000000000", pow2(int(4-numericID)))) // ensures first 2 nodes weight is greater than twice the sum of the others'
				}),
				NumClients:      0,
				Transport:       simTransportName,
				NumFakeTXs:      100,
				Duration:        60 * time.Second,
				ErrorExpected:   es.Errorf("no transactions were delivered"),
				CrashedReplicas: map[int]bool{2: true, 3: true},
				CheckFunc: func(tb testing.TB, deployment *deploytest.Deployment, conf *TestConfig) {
					require.Error(tb, conf.ErrorExpected)
					for replica := range conf.NodeIDsWeight {
						app := deployment.TestConfig.FakeApps[replica]
						require.Equal(tb, conf.NumNetTXs+conf.NumFakeTXs, int(app.TransactionsProcessed))
					}
				},
			}},
	}

	for i, test := range tests {
		i, test := i, test

		var lock sync.Mutex
		tt.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {

			lock.Lock()
			defer lock.Unlock()

			if test.Config.Skip {
				t.Logf("Skipping test: %s", t.Name())
				return
			}

			defer func() {
				if err := recover(); err != nil || t.Failed() {
					t.Logf("Test #%03d (%s) failed", i, test.Desc)
					if test.Config.Transport == simTransportName {
						t.Logf("Reproduce with RANDOM_SEED=%d", test.Config.RandomSeed)
					}
					// Save the test data.
					testRelDir, err := filepath.Rel(os.TempDir(), test.Config.Directory)
					require.NoError(t, err)
					retainedDir := filepath.Join(failedTestDir, testRelDir)

					t.Logf("Saving deployment data to: %s\n", retainedDir)
					err = copy.Copy(test.Config.Directory, retainedDir)
					require.NoError(t, err)
				}
			}()

			// Create a directory for the deployment-generated files and set the test directory name.
			// The directory will be automatically removed when the outer test function exits.
			createDeploymentDir(t, test.Config)

			simMode := test.Config.Transport == simTransportName
			if testing.Short() && !simMode {
				t.SkipNow()
			}

			if simMode {
				if v := os.Getenv("RANDOM_SEED"); v != "" {
					var err error
					test.Config.RandomSeed, err = strconv.ParseInt(v, 10, 64)
					require.NoError(t, err)
				} else {
					test.Config.RandomSeed = time.Now().UnixNano()
				}
				t.Logf("Random seed = %d", test.Config.RandomSeed)
			}

			runIntegrationWithISSConfig(t, test.Config)
		})
	}
}

func benchmarkIntegrationWithISS(b *testing.B) {
	benchmarks := []struct {
		Desc   string // test description
		Config *TestConfig
	}{
		0: {"Runs for 10s with 4 nodes",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(4),
				NumClients:    1,
				Transport:     "fake",
				Duration:      10 * time.Second,
				Logger:        logging.ConsoleErrorLogger,
			}},
		1: {"Runs for 100s with 4 nodes",
			&TestConfig{
				NodeIDsWeight: deploytest.NewNodeIDsDefaultWeights(4),
				NumClients:    1,
				Transport:     "fake",
				Duration:      100 * time.Second,
				Logger:        logging.ConsoleErrorLogger,
			}},
	}

	for i, bench := range benchmarks {
		i, bench := i, bench
		b.Run(fmt.Sprintf("%03d", i), func(b *testing.B) {
			b.ReportAllocs()

			var totalHeapObjects, totalHeapAlloc float64
			for i := 0; i < b.N; i++ {
				createDeploymentDir(b, bench.Config)
				heapObjects, heapAlloc := runIntegrationWithISSConfig(b, bench.Config)
				totalHeapObjects += float64(heapObjects)
				totalHeapAlloc += float64(heapAlloc)
			}
			b.ReportMetric(totalHeapObjects/float64(b.N), "heapObjects/op")
			b.ReportMetric(totalHeapAlloc/float64(b.N), "heapAlloc/op")

			if b.Failed() {
				b.Logf("Benchmark #%03d (%s) failed", i, bench.Desc)
			} else {
				b.Logf("Benchmark #%03d (%s) done", i, bench.Desc)
			}
		})
	}
}

func runIntegrationWithISSConfig(tb testing.TB, conf *TestConfig) (heapObjects int64, heapAlloc int64) {
	tb.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create new test deployment.
	deployment, err := newDeployment(conf)
	require.NoError(tb, err)

	defer deployment.TestConfig.TransportLayer.Close()

	// Schedule shutdown of test deployment
	if conf.Duration > 0 {
		go func() {
			if deployment.Simulation != nil {
				deployment.Simulation.RunFor(conf.Duration)
			} else {
				time.Sleep(conf.Duration)
			}
			cancel()
		}()
	}

	// Run deployment until it stops and returns final node errors.
	var nodeErrors []error
	nodeErrors, heapObjects, heapAlloc = deployment.Run(ctx)

	// Check whether all the test replicas exited correctly.
	assert.Len(tb, nodeErrors, len(conf.NodeIDsWeight))
	for _, err := range nodeErrors {
		if err != nil {
			assert.Equal(tb, mir.ErrStopped, err)
		}
	}

	// Check event logs
	if conf.CheckFunc != nil {
		conf.CheckFunc(tb, deployment, conf)
		return heapObjects, heapAlloc
	}

	if conf.ErrorExpected != nil {
		require.Error(tb, conf.ErrorExpected)
		return heapObjects, heapAlloc
	}

	require.NoError(tb, checkEventTraces(deployment.EventLogFiles(), conf.NumNetTXs+conf.NumFakeTXs))

	// Check if all transactions were delivered.
	for _, replica := range deployment.TestReplicas {
		app := deployment.TestConfig.FakeApps[replica.ID]
		assert.Equal(tb, conf.NumNetTXs+conf.NumFakeTXs, int(app.TransactionsProcessed))
	}

	// If the test failed, keep the generated data.
	if tb.Failed() {

		// Save the test data.
		testRelDir, err := filepath.Rel(os.TempDir(), conf.Directory)
		require.NoError(tb, err)
		retainedDir := filepath.Join(failedTestDir, testRelDir)

		tb.Logf("Test failed. Saving deployment data to: %s\n", retainedDir)
		err = copy.Copy(conf.Directory, retainedDir)
		require.NoError(tb, err)
	}

	return heapObjects, heapAlloc
}

// If conf.Directory is not empty, creates a directory with that path if it does not yet exist.
// If conf.Directory is empty, creates a directory in the OS-default temporary path
// and sets conf.Directory to that path.
func createDeploymentDir(tb testing.TB, conf *TestConfig) {
	tb.Helper()
	if conf == nil {
		return
	}

	if conf.Directory != "" {
		conf.Directory = filepath.Join(os.TempDir(), conf.Directory)
	} else {
		// If no directory is configured, create a temporary directory in the OS-default location.
		conf.Directory = filepath.Join(tb.TempDir(), tb.Name())
	}

	tb.Logf("Using deployment dir: %s\n", conf.Directory)
	err := os.MkdirAll(conf.Directory, 0777)
	require.NoError(tb, err)
	tb.Cleanup(func() { os.RemoveAll(conf.Directory) })
}

func newDeployment(conf *TestConfig) (*deploytest.Deployment, error) {
	nodeIDs := maputil.GetSortedKeys(conf.NodeIDsWeight)
	logger := deploytest.NewLogger(conf.Logger)

	var simulation *deploytest.Simulation
	if conf.Transport == simTransportName {
		r := rand.New(rand.NewSource(conf.RandomSeed)) // nolint: gosec
		eventDelayFn := func(e stdtypes.Event) time.Duration {
			// TODO: Make min and max event processing delay configurable
			return testsim.RandDuration(r, 0, time.Microsecond)
		}
		simulation = deploytest.NewSimulation(r, nodeIDs, eventDelayFn)
	}
	transportLayer, err := deploytest.NewLocalTransportLayer(simulation, conf.Transport, conf.NodeIDsWeight, logger)
	if err != nil {
		return nil, es.Errorf("error creating local transport system: %w", err)
	}
	cryptoSystem, err := deploytest.NewLocalCryptoSystem("pseudo", nodeIDs, logger)
	if err != nil {
		return nil, es.Errorf("could not create a local crypto system: %w", err)
	}

	nodeModules := make(map[stdtypes.NodeID]modules.Modules)
	fakeApps := make(map[stdtypes.NodeID]*deploytest.FakeApp)

	for i, nodeID := range nodeIDs {
		nodeLogger := logging.Decorate(logger, fmt.Sprintf("Node %d: ", i))
		// Dummy application
		fakeApp := deploytest.NewFakeApp()
		initialSnapshot, err := fakeApp.Snapshot()
		if err != nil {
			return nil, err
		}

		// Trantor configuration
		tConf := trantor.DefaultParams(transportLayer.Membership())
		if conf.SlowProposeReplicas[i] {
			// Increase MaxProposeDelay such that it is likely to trigger view change by the SN timeout.
			// Since a sensible value for the segment timeout needs to be stricter than the SN timeout,
			// in the worst case, it will trigger view change by the segment timeout.
			tConf.Iss.MaxProposeDelay = tConf.Iss.PBFTViewChangeSNTimeout
		}

		transport, err := transportLayer.Link(nodeID)
		if err != nil {
			return nil, es.Errorf("error initializing Mir transport: %w", err)
		}
		stateSnapshotpb, err := iss.InitialStateSnapshot(initialSnapshot, tConf.Iss)
		if err != nil {
			return nil, es.Errorf("error initializing Mir state snapshot: %w", err)
		}

		localCrypto, err := cryptoSystem.Crypto(nodeID)
		if err != nil {
			return nil, es.Errorf("error creating local crypto system for node %v: %w", nodeID, err)
		}

		// Use small batches so even a few transactions keep being proposed even after epoch transitions.
		mempoolParams := simplemempool.DefaultModuleParams()
		mempoolParams.MaxTransactionsInBatch = 10
		avParamsTemplate := multisigcollector.DefaultParamsTemplate()
		avParamsTemplate.Limit = 1 // This prevents "batching of batches" by the availability component.

		system, err := trantor.New(
			nodeID,
			transport,
			checkpoint.Genesis(stateSnapshotpb),
			localCrypto,
			appmodule.AppLogicFromStatic(fakeApp, transportLayer.Membership()),
			trantor.Params{
				Mempool:      mempoolParams,
				Iss:          tConf.Iss,
				Net:          libp2p.Params{},
				Availability: avParamsTemplate,
			},
			nodeLogger,
		)
		if err != nil {
			return nil, err
		}

		if conf.CrashedReplicas[i] {
			// Set MaxProposeDelay to 0 such that it is likely to trigger view change by the SN timeout.
			// Since a sensible value for the segment timeout needs to be stricter than the SN timeout,
			// in the worst case, it will trigger view change by the segment timeout.
			err := trantor.PerturbMessages(&eventmangler.ModuleParams{
				DropRate: 1,
			}, trantor.DefaultModuleConfig().Net, system)
			if err != nil {
				return nil, err
			}
		}

		nodeModules[nodeID] = system.Modules()
		fakeApps[nodeID] = fakeApp
	}

	deployConf := &deploytest.TestConfig{
		Info:             conf.Info,
		Simulation:       simulation,
		TransportLayer:   transportLayer,
		NodeIDs:          nodeIDs,
		Membership:       transportLayer.Membership(),
		NodeModules:      nodeModules,
		NumClients:       conf.NumClients,
		NumFakeTXs:       conf.NumFakeTXs,
		NumNetTXs:        conf.NumNetTXs,
		FakeTXDestModule: stdtypes.ModuleID("mempool"),
		Directory:        conf.Directory,
		Logger:           logger,
		FakeApps:         fakeApps,
	}

	return deploytest.NewDeployment(deployConf)
}

func pow2(exp int) uint64 {
	y := uint64(1)

	for i := 0; i < exp; i++ {
		y *= 2
	}

	return y
}
