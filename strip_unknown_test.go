package flago

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripUnknownFlags(t *testing.T) {
	type expected struct {
		res      []string
		stripped []string
	}
	testCases := []struct {
		name                       string
		args                       []string
		ifTreatUnknownAsBool       expected
		ifWaitForValueAfterUnknown expected
	}{
		{
			name: "no args",
			args: []string{},
			ifTreatUnknownAsBool: expected{
				res:      []string{},
				stripped: nil,
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{},
				stripped: nil,
			},
		},
		{
			name: "no unknown flags",
			args: []string{"-s", "some", "-b"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "-b"},
				stripped: nil,
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "-b"},
				stripped: nil,
			},
		},
		{
			name: "unknown flag",
			args: []string{"-s", "some", "-b", "-unknown"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "-b"},
				stripped: []string{"-unknown"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "-b"},
				stripped: []string{"-unknown"},
			},
		},
		{
			name: "unknown flag with value",
			args: []string{"-s", "some", "-b", "-unknown", "value"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "-b", "value"},
				stripped: []string{"-unknown"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "-b"},
				stripped: []string{"-unknown", "value"},
			},
		},
		{
			name: "unknown flag with value and other flag",
			args: []string{"-s", "some", "-b", "-unknown", "value", "-b"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "-b", "value", "-b"},
				stripped: []string{"-unknown"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "-b", "-b"},
				stripped: []string{"-unknown", "value"},
			},
		},
		{
			name: "unknown flag with value and other flag and other unknown flag",
			args: []string{"-s", "some", "-b", "-unknown", "value", "-b", "-unknown2"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "-b", "value", "-b", "-unknown2"}, // "-unknown2" is unnamed arg
				stripped: []string{"-unknown"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "-b", "-b"},
				stripped: []string{"-unknown", "value", "-unknown2"},
			},
		},
		{
			name: "unknown flag with value and other flag and other unknown flag with value",
			args: []string{"-s", "some", "-b", "-unknown", "value", "-b", "-unknown2", "value2"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "-b", "value", "-b", "-unknown2", "value2"},
				stripped: []string{"-unknown"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "-b", "-b"},
				stripped: []string{"-unknown", "value", "-unknown2", "value2"},
			},
		},
		{
			args: []string{"--unknown=value", "-s", "some"},
			name: "dd unknown flag with value and other flag",
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some"},
				stripped: []string{"--unknown=value"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some"},
				stripped: []string{"--unknown=value"},
			},
		},
		{
			args: []string{"-unknown=value", "-s=a"},
			name: "dd unknown flag with value and other flag",
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s=a"},
				stripped: []string{"-unknown=value"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s=a"},
				stripped: []string{"-unknown=value"},
			},
		},
		{
			name: "unknown flag after terminator",
			args: []string{"-s", "some", "--", "-unknown", "value"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "--", "-unknown", "value"},
				stripped: nil,
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "--", "-unknown", "value"},
				stripped: nil,
			},
		},
		{
			name: "known flag after terminator",
			args: []string{"-s", "some", "--", "-b"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "some", "--", "-b"},
				stripped: nil,
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "some", "--", "-b"},
				stripped: nil,
			},
		},
		{
			name: "double-dashed value of known flag",
			args: []string{"-s", "--unk"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "--unk"},
				stripped: nil,
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "--unk"},
				stripped: nil,
			},
		},
		{
			name: "double-dashed value of known flag with the following flags",
			args: []string{"-s", "--unk", "-b", "--unk2"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "--unk", "-b"},
				stripped: []string{"--unk2"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"-s", "--unk", "-b"},
				stripped: []string{"--unk2"},
			},
		},
		{
			name: "unknown bool flag before known flag",
			args: []string{"--unk", "-s", "abc"},
			ifTreatUnknownAsBool: expected{
				res:      []string{"-s", "abc"},
				stripped: []string{"--unk"},
			},
			ifWaitForValueAfterUnknown: expected{
				res:      []string{"abc"},
				stripped: []string{"--unk", "-s"},
			},
		},
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String("s", "", "")
	fs.Bool("b", false, "")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, stripped := StripUnknownFlags(fs, tc.args, true)
			require.ElementsMatch(t, tc.ifTreatUnknownAsBool.res, actual, "ifTreatUnknownAsBool.res")
			require.ElementsMatch(t, tc.ifTreatUnknownAsBool.stripped, stripped, "ifTreatUnknownAsBool.stripped")

			actual, stripped = StripUnknownFlags(fs, tc.args, false)
			require.ElementsMatch(t, tc.ifWaitForValueAfterUnknown.res, actual, "ifWaitForValueAfterUnknown.res")
			require.ElementsMatch(t, tc.ifWaitForValueAfterUnknown.stripped, stripped, "ifWaitForValueAfterUnknown.stripped")
		})
	}
}
