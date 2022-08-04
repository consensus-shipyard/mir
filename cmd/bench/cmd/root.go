// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	id      string
	nrNodes int
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "bench",
		Short: "Mir benchmarking tool",
		Long: "Mir benchmarking tool can run a Mir nodes, measuring latency and" +
			"throughput. The tool can also generate and submit requests to the Mir" +
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
	rootCmd.PersistentFlags().StringVarP(&id, "id", "i", "", "node/client ID")
	_ = rootCmd.MarkPersistentFlagRequired("id")
	rootCmd.PersistentFlags().IntVarP(&nrNodes, "nrNodes", "n", 4, "total number of nodes")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
}
