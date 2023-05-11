package fwinfo

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/commands"
	"github.com/immune-gmbh/AttestationFailureAnalysisService/pkg/dmidecode"
)

// Command is the implementation of `commands.Command`.
type Command struct {
	firmwareAnalysisAddress *string
	firmwareVersion         *string
	firmwareDate            *string
	printReport             *bool
	tags                    *string
}

// Usage prints the syntax of arguments for this command
func (cmd Command) Usage() string {
	return "<path to the image>"
}

// Description explains what this verb commands to do
func (cmd Command) Description() string {
	return "display information about firmware image"
}

// SetupFlagSet is called to allow the command implementation
// to setup which option flags it has.
func (cmd *Command) SetupFlagSet(flag *flag.FlagSet) {
}

// Execute is the main function here. It is responsible to
// start the execution of the command.
//
// `args` are the arguments left unused by verb itself and options.
func (cmd Command) Execute(cfg commands.Config, args []string) error {
	if len(args) < 1 {
		return commands.ErrArgs{Err: fmt.Errorf("error: no path to the firmare was specified")}
	}
	if len(args) > 1 {
		return commands.ErrArgs{Err: fmt.Errorf("error: too many parameters")}
	}
	imagePath := args[0]

	imageBytes, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("unable to read image '%s': %w", imagePath, err)
	}

	dmiTable, err := dmidecode.DMITableFromFirmwareImage(imageBytes)
	if err != nil {
		return fmt.Errorf("unable to parse the image info: '%w'", err)
	}

	b, err := json.Marshal(dmiTable.BIOSInfo())
	if err != nil {
		return fmt.Errorf("unable to serialize BIOSInfo: %w", err)
	}
	fmt.Printf("%s\n", b)

	return nil
}