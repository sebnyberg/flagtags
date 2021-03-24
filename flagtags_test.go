package flagtags_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sebnyberg/flagtags"
	"github.com/urfave/cli/v2"
)

type NilInterface interface{}

func Test_ParseFlags_Validation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		in        interface{}
		wantFlags []cli.Flag
		wantErr   error
	}{
		{"nil reference", nil, nil, flagtags.ErrNilValue},
		{"interface containing nil", NilInterface(nil), nil, flagtags.ErrNilValue},
		{"argument must be pointer", struct{}{}, nil, flagtags.ErrMustBePtr},
		{"map should return invalid struct", &map[string]bool{}, nil, flagtags.ErrInvalidStruct},
		{
			"private field should err",
			&struct {
				name string `name:"a" env:"b"`
			}{"a"},
			nil,
			flagtags.ErrPrivateField,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotFlags, gotErr := flagtags.ParseFlags(tc.in)
			if !cmp.Equal(gotErr, tc.wantErr, cmpopts.EquateErrors()) {
				t.Errorf("expected err: %v, got: %v", tc.wantErr, gotErr)
			}
			if !cmp.Equal(gotFlags, tc.wantFlags) {
				t.Errorf("expected flags did not match result\n%v", cmp.Diff(tc.wantFlags, gotFlags))
			}
		})
	}
}

func Test_ParseFlags(t *testing.T) {
	var testStruct struct {
		S           string  `name:"s" env:"S" value:"s" usage:"S"`
		I           int     `name:"i" env:"I" value:"1" usage:"I"`
		B           bool    `name:"b" env:"B" value:"true" usage:"B"`
		F64         float64 `name:"f64" env:"F64" value:"0.5" usage:"F64"`
		HostURL     string
		GRPCEnabled bool
	}

	expected := []cli.Flag{
		&cli.StringFlag{
			Name:        "s",
			EnvVars:     []string{"S"},
			Value:       "s",
			Destination: &testStruct.S,
			Usage:       "S",
		},
		&cli.IntFlag{
			Name:        "i",
			EnvVars:     []string{"I"},
			Value:       1,
			Destination: &testStruct.I,
			Usage:       "I",
		},
		&cli.BoolFlag{
			Name:        "b",
			EnvVars:     []string{"B"},
			Value:       true,
			Destination: &testStruct.B,
			Usage:       "B",
		},
		&cli.Float64Flag{
			Name:        "f64",
			EnvVars:     []string{"F64"},
			Value:       0.5,
			Destination: &testStruct.F64,
			Usage:       "F64",
		},
		&cli.StringFlag{
			Name:        "host-url",
			EnvVars:     []string{"HOST_URL"},
			Value:       "",
			Destination: &testStruct.HostURL,
			Usage:       "",
		},
		&cli.BoolFlag{
			Name:        "grpc-enabled",
			EnvVars:     []string{"GRPC_ENABLED"},
			Value:       false,
			Destination: &testStruct.GRPCEnabled,
			Usage:       "",
		},
	}

	flags, err := flagtags.ParseFlags(&testStruct)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !cmp.Equal(flags, expected) {
		t.Errorf("parsed flag did not match expected, got/want\n%v", cmp.Diff(expected, flags))
	}
}

func Test_ParseFlags_EmbeddedStruct(t *testing.T) {
	type EmbeddedStruct struct {
		A string `name:"a"`
	}

	var testStruct struct {
		EmbeddedStruct
	}

	expected := []cli.Flag{
		&cli.StringFlag{
			Name:        "a",
			Destination: &testStruct.A,
			EnvVars:     []string{"A"},
		},
	}

	flags, err := flagtags.ParseFlags(&testStruct)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !cmp.Equal(flags, expected) {
		t.Errorf("parsed flag did not match expected, got/want\n%v", cmp.Diff(expected, flags))
	}
}
