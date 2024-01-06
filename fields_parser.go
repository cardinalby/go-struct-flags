package flago

import (
	"fmt"
	"reflect"
)

// fieldInfo contains info about a struct field that should be handled by FlagSet
type fieldInfo struct {
	fieldName     string
	namedFlagRole *namedFlagRole
	isFlagArgs    bool
	fieldValue    reflect.Value
}

// collectFieldsInfoRecursive collects info about all fields of the given struct including nested
// structs. It validates the types of the fields and their tags and returns an error if any of them
// is invalid.
func collectFieldsInfoRecursive(
	structValue reflect.Value,
	parentFlagPrefix string,
	parentUsagePrefix string,
	parentFieldName string,
) (res []fieldInfo, err error) {
	sValType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		field := sValType.Field(i)
		fieldName := getFieldName(parentFieldName, field.Name)
		fieldRole, err := getFieldRole(field)
		if err != nil {
			return nil, fmt.Errorf(`field "%s": %w`, fieldName, err)
		}
		if fieldRole == nil {
			continue
		}
		if fieldInfo, err := collectFieldInfo(
			field,
			structValue.Field(i),
			fieldName,
			parentFlagPrefix,
			parentUsagePrefix,
			fieldRole,
		); err != nil {
			return nil, err
		} else {
			res = append(res, fieldInfo...)
		}
	}
	return res, nil
}

func collectFieldInfo(
	field reflect.StructField,
	fieldValue reflect.Value,
	fieldName string,
	parentFlagPrefix string,
	parentUsagePrefix string,
	fieldRole fieldRole,
) (res []fieldInfo, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf(`field "%s" tagged with "%s": %w`, fieldName, fieldRole.getRoleTagName(), err)
		}
	}()

	fieldType := field.Type
	switch role := fieldRole.(type) {
	case nestedStructRole:
		if err := checkPrefixedFieldType(fieldType); err != nil {
			return nil, err
		}
		if nestedRes, err := collectFieldsInfoRecursive(
			fieldValue,
			parentFlagPrefix+role.flagPrefix,
			parentUsagePrefix+role.usagePrefix,
			fieldName,
		); err != nil {
			return nil, err
		} else {
			res = append(res, nestedRes...)
		}
	case flagArgsRole:
		if err := checkFlagArgsFieldType(fieldType); err != nil {
			return nil, err
		}
		res = append(res, fieldInfo{
			fieldName:  fieldName,
			isFlagArgs: true,
			fieldValue: fieldValue,
		})
	case namedFlagRole:
		varRegister, err := getVarRegister(fieldValue)
		if err != nil {
			return nil, err
		}
		role = role.withPrefixes(parentFlagPrefix, parentUsagePrefix)
		role.varRegister = varRegister
		res = append(res, fieldInfo{
			fieldName:     fieldName,
			namedFlagRole: &role,
			fieldValue:    fieldValue,
		})
	}
	return res, nil
}

func checkPrefixedFieldType(fieldType reflect.Type) error {
	if fieldType.Kind() != reflect.Struct {
		return fmt.Errorf("struct expected, got %s", fieldType.Name())
	}
	return nil
}

func checkFlagArgsFieldType(fieldType reflect.Type) error {
	if fieldType.Kind() != reflect.Slice || fieldType.Elem().Kind() != reflect.String {
		return fmt.Errorf("[]string expected, got %s", fieldType.Name())
	}
	return nil
}

func getFieldName(parentFieldName, fieldName string) string {
	if parentFieldName == "" {
		return fieldName
	}
	return fmt.Sprintf("%s.%s", parentFieldName, fieldName)
}
