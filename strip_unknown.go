package flago

import (
	"flag"

	"github.com/cardinalby/go-struct-flags/iterator"
)

func StripUnknownFlags(flagSet *flag.FlagSet, args []string, treatUnknownFlagsAsBool bool) (res, stripped []string) {
	yieldTransform := iterator.TreatUnknownFlagsAsNonBool
	if treatUnknownFlagsAsBool {
		yieldTransform = iterator.TreatUnknownFlagsAsBool
	}
	iterator.Iterate(args, flagSet, yieldTransform(func(info iterator.ArgInfo) bool {
		if !info.Role.Has(iterator.ArgRoleFlag) &&
			!info.Role.Has(iterator.ArgRoleFlagValue) ||
			info.Role.Has(iterator.ArgRoleKnown) {
			res = append(res, info.Arg)
		} else {
			stripped = append(stripped, info.Arg)
		}
		return true
	}))
	return res, stripped
}
