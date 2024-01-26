package cmdargs

import "strings"

type UnnamedArgsEntry []string

func (ua UnnamedArgsEntry) String() string {
	return strings.Join(ua, " ")
}

func (ua UnnamedArgsEntry) TokenStrings() []string {
	return ua
}

func (ua UnnamedArgsEntry) TokensCount() int {
	return len(ua)
}

func (ua UnnamedArgsEntry) Kind() EntryKind {
	return EntryKindUnnamedArgs
}
