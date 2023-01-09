# flagtags

[![Go Reference](https://pkg.go.dev/badge/github.com/sebnyberg/flagtags.svg)](https://pkg.go.dev/github.com/sebnyberg/flagtags)

Each configurable option should exist in three forms:

1. Struct field (PascalCase)
2. CLI flag (--kebab-case)
3. Environment variable (SCREAMING_SNAKE_CASE)

Keeping the three in sync requires a lot of boilerplate code.

This package uses a struct as the source of truth to generate `urfave/cli/v2`
flags, providing tags for overrides and additional options. 

## Why?

Because writing out config options one-by-one is error-prone.

## Example

```go
package main

import (
	"log"
	"os"

	"github.com/sebnyberg/flagtags"
	"github.com/urfave/cli/v2"
)

type config struct {
	Port        int `value:"3001"`
	DisableAuth bool
	JWTSignKey  string
}

func main() {
	var opts flagtags.Options
	opts.EnvPrefix = "MYAPP_"
	flags := flagtags.ParseFlagsWithOpts(&config, opts)
	app := &cli.App{flags: flags}
	app.Run(os.Args)
}
```

### Supported flag fields (tags)

| Tag | Flag field | Description | Default value |
|---|---| --- | --- |
| `name` | `flag.Name` | Flag name |  Struct field name in kebab-case |
| `env` | `flag.EnvVars[0]` | Environment variable name | Struct field name in SCREAMING_SNAKE_CASE |
| `value` | `flag.Value` | Default flag value | Default initialized value for the corresponding primitive type |
| `usage` | `flag.Usage` | Usage | Empty string |

### Supported types

Currently this library only supports struct fields of type:

* `int`
* `string`
* `bool`

## Complete example

```go
package main

import (
	"fmt"
	"os"

	"github.com/kr/pretty"
	"github.com/sebnyberg/flagtags"
	"github.com/urfave/cli/v2"
)

type config struct {
	Port             int    `value:"3001"`
	DisableAuth      bool   `usage:"Disable authentication"`
	JWTSignKey       string `usage:"JWT signing key"`
	DatabaseHost     string `name:"pghost" env:"PGHOST" usage:"Postgres hostname"`
	DatabasePassword string `name:"pgpassword" env:"PGPASSWORD" usage:"Postgres hostname"`
}

func main() {
	var conf config
	var parseOpts flagtags.Options
	parseOpts.EnvPrefix = "MYAPP_"
	flags := flagtags.MustParseFlagsWithOpts(&conf, parseOpts)

	app := &cli.App{
		Name:     "server",
		HelpName: "server",
		Usage:    "run the server",
		Flags:    flags,
		Action: func(c *cli.Context) error {
			pretty.Println(conf)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

Example output from the CLI invocation:

```bash
$ go run main.go --help
NAME:
   server - run the server

USAGE:
   server [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port value          (default: 3001) [$PORT]
   --disable-auth        Disable authentication (default: false) [$DISABLE_AUTH]
   --jwt-sign-key value  JWT signing key [$JWT_SIGN_KEY]
   --pghost value        Postgres hostname (default: "test") [$PGHOST]
   --pgpassword value    Postgres hostname [$PGPASSWORD]
   --help, -h            show help (default: false)
```
