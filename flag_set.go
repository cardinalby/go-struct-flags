package flago

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/cardinalby/go-struct-flags/cmdargs"
	"github.com/cardinalby/go-struct-flags/stdutil"
)

var ErrFlagRedefined = errors.New("flag redefined")
var ErrIsRequired = errors.New("flag is required")
var ErrMultipleAliases = errors.New("multiple aliases for the same flag are used")

type registeredNamedFlagField struct {
	flagName     string
	postParseClb postParseClb
}

type registeredNamedFlagsField struct {
	fields     []registeredNamedFlagField
	isRequired bool
	isZero     bool
}

// structRegisteredFields contains instruction for finishing parsing of a struct
// after all flags are parsed.
type structRegisteredFields struct {
	// keys: field names, values: slice where each element corresponds to a registered flag (with different names)
	namedFlagFields map[string]registeredNamedFlagsField
	// keys: field names, values: field values that should be assigned with FlagSet.ArgStrings()
	flagArgsToSet map[string]reflect.Value
}

func newStructRegisteredFields() structRegisteredFields {
	return structRegisteredFields{
		namedFlagFields: make(map[string]registeredNamedFlagsField),
		flagArgsToSet:   make(map[string]reflect.Value),
	}
}

// FlagSet is a wrapper around *flag.FlagSet that allows to register structs parsing their fields as flags.
type FlagSet struct {
	*flag.FlagSet
	// registeredFields contains instructions for finishing parsing of the registered structs
	// key is a pointer to a struct
	registeredFields                  map[any]structRegisteredFields
	requiredFlagNames                 map[string]struct{}
	ignoreUnknown                     bool
	ignoreUnknownTreatAmbiguousAsBool bool
	flagsToIgnore                     stdutil.FormalTagNames
	allowParsingMultipleAliases       bool
	ignoredArgs                       []string
}

// Wrap creates a new FlagSet wrapping the given `stdFlagSet` and does not set stdFlagSet.Usage
func Wrap(stdFlagSet *flag.FlagSet) *FlagSet {
	return &FlagSet{
		FlagSet:           stdFlagSet,
		registeredFields:  make(map[any]structRegisteredFields),
		flagsToIgnore:     make(stdutil.FormalTagNames),
		requiredFlagNames: make(map[string]struct{}),
	}
}

// NewFlagSet creates a new FlagSet wrapping new flag.FlagSet with the given name and error handling policy and
// assigns its Usage to the own implementation that groups alternative flag names
func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	stdFlagSet := flag.NewFlagSet(name, errorHandling)
	wrapped := Wrap(stdFlagSet)
	stdFlagSet.Usage = wrapped.defaultUsage

	return wrapped
}

// SetAllowParsingMultipleAliases sets the behavior of Parse() when multiple tag names
// assigned to same field are passed.
// If `true`, it will be ignored and only the last value will be used.
// If `false`, Parse() will return an error.
// Default value is `false`.
func (fls *FlagSet) SetAllowParsingMultipleAliases(allow bool) {
	fls.allowParsingMultipleAliases = allow
}

// SetIgnoreUnknown sets the behavior of Parse() when unknown flags are passed.
// If `true`, they will be ignored.
// If `false`, Parse() will return an error.
// Default value is `false`.
func (fls *FlagSet) SetIgnoreUnknown(ignore bool) {
	fls.ignoreUnknown = ignore
}

// SetIgnoreUnknownAmbiguousAsBoolFlags specifies behaviour of stripping ambiguous unknown flags if
// SetIgnoreUnknown(true) is set.
// Unknown flag without inline value can be treated as bool flag or as flag name that should be followed by value.
// If the flag is unknown, we can't know if it's bool flag or not.
// Example: "-a -b" can be treated as:
// 1. {a: "-b"} - id - if treatAsBool == false
// OR
// 2. {a: true, b: true} - if treatAsBool == true
// Default values is `false`.
// It doesn't affect flags from fields that are defined in a registered struct but were passed as ignored to
// Parse() call. Information about their type will be used to determine if they are bool flags or not.
func (fls *FlagSet) SetIgnoreUnknownAmbiguousAsBoolFlags(treatAsBool bool) {
	fls.ignoreUnknownTreatAmbiguousAsBool = treatAsBool
}

