package flago

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"time"
)

type partialVarRegister func(flagSet *flag.FlagSet, name, usage string) (isZero bool)

// postParseClb is a callback that should be called after Parse() if flag is present
type postParseClb func()

// varRegister is a callback that registers a field as a flag in the given FlagSet
type varRegister func(flagSet *flag.FlagSet, name, usage string) (clb postParseClb, isZero bool)

func getVarRegister(fieldValue reflect.Value) (varRegister, error) {
	valueType := fieldValue.Type()

	if valueType.Kind() == reflect.Ptr {
		valueToParsePtr := reflect.New(valueType.Elem())
		primitiveVarRegister := getPrimitiveVarRegister(valueToParsePtr.Elem(), reflect.Zero(valueType))
		if primitiveVarRegister != nil {
			return func(flagSet *flag.FlagSet, name, usage string) (postParseClb, bool) {
				_ = primitiveVarRegister(flagSet, name, usage)
				return func() {
					fieldValue.Set(valueToParsePtr)
				}, fieldValue.IsNil()
			}, nil
		}
	}

	primitiveVarRegister := getPrimitiveVarRegister(fieldValue, fieldValue)
	if primitiveVarRegister != nil {
		return func(flagSet *flag.FlagSet, name, usage string) (postParseClb, bool) {
			return nil, primitiveVarRegister(flagSet, name, usage)
		}, nil
	}

	if flagValue, isFlagValue := fieldValue.Interface().(flag.Value); isFlagValue {
		return func(flagSet *flag.FlagSet, name, usage string) (postParseClb, bool) {
			flagSet.Var(flagValue, name, usage)
			return nil, true
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
		return func(flagSet *flag.FlagSet, name, usage string) (postParseClb, bool) {
			flagSet.TextVar(textUnmarshaler, name, textMarshaler, usage)
			return nil, true
		}, nil
	}

	if fnc, ok := fieldValue.Interface().(func(string) error); ok {
		if fnc == nil {
			return nil, errors.New("func is nil")
		}
		return func(flagSet *flag.FlagSet, name, usage string) (postParseClb, bool) {
			flagSet.Func(name, usage, fnc)
			return nil, true
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
		return func(flagSet *flag.FlagSet, name, usage string) bool {
			defVal, _ := defaultValue.Interface().(int)
			flagSet.IntVar((*int)(valuePtr), name, defVal, usage)
			return defVal == 0
		}
	case reflect.Uint:
		return func(flagSet *flag.FlagSet, name, usage string) bool {
			defVal, _ := defaultValue.Interface().(uint)
			flagSet.UintVar((*uint)(valuePtr), name, defVal, usage)
			return defVal == 0
		}
	case reflect.Int64:
		if valueType == reflect.TypeOf(time.Duration(0)) {
			return func(flagSet *flag.FlagSet, name, usage string) bool {
				defVal, _ := defaultValue.Interface().(time.Duration)
				flagSet.DurationVar((*time.Duration)(valuePtr), name, defVal, usage)
				return defVal == 0
			}
		} else {
			return func(flagSet *flag.FlagSet, name, usage string) bool {
				defVal, _ := defaultValue.Interface().(int64)
				flagSet.Int64Var((*int64)(valuePtr), name, defVal, usage)
				return defVal == 0
			}
		}
	case reflect.Uint64:
		return func(flagSet *flag.FlagSet, name, usage string) bool {
			defVal, _ := defaultValue.Interface().(uint64)
			flagSet.Uint64Var((*uint64)(valuePtr), name, defVal, usage)
			return defVal == 0
		}
	case reflect.Float64:
		return func(flagSet *flag.FlagSet, name, usage string) bool {
			defVal, _ := defaultValue.Interface().(float64)
			flagSet.Float64Var((*float64)(valuePtr), name, defVal, usage)
			return defVal == 0
		}
	case reflect.String:
		return func(flagSet *flag.FlagSet, name, usage string) bool {
			defVal, _ := defaultValue.Interface().(string)
			flagSet.StringVar((*string)(valuePtr), name, defVal, usage)
			return defVal == ""
		}
	case reflect.Bool:
		return func(flagSet *flag.FlagSet, name, usage string) bool {
			defVal, _ := defaultValue.Interface().(bool)
			flagSet.BoolVar((*bool)(valuePtr), name, defVal, usage)
			return !defVal
		}
	default:
		return nil
	}
}
