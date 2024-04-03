package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	es "github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v3"

	"github.com/filecoin-project/mir/cmd/bench/localtxgenerator"
	"github.com/filecoin-project/mir/cmd/bench/parameterset"
	issconfig "github.com/filecoin-project/mir/pkg/iss/config"
	"github.com/filecoin-project/mir/pkg/membership"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/trantor"
)

const (
	defaultOutFileWithSettings = "node-config.json"
)

var (
	inFile       string
	outFile      string
	destDir      string
	settingsFile string
	expIDDigits  int

	paramsCmd = &cobra.Command{
		Use:   "params",
		Short: "generate parameters",
		RunE: func(_ *cobra.Command, args []string) error {
			return generateParams(args)
		},
	}
)

func init() {
	rootCmd.AddCommand(paramsCmd)
	paramsCmd.Flags().StringVarP(&membershipFile, "membership", "m", "", "File specifying the system membership.")
	paramsCmd.Flags().StringVarP(&inFile, "input-file", "i", "", "File with parameters to start from. Default parameters used if omitted.")
	paramsCmd.Flags().StringVarP(&outFile, "output-file", "o", "", "File to write parameters to. Standard output used if omitted.")
	paramsCmd.Flags().StringVarP(&destDir, "output-dir", "d", ".", "If a settings file is specified, generated output will be placed inside this directory.")
	paramsCmd.Flags().StringVarP(&settingsFile, "settings", "s", "", "Experiment set description. Generates one configuration file per experiment described in the settings file.")
	paramsCmd.Flags().IntVarP(&expIDDigits, "digits", "w", 4, "Digits used for experiment id (only used with a settings file).")
}

type BenchParams struct {
	Trantor  trantor.Params
	TxGen    localtxgenerator.ModuleParams
	Duration time.Duration `json:",string"`
}

func generateParams(args []string) error {

	// Load membership from file
	var memb *trantorpbtypes.Membership
	if membershipFile != "" {
		nodeAddrs, err := membership.FromFileName(membershipFile)
		if err != nil {
			return es.Errorf("could not load membership: %w", err)
		}
		memb, err = membership.DummyMultiAddrs(nodeAddrs)
		if err != nil {
			return es.Errorf("could not create dummy multiaddrs: %w", err)
		}
	}

	// Load initial params.
	// Either from a file or, if no file was given, use Trantor's default params.
	var params BenchParams
	if inFile != "" {
		if err := loadFromFile(inFile, &params); err != nil {
			return err
		}
		if memb != nil {
			params.Trantor.Iss.InitialMembership = memb
		}
	} else {
		if memb == nil {
			return es.Errorf("neither input file nor membership file specified")
		}

		// Use default parameters.
		params = BenchParams{
			Trantor:  trantor.DefaultParams(memb),
			TxGen:    localtxgenerator.DefaultModuleParams("0"),
			Duration: 0,
		}
	}

	// Represent the parameters as json data. even if we have just loaded them from a json input file,
	// converting them back and forth once makes sure that the json file contains (syntactically) valid params.
	paramsData, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return es.Errorf("could not marshal params to json: %w", err)
	}
	paramsJSON := string(paramsData)

	// Update parameter values as specified on the command line.
	if len(args)%2 != 0 {
		return es.Errorf("number of positional arguments must be even (key-value pairs)")
	}
	for i := 0; i < len(args); i += 2 {
		if paramsJSON, err = setParam(paramsJSON, args[i], args[i+1]); err != nil {
			return es.Errorf("could not set parameter '%s' to value '%s': %w", args[i], args[i+1], err)
		}
	}

	// Check if initial parameters are valid.
	if err := checkParams(paramsJSON); err != nil {
		return es.Errorf("generated parameters in valid: %w", err)
	}

	if settingsFile != "" {
		// If a settings file was given, interpret the `output` argument as a directory.
		// For each parameter set generated from the settings file, create a numbered experiment,
		// create a corresponding subdirectory, and save the parameters inside.

		settings, err := loadSettingsFile(settingsFile)
		if err != nil {
			return es.Errorf("error loading settings: %w", err)
		}

		// Create main directory in which all configurations will be placed as subdirectories.
		if err := os.MkdirAll(destDir, 0777); err != nil {
			return es.Errorf("failed creating destination directory %s: %w", destDir, err)
		}

		// When using a settings file, the default output cannot be stdout, since multiple configurations are created
		// in different directories. We thus use a hard-coded file name to use.
		if outFile == "" {
			outFile = defaultOutFileWithSettings
		}

		// Each generated configuration is assigned a unique experiment ID.
		expID := 0

		for _, item := range settings.Elements() {

			// Create a copy of the base configuration and update it with the generated parameters.
			newJSON := paramsJSON
			for _, setting := range item {
				if newJSON, err = setParam(newJSON, setting.Key, setting.Val); err != nil {
					return es.Errorf("could not set param '%v' to '%v': %w", setting.Key, setting.Val, err)
				}
			}

			// Check if initial parameters are valid.
			if err := checkParams(newJSON); err != nil {
				return es.Errorf("generated parameters for experiment %d in valid: %w", expID, err)
			}

			// Write configuration file to its own subdirectory.
			expDirName := fmt.Sprintf(fmt.Sprintf(destDir+"/%%0%dd", expIDDigits), expID)
			destFileName := expDirName + "/" + outFile
			if err := os.Mkdir(expDirName, 0777); err != nil {
				return es.Errorf("failed creating experiment subdirectory %s: %w", expDirName, err)
			}
			if err := os.WriteFile(destFileName, []byte(fmt.Sprintf("%s\n", newJSON)), 0600); err != nil {
				return es.Errorf("could not write output to file '%s': %w", destFileName, err)
			}

			expID++
		}
	} else {
		// If no settings file was given, simply write the single generated configuration file to `output`
		// (or standard output if no `output` was specified).

		if outFile != "" {
			if err := os.WriteFile(outFile, []byte(fmt.Sprintf("%s\n", paramsJSON)), 0600); err != nil {
				return es.Errorf("could not write output to file '%s': %w", outFile, err)
			}
		} else {
			fmt.Println(paramsJSON)
		}
	}

	return nil
}

func setParam(paramsJSON string, paramName string, value string) (string, error) {
	param := gjson.Get(paramsJSON, paramName)
	if !param.Exists() {
		return "", es.Errorf("parameter does not exist: %s", paramName)
	}

	return sjson.Set(paramsJSON, paramName, value)
}

func checkParams(paramsJSON string) error {

	// Unmarshalling the parameters also serves as a (syntactic) sanity check.
	var params BenchParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return es.Errorf("generated parameters in valid: %w", err)
	}

	// Check ISS parameters for consistency
	if err := issconfig.CheckParams(params.Trantor.Iss); err != nil {
		return es.Errorf("invalid ISS params: %w", err)
	}

	return nil
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

func loadSettingsFile(fileName string) (parameterset.Set, error) {
	// Open file.
	f, err := os.Open(fileName)
	if err != nil {
		return nil, es.Errorf("could not open settings file: %w", err)
	}

	// Schedule closing file.
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Could not close settings file: %s\n", fileName)
		}
	}()

	// Read params file.
	var parameterSet map[string]any
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&parameterSet); err != nil {
		return nil, es.Errorf("failed loading settings from file: %w", err)
	}

	return parameterSet, nil
}