// GetIgnoredArgs returns a slice of arguments that were ignored during the last call to Parse()
// because of SetIgnoreUnknown(true), nil otherwise
func (fls *FlagSet) GetIgnoredArgs() []string {
	return fls.ignoredArgs
}

// Parse parses the command-line flags calling Parse on the wrapped FlagSet
// and then sets values of the registered structs fields for flags that were actually parsed.
func (fls *FlagSet) Parse(arguments []string) error {
	if fls.FlagSet == nil {
		return errors.New("wrapped FlagSet is nil")
	}
	fls.ignoredArgs = nil
	if fls.ignoreUnknown {
		argsPassed, argsIgnored := cmdargs.NewArgs(arguments).
			WithFlagSet(fls.FlagSet).
			WithAmbiguousAsBool(fls.ignoreUnknownTreatAmbiguousAsBool).
			StripUnknownFlags(
				fls.flagsToIgnore,
			)
		arguments, fls.ignoredArgs = argsPassed.Args, argsIgnored.Args
	}
	if err := fls.FlagSet.Parse(arguments); err != nil {
		return err
	}
	if err := fls.postProcessRegisteredFields(); err != nil {
		// follow the same error handling policy as the wrapped FlagSet
		_, _ = fmt.Fprintln(fls.Output(), err.Error())
		fls.usage()

		switch fls.ErrorHandling() {
		case flag.ContinueOnError:
			return err
		case flag.ExitOnError:
			if errors.Is(err, flag.ErrHelp) {
				os.Exit(0)
			}
			os.Exit(2)
		case flag.PanicOnError:
			panic(err)
		}
		return err
	}
	return nil
}

// StructVar registers the fields of the given struct as a flags
// `ignoredFields` is a slice of pointers to fields that should be ignored and not registered as flags
func (fls *FlagSet) StructVar(p any, ignoredFields ...any) error {
	return fls.StructVarWithPrefix(p, "", ignoredFields...)
}

// StructVarWithPrefix registers the fields of the given struct as a flags
// with names prefixed with `flagsPrefix`
// `ignoredFields` is a slice of pointers to fields that should be ignored and not registered as flags
func (fls *FlagSet) StructVarWithPrefix(p any, flagsPrefix string, ignoredFields ...any) (err error) {
	defer func() {
		err = fls.recoverParsePanic(recover(), err)
	}()

	if fls.FlagSet == nil {
		return errors.New("wrapped FlagSet is nil")
	}
	structValue, err := getStructPointerElem(p)
	if err != nil {
		return err
	}
	ignoredFieldsMap, err := newIgnoredFieldsMap(ignoredFields)
	if err != nil {
		return fmt.Errorf("invalid ignoredFields: %w", err)
	}

	// collect fields info but don't register flags until all fields are validated
	fieldsInfo, err := collectFieldsInfoRecursive(
		structValue,
		flagsPrefix,
		"",
		"",
		ignoredFieldsMap,
		fls.flagsToIgnore,
	)
	if err != nil {
		return err
	}

	postParseActions := newStructRegisteredFields()
	for _, info := range fieldsInfo {
		if info.isFlagArgs {
			postParseActions.flagArgsToSet[info.fieldName] = info.fieldValue
		} else if info.namedFlagRole != nil {
			postParseActions.namedFlagFields[info.fieldName] = fls.registerNamedFlagField(info)
		}
	}
	// add registeredFields to the map only if all fields are valid and registered
	fls.registeredFields[p] = postParseActions

	return nil
}

// PrintDefaults prints the default FlagSet usage to wrapped FlagSet.Output grouping alternative flag names
func (fls *FlagSet) PrintDefaults() {
	PrintFlagSetDefaults(fls)
}

