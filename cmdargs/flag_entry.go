package cmdargs

import (
	"strconv"
	"strings"
)

type FlagEntry struct {
	name           string
	value          string
	isInline       bool
	isDoubleDashed bool
	isBool         bool
}

func NewBoolFlagEntry(flagName, flagValue string) FlagEntry {
	return FlagEntry{
		name:     flagName,
		value:    flagValue,
		isInline: flagValue != "",
		isBool:   true,
	}
}

func NewFlagEntry(flagName, flagValue string) FlagEntry {
	return FlagEntry{
		name:  flagName,
		value: flagValue,
	}
}

func (f FlagEntry) TokenStrings() []string {
	var name string
	if f.isDoubleDashed {
		name = "--" + f.name
	} else {
		name = "-" + f.name
	}

	switch {
	case f.isInline:
		return []string{name + "=" + f.value}
	case f.isBool: // !inline
		return []string{name}
	default:
		return []string{name, f.value}
	}
}

func (f FlagEntry) TokensCount() int {
	if f.isInline || f.isBool {
		return 1
	}
	return 2
}

func (f FlagEntry) Kind() EntryKind {
	return EntryKindFlag
}

func (f FlagEntry) String() string {
	return strings.Join(f.TokenStrings(), " ")
}

func (f FlagEntry) Name() string {
	return f.name
}

func (f FlagEntry) Value() string {
	return f.value
}

func (f FlagEntry) IsInline() bool {
	return f.isInline
}

func (f FlagEntry) IsBool() bool {
	return f.isBool
}

func (f FlagEntry) IsDoubleDashed() bool {
	return f.isDoubleDashed
}

func (f FlagEntry) Equals(other FlagEntry) bool {
	return f.name == other.name &&
		f.value == other.value &&
		f.isBool == other.isBool
}

func (f FlagEntry) WithName(name string) FlagEntry {
	f.name = name
	return f
}

func (f FlagEntry) WithValue(value string) FlagEntry {
	if f.value == value {
		return f
	}
	f.value = value
	if f.isBool {
		f.isInline = value != ""
		if f.isInline {
			if _, err := strconv.ParseBool(value); err != nil {
				f.isBool = false
			}
		}
	}
	return f
}

// WithNoValue makes the flag a bool flag with no value (equals true)
func (f FlagEntry) WithNoValue() FlagEntry {
	f.isBool = true
	f.isInline = false
	f.value = ""
	return f
}

// WithInline makes the flag inline or not inline
// If it was a bool flag and `isInline = false`, it becomes a non-bool flag with string value. In this case,
// if bool flag didn't have inline value, it will have empty string non-inline value
// If it was a bool flag with no value and `isInline = true`, explicit value "true" will be used
func (f FlagEntry) WithInline(isInline bool) FlagEntry {
	if f.isInline == isInline {
		return f
	}
	if isInline {
		if f.isBool { // not inline bool with implicit "true". Make it explicit
			f.value = "true"
		}
	} else {
		f.isBool = false
	}
	f.isInline = isInline
	return f
}

func (f FlagEntry) WithDoubleDashes(isDoubleDashed bool) FlagEntry {
	f.isDoubleDashed = isDoubleDashed
	return f
}
