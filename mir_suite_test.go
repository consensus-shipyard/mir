/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0

Refactored: 1
*/

package mir_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TODO: Reassess the following comment, the need for ContextTimeout itself, and the contents of the init() function.
// ContextTimeout is an unfortunate parameter included for executing
// the stress related tests.  Most of the testing we try to make deterministic
// and independent of time (for instance, by specifying step counts), but for
// the more 'real' integration stress tests, this is not possible.  Since
// the CI hardware is weak, and, the race detector slows testing considerably,
// this value is overridden via MIR_TEST_CONTEXT_TIMEOUT in CI.
var ContextTimeout = 30 * time.Second

func init() {
	val := os.Getenv("MIR_TEST_STRESS_TICK_INTERVAL")
	if val != "" {
		dur, err := time.ParseDuration(val)
		if err != nil {
			fmt.Printf("Could not parse duration for stress tick interval: %s\n", err)
			return
		}
		fmt.Printf("Setting tick interval to be %v\n", dur)
		tickInterval = dur
	}

	val = os.Getenv("MIR_TEST_STRESS_TEST_TIMEOUT")
	if val != "" {
		dur, err := time.ParseDuration(val)
		if err != nil {
			fmt.Printf("Could not parse duration for stress tick interval: %s\n", err)
			return
		}
		fmt.Printf("Setting test timeout to be %v\n", dur)
		testTimeout = dur
	}
}

// Runs the tests specified (in separate files) using the Ginkgo testing framework.
func TestMir(t *testing.T) {
	RegisterFailHandler(Fail)

	// Override the ContextTimeout value based on an environment variable.
	val := os.Getenv("MIR_TEST_CONTEXT_TIMEOUT")
	if val != "" {
		dur, err := time.ParseDuration(val)
		Expect(err).NotTo(HaveOccurred())
		ContextTimeout = dur
	}

	RunSpecs(t, "Mir Suite")
}