func (fls *FlagSet) postProcessRegisteredFields() error {
	existingFlagNames := stdutil.GetExistingFlagNames(fls.FlagSet)
	var errs []error

	for _, structFields := range fls.registeredFields {
		for _, namedFlagsField := range structFields.namedFlagFields {
			var fieldFirstFoundFlagName string
			isAnyFieldFlagFound := false
			for _, namedFlagField := range namedFlagsField.fields {
				if _, exists := existingFlagNames[namedFlagField.flagName]; !exists {
					continue
				}
				isAnyFieldFlagFound = true
				if !fls.allowParsingMultipleAliases {
					if fieldFirstFoundFlagName == "" {
						fieldFirstFoundFlagName = namedFlagField.flagName
					} else if fieldFirstFoundFlagName != namedFlagField.flagName {
						errs = append(errs, fmt.Errorf(
							`%w: "%s" and "%s"`,
							ErrMultipleAliases,
							fieldFirstFoundFlagName,
							namedFlagField.flagName,
						))
						continue
					}
				}
				if len(errs) == 0 && namedFlagField.postParseClb != nil {
					namedFlagField.postParseClb()
				}
			}
			if namedFlagsField.isRequired && !isAnyFieldFlagFound {
				names := make([]string, len(namedFlagsField.fields))
				for i, namedFlagField := range namedFlagsField.fields {
					names[i] = namedFlagField.flagName
				}
				errs = append(errs, fmt.Errorf(`%w: "%s"`, ErrIsRequired, strings.Join(names, `"/"`)))
			}
		}
		if len(errs) == 0 {
			for _, fieldValue := range structFields.flagArgsToSet {
				fieldValue.Set(reflect.ValueOf(fls.FlagSet.Args()))
			}
		}
	}
	if len(errs) > 0 {
		return joinErr(errs...)
	}
	return nil
}

func (fls *FlagSet) usage() {
	if fls.Usage == nil {
		printUsageTitle(fls.FlagSet, fls.FlagSet.Name())
		fls.FlagSet.PrintDefaults()
	} else {
		fls.Usage()
	}
}

func (fls *FlagSet) defaultUsage() {
	DefaultUsage(fls)
}

// sprintf formats the message, prints it to output, and returns it.
func (fls *FlagSet) sprintf(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintln(fls.Output(), msg)
	return msg
}

func (fls *FlagSet) registerNamedFlagField(info fieldInfo) (res registeredNamedFlagsField) {
	isZero := false
	for _, flagName := range info.namedFlagRole.flagNames {
		registeredNamedFlagField := registeredNamedFlagField{
			flagName: flagName,
		}
		registeredNamedFlagField.postParseClb, isZero = info.namedFlagRole.varRegister(
			fls.FlagSet, flagName, info.namedFlagRole.usage,
		)
		res.fields = append(res.fields, registeredNamedFlagField)
	}
	res.isRequired = info.namedFlagRole.isRequired && isZero
	if res.isRequired {
		for _, flagName := range info.namedFlagRole.flagNames {
			fls.requiredFlagNames[flagName] = struct{}{}
		}
	}
	return res
}

func (fls *FlagSet) recoverParsePanic(panicMsg any, existingErr error) error {
	if panicMsg == nil {
		return existingErr
	}
	var panicErr error
	panicStrMsg, isStr := panicMsg.(string)
	if isStr && strings.HasPrefix(panicStrMsg, "flag redefined") {
		panicErr = ErrFlagRedefined
	} else {
		panicErr = fmt.Errorf("%v", panicErr)
	}
	return joinErr(existingErr, panicErr)
}

func getStructPointerElem(p any) (res reflect.Value, err error) {
	val := reflect.ValueOf(p)
	if val.Kind() != reflect.Pointer {
		return reflect.Value{}, fmt.Errorf("expected pointer to struct, got %s", val.Type().Name())
	}
	res = reflect.ValueOf(p).Elem()
	if res.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("expected struct, got %v", val.Type().Name())
	}
	return res, nil
}

func newIgnoredFieldsMap(fields []any) (map[unsafe.Pointer]struct{}, error) {
	res := make(map[unsafe.Pointer]struct{})
	for i, field := range fields {
		val := reflect.ValueOf(field)
		if val.Kind() != reflect.Ptr {
			return nil, fmt.Errorf(`element %d: pointer expected, got %s`, i, val.Type().Name())
		}
		res[val.UnsafePointer()] = struct{}{}
	}
	return res, nil
}
