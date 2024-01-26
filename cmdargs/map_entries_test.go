package cmdargs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArgs_MapEntries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		args   []string
		mapper func(token Entry) Entry
		expRes []string
	}{
		{
			name: "no args",
			args: []string{},
			mapper: func(token Entry) Entry {
				return token
			},
			expRes: []string(nil),
		},
		{
			name: "no changes",
			args: []string{"-s", "some", "--", "abc"},
			mapper: func(token Entry) Entry {
				return token
			},
			expRes: []string{"-s", "some", "--", "abc"},
		},
		{
			name: "replace terminator and rm unnamed args",
			args: []string{"-s", "some", "--", "abc"},
			mapper: func(token Entry) Entry {
				if token.Kind() == EntryKindTerminator {
					return UnnamedArgsEntry{"xyz", "123"}
				}
				if token.Kind() == EntryKindUnnamedArgs {
					return nil
				}
				return token
			},
			expRes: []string{"-s", "some", "xyz", "123"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := Args{Args: tc.args}
			res := args.MapEntries(tc.mapper)
			require.Equal(t, tc.expRes, res.Args)
		})
	}
}
