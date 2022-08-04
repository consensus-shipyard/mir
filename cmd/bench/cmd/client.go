// Copyright Contributors to the Mir project
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"math/rand"
	"time"

	rateLimiter "golang.org/x/time/rate"

	"github.com/spf13/cobra"

	"github.com/filecoin-project/mir/pkg/dummyclient"
	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
)

var (
	reqSize  int
	rate     float64
	burst    int
	duration time.Duration

	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Generate and submit requests to a Mir cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClient()
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().IntVarP(&reqSize, "reqSize", "s", 256, "size of each request in bytes")
	clientCmd.Flags().Float64VarP(&rate, "rate", "r", 1000, "average number of requests per second")
	clientCmd.Flags().IntVarP(&burst, "burst", "b", 1, "maximum number of requests in a burst")
	clientCmd.Flags().DurationVarP(&duration, "duration", "T", 10*time.Second, "benchmarking duration")
}

func runClient() error {
	var logger logging.Logger
	if verbose {
		logger = logging.ConsoleDebugLogger
	} else {
		logger = logging.ConsoleWarnLogger
	}

	nodeIDs := make([]t.NodeID, nrNodes)
	for i := range nodeIDs {
		nodeIDs[i] = t.NewNodeIDFromInt(i)
	}

	reqReceiverAddrs := make(map[t.NodeID]string)
	for i := range nodeIDs {
		reqReceiverAddrs[t.NewNodeIDFromInt(i)] = fmt.Sprintf("127.0.0.1:%d", ReqReceiverBasePort+i)
	}

	ctx, stop := context.WithCancel(context.Background())

	client := dummyclient.NewDummyClient(
		t.ClientID(id),
		crypto.SHA256,
		logger,
	)
	client.Connect(ctx, reqReceiverAddrs)
	defer client.Disconnect()

	go func() {
		time.Sleep(duration)
		stop()
	}()

	limiter := rateLimiter.NewLimiter(rateLimiter.Limit(rate), 1)
	reqBytes := make([]byte, reqSize)
	for i := 0; ; i++ {
		if err := limiter.Wait(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				err = nil
			}
			return err
		}
		rand.Read(reqBytes) //nolint:gosec
		logger.Log(logging.LevelDebug, fmt.Sprintf("Submitting request #%d", i))
		if err := client.SubmitRequest(reqBytes); err != nil {
			return err
		}
	}
}
