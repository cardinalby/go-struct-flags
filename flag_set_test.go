package flago

import (
	"bufio"
	"bytes"
	"flag"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	NotTagged  int8
	NotTagged2 int8          `flag:"-"`
	Str        string        `flag:"s" flagUsage:"usage1"`
	Int        int           `flags:"i,ii" flagUsage:"usage2"`
	Bool       bool          `flag:"b" flagUsage:"usage3"`
	Duration   time.Duration `flag:"d" flagUsage:"usage4"`
	StrP       *string       `flag:"sp" flagUsage:"usage1"`
	IntP       *int          `flag:"ip" flagUsage:"usage2"`
	BoolP      *bool         `flag:"bp" flagUsage:"usage3"`
	Extra      []string      `flagArgs:"true"`
}

func requireEqualPtr[T any](t *testing.T, exp, act *T) {
	t.Helper()
	if exp == nil {
		require.Nil(t, act)
		return
	}
	require.NotNil(t, act)
	require.Equal(t, *exp, *act)
}

func ptr[T any](v T) *T {
	return &v
}

func captureOutput(flagSet *FlagSet, f func()) string {
	oldOutput := flagSet.Output()
	defer flagSet.SetOutput(oldOutput)

	var buff bytes.Buffer
	buffWriter := bufio.NewWriter(&buff)
	flagSet.SetOutput(buffWriter)
	f()
	_ = buffWriter.Flush()
	return buff.String()
}

func TestParseValidStruct(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name          string
		args          []string
		val           testStruct
		prefix        string
		parseErr      bool
		expStr        string
		expInt        int
		expBool       bool
		expDuration   time.Duration
		expStrP       *string
		expIntP       *int
		expBoolP      *bool
		expExtra      []string
		expNotTagged  int8
		expNotTagged2 int8
	}

	testCases := []testCase{
		{
			name: "all_mandatory",
			args: []string{"--s", "val2", "--ii=2", "--b", "--d", "1h"},
			val: testStruct{
				NotTagged:  1,
				NotTagged2: 2,
			},
			prefix:        "",
			parseErr:      false,
			expStr:        "val2",
			expInt:        2,
			expBool:       true,
			expDuration:   time.Hour,
			expStrP:       nil,
			expIntP:       nil,
			expBoolP:      nil,
			expExtra:      nil,
			expNotTagged:  1,
			expNotTagged2: 2,
		},
		{
			name: "some_mandatory",
			args: []string{"--s", "val2", "abc", "def"},
			val: testStruct{
				NotTagged:  1,
				NotTagged2: 2,
				Int:        41,
				IntP:       ptr(42),
			},
			prefix:        "",
			parseErr:      false,
			expStr:        "val2",
			expInt:        41,
			expBool:       false,
			expDuration:   0,
			expStrP:       nil,
			expIntP:       ptr(42),
			expBoolP:      nil,
			expExtra:      []string{"abc", "def"},
			expNotTagged:  1,
			expNotTagged2: 2,
		},
		{
			name: "all_optional",
			args: []string{"--sp", "val2", "--ip", "2", "--bp"},
			val: testStruct{
				NotTagged:  1,
				NotTagged2: 2,
				Duration:   time.Hour,
			},
			prefix:        "",
			parseErr:      false,
			expStr:        "",
			expInt:        0,
			expBool:       false,
			expDuration:   time.Hour,
			expStrP:       ptr("val2"),
			expIntP:       ptr(2),
			expBoolP:      ptr(true),
			expExtra:      nil,
			expNotTagged:  1,
			expNotTagged2: 2,
		},
		{
			name: "some_optional",
			args: []string{"--sp", "val2"},
			val: testStruct{
				Str:        "default",
				NotTagged:  1,
				NotTagged2: 2,
			},
			prefix:        "",
			parseErr:      false,
			expStr:        "default",
			expInt:        0,
			expBool:       false,
			expDuration:   0,
			expStrP:       ptr("val2"),
			expIntP:       nil,
			expBoolP:      nil,
			expExtra:      nil,
			expNotTagged:  1,
			expNotTagged2: 2,
		},
		{
			name: "with prefix",
			args: []string{"--pr-sp", "val2", "--pr-i", "2", "abc"},
			val: testStruct{
				Str:        "default",
				NotTagged:  1,
				NotTagged2: 2,
			},
			prefix:        "pr-",
			parseErr:      false,
			expStr:        "default",
			expInt:        2,
			expBool:       false,
			expDuration:   0,
			expStrP:       ptr("val2"),
			expIntP:       nil,
			expBoolP:      nil,
			expExtra:      []string{"abc"},
			expNotTagged:  1,
			expNotTagged2: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
			require.NoError(t, fls.StructVarWithPrefix(&tc.val, tc.prefix))
			err := fls.Parse(tc.args)
			if tc.parseErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expNotTagged, tc.val.NotTagged)
			require.Equal(t, tc.expStr, tc.val.Str)
			require.Equal(t, tc.expInt, tc.val.Int)
			require.Equal(t, tc.expBool, tc.val.Bool)
			require.Equal(t, tc.expDuration, tc.val.Duration)
			requireEqualPtr(t, tc.expStrP, tc.val.StrP)
			requireEqualPtr(t, tc.expIntP, tc.val.IntP)
			requireEqualPtr(t, tc.expBoolP, tc.val.BoolP)
			require.ElementsMatch(t, tc.expExtra, tc.val.Extra)
		})
	}
}

