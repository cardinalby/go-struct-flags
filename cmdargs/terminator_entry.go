package cmdargs

type TerminatorEntry struct{}

func NewTerminatorEntry() Entry {
	return TerminatorEntry{}
}

func (tt TerminatorEntry) String() string {
	return "--"
}

func (tt TerminatorEntry) TokenStrings() []string {
	return []string{tt.String()}
}

func (tt TerminatorEntry) TokensCount() int {
	return 1
}

func (tt TerminatorEntry) Kind() EntryKind {
	return EntryKindTerminator
}
