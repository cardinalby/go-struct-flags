[![test](https://github.com/cardinalby/go-struct-flags/actions/workflows/test.yml/badge.svg)](https://github.com/cardinalby/go-struct-flags/actions/workflows/test.yml)
[![list](https://github.com/cardinalby/go-struct-flags/actions/workflows/list.yml/badge.svg)](https://github.com/cardinalby/go-struct-flags/actions/workflows/list.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cardinalby/go-struct-flags.svg)](https://pkg.go.dev/github.com/cardinalby/go-struct-flags)

# Parse command line flags into a struct with tags

The library is a wrapper around the standard library's `flag` package providing a way 
to parse command line flags into a struct fields

## The main features are:
- Ignore unknown flags
- Distinguish **not passed** flags from **default** values that is difficult with the standard `flag` package
- More convenient way to define flags using **struct tags**
- Using nested structs for similar flag groups

## Example

```shell
go get github.com/cardinalby/go-struct-flags
```

### Define your flags struct with tags:
```go
import (
    flago "github.com/cardinalby/go-struct-flags"
)

type MyFlags struct {
    // Verbose can be filled both by "-v" and "-verbose" flags
    Verbose   int      `flags:"verbose,v" flagUsage:"verbose mode: 1,2,3"`
	
    // Login is optional, it will be set only if passed
    Login     *string  `flag:"login" flagUsage:"user login"`
	
    // FileNames will contain flag.Args() after Parse()
    FileNames []string `flagArgs:"true"`
}

// This is the way to set defaults:
myFlags := MyFlags{}
```

### Create new FlagSet and register the struct:
```go
flagSet := flago.NewFlagSet("myApp", flag.ExitOnError)
if err := flagSet.StructVar(&myFlags); err != nil {
    // error can happen if struct tags are invalid
    panic(err)
}
```
### Parse command line flags:
```go
// Use short form of "-v"
// Normally you would use os.Args[1:] instead of hardcoded values
if err := flagSet.Parse([]string{"-v", "2", "--login", "user1", "file1", "file2"}); err != nil {
    // Parse has default behavior
    panic(err)
}

// myFlags.Verbose == 2
// *myFlags.Login == "user1"
// myFlags.FileNames == []string{"file1", "file2"}
```

### Check if flag is passed
Using pointer fields enables you to distinguish **not passed** flags from **default** values:

```go
// Use long form of "--verbose" and don't pass "--login"
if err := flagSet.Parse([]string{"--verbose", "2"}); err != nil {
    // Parse has default behavior
    panic(err)
}

// myFlags.Verbose == 2
// myFlags.Login == nil
// len(myFlags.FileNames) == 0
```

### Set defaults
Unlike std `flag` package the library doesn't provide a way to explicitly set "default" values.

You can do it in native way just assigning them to struct fields before parsing (both for pointer and non-pointer fields)
```go
defaultLogin := "admin"
myFlags := MyFlags{
    Verbose: 1,
    Login: &defaultLogin,
}

// No flags are passed
if err := flagSet.Parse("file1", "file2"); err != nil {
    // Parse has default behavior
    panic(err)
}

// myFlags.Verbose == 1       <--- hasn't been changed
// *myFlags.Login == "admin"  <--- hasn't been changed
// myFlags.FileNames == []string{"file1", "file2"}
```

### Using nested structs

You can create logically similar flag groups assigning them a prefix using nested structs:
```go
type PersonFlags struct {
    Name  string `flag:"name" flagUsage:"person name"`
    Email string `flag:"email" flagUsage:"person email"`
}
type MyParentFlags struct {
    Sender   PersonFlags `flagPrefix:"sender-" flagUsagePrefix:"sender "`
    Receiver PersonFlags `flagPrefix:"receiver-" flagUsagePrefix:"receiver "`
}

flagSet := NewFlagSet("myApp", flag.ExitOnError)
myFlags := MyParentFlags{}
if err := flagSet.StructVar(&myFlags); err != nil {
    panic(err)
}

if err := flagSet.Parse([]string{
    "--sender-name", "John",
    "--sender-email", "j@email.com",
    "--receiver-name", "Dave",
    "--receiver-email", "d@email.com",
}); err != nil {
    panic(err)
}

// myFlags.Sender.Name == "John"
// myFlags.Sender.Email == "j@email.com"
// myFlags.Receiver.Name == "Dave"
// myFlags.Receiver.Email == "d@email.com"
```

See tests for more examples.

## Constructing
The library provides two constructors for `flago.FlagSet`: 
- `Wrap(*flag.FlagSet)` that wraps the existing std `flag.FlagSet` instance that can have some flags already registered
or be used to register new flags using its methods.
- `NewFlagSet()` creates new `flag.FlagSet` instance and sets its `Usage` to [`flago.DefaultUsage`](#usage-help-message)
- Functions with the names matching `flago.FlagSet` method names for a default `flago.CommandLine` 
FlagSet instance are available in the same manner as in the standard `flag` package.

## Configure `Parse()` behavior

### 🔹 Ignore unknown flags
`SetIgnoreUnknown(true)` method call will make `Parse()` ignore unknown flags instead of returning an error.

To retrieve unknown flags that have been ignored, call `GetIgnoredArgs()` after `Parse()`.

### 🔹 Allow parsing multiple aliases
`SetAllowParsingMultipleAliases(true)` method call will make `Parse()` not return an error if multiple aliases
of the same field are passed. The last passed value will be used.

# Supported struct tags
To parse flags and args to struct fields you should use `StructVar()` or `StructVarWithPrefix()` methods.

The methods accept optional arguments of pointers to ignored fields. These fields will not be registered as flags.

The struct fields that don't have any "flag" tags will be ignored. 

## Define named flag(s) for a field

### 🔻 `flag="name"`
Defines flag name for a field. Unlike `json` package, if the name is not set, the field will be ignored.

- Field should have type supported by `flag` package or be pointer to such type.
- If the field is a **pointer**, it will be set only if the flag is passed to `Parse()`.
- If it's not a pointer and is the correspondent flag is not passed, its default value will remain.

### 🔻 `flags="name1,name2"`

Same as `flag` but defines multiple comma-separated flag names (aliases) for a field.

- You should use either `flag` or `flags` tag for a field, not both.
- By default, `Parse()` will return an **error** if **multiple aliases** of the same field are passed. You can 
change this behavior using `SetAllowParsingMultipleAliases(true)`

### 🔸 `flagUsage="message"`

Defines flag usage message for a field. Should be used only for fields with `flag` or `flags` tag.

## Assign remaining args

### 🔻 `flagArgs="true"`

Field will be filled with `flag.Args()` (remaining command line args after named flags) after `Parse()`. 

- Other "flag" tags should not be used for such fields.

## Describe nested structs

The library parses fields in **nested structs** if explicitly instructed with `flagPrefix` tag on a
field containing another struct.

### 🔻 `flagPrefix="pref"`

Instructs the library to parse fields in nested struct.

- All resulting flag names (specified by `flag` and `flags` tags) in the nested struct will have the specified prefix.
- With `flagPrefix=""` nested struct will still be parsed but without using prefix for its fields.
- The prefix does not affect fields tagged with `flagArgs`.

### 🔸 `flagUsagePrefix="usage_pref"`

This tag is used only for nested struct fields with `flagPrefix` tag.

Instructs the library to use the specified prefix for flag usage messages of fields in nested struct.

### Usage help message

If you use `flago.NewFlagSet()` constructor, resulting FlagSet will assign own default implementation
of `Usage` that prints help message in the same manner as the standard `flag` package but
**grouping aliases** assigned to a field using `flags` tag in one line.

If you use `flago.Wrap()` constructor, it doesn't override default `Usage` of the standard `FlagSet`.
You can do it manually: `flagSet.Usage = flago.DefaultUsage`

### Field types support

- `StructVar()` method parses fields and their tags and calls the correspondent `FlagSet.***Var()` methods
depending on the field type. 
- So fields should have types supported by `flag` package or be pointers to such types.
- Fields  implementing `flag.Value` and `func(string) error` fields are also supported (but can't be pointers).

#### Special case
If a field has `encoding.TextUnmarshaler` interface, it also should implement `encoding.TextMarshaler`.

The library will call `FlagSet.TextVar()` on such fields that requires a default "marshaler" value.







