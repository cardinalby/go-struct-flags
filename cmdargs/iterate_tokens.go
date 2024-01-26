package cmdargs

import (
	"strings"
)

func (args Args) IterateTokens(yield func(info Token) bool) {
	args.iterateTokensImpl(func(token Token) yieldInstr {
		continueInstr := yieldNext
		if token.Role == RoleFlag {
			// Ambiguous unknown flag
			isBool, hasKnownType := args.knownFlags[token.FlagName]
			if (!hasKnownType && args.ambiguousAsBool) || (hasKnownType && isBool) {
				token.Role |= RoleBoolFlag
			} else {
				continueInstr |= yieldExpectValue
			}
		}

		if !yield(token) {
			return yieldStop
		}

		return continueInstr
	})
}

type yieldInstr int

func (i yieldInstr) has(instr yieldInstr) bool {
	return i&instr != 0
}

const (
	// yieldNext instructs to continue iteration
	yieldNext yieldInstr = 1 << iota

	// yieldExpectValue in combination with yieldNext instructs to treat the current RoleFlag result
	// as a flag name that should be followed by value (not a bool flag)
	yieldExpectValue = 1 << iota

	// yieldStop instructs to stop iteration
	yieldStop = 1 << iota
)

// yield is a function that is called for each arg.
// If it receives info.Role == RoleFlag it means that the role of the arg is ambiguous,
// and it can be either bool flag or flag name that should be followed by value.
// In this case the further behavior of the iterator depends on YieldFunc return value.
// - If it's yieldNext, the current arg will be treated as bool flag.
// - If it's yieldNext | yieldExpectValue, the current arg will be treated as flag name that
// and next arg will be treated as value.
func (args Args) iterateTokensImpl(yield func(token Token) yieldInstr) {
	expRole := Role(0)

	for i, arg := range args.Args {
		token := Token{
			Arg: arg,
		}

		if expRole == RoleUnnamed {
			token.Role = expRole
			if yield(token).has(yieldStop) {
				return
			}
			continue
		}

		if expRole.Has(RoleFlagValue) {
			token.Role = expRole
			token.FlagValue = arg
			expRole = 0
			if yield(token).has(yieldStop) {
				return
			}
			continue
		}

		parsed := parseArg(token.Arg)

		if parsed.isTerminator {
			token.Role = RoleTerminator
			expRole = RoleUnnamed
			if yield(token).has(yieldStop) {
				return
			}
			continue
		}

		if !parsed.isFlag {
			token.Role = RoleUnnamed
			expRole = RoleUnnamed
			if yield(token).has(yieldStop) {
				return
			}
			continue
		}

		token.FlagName = parsed.flagName
		token.FlagValue = parsed.inlineValue
		token.Role = RoleFlag

		isBoolFlag, isKnown := args.knownFlags[parsed.flagName]
		if isKnown {
			token.Role |= RoleKnown
		}
		if parsed.inlineValue != "" {
			token.Role |= RoleInline
		}
		isLastFlag := func(i int) bool {
			argsLen := len(args.Args)
			return i == argsLen-1 || (argsLen >= i && args.Args[i+1] == "--")
		}
		if !isKnown && isLastFlag(i) && parsed.inlineValue == "" {
			isBoolFlag = true
		}
		if isBoolFlag {
			token.Role |= RoleBoolFlag
		}

		yieldRes := yield(token)
		if yieldRes.has(yieldStop) {
			return
		}
		if parsed.inlineValue == "" && !isBoolFlag && (isKnown || yieldRes.has(yieldExpectValue)) {
			expRole = RoleFlagValue
			if isKnown {
				expRole |= RoleKnown
			}
		}
	}
}

type parsedArg struct {
	isFlag       bool
	isTerminator bool
	flagName     string
	inlineValue  string
}

func parseArg(arg string) (res parsedArg) {
	if len(arg) < 2 || arg[0] != '-' {
		return res
	}
	argFlagNameStartIndex := 1
	if arg[1] == '-' {
		argFlagNameStartIndex++
		if len(arg) == 2 { // "--" terminates the flags
			res.isTerminator = true
			return res
		}
	}
	flagNameValue := arg[argFlagNameStartIndex:]
	equalsSignIndex := strings.Index(flagNameValue, "=")
	if equalsSignIndex == 0 {
		// std FlagSet.Parse() will return "bad flag syntax" error
		return res
	}
	if equalsSignIndex > 0 {
		res.flagName = flagNameValue[:equalsSignIndex]
		res.inlineValue = flagNameValue[equalsSignIndex+1:]
	} else {
		res.flagName = flagNameValue
	}
	res.isFlag = true
	return res
}
