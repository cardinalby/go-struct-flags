package iterator

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func getTestFlagSet() *flag.FlagSet {
	fls := flag.NewFlagSet("", flag.ContinueOnError)
	fls.String("s", "", "")
	fls.Bool("b", false, "")
	return fls
}

func testIterate(
	args []string,
	yieldInstructions []YieldInstr,
	expected []ArgInfo,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		var actual []ArgInfo
		i := 0
		Iterate(args, getTestFlagSet(), func(info ArgInfo) YieldInstr {
			require.Less(t, i, len(yieldInstructions))
			actual = append(actual, info)
			res := yieldInstructions[i]
			i++
			return res
		})
		require.Equal(t, expected, actual)
	}
}

func TestIterate(t *testing.T) {
	t.Parallel()
	// s - known string flag
	// b - known bool flag

	t.Run("1", testIterate(
		[]string{"-s", "some", "--b=true", "--", "abc"},
		[]YieldInstr{YieldNext, YieldNext, YieldNext, YieldNext, YieldNext},
		[]ArgInfo{
			{
				Arg:      "-s",
				FlagName: "s",
				Role:     ArgRoleFlag | ArgRoleKnown,
			},
			{
				Arg:       "some",
				FlagValue: "some",
				Role:      ArgRoleFlagValue | ArgRoleKnown,
			},
			{
				Arg:       "--b=true",
				FlagName:  "b",
				FlagValue: "true",
				Role:      ArgRoleFlag | ArgRoleKnown | ArgRoleInline | ArgRoleBoolFlag,
			},
			{
				Arg:  "--",
				Role: ArgRoleTerminator,
			},
			{
				Arg:  "abc",
				Role: ArgRoleUnnamed,
			},
		}))

	t.Run("stop", testIterate(
		[]string{"-s", "some", "--b=true", "--", "abc"},
		[]YieldInstr{YieldStop},
		[]ArgInfo{
			{
				Arg:      "-s",
				FlagName: "s",
				Role:     ArgRoleFlag | ArgRoleKnown,
			},
		}))

	t.Run("2", testIterate(
		[]string{"--s=some", "-b", "abc", "--", "def"},
		[]YieldInstr{YieldNext, YieldNext, YieldNext, YieldNext, YieldNext},
		[]ArgInfo{
			{
				Arg:       "--s=some",
				FlagName:  "s",
				FlagValue: "some",
				Role:      ArgRoleFlag | ArgRoleKnown | ArgRoleInline,
			},
			{
				Arg:      "-b",
				FlagName: "b",
				Role:     ArgRoleFlag | ArgRoleKnown | ArgRoleBoolFlag,
			},
			{
				Arg:  "abc",
				Role: ArgRoleUnnamed,
			},
			{
				Arg:  "--",
				Role: ArgRoleUnnamed,
			},
			{
				Arg:  "def",
				Role: ArgRoleUnnamed,
			},
		},
	))

	t.Run("3", testIterate(
		[]string{"--s", "--", "-b=0", "--", "--abc"},
		[]YieldInstr{YieldNext, YieldNext, YieldNext, YieldNext, YieldNext},
		[]ArgInfo{
			{
				Arg:      "--s",
				FlagName: "s",
				Role:     ArgRoleFlag | ArgRoleKnown,
			},
			{
				Arg:       "--",
				FlagValue: "--",
				Role:      ArgRoleFlagValue | ArgRoleKnown,
			},
			{
				Arg:       "-b=0",
				FlagName:  "b",
				FlagValue: "0",
				Role:      ArgRoleFlag | ArgRoleKnown | ArgRoleInline | ArgRoleBoolFlag,
			},
			{
				Arg:  "--",
				Role: ArgRoleTerminator,
			},
			{
				Arg:  "--abc",
				Role: ArgRoleUnnamed,
			},
		},
	))

	t.Run("4", testIterate(
		[]string{"--x", "-s", "--"},
		[]YieldInstr{YieldNext, YieldNext, YieldNext},
		[]ArgInfo{
			{
				Arg:      "--x",
				FlagName: "x",
				Role:     ArgRoleFlag,
			},
			{
				Arg:      "-s",
				FlagName: "s",
				Role:     ArgRoleFlag | ArgRoleKnown,
			},
			{
				Arg:       "--",
				FlagValue: "--",
				Role:      ArgRoleFlagValue | ArgRoleKnown,
			},
		},
	))

	t.Run("5", testIterate(
		[]string{"--x", "-s", "--"},
		[]YieldInstr{YieldNext | YieldExpectValue, YieldNext, YieldNext},
		[]ArgInfo{
			{
				Arg:      "--x",
				FlagName: "x",
				Role:     ArgRoleFlag,
			},
			{
				Arg:       "-s",
				Role:      ArgRoleFlagValue,
				FlagValue: "-s",
			},
			{
				Arg:  "--",
				Role: ArgRoleTerminator,
			},
		},
	))

	t.Run("last_unknown", testIterate(
		[]string{"--s", "abc", "--x"},
		[]YieldInstr{YieldNext, YieldNext, YieldNext},
		[]ArgInfo{
			{
				Arg:      "--s",
				FlagName: "s",
				Role:     ArgRoleFlag | ArgRoleKnown,
			},
			{
				Arg:       "abc",
				FlagValue: "abc",
				Role:      ArgRoleFlagValue | ArgRoleKnown,
			},
			{
				Arg:      "--x",
				FlagName: "x",
				Role:     ArgRoleFlag | ArgRoleBoolFlag,
			},
		},
	))
}
