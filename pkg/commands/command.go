package commands

import (
	"context"
	"flag"
)

// Command is an interface of implementations of command verbs
// (like "dump", "verify" etc of "afascli dump"/"afascli verify")
type Command interface {
	// Description explains what this verb commands to do
	Description() string

	// Usage prints the syntax of arguments for this command
	Usage() string

	// SetupFlagSet is called to allow the command implementation
	// to setup which option flags it has.
	SetupFlagSet(flagSet *flag.FlagSet)

	// Execute is the main function here. It is responsible to
	// start the execution of the command.
	//
	// `args` are the arguments left unused by verb itself and options.
	Execute(ctx context.Context, cfg Config, args []string) error
}
