package flagtags

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/urfave/cli/v2"
)

// Use errors.Is(err, flagtags.ErrX) to check error values
var (
	ErrNilValue      = errors.New("nil value")
	ErrMustBePtr     = errors.New("argument must be a pointer to a struct")
	ErrInvalidStruct = errors.New("provided value was not a struct")
	ErrPrivateField  = errors.New("private field")
	ErrNotSupported  = errors.New("unsupported type")
)

// MustParseFlags parses flag tags from the provided struct.
// If ParseFlags returns an error, this panics.
// See ParseFlags for more info.
func MustParseFlags(s interface{}) []cli.Flag {
	flags, err := ParseFlags(s)
	if err != nil {
		panic(err)
	}
	return flags
}

// ParseFlags parses flags from the struct fields and their tags.
//
// Supported field types are "int", "string", "bool".
// Supported tags are "name", "env", "value", "usage".
//
// By default (without tags),
// value is the default value of the primitive type,
// usage is empty,
// env is the field name in SCREAMING_SNAKE_CASE,
// name is the field name in kebab-case.
//
func ParseFlags(s interface{}) ([]cli.Flag, error) {
	// Return error if reference is nil
	if s == nil {
		return nil, ErrNilValue
	}

	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return nil, ErrMustBePtr
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return nil, ErrInvalidStruct
	}

	t := reflect.TypeOf(s).Elem()

	flags := make([]cli.Flag, v.NumField())
	var err error
	for i := 0; i < v.NumField(); i++ {
		flags[i], err = parseFlag(t.Field(i), v.Field(i))
		if err != nil {
			return nil, err
		}
	}

	return flags, nil
}

func parseFlag(t reflect.StructField, v reflect.Value) (cli.Flag, error) {
	var name string
	name, ok := t.Tag.Lookup("name")
	// If not set, infer from struct field name
	if !ok {
		name = ToKebabCase(t.Name)
	}

	// If not set, infer from struct field name
	var env string
	env, ok = t.Tag.Lookup("env")
	if !ok {
		env = ToScreamingSnakeCase(t.Name)
	}

	if !v.CanSet() {
		return nil, fmt.Errorf("%w: field '%v' must be made public", ErrPrivateField, t.Name)
	}

	strValue, _ := t.Tag.Lookup("value")
	usage, _ := t.Tag.Lookup("usage")
	iface := v.Addr().Interface()

	switch v.Kind() {
	case reflect.String:
		dst, ok := iface.(*string)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		return stringFlag(name, strValue, dst, env, usage), nil
	case reflect.Int:
		dst, ok := iface.(*int)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		// Parse value
		if len(strValue) == 0 {
			return intFlag(name, 0, dst, env, usage), nil
		}
		i, err := strconv.Atoi(strValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided value '%v' as an int, err: %w", strValue, err)
		}
		return intFlag(name, i, dst, env, usage), nil
	case reflect.Bool:
		dst, ok := iface.(*bool)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		// Parse value
		if len(strValue) == 0 {
			return boolFlag(name, false, dst, env, usage), nil
		}
		b, err := strconv.ParseBool(strValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided value '%v' as an int, err: %w", strValue, err)
		}
		return boolFlag(name, b, dst, env, usage), nil
	default:
		return nil, fmt.Errorf("%w: type '%v' is not supported yet", ErrNotSupported, v.Kind())
	}
}

func stringFlag(name string, value string, destination *string, envVar string, usage string) *cli.StringFlag {
	return &cli.StringFlag{Name: name, Value: value, EnvVars: []string{envVar}, Destination: destination, Usage: usage}
}

func boolFlag(name string, value bool, destination *bool, envVar string, usage string) *cli.BoolFlag {
	return &cli.BoolFlag{Name: name, Value: value, EnvVars: []string{envVar}, Destination: destination, Usage: usage}
}

func intFlag(name string, value int, destination *int, envVar string, usage string) *cli.IntFlag {
	return &cli.IntFlag{Name: name, Value: value, EnvVars: []string{envVar}, Destination: destination, Usage: usage}
}
