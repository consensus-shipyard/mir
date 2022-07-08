/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mir_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
)

const (
	failedTestDir = "failed-test-data"
)

func TestIntegration(t *testing.T) {
	t.Run("ISS", testIntegrationWithISS)
}

func BenchmarkIntegration(b *testing.B) {
	b.Run("ISS", benchmarkIntegrationWithISS)
}

func testIntegrationWithISS(t *testing.T) {
	tests := []struct {
		Desc   string // test description
		Config *deploytest.TestConfig
	}{
		0: {"Do nothing with 1 node",
			&deploytest.TestConfig{
				NumReplicas: 1,
				Transport:   "fake",
				Duration:    4 * time.Second,
			}},
		1: {"Do nothing with 4 nodes",
			&deploytest.TestConfig{
				NumReplicas:         4,
				Transport:           "fake",
				Duration:            20 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		2: {"Submit 10 fake requests with 1 node",
			&deploytest.TestConfig{
				NumReplicas:     1,
				Transport:       "fake",
				NumFakeRequests: 10,
				Directory:       "mirbft-deployment-test",
				Duration:        4 * time.Second,
			}},
		3: {"Submit 10 fake requests with 1 node, loading WAL",
			&deploytest.TestConfig{
				NumReplicas:     1,
				NumClients:      1,
				Transport:       "fake",
				NumFakeRequests: 10,
				Directory:       "mirbft-deployment-test",
				Duration:        4 * time.Second,
			}},
		4: {"Submit 100 fake requests with 4 nodes",
			&deploytest.TestConfig{
				NumReplicas:         4,
				NumClients:          0,
				Transport:           "fake",
				NumFakeRequests:     100,
				Duration:            15 * time.Second,
				SlowProposeReplicas: map[int]bool{0: true},
			}},
		5: {"Submit 10 fake requests with 4 nodes and gRPC networking",
			&deploytest.TestConfig{
				NumReplicas:     4,
				NumClients:      1,
				Transport:       "grpc",
				NumFakeRequests: 10,
				Duration:        4 * time.Second,
			}},
		6: {"Submit 10 requests with 1 node and gRPC networking",
			&deploytest.TestConfig{
				NumReplicas:    1,
				NumClients:     1,
				Transport:      "grpc",
				NumNetRequests: 10,
				Duration:       4 * time.Second,
			}},
		7: {"Submit 10 requests with 4 nodes and gRPC networking",
			&deploytest.TestConfig{
				NumReplicas:    4,
				NumClients:     1,
				Transport:      "grpc",
				NumNetRequests: 10,
				Duration:       4 * time.Second,
			}},
		8: {"Submit 10 fake requests with 4 nodes and libp2p networking",
			&deploytest.TestConfig{
				NumReplicas:     4,
				NumClients:      1,
				Transport:       "libp2p",
				NumFakeRequests: 10,
				Duration:        10 * time.Second,
			}},
		9: {"Submit 10 requests with 1 node and libp2p networking",
			&deploytest.TestConfig{
				NumReplicas:    1,
				NumClients:     1,
				Transport:      "libp2p",
				NumNetRequests: 10,
				Duration:       10 * time.Second,
			}},
		10: {"Submit 10 requests with 4 nodes and libp2p networking",
			&deploytest.TestConfig{
				NumReplicas:    4,
				NumClients:     1,
				Transport:      "libp2p",
				NumNetRequests: 10,
				Duration:       30 * time.Second,
			}},
	}

	for i, test := range tests {
		// Create a directory for the deployment-generated files and set the test directory name.
		// The directory will be automatically removed when the outer test function exits.
		createDeploymentDir(t, test.Config)

		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			runIntegrationWithISSConfig(t, test.Config)

			if t.Failed() {
				t.Logf("Test #%03d (%s) failed", i, test.Desc)
			}
		})
	}
}

func benchmarkIntegrationWithISS(b *testing.B) {
	benchmarks := []struct {
		Desc   string // test description
		Config *deploytest.TestConfig
	}{
		0: {"Runs for 10s with 4 nodes",
			&deploytest.TestConfig{
				NumReplicas: 4,
				NumClients:  1,
				Transport:   "fake",
				Duration:    10 * time.Second,
				Logger:      logging.ConsoleErrorLogger,
			}},
		1: {"Runs for 100s with 4 nodes",
			&deploytest.TestConfig{
				NumReplicas: 4,
				NumClients:  1,
				Transport:   "fake",
				Duration:    100 * time.Second,
				Logger:      logging.ConsoleErrorLogger,
			}},
	}

	for i, bench := range benchmarks {
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

func runIntegrationWithISSConfig(tb testing.TB, conf *deploytest.TestConfig) (heapObjects int64, heapAlloc int64) {
	tb.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create new test deployment.
	deployment, err := deploytest.NewDeployment(conf)
	require.NoError(tb, err)

	// Schedule shutdown of test deployment
	if conf.Duration > 0 {
		go func() {
			time.Sleep(conf.Duration)
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

	// Check if all requests were delivered.
	for _, replica := range deployment.TestReplicas {
		assert.Equal(tb, conf.NumNetRequests+conf.NumFakeRequests, int(replica.App.RequestsProcessed))
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

	return
}

// If conf.Directory is not empty, creates a directory with that path if it does not yet exist.
// If conf.Directory is empty, creates a directory in the OS-default temporary path
// and sets conf.Directory to that path.
func createDeploymentDir(tb testing.TB, conf *deploytest.TestConfig) {
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
