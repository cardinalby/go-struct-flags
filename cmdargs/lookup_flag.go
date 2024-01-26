package cmdargs

func (args Args) LookupFlag(flagName string) (res FlagEntry, has bool) {
	args.IterateEntries(func(entry Entry) bool {
		if f, isFlag := entry.(FlagEntry); isFlag && f.Name() == flagName {
			res = f
			has = true
			return false
		}
		return true
	})
	return res, has
}

func (args Args) DeleteFlag(flagName string) (res Args, deleted bool) {
	res = args.MapEntries(func(entry Entry) Entry {
		if f, isFlag := entry.(FlagEntry); isFlag && f.Name() == flagName {
			deleted = true
			return nil
		}
		return entry
	})
	return res, deleted
}

func (args Args) UpsertFlag(
	insert FlagEntry,
	update func(old FlagEntry) (updated FlagEntry),
) Args {
	wasUpdated := false
	if res := args.MapEntries(func(entry Entry) Entry {
		if f, isFlag := entry.(FlagEntry); isFlag && f.Name() == insert.name {
			wasUpdated = true
			return update(f)
		}
		return entry
	}); wasUpdated {
		return res
	}

	res := Args{
		Args:            append(insert.TokenStrings(), args.Args...),
		knownFlags:      args.knownFlags.Clone(),
		ambiguousAsBool: args.ambiguousAsBool,
	}
	res.knownFlags[insert.name] = insert.IsBool()
	return res
}
