package flago

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	flagNameTag     = "flag"
	flagRequiredTag = "flagRequired"
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
	isRequired  bool
	isIgnored   bool
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
		flagName        string
		flagNames       []string
		flagArgs        bool
		flagRequired    bool
		hasFlagRequired bool
		flagPrefix      string
		hasFlagPrefix   bool
		err             error
	)
	tags := field.Tag

	if flagName = tags.Get(flagNameTag); flagName == "-" {
		flagName = ""
	}
	flagNames = getFlagNames(tags)

	if flagArgs, _, err = getBoolTag(tags, flagArgsTag); err != nil {
		return nil, err
	}
	if flagRequired, hasFlagRequired, err = getBoolTag(tags, flagRequiredTag); err != nil {
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
			usage:      usage,
			isRequired: flagRequired,
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

	for tagName, hasTag := range map[string]bool{
		flagUsageTag:    hasUsage,
		flagRequiredTag: hasFlagRequired,
	} {
		if hasTag {
			return nil, fmt.Errorf(
				`"%s" tag can be used only with "%s" or "%s" tags`,
				tagName, flagNameTag, flagNamesTag,
			)
		}
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

func getBoolTag(tags reflect.StructTag, tagName string) (val bool, exists bool, err error) {
	var strVal string
	if strVal, exists = tags.Lookup(tagName); strVal != "" {
		if val, err = strconv.ParseBool(strVal); err != nil {
			return false, exists,
				fmt.Errorf(`invalid "%s" tag bool value: "%s"`, tagName, strVal)
		}
	}
	return val, exists, nil
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
