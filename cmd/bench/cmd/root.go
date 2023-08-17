// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	id             string
	membershipFile string
	verbose        bool

	rootCmd = &cobra.Command{
		Use:   "bench",
		Short: "Mir benchmarking tool",
		Long: "Mir benchmarking tool can run a Mir nodes, measuring latency and" +
			"throughput. The tool can also generate and submit transactions to the Mir" +
			"cluster at a specified rate.",
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
}
