package cmdargs

import (
	"testing"

	"github.com/cardinalby/go-struct-flags/stdutil"
	"github.com/stretchr/testify/require"
)

func TestArgs_LookupFlag(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		args := Args{}
		res, ok := args.LookupFlag("x")
		require.False(t, ok)
		require.Equal(t, FlagEntry{}, res)
	})

	t.Run("simple", func(t *testing.T) {
		args := Args{
			Args: []string{"-s", "abc", "--x=4"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
			},
		}
		// no match
		res, ok := args.LookupFlag("y")
		require.False(t, ok)
		require.Equal(t, FlagEntry{}, res)
		// has match
		res, ok = args.LookupFlag("x")
		require.True(t, ok)
		require.Equal(t, FlagEntry{
			name:           "x",
			value:          "4",
			isInline:       true,
			isDoubleDashed: true,
			isBool:         false,
		}, res)

	})
}

func TestArgs_DeleteFlag(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		args := Args{}
		res, deleted := args.DeleteFlag("x")
		require.False(t, deleted)
		require.Equal(t, args, res)
	})

	t.Run("simple", func(t *testing.T) {
		args := Args{
			Args: []string{"-s", "abc", "--x=4", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
			},
		}
		// no match
		res, deleted := args.DeleteFlag("y")
		require.False(t, deleted)
		require.Equal(t, args, res)
		// has match
		res, deleted = args.DeleteFlag("x")
		require.True(t, deleted)
		require.Equal(t, Args{
			Args: []string{"-s", "abc", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
			},
		}, res)
	})
}

func TestArgs_UpsertFlag(t *testing.T) {
	t.Run("insert", func(t *testing.T) {
		args := Args{
			Args: []string{"-s", "abc", "xyz"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
			},
		}
		res := args.UpsertFlag(
			NewFlagEntry("x", "4"),
			func(old FlagEntry) FlagEntry {
				require.Fail(t, "should not be called")
				return old
			},
		)
		require.Equal(t, Args{
			Args: []string{"-x", "4", "-s", "abc", "xyz"},
			knownFlags: stdutil.FormalTagNames{
				"x": false,
				"s": false,
			},
		}, res)
		require.Equal(t, stdutil.FormalTagNames{
			"s": false,
		}, args.knownFlags)
	})

	t.Run("update", func(t *testing.T) {
		args := Args{
			Args: []string{"-s", "abc", "--x", "5", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
				"x": false,
			},
		}

		res := args.UpsertFlag(
			NewFlagEntry("x", "10"),
			func(old FlagEntry) FlagEntry {
				return old.WithValue("11").WithInline(true)
			})

		require.Equal(t, Args{
			Args: []string{"-s", "abc", "--x=11", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
				"x": false,
			},
		}, res)
	})

	t.Run("update changing type", func(t *testing.T) {
		args := Args{
			Args: []string{"-s", "abc", "--x", "5", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
				"x": false,
			},
		}

		res := args.UpsertFlag(
			NewFlagEntry("x", "10"),
			func(old FlagEntry) FlagEntry {
				return old.WithNoValue()
			})

		require.Equal(t, Args{
			Args: []string{"-s", "abc", "--x", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
				"x": true,
			},
		}, res)
		require.Equal(t, stdutil.FormalTagNames{
			"s": false,
			"x": false,
		}, args.knownFlags)
	})

	t.Run("update changing name", func(t *testing.T) {
		args := Args{
			Args: []string{"-s", "abc", "--x", "5", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
				"x": false,
			},
		}

		res := args.UpsertFlag(
			NewFlagEntry("x", "10"),
			func(old FlagEntry) FlagEntry {
				return old.WithName("y")
			})

		require.Equal(t, Args{
			Args: []string{"-s", "abc", "--y", "5", "def"},
			knownFlags: stdutil.FormalTagNames{
				"s": false,
				"x": false,
				"y": false,
			},
		}, res)
		require.Equal(t, stdutil.FormalTagNames{
			"s": false,
			"x": false,
		}, args.knownFlags)
	})
}
