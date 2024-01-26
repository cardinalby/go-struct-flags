package cmdargs

type EntryKind int

const (
	EntryKindFlag EntryKind = iota
	EntryKindTerminator
	EntryKindUnnamedArgs
)

type Entry interface {
	String() string
	TokenStrings() []string
	TokensCount() int
	Kind() EntryKind
}
