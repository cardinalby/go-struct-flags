package stdutil

import "flag"

func GetExistingFlagNames(flagSet *flag.FlagSet) map[string]struct{} {
	flags := make(map[string]struct{})
	flagSet.Visit(func(f *flag.Flag) {
		flags[f.Name] = struct{}{}
	})
	return flags
}

type boolFlag interface {
	IsBoolFlag() bool
}

// FormalTagNames is a map where key is a flag name and value indicates it's a bool flag
type FormalTagNames map[string]bool

func (flags FormalTagNames) Clone() FormalTagNames {
	clone := make(FormalTagNames, len(flags))
	for flagName, isBoolFlag := range flags {
		clone[flagName] = isBoolFlag
	}
	return clone
}

// GetFormalFlagNames returns a map where key is a flag name and value indicates it's a bool flag
func GetFormalFlagNames(flagSet *flag.FlagSet) FormalTagNames {
	flags := make(FormalTagNames)
	flagSet.VisitAll(func(f *flag.Flag) {
		isBoolFlag := false
		if boolFlag, ok := f.Value.(boolFlag); ok {
			isBoolFlag = boolFlag.IsBoolFlag()
		}
		flags[f.Name] = isBoolFlag
	})
	return flags
}
