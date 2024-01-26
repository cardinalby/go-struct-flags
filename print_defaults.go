package flago

import (
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
)

func printUsageTitle(flagSet *flag.FlagSet, name string) {
	if name == "" {
		_, _ = fmt.Fprintf(flagSet.Output(), "Usage:\n")
	} else {
		_, _ = fmt.Fprintf(flagSet.Output(), "Usage of %s:\n", name)
	}
}

// DefaultUsage prints the default FlagSet usage to flagSet.Output grouping alternative flag names
func DefaultUsage(flagSet *FlagSet) {
	printUsageTitle(flagSet.FlagSet, flagSet.Name())
	PrintFlagSetDefaults(flagSet)
}

type mutatorWriter struct {
	writer  io.Writer
	mutator func(string) string
}

func newFilteringWriter(writer io.Writer, filter func(string) string) *mutatorWriter {
	return &mutatorWriter{
		writer:  writer,
		mutator: filter,
	}
}

func (w *mutatorWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	if res := w.mutator(s); len(res) > 0 {
		return w.writer.Write([]byte(res))
	}
	return len(p), nil
}

type flagNames struct {
	f          *flag.Flag
	isRequired bool
	names      []string
}

// indexFormalFlagNames returns a map of flag names to flag names grouped by flag value
func indexFormalFlagNames(flagSet *FlagSet) map[string]*flagNames {
	namesByValue := make(map[flag.Value]*flagNames)
	flagSet.VisitAll(func(f *flag.Flag) {
		fNames, ok := namesByValue[f.Value]
		if !ok {
			fNames = &flagNames{
				f: f,
			}
			if _, isRequired := flagSet.requiredFlagNames[f.Name]; isRequired {
				fNames.isRequired = true
			}
			namesByValue[f.Value] = fNames
		}
		fNames.names = append(fNames.names, f.Name)
	})
	res := make(map[string]*flagNames)
	for _, fNames := range namesByValue {
		for _, name := range fNames.names {
			res[name] = fNames
		}
	}
	return res
}

var defaultsOutputItemNameSearchRegexp = regexp.MustCompile(`(?m)^\s*?-(.*?)\s`)
var defaultsOutputItemNameReplaceRegexp = regexp.MustCompile(`(?m)^(\s*?)(-.*?)(\s.*?)$`)

func getDefaultsOutputItemName(outputItem string) string {
	if matches := defaultsOutputItemNameSearchRegexp.FindStringSubmatch(outputItem); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func replaceDefaultsOutputItemName(outputItem, newName string) string {
	s := defaultsOutputItemNameReplaceRegexp.ReplaceAllString(
		outputItem,
		fmt.Sprintf("$1%s$3", newName),
	)
	return s
}

func addDefaultsRequiredMark(outputItem string) string {
	return strings.Replace(outputItem, "\t", "\t* ", 1)
}

// PrintFlagSetDefaults prints flag names and usage grouping alternative flag names
func PrintFlagSetDefaults(flagSet *FlagSet) {
	// The implementation of this method is dirty and relies on the internal implementation details
	// of the flag package. Otherwise, it would be difficult (or impossible if following the same API)
	// to provide own implementation since the flag package checks flag.Value types against unexported
	// types to output flag names and usage.
	// This approach assumes that standard FlagSet.PrintDefaults() implementation calls output.Write()
	// once for each flag in the fixed format.
	// The idea is to replace the fist occurrence of the flag name (corresponding some field) with all
	// alternative flag names for the field and skip the rest of the occurrences.
	indexedFlagNames := indexFormalFlagNames(flagSet)
	seenFlags := make(map[*flagNames]struct{})

	originalOutput := flagSet.Output()
	defer flagSet.SetOutput(originalOutput)

	flagSet.SetOutput(newFilteringWriter(originalOutput, func(s string) string {
		if name := getDefaultsOutputItemName(s); name != "" {
			if fNames, ok := indexedFlagNames[name]; ok {
				if _, seen := seenFlags[fNames]; !seen {
					seenFlags[fNames] = struct{}{}
					if len(fNames.names) > 1 {
						names := strings.Builder{}
						for i, name := range fNames.names {
							if i > 0 {
								names.WriteString(" ")
							}
							names.WriteString("-")
							names.WriteString(name)
						}
						s = replaceDefaultsOutputItemName(s, names.String())
					}
					if fNames.isRequired {
						s = addDefaultsRequiredMark(s)
					}
					return s
				}
				return "" // skip
			}
		}
		return s
	}))

	flagSet.FlagSet.PrintDefaults()
}
