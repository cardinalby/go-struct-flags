package flago

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func getExistingFlagNames(flagSet *flag.FlagSet) map[string]struct{} {
	flags := make(map[string]struct{})
	flagSet.Visit(func(f *flag.Flag) {
		flags[f.Name] = struct{}{}
	})
	return flags
}

type boolFlag interface {
	IsBoolFlag() bool
}

// getFormalFlagNames returns a map where key is a flag name and value indicates it's a bool flag
func getFormalFlagNames(flagSet *flag.FlagSet) map[string]bool {
	flags := make(map[string]bool)
	flagSet.VisitAll(func(f *flag.Flag) {
		isBoolFlag := false
		if boolFlag, ok := f.Value.(boolFlag); ok {
			isBoolFlag = boolFlag.IsBoolFlag()
		}
		flags[f.Name] = isBoolFlag
	})
	return flags
}

type partialVarRegister func(flagSet *flag.FlagSet, name, usage string)

// postParseClb is a callback that should be called after Parse() if flag is present
type postParseClb func()

// varRegister is a callback that registers a field as a flag in the given FlagSet
type varRegister func(flagSet *flag.FlagSet, name, usage string) postParseClb

func getVarRegister(fieldValue reflect.Value) (varRegister, error) {
	valueType := fieldValue.Type()

	if valueType.Kind() == reflect.Ptr {
		valueToParsePtr := reflect.New(valueType.Elem())
		primitiveVarRegister := getPrimitiveVarRegister(valueToParsePtr.Elem(), reflect.Zero(valueType))
		if primitiveVarRegister != nil {
			return func(flagSet *flag.FlagSet, name, usage string) postParseClb {
				primitiveVarRegister(flagSet, name, usage)
				return func() {
					fieldValue.Set(valueToParsePtr)
				}
			}, nil
		}
	}

	primitiveVarRegister := getPrimitiveVarRegister(fieldValue, fieldValue)
	if primitiveVarRegister != nil {
		return func(flagSet *flag.FlagSet, name, usage string) postParseClb {
			primitiveVarRegister(flagSet, name, usage)
			return nil
		}, nil
	}

	if flagValue, isFlagValue := fieldValue.Interface().(flag.Value); isFlagValue {
		return func(flagSet *flag.FlagSet, name, usage string) postParseClb {
			flagSet.Var(flagValue, name, usage)
			return nil
		}, nil
	}

	if textUnmarshaler, ok := fieldValue.Interface().(encoding.TextUnmarshaler); ok {
		if fieldValue.IsNil() {
			return nil, errors.New("implements encoding.TextUnmarshaler but is nil")
		}
		textMarshaler, isTextMarshaler := fieldValue.Interface().(encoding.TextMarshaler)
		if !isTextMarshaler {
			return nil, errors.New("implements encoding.TextUnmarshaler but not encoding.TextMarshaler")
		}
		return func(flagSet *flag.FlagSet, name, usage string) postParseClb {
			flagSet.TextVar(textUnmarshaler, name, textMarshaler, usage)
			return nil
		}, nil
	}

	if fnc, ok := fieldValue.Interface().(func(string) error); ok {
		if fnc == nil {
			return nil, errors.New("func is nil")
		}
		return func(flagSet *flag.FlagSet, name, usage string) postParseClb {
			flagSet.Func(name, usage, fnc)
			return nil
		}, nil
	}

	return nil, fmt.Errorf("unsupported field type %s", valueType.Name())
}

func getPrimitiveVarRegister(
	value reflect.Value,
	defaultValue reflect.Value,
) partialVarRegister {
	valueType := value.Type()
	valuePtr := value.Addr().UnsafePointer()

	switch valueType.Kind() {
	case reflect.Int:
		return func(flagSet *flag.FlagSet, name, usage string) {
			defVal, _ := defaultValue.Interface().(int)
			flagSet.IntVar((*int)(valuePtr), name, defVal, usage)
		}
	case reflect.Uint:
		return func(flagSet *flag.FlagSet, name, usage string) {
			defVal, _ := defaultValue.Interface().(uint)
			flagSet.UintVar((*uint)(valuePtr), name, defVal, usage)
		}
	case reflect.Int64:
		if valueType == reflect.TypeOf(time.Duration(0)) {
			return func(flagSet *flag.FlagSet, name, usage string) {
				defVal, _ := defaultValue.Interface().(time.Duration)
				flagSet.DurationVar((*time.Duration)(valuePtr), name, defVal, usage)
			}
		} else {
			return func(flagSet *flag.FlagSet, name, usage string) {
				defVal, _ := defaultValue.Interface().(int64)
				flagSet.Int64Var((*int64)(valuePtr), name, defVal, usage)
			}
		}
	case reflect.Uint64:
		return func(flagSet *flag.FlagSet, name, usage string) {
			defVal, _ := defaultValue.Interface().(uint64)
			flagSet.Uint64Var((*uint64)(valuePtr), name, defVal, usage)
		}
	case reflect.Float64:
		return func(flagSet *flag.FlagSet, name, usage string) {
			defVal, _ := defaultValue.Interface().(float64)
			flagSet.Float64Var((*float64)(valuePtr), name, defVal, usage)
		}
	case reflect.String:
		return func(flagSet *flag.FlagSet, name, usage string) {
			defVal, _ := defaultValue.Interface().(string)
			flagSet.StringVar((*string)(valuePtr), name, defVal, usage)
		}
	case reflect.Bool:
		return func(flagSet *flag.FlagSet, name, usage string) {
			defVal, _ := defaultValue.Interface().(bool)
			flagSet.BoolVar((*bool)(valuePtr), name, defVal, usage)
		}
	default:
		return nil
	}
}

func parseArg(arg string) (isFlag bool, isTerminator bool, flagName string, hasInlineValue bool) {
	if len(arg) < 2 || arg[0] != '-' {
		return false, false, "", false
	}
	numMinuses := 1
	if arg[1] == '-' {
		numMinuses++
		if len(arg) == 2 { // "--" terminates the flags
			return false, true, "", false
		}
	}
	flagName = arg[numMinuses:]

	if equalsSignIndex := strings.Index(flagName, "="); equalsSignIndex == 0 {
		// std FlagSet.Parse() will return "bad flag syntax" error
		return false, false, "", false
	} else if equalsSignIndex > 0 {
		flagName = flagName[:equalsSignIndex]
		hasInlineValue = true
	}
	return true, false, flagName, hasInlineValue
}

func StripUnknownFlags(flagSet *flag.FlagSet, args []string) (res, stripped []string) {
	formalFlagNames := getFormalFlagNames(flagSet)

	res = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		isFlag, isTerminator, flagName, hasInlineValue := parseArg(arg)
		if isTerminator {
			res = append(res, args[i:]...)
			break
		}
		if !isFlag {
			res = append(res, arg)
			continue
		}
		isBoolFlag, exists := formalFlagNames[flagName]
		var appendTo *[]string
		if exists {
			appendTo = &res
		} else {
			appendTo = &stripped
		}

		*appendTo = append(*appendTo, arg)
		if !hasInlineValue && !isBoolFlag {
			// next arg is supposed to be the flag value
			if i+1 < len(args) {
				*appendTo = append(*appendTo, args[i+1])
			}
			i++ // skip the flag value
		}
	}
	return res, stripped
}
