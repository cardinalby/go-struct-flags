package cmdargs

func (args Args) MapFlags(
	mapper func(flag FlagEntry) (mapped Entry),
) Args {
	return args.MapEntries(func(entry Entry) Entry {
		if flag, isFlag := entry.(FlagEntry); isFlag {
			return mapper(flag)
		}
		return entry
	})
}
