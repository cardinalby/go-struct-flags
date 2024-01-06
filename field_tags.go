package flago

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	flagNameTag     = "flag"
	flagNamesTag    = "flags"
	flagArgsTag     = "flagArgs"
	flagUsageTag    = "flagUsage"
	flagUsagePrefix = "flagUsagePrefix"
	flagPrefixTag   = "flagPrefix"
)

type fieldRole interface {
	getRoleTagName() string
}

type namedFlagRole struct {
	flagNames   []string
	varRegister varRegister
	usage       string
	roleTagName string
}

func (r namedFlagRole) getRoleTagName() string {
	return r.roleTagName
}

func (r namedFlagRole) withPrefixes(namePrefix string, usagePrefix string) namedFlagRole {
	if namePrefix != "" {
		for i, name := range r.flagNames {
			r.flagNames[i] = namePrefix + name
		}
	}
	r.usage = usagePrefix + r.usage
	return r
}

type flagArgsRole struct {
}

func (r flagArgsRole) getRoleTagName() string {
	return flagArgsTag
}

type nestedStructRole struct {
	flagPrefix  string
	usagePrefix string
}

func (r nestedStructRole) getRoleTagName() string {
	return flagPrefixTag
}

func getFieldRole(field reflect.StructField) (fieldRole, error) {
	var (
		flagName      string
		flagNames     []string
		flagArgs      bool
		flagPrefix    string
		hasFlagPrefix bool
		err           error
	)
	tags := field.Tag

	if flagName = tags.Get(flagNameTag); flagName == "-" {
		flagName = ""
	}
	flagNames = getFlagNames(tags)

	if flagArgs, err = getIsFlagArgs(tags); err != nil {
		return nil, err
	}

	flagPrefix, hasFlagPrefix = tags.Lookup(flagPrefixTag)

	hasFlagName := flagName != ""
	hasFlagNames := len(flagNames) > 0

	behaviorTagsCount := trueCount(
		hasFlagName,
		hasFlagNames,
		hasFlagPrefix,
		flagArgs,
	)
	if behaviorTagsCount == 0 {
		return nil, nil
	}
	if behaviorTagsCount > 1 {
		return nil, fmt.Errorf(
			`only one of "%s", "%s", "%s", "%s" tags can be used`,
			flagNameTag, flagNamesTag, flagArgsTag, flagPrefixTag,
		)
	}

	usage, hasUsage := tags.Lookup(flagUsageTag)
	usagePrefix, hasUsagePrefix := tags.Lookup(flagUsagePrefix)

	if hasUsagePrefix && !hasFlagPrefix {
		return nil, fmt.Errorf(`"%s" tag can be used only with "%s" tag`, flagUsagePrefix, flagPrefixTag)
	}

	if hasFlagName || hasFlagNames {
		role := namedFlagRole{
			usage: usage,
		}
		if hasFlagName {
			role.flagNames = []string{flagName}
			role.roleTagName = flagNameTag
		} else {
			role.flagNames = flagNames
			role.roleTagName = flagNamesTag
		}
		return role, nil
	}

	if hasUsage {
		return nil, fmt.Errorf(
			`"%s" tag can be used only with "%s" or "%s" tags`,
			flagUsageTag, flagNameTag, flagNamesTag,
		)
	}

	if hasFlagPrefix {
		return nestedStructRole{
			flagPrefix:  flagPrefix,
			usagePrefix: usagePrefix,
		}, nil
	}

	if flagArgs {
		return flagArgsRole{}, nil
	}

	// should never happen
	return nil, nil
}

func getIsFlagArgs(tags reflect.StructTag) (isFlagArgs bool, err error) {
	if flagArgs := tags.Get(flagArgsTag); flagArgs != "" {
		if isFlagArgs, err = strconv.ParseBool(flagArgs); err != nil {
			return false,
				fmt.Errorf(`invalid "%s" tag bool value: "%s"`, flagArgsTag, flagArgs)
		}
	}
	return isFlagArgs, nil
}

func getFlagNames(tags reflect.StructTag) []string {
	namesStr := tags.Get(flagNamesTag)
	if namesStr == "" {
		return nil
	}
	names := strings.Split(namesStr, ",")
	for i := range names {
		if trimmed := strings.TrimSpace(names[i]); trimmed != "" {
			names[i] = trimmed
		}
	}
	return names
}

func trueCount(values ...bool) (res int) {
	for _, v := range values {
		if v {
			res++
		}
	}
	return res
}
