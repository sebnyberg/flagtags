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
// Supported field types are "int", "string", "bool", "float64".
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

	var err error
	flags := make([]cli.Flag, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		newFlags, fieldErr := flagsFromField(t.Field(i), v.Field(i))
		if fieldErr != nil && err == nil {
			err = fieldErr
		}
		flags = append(flags, newFlags...)
	}

	return flags, err
}

func flagsFromField(t reflect.StructField, v reflect.Value) ([]cli.Flag, error) {
	var name string
	name, ok := t.Tag.Lookup("name")
	// If not set, infer from struct field name
	if !ok {
		name = toKebabCase(t.Name)
	}

	// If not set, infer from struct field name
	var env string
	env, ok = t.Tag.Lookup("env")
	if !ok {
		env = toScreamingSnakeCase(t.Name)
	}

	if !v.CanSet() {
		return nil, fmt.Errorf("%w: field '%v' must be made public", ErrPrivateField, t.Name)
	}

	strValue, _ := t.Tag.Lookup("value")
	usage, _ := t.Tag.Lookup("usage")
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	iface := v.Addr().Interface()

	switch v.Kind() {
	case reflect.Struct:
		return ParseFlags(iface)
	case reflect.String:
		dst, ok := iface.(*string)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		return []cli.Flag{stringFlag(name, strValue, dst, env, usage)}, nil
	case reflect.Int:
		dst, ok := iface.(*int)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		if strValue == "" {
			return []cli.Flag{intFlag(name, 0, dst, env, usage)}, nil
		}
		i, err := strconv.Atoi(strValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided value '%v' as an int, err: %w", strValue, err)
		}
		return []cli.Flag{intFlag(name, i, dst, env, usage)}, nil
	case reflect.Float64:
		dst, ok := iface.(*float64)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		if strValue == "" {
			return []cli.Flag{float64Flag(name, 0, dst, env, usage)}, nil
		}
		f, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided value '%v' as an int, err: %w", strValue, err)
		}
		return []cli.Flag{float64Flag(name, f, dst, env, usage)}, nil
	case reflect.Bool:
		dst, ok := iface.(*bool)
		if !ok {
			return nil, fmt.Errorf("failed to parse address of field %v", t.Name)
		}
		if len(strValue) == 0 {
			return []cli.Flag{boolFlag(name, false, dst, env, usage)}, nil
		}
		b, err := strconv.ParseBool(strValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse provided value '%v' as an int, err: %w", strValue, err)
		}
		return []cli.Flag{boolFlag(name, b, dst, env, usage)}, nil
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

func float64Flag(name string, value float64, destination *float64, envVar string, usage string) *cli.Float64Flag {
	return &cli.Float64Flag{Name: name, Value: value, EnvVars: []string{envVar}, Destination: destination, Usage: usage}
}
