package iterator

import (
	"flag"
	"strings"

	"github.com/cardinalby/go-struct-flags/stdutil"
)

type ArgRole int

func (r ArgRole) Has(role ArgRole) bool {
	return r&role != 0
}

const (
	ArgRoleFlag       ArgRole = 1 << iota
	ArgRoleBoolFlag           = 1 << iota // modifies ArgRoleFlag
	ArgRoleKnown              = 1 << iota // modifies ArgRoleFlag or ArgRoleFlagValue
	ArgRoleInline             = 1 << iota // modifies ArgRoleFlag
	ArgRoleFlagValue          = 1 << iota
	ArgRoleUnnamed            = 1 << iota
	ArgRoleTerminator         = 1 << iota
)

type ArgInfo struct {
	Arg       string
	FlagName  string
	FlagValue string
	// Role is sum of ArgRole constants. Possible values:
	// ArgRoleFlag | ArgRoleKnown | ArgRoleInline // contains FlagValue
	// ArgRoleFlag | ArgRoleKnown | ArgRoleBoolFlag
	// ArgRoleFlag | ArgRoleKnown   // will be followed by value
	// ArgRoleFlag | ArgRoleInline | ArgRoleBoolFlag  // unknown, contains FlagValue
	// ArgRoleFlag | ArgRoleInline  // unknown, contains FlagValue
	// ArgRoleFlag                  // unknown, it's ambiguous whether it's bool flag or flag name
	// ArgRoleUnnamed			    // On receiving it, yield function should decide how to treat it
	// ArgRoleFlagValue
	// ArgRoleUnnamed
	// ArgRoleTerminator
	Role ArgRole
}

type YieldInstr int

func (i YieldInstr) Has(instr YieldInstr) bool {
	return i&instr != 0
}

const (
	// YieldNext instructs to continue iteration
	YieldNext YieldInstr = 1 << iota

	// YieldExpectValue in combination with YieldNext instructs to treat the current ArgRoleFlag result
	// as a flag name that should be followed by value (not a bool flag)
	YieldExpectValue = 1 << iota

	// YieldStop instructs to stop iteration
	YieldStop = 1 << iota
)

// YieldFunc is a function that is called for each arg.
// If it receives info.Role == ArgRoleFlag it means that the role of the arg is ambiguous,
// and it can be either bool flag or flag name that should be followed by value.
// In this case the further behavior of the iterator depends on YieldFunc return value.
// - If it's YieldNext, the current arg will be treated as bool flag.
// - If it's YieldNext | YieldExpectValue, the current arg will be treated as flag name that
// and next arg will be treated as value.
type YieldFunc func(info ArgInfo) YieldInstr

// TreatUnknownFlagsAsBool transforms yield func with bool return value to YieldFunc that always
// treats unknown flags as bool flags if ambiguous
func TreatUnknownFlagsAsBool(yield func(info ArgInfo) bool) YieldFunc {
	return func(info ArgInfo) YieldInstr {
		if yield(info) {
			return YieldNext
		}
		return YieldStop
	}
}

// TreatUnknownFlagsAsNonBool transforms yield func with bool return value to YieldFunc that always
// treats unknown flags as non-bool flags if ambiguous
func TreatUnknownFlagsAsNonBool(yield func(info ArgInfo) bool) YieldFunc {
	return func(info ArgInfo) YieldInstr {
		if yield(info) {
			if info.Role == ArgRoleFlag {
				return YieldNext | YieldExpectValue
			}
			return YieldNext
		}
		return YieldStop
	}
}

func Iterate(
	args []string,
	flagSet *flag.FlagSet,
	yield YieldFunc,
) {
	var knownFlags map[string]bool
	if flagSet != nil {
		knownFlags = stdutil.GetFormalFlagNames(flagSet)
	}
	expRole := ArgRole(0)

	for i, arg := range args {
		info := ArgInfo{
			Arg: arg,
		}

		if expRole == ArgRoleUnnamed {
			info.Role = expRole
			if yield(info).Has(YieldStop) {
				return
			}
			continue
		}

		if expRole.Has(ArgRoleFlagValue) {
			info.Role = expRole
			info.FlagValue = arg
			expRole = 0
			if yield(info).Has(YieldStop) {
				return
			}
			continue
		}

		isFlag, isTerminator, flagName, inlineValue := parseArg(info.Arg)

		if isTerminator {
			info.Role = ArgRoleTerminator
			expRole = ArgRoleUnnamed
			if yield(info).Has(YieldStop) {
				return
			}
			continue
		}

		if !isFlag {
			info.Role = ArgRoleUnnamed
			expRole = ArgRoleUnnamed
			if yield(info).Has(YieldStop) {
				return
			}
			continue
		}

		info.FlagName = flagName
		info.FlagValue = inlineValue
		info.Role = ArgRoleFlag

		isBoolFlag, isKnown := knownFlags[flagName]
		if isKnown {
			info.Role |= ArgRoleKnown
		}
		if inlineValue != "" {
			info.Role |= ArgRoleInline
		}
		if isBoolFlag || (!isKnown && i == len(args)-1) {
			info.Role |= ArgRoleBoolFlag
		}

		yieldRes := yield(info)
		if yieldRes.Has(YieldStop) {
			return
		}
		if inlineValue == "" && !isBoolFlag && (isKnown || yieldRes.Has(YieldExpectValue)) {
			expRole = ArgRoleFlagValue
			if isKnown {
				expRole |= ArgRoleKnown
			}
		}
	}
}

func parseArg(arg string) (isFlag bool, isTerminator bool, flagNameValue string, inlineValue string) {
	if len(arg) < 2 || arg[0] != '-' {
		return false, false, "", ""
	}
	numMinuses := 1
	if arg[1] == '-' {
		numMinuses++
		if len(arg) == 2 { // "--" terminates the flags
			return false, true, "", ""
		}
	}
	flagNameValue = arg[numMinuses:]
	flagName := flagNameValue
	if equalsSignIndex := strings.Index(flagNameValue, "="); equalsSignIndex == 0 {
		// std FlagSet.Parse() will return "bad flag syntax" error
		return false, false, "", ""
	} else if equalsSignIndex > 0 {
		flagName = flagNameValue[:equalsSignIndex]
		inlineValue = flagNameValue[equalsSignIndex+1:]
	}
	return true, false, flagName, inlineValue
}
