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
		stripped []string
	}

	testCases := []testCase{
		{
			name:     "no args",
			args:     []string{},
			expected: []string{},
			stripped: nil,
		},
		{
			name:     "no unknown flags",
			args:     []string{"-s", "some", "-b"},
			expected: []string{"-s", "some", "-b"},
			stripped: nil,
		},
		{
			name:     "unknown flag",
			args:     []string{"-s", "some", "-b", "-unknown"},
			expected: []string{"-s", "some", "-b"},
			stripped: []string{"-unknown"},
		},
		{
			name:     "unknown flag with value",
			args:     []string{"-s", "some", "-b", "-unknown", "value"},
			expected: []string{"-s", "some", "-b"},
			stripped: []string{"-unknown", "value"},
		},
		{
			name:     "unknown flag with value and other flag",
			args:     []string{"-s", "some", "-b", "-unknown", "value", "-b"},
			expected: []string{"-s", "some", "-b", "-b"},
			stripped: []string{"-unknown", "value"},
		},
		{
			name:     "unknown flag with value and other flag and other unknown flag",
			args:     []string{"-s", "some", "-b", "-unknown", "value", "-b", "-unknown2"},
			expected: []string{"-s", "some", "-b", "-b"},
			stripped: []string{"-unknown", "value", "-unknown2"},
		},
		{
			name:     "unknown flag with value and other flag and other unknown flag with value",
			args:     []string{"-s", "some", "-b", "-unknown", "value", "-b", "-unknown2", "value2"},
			expected: []string{"-s", "some", "-b", "-b"},
			stripped: []string{"-unknown", "value", "-unknown2", "value2"},
		},
		{
			args:     []string{"--unknown=value", "-s", "some"},
			name:     "dd unknown flag with value and other flag",
			expected: []string{"-s", "some"},
			stripped: []string{"--unknown=value"},
		},
		{
			args:     []string{"-unknown=value", "-s=a"},
			name:     "dd unknown flag with value and other flag",
			expected: []string{"-s=a"},
			stripped: []string{"-unknown=value"},
		},
		{
			name:     "unknown flag after terminator",
			args:     []string{"-s", "some", "--", "-unknown", "value"},
			expected: []string{"-s", "some", "--", "-unknown", "value"},
			stripped: nil,
		},
		{
			name:     "known flag after terminator",
			args:     []string{"-s", "some", "--", "-b"},
			expected: []string{"-s", "some", "--", "-b"},
			stripped: nil,
		},
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String("s", "", "")
	fs.Bool("b", false, "")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, stripped := StripUnknownFlags(fs, tc.args)
			require.ElementsMatch(t, tc.expected, actual)
			require.ElementsMatch(t, tc.stripped, stripped)
		})
	}
}
