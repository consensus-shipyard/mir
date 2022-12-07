package trantor

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/filecoin-project/mir/pkg/util/issutil"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/availability/batchdb/fakebatchdb"
	"github.com/filecoin-project/mir/pkg/availability/multisigcollector"
	"github.com/filecoin-project/mir/pkg/batchfetcher"
	"github.com/filecoin-project/mir/pkg/checkpoint"
	"github.com/filecoin-project/mir/pkg/clientprogress"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/iss"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/mempool/simplemempool"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/testsim"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (
	failedTestDir = "failed-test-data"
)

func TestIntegration(t *testing.T) {
	defer goleak.VerifyNone(t,
		goleak.IgnoreTopFunction("github.com/filecoin-project/mir/pkg/deploytest.newSimModule"),
		goleak.IgnoreTopFunction("github.com/filecoin-project/mir/pkg/testsim.(*Chan).recv"),

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
	NumReplicas         int
	NumClients          int
	Transport           string
	NumFakeRequests     int
	NumNetRequests      int
	Duration            time.Duration
	Directory           string
	SlowProposeReplicas map[int]bool
	Logger              logging.Logger
}

func testIntegrationWithISS(t *testing.T) {
	tests := []struct {
		Desc   string // test description
		Config *TestConfig
	}{
		0: {"Do nothing with 1 node",
			&TestConfig{
				NumReplicas: 1,
				Transport:   "fake",
				Duration:    4 * time.Second,
			}},
		1: {"Do nothing with 4 nodes, one of them slow",
			&TestConfig{
				NumReplicas:         4,
				Transport:           "fake",
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		2: {"Submit 10 fake requests with 1 node",
			&TestConfig{
				NumReplicas:     1,
				Transport:       "fake",
				NumFakeRequests: 10,
				Directory:       "mirbft-deployment-test",
				Duration:        4 * time.Second,
			}},
		3: {"Submit 10 fake requests with 1 node, loading WAL",
			&TestConfig{
				NumReplicas:     1,
				NumClients:      1,
				Transport:       "fake",
				NumFakeRequests: 10,
				Directory:       "mirbft-deployment-test",
				Duration:        4 * time.Second,
			}},
		4: {"Submit 100 fake requests with 4 nodes, one of them slow",
			&TestConfig{
				NumReplicas:         4,
				NumClients:          0,
				Transport:           "fake",
				NumFakeRequests:     100,
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},

		5: {"Submit 10 fake requests with 4 nodes and gRPC networking",
			&TestConfig{
				NumReplicas:     4,
				NumClients:      1,
				Transport:       "grpc",
				NumFakeRequests: 10,
				Duration:        4 * time.Second,
			}},
		6: {"Submit 10 requests with 1 node and gRPC networking",
			&TestConfig{
				NumReplicas:    1,
				NumClients:     1,
				Transport:      "grpc",
				NumNetRequests: 10,
				Duration:       4 * time.Second,
			}},
		7: {"Submit 10 requests with 4 nodes and gRPC networking",
			&TestConfig{
				Info:           "grpc 10 requests and 4 nodes",
				NumReplicas:    4,
				NumClients:     1,
				Transport:      "grpc",
				NumNetRequests: 10,
				Duration:       4 * time.Second,
			}},
		8: {"Submit 10 fake requests with 4 nodes and libp2p networking",
			&TestConfig{
				NumReplicas:     4,
				NumClients:      1,
				Transport:       "libp2p",
				NumFakeRequests: 10,
				Duration:        10 * time.Second,
			}},
		9: {"Submit 10 requests with 1 node and libp2p networking",
			&TestConfig{
				NumReplicas:    1,
				NumClients:     1,
				Transport:      "libp2p",
				NumNetRequests: 10,
				Duration:       10 * time.Second,
			}},
		10: {"Submit 10 requests with 4 nodes and libp2p networking",
			&TestConfig{
				Info:           "libp2p 10 requests and 4 nodes",
				NumReplicas:    4,
				NumClients:     1,
				Transport:      "libp2p",
				NumNetRequests: 10,
				Duration:       15 * time.Second,
			}},
		11: {"Do nothing with 1 node in simulation",
			&TestConfig{
				NumReplicas: 1,
				Transport:   "sim",
				Duration:    4 * time.Second,
			}},

		12: {"Do nothing with 4 nodes in simulation, one of them slow",
			&TestConfig{
				NumReplicas:         4,
				Transport:           "sim",
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		13: {"Submit 10 fake requests with 1 node in simulation",
			&TestConfig{
				NumReplicas:     1,
				Transport:       "sim",
				NumFakeRequests: 10,
				Directory:       "mirbft-deployment-test",
				Duration:        4 * time.Second,
			}},
		14: {"Submit 10 fake requests with 1 node in simulation, loading WAL",
			&TestConfig{
				NumReplicas:     1,
				NumClients:      1,
				Transport:       "sim",
				NumFakeRequests: 10,
				Directory:       "mirbft-deployment-test",
				Duration:        4 * time.Second,
			}},
		15: {"Submit 100 fake requests with 1 node in simulation",
			&TestConfig{
				NumReplicas:     1,
				NumClients:      0,
				Transport:       "sim",
				NumFakeRequests: 100,
				Duration:        20 * time.Second,
			}},
		16: {"Submit 100 fake requests with 4 nodes in simulation, one of them slow",
			&TestConfig{
				NumReplicas:         4,
				NumClients:          0,
				Transport:           "sim",
				NumFakeRequests:     100,
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
	}

	for i, test := range tests {
		i, test := i, test

		// Create a directory for the deployment-generated files and set the test directory name.
		// The directory will be automatically removed when the outer test function exits.
		createDeploymentDir(t, test.Config)

		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			simMode := (test.Config.Transport == "sim")
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

			if t.Failed() {
				t.Logf("Test #%03d (%s) failed", i, test.Desc)
				if simMode {
					t.Logf("Reproduce with RANDOM_SEED=%d", test.Config.RandomSeed)
				}
			}
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
				NumReplicas: 4,
				NumClients:  1,
				Transport:   "fake",
				Duration:    10 * time.Second,
				Logger:      logging.ConsoleErrorLogger,
			}},
		1: {"Runs for 100s with 4 nodes",
			&TestConfig{
				NumReplicas: 4,
				NumClients:  1,
				Transport:   "fake",
				Duration:    100 * time.Second,
				Logger:      logging.ConsoleErrorLogger,
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
	assert.Len(tb, nodeErrors, conf.NumReplicas)
	for _, err := range nodeErrors {
		if err != nil {
			assert.Equal(tb, mir.ErrStopped, err)
		}
	}

	// Check event logs
	require.NoError(tb, checkEventTraces(deployment.EventLogFiles(), conf.NumNetRequests+conf.NumFakeRequests))

	// Check if all requests were delivered.
	for _, replica := range deployment.TestReplicas {
		app := replica.Modules["app"].(*deploytest.FakeApp)
		assert.Equal(tb, conf.NumNetRequests+conf.NumFakeRequests, int(app.RequestsProcessed))
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
		tb.Logf("Using deployment dir: %s\n", conf.Directory)
		err := os.MkdirAll(conf.Directory, 0777)
		require.NoError(tb, err)
		tb.Cleanup(func() { os.RemoveAll(conf.Directory) })
	} else {
		// If no directory is configured, create a temporary directory in the OS-default location.
		conf.Directory = tb.TempDir()
		tb.Logf("Created temp dir: %s\n", conf.Directory)
	}
}

func newDeployment(conf *TestConfig) (*deploytest.Deployment, error) {
	nodeIDs := deploytest.NewNodeIDs(conf.NumReplicas)
	logger := deploytest.NewLogger(conf.Logger)

	var simulation *deploytest.Simulation
	if conf.Transport == "sim" {
		r := rand.New(rand.NewSource(conf.RandomSeed)) // nolint: gosec
		eventDelayFn := func(e *eventpb.Event) time.Duration {
			// TODO: Make min and max event processing delay configurable
			return testsim.RandDuration(r, 0, time.Microsecond)
		}
		simulation = deploytest.NewSimulation(r, nodeIDs, eventDelayFn)
	}
	transportLayer := deploytest.NewLocalTransportLayer(simulation, conf.Transport, nodeIDs, logger)
	cryptoSystem := deploytest.NewLocalCryptoSystem("pseudo", nodeIDs, logger)

	nodeModules := make(map[t.NodeID]modules.Modules)

	for i, nodeID := range nodeIDs {
		nodeLogger := logging.Decorate(logger, fmt.Sprintf("Node %d: ", i))
		// Dummy application
		fakeApp := deploytest.NewFakeApp("iss", transportLayer.Nodes())

		// ISS configuration
		issConfig := issutil.DefaultParams(transportLayer.Nodes())
		if conf.SlowProposeReplicas[i] {
			// Increase MaxProposeDelay such that it is likely to trigger view change by the SN timeout.
			// Since a sensible value for the segment timeout needs to be stricter than the SN timeout,
			// in the worst case, it will trigger view change by the segment timeout.
			issConfig.MaxProposeDelay = issConfig.PBFTViewChangeSNTimeout
		}

		// ISS instantiation
		issProtocol, err := iss.New(
			nodeID,
			iss.DefaultModuleConfig(),
			issConfig,
			checkpoint.Genesis(iss.InitialStateSnapshot(fakeApp.Snapshot(), issConfig)),
			logging.Decorate(nodeLogger, "ISS: "),
		)
		if err != nil {
			return nil, fmt.Errorf("error creating ISS protocol module: %w", err)
		}

		checkpointing := checkpoint.Factory(
			checkpoint.DefaultModuleConfig(),
			nodeID,
			logging.Decorate(nodeLogger, "CHKP: "),
		)

		transport, err := transportLayer.Link(nodeID)
		if err != nil {
			return nil, fmt.Errorf("error initializing Mir transport: %w", err)
		}

		// Use a simple mempool for incoming requests.
		mempool := simplemempool.NewModule(
			&simplemempool.ModuleConfig{
				Self:   "mempool",
				Hasher: "hasher",
			},
			&simplemempool.ModuleParams{
				MaxTransactionsInBatch: 10,
			},
		)

		// Use fake batch database.
		batchdb := fakebatchdb.NewModule(
			&fakebatchdb.ModuleConfig{
				Self: "batchdb",
			},
		)

		// Instantiate the availability layer.
		availability := multisigcollector.NewReconfigurableModule(
			&multisigcollector.ModuleConfig{
				Self:    "availability",
				Mempool: "mempool",
				BatchDB: "batchdb",
				Net:     "net",
				Crypto:  "crypto",
			},
			nodeID,
			nodeLogger,
		)

		batchFetcher := batchfetcher.NewModule(
			batchfetcher.DefaultModuleConfig(),
			0,
			clientprogress.NewClientProgress(nodeLogger),
		)

		modulesWithDefaults, err := iss.DefaultModules(map[t.ModuleID]modules.Module{
			"app":          fakeApp,
			"crypto":       cryptoSystem.Module(nodeID),
			"iss":          issProtocol,
			"checkpoint":   checkpointing,
			"batchfetcher": batchFetcher,
			"net":          transport,
			"mempool":      mempool,
			"batchdb":      batchdb,
			"availability": availability,
		}, iss.DefaultModuleConfig())
		if err != nil {
			return nil, fmt.Errorf("error initializing the Mir modules: %w", err)
		}

		nodeModules[nodeID] = modulesWithDefaults
	}

	deployConf := &deploytest.TestConfig{
		Info:                   conf.Info,
		Simulation:             simulation,
		TransportLayer:         transportLayer,
		NodeIDs:                nodeIDs,
		Nodes:                  transportLayer.Nodes(),
		NodeModules:            nodeModules,
		NumClients:             conf.NumClients,
		NumFakeRequests:        conf.NumFakeRequests,
		NumNetRequests:         conf.NumNetRequests,
		FakeRequestsDestModule: t.ModuleID("mempool"),
		Directory:              conf.Directory,
		Logger:                 logger,
	}

	return deploytest.NewDeployment(deployConf)
}
