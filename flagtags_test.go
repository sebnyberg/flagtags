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
		A           string `name:"a" env:"A" value:"a" usage:"use A"`
		B           int    `name:"b" env:"B" value:"1" usage:"use B"`
		C           bool   `name:"c" env:"C" value:"true" usage:"use C"`
		HostURL     string
		GRPCEnabled bool
	}

	expected := []cli.Flag{
		&cli.StringFlag{
			Name:        "a",
			EnvVars:     []string{"A"},
			Value:       "a",
			Destination: &testStruct.A,
			Usage:       "use A",
		},
		&cli.IntFlag{
			Name:        "b",
			EnvVars:     []string{"B"},
			Value:       1,
			Destination: &testStruct.B,
			Usage:       "use B",
		},
		&cli.BoolFlag{
			Name:        "c",
			EnvVars:     []string{"C"},
			Value:       true,
			Destination: &testStruct.C,
			Usage:       "use C",
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
		t.Errorf("expected nil error, got: %v", err)
	}
	if !cmp.Equal(flags, expected) {
		t.Errorf("parsed flag did not match expected, got\n%v", cmp.Diff(expected, flags))
	}
}
