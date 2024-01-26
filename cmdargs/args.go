package cmdargs

import (
	"flag"

	"github.com/cardinalby/go-struct-flags/stdutil"
)

type Args struct {
	Args            []string
	knownFlags      stdutil.FormalTagNames
	ambiguousAsBool bool
}

func NewArgs(args []string) Args {
	return Args{
		Args: args,
	}
}

func (args Args) WithAmbiguousAsBool(ambiguousAsBool bool) Args {
	args.ambiguousAsBool = ambiguousAsBool
	return args
}

func (args Args) WithFlagSet(flagSets ...*flag.FlagSet) Args {
	args.knownFlags = args.knownFlags.Clone()
	for _, fls := range flagSets {
		for flagName, isBoolFlag := range stdutil.GetFormalFlagNames(fls) {
			args.knownFlags[flagName] = isBoolFlag
		}
	}
	return args
}

func (args Args) WithKnownFlags(knownFlags stdutil.FormalTagNames) Args {
	args.knownFlags = args.knownFlags.Clone()
	for flagName, isBoolFlag := range knownFlags {
		args.knownFlags[flagName] = isBoolFlag
	}
	return args
}

func (args Args) WithoutKnownFlags(knownFlagsToRemove stdutil.FormalTagNames) Args {
	args.knownFlags = args.knownFlags.Clone()
	for flagName := range knownFlagsToRemove {
		delete(args.knownFlags, flagName)
	}
	return args
}
