package flago

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripUnknownFlags(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		expected []string
	}

	testCases := []testCase{
		{
			name:     "no args",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "no unknown flags",
			args:     []string{"-s", "some", "-b"},
			expected: []string{"-s", "some", "-b"},
		},
		{
			name:     "unknown flag",
			args:     []string{"-s", "some", "-b", "-unknown"},
			expected: []string{"-s", "some", "-b"},
		},
		{
			name:     "unknown flag with value",
			args:     []string{"-s", "some", "-b", "-unknown", "value"},
			expected: []string{"-s", "some", "-b"},
		},
		{
			name:     "unknown flag with value and other flag",
			args:     []string{"-s", "some", "-b", "-unknown", "value", "-b"},
			expected: []string{"-s", "some", "-b", "-b"},
		},
		{
			name:     "unknown flag with value and other flag and other unknown flag",
			args:     []string{"-s", "some", "-b", "-unknown", "value", "-b", "-unknown2"},
			expected: []string{"-s", "some", "-b", "-b"},
		},
		{
			name:     "unknown flag with value and other flag and other unknown flag with value",
			args:     []string{"-s", "some", "-b", "-unknown", "value", "-b", "-unknown2", "value2"},
			expected: []string{"-s", "some", "-b", "-b"},
		},
		// check double-dash form: --flag=value and -flag=value
		{
			args:     []string{"--unknown=value", "-s", "some"},
			name:     "dd unknown flag with value and other flag",
			expected: []string{"-s", "some"},
		},
		{
			args:     []string{"-unknown=value", "-s=a"},
			name:     "dd unknown flag with value and other flag",
			expected: []string{"-s=a"},
		},
		{
			name:     "unknown flag after terminator",
			args:     []string{"-s", "some", "--", "-unknown", "value"},
			expected: []string{"-s", "some", "--", "-unknown", "value"},
		},
		{
			name:     "known flag after terminator",
			args:     []string{"-s", "some", "--", "-b"},
			expected: []string{"-s", "some", "--", "-b"},
		},
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String("s", "", "")
	fs.Bool("b", false, "")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := StripUnknownFlags(fs, tc.args)
			require.ElementsMatch(t, tc.expected, actual)
		})
	}
}
