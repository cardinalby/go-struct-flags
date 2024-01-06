package flago

import (
	"flag"
	"os"
)

// CommandLine is a default FlagSet that is used by the package functions.
// It's a wrapper around flag.CommandLine to follow the same pattern as in the stdlib.
var CommandLine = Wrap(flag.CommandLine)

func init() {
	// Override generic FlagSet default Usage with call to global Usage.
	// Note: This is not CommandLine.Usage = Usage,
	// because we want any eventual call to use any updated value of Usage,
	// not the value it has when this line is run.
	CommandLine.Usage = commandLineUsage
}

func commandLineUsage() {
	Usage()
}

// StructVar registers the given struct with the default FlagSet
// See FlagSet.StructVar
func StructVar(p any) error {
	return CommandLine.StructVar(p)
}

// StructVarWithPrefix registers the given struct with the default FlagSet using
// the given prefix for flag names.
// See FlagSet.StructVarWithPrefix
func StructVarWithPrefix(p any, flagsPrefix string) error {
	return CommandLine.StructVarWithPrefix(p, flagsPrefix)
}

func SetAllowParsingMultipleAliases(allow bool) {
	CommandLine.SetAllowParsingMultipleAliases(allow)
}

// Parse parses the command-line flags using the default FlagSet
func Parse() error {
	return CommandLine.Parse(os.Args[1:])
}

// PrintDefaults prints the default FlagSet usage to stdout grouping alternative flag names
func PrintDefaults() {
	CommandLine.PrintDefaults()
}

var Usage = func() {
	printUsageTitle(CommandLine.FlagSet, os.Args[0])
	PrintDefaults()
}
