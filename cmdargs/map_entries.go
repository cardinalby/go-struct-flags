package cmdargs

func (args Args) MapEntries(
	mapper func(Entry) Entry,
) Args {
	res := Args{
		ambiguousAsBool: args.ambiguousAsBool,
	}

	args.IterateEntries(func(entry Entry) bool {
		if mapped := mapper(entry); mapped != nil {
			res.Args = append(res.Args, mapped.TokenStrings()...)
			flagToken, isFlagToken := entry.(FlagEntry)
			mappedFlagToken, isMappedFlagToken := mapped.(FlagEntry)
			if isFlagToken && isMappedFlagToken &&
				(flagToken.Name() != mappedFlagToken.Name() ||
					flagToken.IsBool() != mappedFlagToken.IsBool()) {
				// Flag name or type changed
				if res.knownFlags == nil {
					res.knownFlags = args.knownFlags.Clone()
				}
				res.knownFlags[mappedFlagToken.Name()] = mappedFlagToken.IsBool()
			}
		}
		return true
	})

	if res.knownFlags == nil {
		res.knownFlags = args.knownFlags
	}

	return res
}
