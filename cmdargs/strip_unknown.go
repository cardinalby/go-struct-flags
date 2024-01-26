package cmdargs

import (
	"github.com/cardinalby/go-struct-flags/stdutil"
)

func (args Args) StripUnknownFlags(
	ignoredFlagsWithKnownType stdutil.FormalTagNames,
) (res, stripped Args) {
	c := args.WithKnownFlags(ignoredFlagsWithKnownType)
	res.knownFlags = args.knownFlags
	res.ambiguousAsBool = args.ambiguousAsBool
	stripped.knownFlags = args.knownFlags
	stripped.ambiguousAsBool = args.ambiguousAsBool

	isKnownFlag := func(flagName string) bool {
		_, has := args.knownFlags[flagName]
		return has
	}
	c.IterateEntries(func(entry Entry) bool {
		if f, isFlag := entry.(FlagEntry); isFlag && !isKnownFlag(f.Name()) {
			stripped.Args = append(stripped.Args, entry.TokenStrings()...)
		} else {
			res.Args = append(res.Args, entry.TokenStrings()...)
		}
		return true
	})

	return res, stripped
}
