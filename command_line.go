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

// SetAllowParsingMultipleAliases sets the behavior of Parse() when multiple tag names
// assigned to same field are passed.
// If `true`, it will be ignored and only the last value will be used.
// If `false`, Parse() will return an error.
// Default value is `false`.
func SetAllowParsingMultipleAliases(allow bool) {
	CommandLine.SetAllowParsingMultipleAliases(allow)
}

// SetIgnoreUnknown sets the behavior of Parse() when unknown flags are passed.
// If `true`, they will be ignored.
// If `false`, Parse() will return an error.
// Default value is `false`.
func SetIgnoreUnknown(ignore bool) {
	CommandLine.SetIgnoreUnknown(ignore)
}

// GetIgnoredArgs returns a slice of arguments that were ignored during the last call to Parse()
// because of SetIgnoreUnknown(true), nil otherwise
func GetIgnoredArgs() []string {
	return CommandLine.GetIgnoredArgs()
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
