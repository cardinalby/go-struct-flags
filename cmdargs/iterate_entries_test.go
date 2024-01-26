package cmdargs

import (
	"testing"

	"github.com/cardinalby/go-struct-flags/stdutil"
	"github.com/stretchr/testify/require"
)

func TestArgs_IterateEntries(t *testing.T) {
	t.Parallel()
	knownFlags := stdutil.FormalTagNames{
		"b":  true,
		"b2": true,
		"s":  false,
	}
	argsArr := []string{
		"-s", "some", "--unk1", "unk_val1", "-b", "--unk2=unk_val2", "--b2=true", "-unk3", "--", "rem", "rem2",
	}
	expSeen := []Entry{
		NewFlagEntry("s", "some"),
		NewFlagEntry("unk1", "unk_val1").WithDoubleDashes(true),
		NewBoolFlagEntry("b", ""),
		NewFlagEntry("unk2", "unk_val2").WithInline(true).WithDoubleDashes(true),
		NewBoolFlagEntry("b2", "true").WithDoubleDashes(true),
		NewBoolFlagEntry("unk3", ""),
		NewTerminatorEntry(),
		UnnamedArgsEntry{"rem", "rem2"},
	}
	var seen []Entry
	NewArgs(argsArr).WithKnownFlags(knownFlags).IterateEntries(func(info Entry) (getNext bool) {
		seen = append(seen, info)
		return true
	})
	require.Equal(t, expSeen, seen)
}
