package cmdargs

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

func testIterateTokensImpl(
	args []string,
	yieldInstructions []yieldInstr,
	expected []Token,
) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		var actual []Token
		i := 0
		NewArgs(args).WithFlagSet(getTestFlagSet()).iterateTokensImpl(func(info Token) yieldInstr {
			require.Less(t, i, len(yieldInstructions))
			actual = append(actual, info)
			res := yieldInstructions[i]
			i++
			return res
		})
		require.Equal(t, expected, actual)
	}
}

func TestIterateImpl(t *testing.T) {
	t.Parallel()
	// s - known string flag
	// b - known bool flag

	t.Run("1", testIterateTokensImpl(
		[]string{"-s", "some", "--b=true", "--", "abc"},
		[]yieldInstr{yieldNext, yieldNext, yieldNext, yieldNext, yieldNext},
		[]Token{
			{
				Arg:      "-s",
				FlagName: "s",
				Role:     RoleFlag | RoleKnown,
			},
			{
				Arg:       "some",
				FlagValue: "some",
				Role:      RoleFlagValue | RoleKnown,
			},
			{
				Arg:       "--b=true",
				FlagName:  "b",
				FlagValue: "true",
				Role:      RoleFlag | RoleKnown | RoleInline | RoleBoolFlag,
			},
			{
				Arg:  "--",
				Role: RoleTerminator,
			},
			{
				Arg:  "abc",
				Role: RoleUnnamed,
			},
		}))

	t.Run("stop", testIterateTokensImpl(
		[]string{"-s", "some", "--b=true", "--", "abc"},
		[]yieldInstr{yieldStop},
		[]Token{
			{
				Arg:      "-s",
				FlagName: "s",
				Role:     RoleFlag | RoleKnown,
			},
		}))

	t.Run("2", testIterateTokensImpl(
		[]string{"--s=some", "-b", "abc", "--", "def"},
		[]yieldInstr{yieldNext, yieldNext, yieldNext, yieldNext, yieldNext},
		[]Token{
			{
				Arg:       "--s=some",
				FlagName:  "s",
				FlagValue: "some",
				Role:      RoleFlag | RoleKnown | RoleInline,
			},
			{
				Arg:      "-b",
				FlagName: "b",
				Role:     RoleFlag | RoleKnown | RoleBoolFlag,
			},
			{
				Arg:  "abc",
				Role: RoleUnnamed,
			},
			{
				Arg:  "--",
				Role: RoleUnnamed,
			},
			{
				Arg:  "def",
				Role: RoleUnnamed,
			},
		},
	))

	t.Run("3", testIterateTokensImpl(
		[]string{"--s", "--", "-b=0", "--", "--abc"},
		[]yieldInstr{yieldNext, yieldNext, yieldNext, yieldNext, yieldNext},
		[]Token{
			{
				Arg:      "--s",
				FlagName: "s",
				Role:     RoleFlag | RoleKnown,
			},
			{
				Arg:       "--",
				FlagValue: "--",
				Role:      RoleFlagValue | RoleKnown,
			},
			{
				Arg:       "-b=0",
				FlagName:  "b",
				FlagValue: "0",
				Role:      RoleFlag | RoleKnown | RoleInline | RoleBoolFlag,
			},
			{
				Arg:  "--",
				Role: RoleTerminator,
			},
			{
				Arg:  "--abc",
				Role: RoleUnnamed,
			},
		},
	))

	t.Run("4", testIterateTokensImpl(
		[]string{"--x", "-s", "--"},
		[]yieldInstr{yieldNext, yieldNext, yieldNext},
		[]Token{
			{
				Arg:      "--x",
				FlagName: "x",
				Role:     RoleFlag,
			},
			{
				Arg:      "-s",
				FlagName: "s",
				Role:     RoleFlag | RoleKnown,
			},
			{
				Arg:       "--",
				FlagValue: "--",
				Role:      RoleFlagValue | RoleKnown,
			},
		},
	))

	t.Run("5", testIterateTokensImpl(
		[]string{"--x", "-s", "--"},
		[]yieldInstr{yieldNext | yieldExpectValue, yieldNext, yieldNext},
		[]Token{
			{
				Arg:      "--x",
				FlagName: "x",
				Role:     RoleFlag,
			},
			{
				Arg:       "-s",
				Role:      RoleFlagValue,
				FlagValue: "-s",
			},
			{
				Arg:  "--",
				Role: RoleTerminator,
			},
		},
	))

	t.Run("last_unknown", testIterateTokensImpl(
		[]string{"--s", "abc", "--x"},
		[]yieldInstr{yieldNext, yieldNext, yieldNext},
		[]Token{
			{
				Arg:      "--s",
				FlagName: "s",
				Role:     RoleFlag | RoleKnown,
			},
			{
				Arg:       "abc",
				FlagValue: "abc",
				Role:      RoleFlagValue | RoleKnown,
			},
			{
				Arg:      "--x",
				FlagName: "x",
				Role:     RoleFlag | RoleBoolFlag,
			},
		},
	))
}

func TestArgs_IterateTokens(t *testing.T) {
	t.Parallel()

	knowFlags := map[string]bool{
		"s": false,
		"b": true,
	}
	testCases := []struct {
		name            string
		args            []string
		ambiguousAsBool bool
		expected        []Token
	}{
		{
			name:            "ambiguousAsBool",
			args:            []string{"-s", "some", "-b", "-unknown", "value"},
			ambiguousAsBool: true,
			expected: []Token{
				{
					Arg:      "-s",
					FlagName: "s",
					Role:     RoleFlag | RoleKnown,
				},
				{
					Arg:       "some",
					FlagValue: "some",
					Role:      RoleFlagValue | RoleKnown,
				},
				{
					Arg:      "-b",
					FlagName: "b",
					Role:     RoleFlag | RoleKnown | RoleBoolFlag,
				},
				{
					Arg:      "-unknown",
					FlagName: "unknown",
					Role:     RoleFlag | RoleBoolFlag,
				},
				{
					Arg:  "value",
					Role: RoleUnnamed,
				},
			},
		},
		{
			name:            "not ambiguousAsBool",
			args:            []string{"-s", "some", "-b", "-unknown", "value"},
			ambiguousAsBool: false,
			expected: []Token{
				{
					Arg:      "-s",
					FlagName: "s",
					Role:     RoleFlag | RoleKnown,
				},
				{
					Arg:       "some",
					FlagValue: "some",
					Role:      RoleFlagValue | RoleKnown,
				},
				{
					Arg:      "-b",
					FlagName: "b",
					Role:     RoleFlag | RoleKnown | RoleBoolFlag,
				},
				{
					Arg:      "-unknown",
					FlagName: "unknown",
					Role:     RoleFlag,
				},
				{
					Arg:       "value",
					FlagValue: "value",
					Role:      RoleFlagValue,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := 0
			NewArgs(tc.args).
				WithKnownFlags(knowFlags).
				WithAmbiguousAsBool(tc.ambiguousAsBool).
				IterateTokens(func(token Token) bool {
					require.Equal(t, tc.expected[i], token, "i=%d", i)
					i++
					return true
				})
		})
	}
}

func TestRole_Has(t *testing.T) {
	fls := flag.NewFlagSet("", flag.ContinueOnError)
	s := fls.String("s", "", "")
	require.NoError(t, fls.Parse([]string{`-s=a b`}))
	println(*s)
}
