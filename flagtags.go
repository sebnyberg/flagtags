package flagtags

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

// Use errors.Is(err, flagtags.ErrX) to check error values
var (
	ErrNilValue      = errors.New("nil value")
	ErrMustBePtr     = errors.New("argument must be a pointer to a struct")
	ErrInvalidStruct = errors.New("provided value was not a struct")
	ErrPrivateField  = errors.New("private field")
	ErrNotSupported  = errors.New("unsupported type")
	ErrNilStructPtr  = errors.New("nested structs must be non-nil")
)

type Options struct {
	// EnvPrefix is put in front of each env var, literally. For example, the
	// prefix "MYAPP_" would give an env var of "MYAPP_MY_FIELD".
	EnvPrefix string

	// FlagPrefix is put in front of each cli flag. For example, the prefix
	// "myapp-" would give a flag name of "myapp-my-field".
	FlagPrefix string
}

var defaultOpts Options

// MustParseFlags parses flag tags from the provided struct and panics on err.
func MustParseFlags(s interface{}) []cli.Flag {
	return MustParseFlagsWithOpts(s, defaultOpts)
}

// MustParseFlagsWithOpts parses flag tags from the provided struct and panics
// on err.
func MustParseFlagsWithOpts(s interface{}, opts Options) []cli.Flag {
	flags, err := ParseFlagsWithOpts(s, opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return flags
}

// ParseFlags parses flags from struct field tags using default options. For
// more info, see ParseFlagsWithOpts.
func ParseFlags(s interface{}) ([]cli.Flag, error) {
	return ParseFlagsWithOpts(s, defaultOpts)
}

// ParseFlagsWithOpts parses flags from the struct fields and their tags.
//
// Supported field types are "int", "string", "bool", "float64".
// Supported tags are "name", "env", "value", "usage".
//
// By default (without tags), value is the default value of the primitive type,
// usage is empty, env is the field name in SCREAMING_SNAKE_CASE, name is the
// field name in kebab-case.
//
// Please see flagtags.Options for customization.
func ParseFlagsWithOpts(s interface{}, opts Options) ([]cli.Flag, error) {
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
		newFlags, fieldErr := flagsFromField(t.Field(i), v.Field(i), &opts)
		if fieldErr != nil && err == nil {
			err = fieldErr
		}
		flags = append(flags, newFlags...)
	}

	if err != nil {
		return nil, err
	}
	return flags, nil
}

func flagsFromField(
	t reflect.StructField,
	v reflect.Value,
	opts *Options,
) ([]cli.Flag, error) {
	// If not set, infer from struct field name
	var name string
	name, ok := t.Tag.Lookup("name")
	if !ok {
		name = toKebabCase(t.Name)
	}
	if opts.FlagPrefix != "" {
		name = opts.FlagPrefix + name
	}

	// If not set, infer from struct field name
	var env string
	env, ok = t.Tag.Lookup("env")
	if !ok {
		// Use custom name if available, otherwise use field name
		var envName string
		envName, hasCustomName := t.Tag.Lookup("name")
		if !hasCustomName {
			envName = t.Name
		}
		env = toScreamingSnakeCase(envName)
	}
	if opts.EnvPrefix != "" {
		env = opts.EnvPrefix + env
	}

	if !v.CanSet() {
		return nil, fmt.Errorf("%w: field '%v' must be made public", ErrPrivateField, t.Name)
	}

	strValue, _ := t.Tag.Lookup("value")
	usage, _ := t.Tag.Lookup("usage")

	// Handle pointer to struct case
	if v.Kind() == reflect.Ptr {
		if v.Type().Elem().Kind() == reflect.Struct && v.IsNil() {
			// Return error if pointer to struct is nil
			return nil, fmt.Errorf("nested structs %w, '%s' was nil", ErrNilStructPtr, t.Name)
		}

		// For non-struct pointers, dereference as before
		v = v.Elem()
	}
	iface := v.Addr().Interface()

	switch v.Kind() {
	case reflect.Struct:
		// For nested structs, use the custom name (if provided) for prefixes
		var prefixName string
		prefixName, ok := t.Tag.Lookup("name")
		if !ok {
			prefixName = t.Name
		}

		optsCopy := *opts
		if !t.Anonymous {
			optsCopy.FlagPrefix = opts.FlagPrefix + toKebabCase(prefixName) + "-"
			optsCopy.EnvPrefix = opts.EnvPrefix + toScreamingSnakeCase(prefixName) + "_"
		}
		return ParseFlagsWithOpts(iface, optsCopy)
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

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	snake = strings.ReplaceAll(snake, "-", "_")
	return strings.ToLower(snake)
}

func toKebabCase(str string) string {
	return strings.ReplaceAll(toSnakeCase(str), "_", "-")
}

func toScreamingSnakeCase(str string) string {
	return strings.ToUpper(toSnakeCase(str))
}
