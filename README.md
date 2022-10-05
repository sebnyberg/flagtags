# flagtags

Tag struct fields to generate `urfave/cli/v2` flags.

## Example

Binding flags in `urfave/cli/v2` to a struct requires the following code:

```go
package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

type config struct {
	Port        int `value:3001`
	DisableAuth bool
	JWTSignKey  string
}

var conf config

func main() {
	flags := []cli.Flag{
		&cli.IntFlag{
			Name:        "port",
			EnvVars:     []string{"PORT"},
			Value:       3001,
			Destination: &conf.Port,
		},
		&cli.BoolFlag{
			Name:        "disable-auth",
			EnvVars:     []string{"DISABLE_AUTH"},
			Value:       false,
			Destination: &conf.DisableAuth,
		},
		&cli.StringFlag{
			Name:        "jwt-sign-key",
			EnvVars:     []string{"JWT_SIGN_KEY"},
			Value:       "",
			Destination: &conf.JWTSignKey,
		},
	}
	app := &cli.App{Flags: flags}
	app.Run(os.Args)
}
```

This package provides some sensible defaults and tags for placing this initialization directly in the struct.

Equivalent code using flagtags:

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
	var conf config
	flags, err := flagtags.ParseFlags(&conf)
	if err != nil {
		log.Fatalln(err)
	}
	app := &cli.App{Flags: flags}
	app.Run(os.Args)
}
```

## Tags and sensible defaults

When a specific tag is missing, sensible defaults are applied based on the name and primitive type of the struct field. See next section.

The presence of a tag takes precendence over sensible defaults.

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

	app := &cli.App{
		Name:     "server",
		HelpName: "server",
		Usage:    "run the server",
		Flags:    flagtags.MustParseFlags(&conf),
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