func TestNoTaggedFields(t *testing.T) {
	type s struct {
		Int int
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := s{}
	require.NoError(t, fls.StructVarWithPrefix(&structVal, ""))
	require.NoError(t, fls.Parse([]string{"ad"}))
	require.Empty(t, structVal.Int)
}

func TestInvalidFieldType(t *testing.T) {
	type invalidStruct struct {
		Time time.Time `flag:"t" flagUsage:"usage1"`
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := invalidStruct{}
	require.Error(t, fls.StructVarWithPrefix(&structVal, ""))
}

func TestInvalidFlagArgsFieldType(t *testing.T) {
	type invalidStruct struct {
		A string `flagArgs:"true"`
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := invalidStruct{}
	require.Error(t, fls.StructVarWithPrefix(&structVal, ""))
}

func TestInvalidFlagArgsTagValue(t *testing.T) {
	type invalidStruct struct {
		A string `flagArgs:"abc"`
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := invalidStruct{}
	require.Error(t, fls.StructVarWithPrefix(&structVal, ""))
}

func TestInvalidFlagArgsTogetherWithFlagName(t *testing.T) {
	type invalidStruct struct {
		A string `flag:"a" flagArgs:"true"`
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := invalidStruct{}
	require.Error(t, fls.StructVarWithPrefix(&structVal, ""))
}

func TestTextVar(t *testing.T) {
	type testStruct struct {
		T *big.Float `flag:"t"`
	}
	fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
	structVal := testStruct{}
	require.Error(t, fls.StructVarWithPrefix(&structVal, ""))

	structVal.T = big.NewFloat(3)
	require.NoError(t, fls.StructVarWithPrefix(&structVal, ""))
	require.NoError(t, fls.Parse(nil))
	require.Equal(t, "3", structVal.T.String())

	require.NoError(t, fls.Parse([]string{"--t", "1.2"}))
	require.Equal(t, "1.2", structVal.T.String())
}

func TestFuncVar(t *testing.T) {
	type testStruct struct {
		F func(string) error `flag:"f"`
	}
	fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
	structVal := testStruct{}
	require.Error(t, fls.StructVarWithPrefix(&structVal, ""))

	var receivedStr *string
	structVal.F = func(s string) error {
		receivedStr = &s
		return nil
	}
	require.NoError(t, fls.StructVarWithPrefix(&structVal, ""))
	require.NoError(t, fls.Parse(nil))
	require.Nil(t, receivedStr)

	require.NoError(t, fls.Parse([]string{"--f", "abc"}))
	require.NotNil(t, receivedStr)
	require.Equal(t, "abc", *receivedStr)
}

func TestIgnoreMultipleAliasesNames(t *testing.T) {
	type testStruct struct {
		Int int `flags:"i1,i2"`
	}
	fls := NewFlagSet("", flag.ContinueOnError)
	structVal := testStruct{}
	require.NoError(t, fls.StructVarWithPrefix(&structVal, ""))

	fls.SetAllowParsingMultipleAliases(true)
	require.NoError(t, fls.Parse([]string{"--i1", "1", "--i2", "2"}))
	require.Equal(t, 2, structVal.Int)
}

func TestErrorOnMultipleAliasesNames(t *testing.T) {
	type testStruct struct {
		Int int `flags:"i1,i2"`
	}
	fls := NewFlagSet("", flag.ContinueOnError)
	structVal := testStruct{}
	require.NoError(t, fls.StructVarWithPrefix(&structVal, ""))

	var parseErr error
	output := captureOutput(fls, func() {
		parseErr = fls.Parse([]string{"--i1", "1", "--i2", "2"})
	})
	expectedErrorMsg := `either "i1" or "i2" flag should be used but not both`
	require.ErrorContains(t, parseErr, expectedErrorMsg)
	expectedUsage := "Usage:\n  -i1 -i2 int\n    \t\n"

	expectedOutput := expectedErrorMsg + "\n" + expectedUsage

	require.Equal(t, expectedOutput, output)
}

type testNestedStruct struct {
	NotTagged int8
	Str       string   `flag:"s" flagUsage:"usage1-n"`
	Args      []string `flagArgs:"true"`
}

type testParentStruct struct {
	NotTagged int8
	Str       string           `flag:"s" flagUsage:"usage1"`
	Nested1   testNestedStruct `flagPrefix:"a-"`
	Nested2   testNestedStruct `flagPrefix:"b-"`
}

func TestParseNestedStruct(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name                string
		args                []string
		val                 testParentStruct
		prefix              string
		parseErr            bool
		expStr              string
		expNotTagged        int8
		expNested1Str       string
		expNested1NotTagged int8
		expNested2Str       string
		expNested2NotTagged int8
		expNestedArgs       []string
	}
	testCases := []testCase{
		{
			name: "all_mandatory",
			args: []string{"--p-s", "val2", "--p-a-s", "val3", "--p-b-s", "val4", "abc", "def"},
			val: testParentStruct{
				NotTagged: 1,
				Nested1: testNestedStruct{
					NotTagged: 2,
					Args:      []string{"default"},
				},
				Nested2: testNestedStruct{
					NotTagged: 3,
				},
			},
			prefix:              "p-",
			parseErr:            false,
			expStr:              "val2",
			expNotTagged:        1,
			expNested1Str:       "val3",
			expNested1NotTagged: 2,
			expNested2Str:       "val4",
			expNested2NotTagged: 3,
			expNestedArgs:       []string{"abc", "def"},
		},
		{
			name: "some_mandatory",
			args: []string{"--s", "val2", "--a-s", "val3"},
			val: testParentStruct{
				NotTagged: 1,
				Nested1: testNestedStruct{
					NotTagged: 2,
				},
				Nested2: testNestedStruct{
					NotTagged: 3,
					Str:       "default",
				},
			},
			prefix:              "",
			parseErr:            false,
			expStr:              "val2",
			expNotTagged:        1,
			expNested1Str:       "val3",
			expNested1NotTagged: 2,
			expNested2Str:       "default",
			expNested2NotTagged: 3,
			expNestedArgs:       nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
			require.NoError(t, fls.StructVarWithPrefix(&tc.val, tc.prefix))
			err := fls.Parse(tc.args)
			if tc.parseErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expNotTagged, tc.val.NotTagged)
			require.Equal(t, tc.expStr, tc.val.Str)
			require.Equal(t, tc.expNested1NotTagged, tc.val.Nested1.NotTagged)
			require.Equal(t, tc.expNested1Str, tc.val.Nested1.Str)
			require.Equal(t, tc.expNested2NotTagged, tc.val.Nested2.NotTagged)
			require.Equal(t, tc.expNested2Str, tc.val.Nested2.Str)
			require.ElementsMatch(t, tc.expNestedArgs, tc.val.Nested1.Args)
			require.ElementsMatch(t, tc.expNestedArgs, tc.val.Nested2.Args)
		})
	}
}

func TestInvalidFlagPrefixFieldType(t *testing.T) {
	type invalidStruct struct {
		A string `flagPrefix:"a"`
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := invalidStruct{}
	require.Error(t, fls.StructVar(&structVal))
}

func TestInvalidNestedStruct(t *testing.T) {
	type nestedStruct2 struct {
		C string `flagArgs:"a"`
	}
	type nestedStruct struct {
		B nestedStruct2 `flagPrefix:"b-"`
	}
	type parentStruct struct {
		A nestedStruct `flagPrefix:"a-"`
	}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	structVal := parentStruct{}
	require.ErrorContains(t, fls.StructVar(&structVal), "A.B.C")
}

func TestSameStructWithDifferentPrefixes(t *testing.T) {
	type simpleStruct struct {
		X string `flag:"x"`
	}
	structVal := simpleStruct{}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	require.NoError(t, fls.StructVarWithPrefix(&structVal, "a-"))
	require.NoError(t, fls.StructVarWithPrefix(&structVal, "b-"))
	require.NoError(t, fls.Parse([]string{"--a-x", "val1", "--b-x", "val2"}))
	// the same semantics as in std flag package
	require.Equal(t, "val2", structVal.X)
}

func TestEmptyPrefix(t *testing.T) {
	type nestedStruct struct {
		Y string `flag:"y"`
	}
	type simpleStruct struct {
		N nestedStruct `flagPrefix:""`
		X string       `flag:"x"`
	}
	structVal := simpleStruct{}
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	fls := Wrap(flagSet)
	require.NoError(t, fls.StructVarWithPrefix(&structVal, "p-"))
	require.NoError(t, fls.Parse([]string{"--p-x", "val1", "--p-y", "val2"}))
	// the same semantics as in std flag package
	require.Equal(t, "val1", structVal.X)
	require.Equal(t, "val2", structVal.N.Y)
}

func TestEmbedStructNoTags(t *testing.T) {
	type embeddedStruct struct {
		Y string `flag:"y"`
	}
	type simpleStruct struct {
		embeddedStruct
		X string `flag:"x"`
	}
	structVal := simpleStruct{}
	fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
	require.NoError(t, fls.StructVarWithPrefix(&structVal, "p-"))
	require.Error(t, fls.Parse([]string{"--p-x", "val1", "--p-y", "val2"}))
}

func TestEmbedStructWithPrefix(t *testing.T) {
	type embeddedStruct struct {
		Y string `flag:"y"`
	}
	type simpleStruct struct {
		embeddedStruct `flagPrefix:"e-"`
		X              string `flag:"x"`
	}
	structVal := simpleStruct{}
	fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
	require.NoError(t, fls.StructVarWithPrefix(&structVal, "p-"))
	require.NoError(t, fls.Parse([]string{"--p-x", "val1", "--p-e-y", "val2"}))
	// the same semantics as in std flag package
	require.Equal(t, "val1", structVal.X)
	require.Equal(t, "val2", structVal.Y)
}

func TestEmbedStructWithEmptyPrefix(t *testing.T) {
	type embeddedStruct struct {
		Y string `flag:"y"`
	}
	type simpleStruct struct {
		embeddedStruct `flagPrefix:""`
		X              string `flag:"x"`
	}
	structVal := simpleStruct{}
	fls := Wrap(flag.NewFlagSet("", flag.ContinueOnError))
	require.NoError(t, fls.StructVarWithPrefix(&structVal, "p-"))
	require.NoError(t, fls.Parse([]string{"--p-x", "val1", "--p-y", "val2"}))
	// the same semantics as in std flag package
	require.Equal(t, "val1", structVal.X)
	require.Equal(t, "val2", structVal.Y)
}

func TestMultiNamesFlagSetUsage(t *testing.T) {
	t.Parallel()

	type nestedStruct struct {
		X string `flag:"x" flagUsage:"usage_x"`
		Y string `flags:"y,yy" flagUsage:"usage_y"`
	}
	type simpleStruct struct {
		S       string       `flag:"s" flagUsage:"usage_s"`
		Nested2 nestedStruct `flagPrefix:"n2-" flagUsagePrefix:"usage_n2_pr: "`
		Sp      *string      `flag:"sp" flagUsage:"usage_sp"`
		Nested  nestedStruct `flagPrefix:"n-"`
	}
	structVal := simpleStruct{
		S:  "s_default",
		Sp: ptr("sp_default"),
	}
	fls := NewFlagSet("fls_name", flag.ContinueOnError)

	require.NoError(t, fls.StructVar(&structVal))

	expectedUsage := `Usage of fls_name:
  -n-x string
    	usage_x
  -n-y -n-yy string
    	usage_y
  -n2-x string
    	usage_n2_pr: usage_x
  -n2-y -n2-yy string
    	usage_n2_pr: usage_y
  -s string
    	usage_s (default "s_default")
  -sp string
    	usage_sp
`
	require.Equal(
		t,
		expectedUsage,
		captureOutput(fls, fls.Usage),
	)

	var parseErr error
	expectedOutput := "flag provided but not defined: -wrong\n" + expectedUsage
	require.Equal(
		t,
		expectedOutput,
		captureOutput(fls, func() {
			parseErr = fls.Parse([]string{"-n-x", "some", "--wrong"})
		}),
	)
	require.Error(t, parseErr)
}
