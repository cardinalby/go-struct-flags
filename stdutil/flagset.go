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

// GetFormalFlagNames returns a map where key is a flag name and value indicates it's a bool flag
func GetFormalFlagNames(flagSet *flag.FlagSet) map[string]bool {
	flags := make(map[string]bool)
	flagSet.VisitAll(func(f *flag.Flag) {
		isBoolFlag := false
		if boolFlag, ok := f.Value.(boolFlag); ok {
			isBoolFlag = boolFlag.IsBoolFlag()
		}
		flags[f.Name] = isBoolFlag
	})
	return flags
}
