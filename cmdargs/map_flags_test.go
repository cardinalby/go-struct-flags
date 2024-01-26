package cmdargs

import (
	"testing"

	"github.com/cardinalby/go-struct-flags/stdutil"
	"github.com/stretchr/testify/require"
)

func TestArgs_MapFlags(t *testing.T) {
	t.Parallel()

	knownFlags := stdutil.FormalTagNames{
		"b":  true,
		"b2": true,
		"s":  false,
	}
	argsArr := []string{
		"-s", "some", "--unk1", "unk_val1", "-b", "--unk2=unk_val2", "--b2=true", "-unk3", "--", "rem",
	}

	testCases := []struct {
		name          string
		mapper        func(f FlagEntry) Entry
		expArgs       []string
		expKnownFlags stdutil.FormalTagNames
	}{
		{
			name: "no changes",
			mapper: func(f FlagEntry) Entry {
				return f
			},
			expArgs: []string{
				"-s", "some", "--unk1", "unk_val1", "-b", "--unk2=unk_val2", "--b2=true", "-unk3", "--", "rem"},
			expKnownFlags: knownFlags,
		},
		{
			name: "change all",
			mapper: func(f FlagEntry) Entry {
				switch f.Name() {
				case "s":
					return f.WithValue("new_val")
				case "unk1":
					return nil
				case "b":
					return f.WithValue("false")
				case "unk2":
					return f.WithValue("new_val2")
				case "b2":
					return f.WithNoValue()
				default:
					return f
				}
			},
			expArgs:       []string{"-s", "new_val", "-b=false", "--unk2=new_val2", "--b2", "-unk3", "--", "rem"},
			expKnownFlags: knownFlags,
		},
		{
			name: "change all and add new",
			mapper: func(f FlagEntry) Entry {
				switch f.Name() {
				case "s":
					return f.WithNoValue()
				case "unk1":
					return f.WithInline(true)
				case "b":
					return f.WithValue("some")
				case "unk2":
					return f.WithInline(false).WithName("unk22")
				case "b2":
					return UnnamedArgsEntry(append(
						f.WithNoValue().TokenStrings(),
						"--new", "val",
					))
				case "unk3":
					return UnnamedArgsEntry(append(
						[]string{"--new2", "val2"},
						f.TokenStrings()...,
					))
				default:
					return f
				}
			},
			expArgs: []string{
				"-s", "--unk1=unk_val1", "-b=some", "--unk22", "unk_val2", "--b2",
				"--new", "val", "--new2", "val2", "-unk3", "--", "rem"},
			expKnownFlags: stdutil.FormalTagNames{
				"b":     false,
				"b2":    true,
				"s":     true,
				"unk22": false,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := 0
			knownFlagsCopy := knownFlags.Clone()
			res := NewArgs(argsArr).WithKnownFlags(knownFlagsCopy).MapFlags(func(f FlagEntry) Entry {
				defer func() {
					i++
				}()
				return tc.mapper(f)
			})
			require.Equal(t, knownFlags, knownFlagsCopy)
			require.Equal(t, tc.expArgs, res.Args)
			require.Equal(t, tc.expKnownFlags, res.knownFlags)
		})
	}
}
