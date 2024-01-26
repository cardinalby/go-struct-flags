package cmdargs

func (args Args) IterateEntries(yield func(entry Entry) (getNext bool)) {
	var unnamedArgs UnnamedArgsEntry
	var prevFlagNameToken Token
	args.IterateTokens(func(token Token) bool {
		switch {
		case token.Role.Has(RoleFlag):
			isBool := token.Role.Has(RoleBoolFlag)
			isInline := token.Role.Has(RoleInline)
			if !isInline && !isBool {
				prevFlagNameToken = token
				return true
			}
			return yield(FlagEntry{
				name:           token.FlagName,
				value:          token.FlagValue,
				isInline:       isInline,
				isDoubleDashed: token.Arg[1] == '-',
				isBool:         isBool,
			})
		case token.Role.Has(RoleFlagValue):
			yieldRes := yield(FlagEntry{
				name:           prevFlagNameToken.FlagName,
				value:          token.FlagValue,
				isInline:       false,
				isDoubleDashed: prevFlagNameToken.Arg[1] == '-',
				isBool:         false,
			})
			prevFlagNameToken = Token{}
			return yieldRes
		case token.Role.Has(RoleTerminator):
			return yield(NewTerminatorEntry())
		default:
			unnamedArgs = append(unnamedArgs, token.Arg)
			return true
		}
	})
	if len(unnamedArgs) > 0 {
		yield(unnamedArgs)
	}
}
