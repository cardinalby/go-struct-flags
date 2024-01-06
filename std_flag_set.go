package flago

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"time"
)

func getExistingFlagNames(flagSet *flag.FlagSet) map[string]struct{} {
	flags := make(map[string]struct{})
	flagSet.Visit(func(f *flag.Flag) {
		flags[f.Name] = struct{}{}
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
