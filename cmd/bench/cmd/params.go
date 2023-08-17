package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/filecoin-project/mir/pkg/localtxgenerator"
	"github.com/filecoin-project/mir/pkg/trantor"
	es "github.com/go-errors/errors"
)

type BenchParams struct {
	Trantor  trantor.Params
	TxGen    localtxgenerator.ModuleParams
	Duration time.Duration `json:",string"`
}

func loadFromFile(fileName string, dest *BenchParams) error {
	// Open file.
	f, err := os.Open(fileName)
	if err != nil {
		return es.Errorf("could not open parameter file: %w", err)
	}

	// Schedule closing file.
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Could not close parameter file: %s\n", fileName)
		}
	}()

	// Read params file.
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(dest); err != nil {
		return es.Errorf("failed loading parameters from file: %w", err)
	}

	return nil
}
