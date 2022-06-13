/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mir_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/filecoin-project/mir/pkg/iss"
	t "github.com/filecoin-project/mir/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
)

var (
	// Map of all the directories accessed by the tests.
	// All of those will be deleted after the tests complete.
	// We are not deleting them on the fly, to make it possible for a test
	// to access a directory created by a previous test.
	tempDirs = make(map[string]struct{})
)

// TODO: Update Jason's comment.
// StressyTest attempts to spin up as 'real' a network as possible, using
// fake links, but real concurrent go routines.  This means the test is non-deterministic
// so we can't make assertions about the state of the network that are as specific
// as the more general single threaded testengine type tests.  Still, there
// seems to be value in confirming that at a basic level, a concurrent network executes
// correctly.
var _ = Describe("Basic test", func() {

	var (
		currentTestConfig *deploytest.TestConfig

		// The deployment used by the test.
		deployment *deploytest.Deployment

		// When the deployment stops, the final node statuses will be written here.
		finalStatuses []deploytest.NodeStatus

		heapObjects int64
		heapAlloc   int64

		ctx context.Context
	)

	// Before each run, clear the test state variables.
	BeforeEach(func() {
		finalStatuses = nil
		ctx = context.Background()
	})

	// Define what happens when the test runs.
	// Each run is parametrized with a TestConfig
	testFunc := func(testConfig *deploytest.TestConfig) {

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// Set current test config so it can be accessed after the test is complete.
		currentTestConfig = testConfig

		// Create a directory for the deployment-generated files
		// and s the test directory name, for later deletion.
		err := createDeploymentDir(testConfig)
		Expect(err).NotTo(HaveOccurred())
		tempDirs[testConfig.Directory] = struct{}{}

		// Create new test deployment.
		deployment, err = deploytest.NewDeployment(testConfig)
		Expect(err).NotTo(HaveOccurred())

		// Schedule shutdown of test deployment
		if testConfig.Duration != 0 {
			go func() {
				time.Sleep(testConfig.Duration)
				cancel()
			}()
		}

		// Run deployment until it stops and returns final node statuses.
		finalStatuses, heapObjects, heapAlloc = deployment.Run(ctx)
		fmt.Printf("Deployment run returned.")

		// Check whether all the test replicas exited correctly.
		Expect(finalStatuses).NotTo(BeNil())
		for _, status := range finalStatuses {
			if status.ExitErr != mir.ErrStopped {
				Expect(status.ExitErr).NotTo(HaveOccurred())
			}
			Expect(status.StatusErr).NotTo(HaveOccurred())
		}

		// Check if all requests were delivered.
		for _, replica := range deployment.TestReplicas {
			Expect(int(replica.App.RequestsProcessed)).To(Equal(testConfig.NumFakeRequests + testConfig.NumNetRequests))
		}

		fmt.Printf("Test finished.\n\n")
	}

	// After each test, check for errors and print them if any occurred.
	AfterEach(func() {
		// If the test failed
		if CurrentSpecReport().Failed() {

			// Keep the generated data
			retainedDir := fmt.Sprintf("failed-test-data/%s", currentTestConfig.Directory)
			fmt.Printf("Test failed. Moving deployment data to: %s\n", retainedDir)
			if err := os.MkdirAll(retainedDir, 0777); err != nil {
				fmt.Printf("Failed to create directlry: %s\n", retainedDir)
			}
			if err := os.Rename(currentTestConfig.Directory, retainedDir); err != nil {
				fmt.Printf("Failed renaming directory %s to %s\n", currentTestConfig.Directory, retainedDir)
			}

			// Print final status of the system.
			fmt.Printf("\n\nPrinting status because of failed test in %s\n",
				CurrentGinkgoTestDescription().TestText)

			for nodeIndex, nodeStatus := range finalStatuses {
				fmt.Printf("\nStatus for node %d\n", nodeIndex)

				// ErrStopped indicates normal termination.
				// If another error is detected, print it.
				if nodeStatus.ExitErr == mir.ErrStopped {
					fmt.Printf("\nStopped normally\n")
				} else {
					fmt.Printf("\nStopped with error: %+v\n", nodeStatus.ExitErr)
				}

				// Print node status if available.
				if nodeStatus.StatusErr != nil {
					// If node status could not be obtained, print the associated error.
					fmt.Printf("Could not obtain final status of node %d: %v", nodeIndex, nodeStatus.StatusErr)
				} else {
					// Otherwise, print the status.
					// TODO: print the status in a readable form.
					fmt.Printf("%v\n", nodeStatus.Status)
				}
			}
		}
	})

	// Create an alternative ISS configuration for a 4-node deployment with a long batch timeout.
	// It will be used to simulate one node being slow.
	membership := make([]t.NodeID, 4)
	for i := 0; i < len(membership); i++ {
		membership[i] = t.NewNodeIDFromInt(i)
	}
	slowProposeConfig := iss.DefaultConfig(membership)
	slowProposeConfig.MaxProposeDelay = 2 * time.Second

	DescribeTable("Simple tests", testFunc,
		Entry("Does nothing with 1 node", &deploytest.TestConfig{
			NumReplicas: 1,
			Transport:   "fake",
			Directory:   "",
			Duration:    4 * time.Second,
		}),
		Entry("Does nothing with 4 nodes", &deploytest.TestConfig{
			NumReplicas:           4,
			Transport:             "fake",
			Directory:             "",
			Duration:              20 * time.Second,
			FirstReplicaISSConfig: slowProposeConfig,
		}),
		Entry("Submits 10 fake requests with 1 node", &deploytest.TestConfig{
			NumReplicas:     1,
			Transport:       "fake",
			NumFakeRequests: 10,
			Directory:       "mirbft-deployment-test",
			Duration:        4 * time.Second,
		}),
		Entry("Submits 10 fake requests with 1 node, loading WAL", &deploytest.TestConfig{
			NumReplicas:     1,
			NumClients:      1,
			Transport:       "fake",
			NumFakeRequests: 10,
			Directory:       "mirbft-deployment-test",
			Duration:        4 * time.Second,
		}),
		Entry("Submits 100 fake requests with 4 nodes", &deploytest.TestConfig{
			NumReplicas:           4,
			NumClients:            0,
			Transport:             "fake",
			NumFakeRequests:       100,
			Directory:             "",
			Duration:              10 * time.Second,
			FirstReplicaISSConfig: slowProposeConfig,
		}),
		Entry("Submits 10 fake requests with 4 nodes and actual networking", &deploytest.TestConfig{
			NumReplicas:     4,
			NumClients:      1,
			Transport:       "grpc",
			NumFakeRequests: 10,
			Directory:       "",
			Duration:        4 * time.Second,
		}),
		Entry("Submits 10 requests with 1 node and actual networking", &deploytest.TestConfig{
			NumReplicas:    1,
			NumClients:     1,
			Transport:      "grpc",
			NumNetRequests: 10,
			Directory:      "",
			Duration:       4 * time.Second,
		}),
		Entry("Submits 10 requests with 4 nodes and actual networking", &deploytest.TestConfig{
			NumReplicas:    4,
			NumClients:     1,
			Transport:      "grpc",
			NumNetRequests: 10,
			Directory:      "",
			Duration:       4 * time.Second,
		}),
	)

	// This Pending table will be skipped.
	// Change to FDescribeTable to make it focused.
	XDescribeTable("Memory usage benchmarks", Serial,
		func(c *deploytest.TestConfig) {
			testFunc(c)
			AddReportEntry("Heap objects", heapObjects)
			AddReportEntry("Heap allocated (KB)",
				fmt.Sprintf("%.3f", float64(heapAlloc)/1024))
		},
		Entry("Runs for 10s with 4 nodes", &deploytest.TestConfig{
			NumReplicas: 4,
			NumClients:  1,
			Transport:   "fake",
			Duration:    10 * time.Second,
			Logger:      logging.ConsoleErrorLogger,
		}),
		Entry("Runs for 100s with 4 nodes", &deploytest.TestConfig{
			NumReplicas: 4,
			NumClients:  1,
			Transport:   "fake",
			Duration:    100 * time.Second,
			Logger:      logging.ConsoleErrorLogger,
		}),
	)
})

// Remove all temporary data produced by the tests at the end.
var _ = AfterSuite(func() {
	for dir := range tempDirs {
		fmt.Printf("Removing temporary test directory: %v\n", dir)
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("Could not remove directory: %v\n", err)
		}
	}
})

// If config.Directory is not empty, creates a directory with that path if it does not yet exist.
// If config.Directory is empty, creates a directory in the OS-default temporary path
// and sets config.Directory to that path.
// If an error occurs, it is returned without modifying config.Directory.
func createDeploymentDir(config *deploytest.TestConfig) error {
	if config.Directory != "" {
		fmt.Printf("Using deployment dir: %s\n", config.Directory)
		// If a directory is configured, use the configured one.
		if err := os.MkdirAll(config.Directory, 0777); err != nil {
			return err
		}

		return nil
	}

	// If no directory is configured, create a temporary directory in the OS-default location.
	tmpDir, err := ioutil.TempDir("", "mir-deployment-test.")
	fmt.Printf("Creating temp dir: %s\n", tmpDir)
	if err != nil {
		return err
	}

	config.Directory = tmpDir

	return nil
}
