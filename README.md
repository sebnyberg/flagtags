# flagtags

[![Go Reference](https://pkg.go.dev/badge/github.com/sebnyberg/flagtags.svg)](https://pkg.go.dev/github.com/sebnyberg/flagtags)

Building command-line applications typically requires maintaining configuration in three places:

- **Struct fields** (`Config.DatabasePort`)
- **CLI flags** (`--database-port`)
- **Environment variables** (`DATABASE_PORT`)

Keeping these synchronized is tedious, error-prone, and leads to inconsistent naming across your application.

`flagtags` uses struct tags as the single source of truth to automatically generate consistent `urfave/cli/v2` flags with proper naming conventions:

- **PascalCase** struct fields → **kebab-case** CLI flags → **SCREAMING_SNAKE_CASE** environment variables

## Installation

```bash
go get github.com/sebnyberg/flagtags
```

## Quick Start

```go
package main

import (
	"log"
	"os"

	"github.com/sebnyberg/flagtags"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Port        int    `value:"3000" usage:"Server port"`
	DatabaseURL string `usage:"Database connection URL"`
	Debug       bool   `usage:"Enable debug mode"`
}

func main() {
	var config Config
	
	flags := flagtags.MustParseFlags(&config)
	
	app := &cli.App{
		Name:  "myapp",
		Flags: flags,
		Action: func(c *cli.Context) error {
			// config struct is automatically populated
			log.Printf("Starting server on port %d", config.Port)
			return nil
		},
	}
	
	app.Run(os.Args)
}
```

This automatically creates:
- `--port` flag with `PORT` env var (default: 3000)
- `--database-url` flag with `DATABASE_URL` env var  
- `--debug` flag with `DEBUG` env var

## Supported Types

| Go Type | CLI Flag Type | Description |
|---------|---------------|-------------|
| `string` | `StringFlag` | Text values |
| `int` | `IntFlag` | Integer values |
| `bool` | `BoolFlag` | Boolean flags |
| `float64` | `Float64Flag` | Floating point values |

## Struct Tags

Customize flag behavior with struct tags:

| Tag | Purpose | Default |
|-----|---------|---------|
| `name` | CLI flag name | Field name in kebab-case |
| `env` | Environment variable name | Field name in SCREAMING_SNAKE_CASE |
| `value` | Default value | Go zero value |
| `usage` | Help text | Empty string |

## Advanced Usage

### Prefixes and Options

Use `flagtags.Options` to add prefixes to flags and environment variables:

```go
type Config struct {
    DatabaseHost string `usage:"Database hostname"`
    DatabasePort int    `value:"5432" usage:"Database port"`
}

var config Config
opts := flagtags.Options{
    EnvPrefix:  "MYAPP_",     // Creates MYAPP_DATABASE_HOST, MYAPP_DATABASE_PORT
    FlagPrefix: "db-",        // Creates --db-database-host, --db-database-port
}

flags := flagtags.MustParseFlagsWithOpts(&config, opts)
```

### Nested Structs

Organize complex configuration with nested structs:

```go
type Config struct {
    Server struct {
        Port    int    `value:"8080" usage:"Server port"`
        Host    string `value:"localhost" usage:"Server host"`
        TLS     struct {
            CertFile string `usage:"TLS certificate file"`
            KeyFile  string `usage:"TLS private key file"`
        }
    }
    Database struct {
        URL      string `usage:"Database connection URL"`
        MaxConns int    `value:"10" usage:"Maximum database connections"`
    } `name:"db"` // Custom prefix: --db-url, --db-max-conns
}

var config Config
flags := flagtags.MustParseFlags(&config)
```

This creates flags like:

- `--server-port` / `SERVER_PORT`
- `--server-tls-cert-file` / `SERVER_TLS_CERT_FILE`  
- `--db-url` / `DB_URL` (using custom `name` tag)

### Custom Tag Examples

```go
type Config struct {
    // Custom flag and env names
    APIKey string `name:"api-key" env:"EXTERNAL_API_KEY" usage:"API authentication key"`
    
    // Different flag name, default env name
    LogLevel string `name:"log-level" usage:"Logging level (debug, info, warn, error)"`
    
    // Custom default value
    Timeout float64 `value:"30.5" usage:"Request timeout in seconds"`
    
    // Boolean with default true
    EnableMetrics bool `value:"true" usage:"Enable metrics collection"`
}
```

## Complete Example

```go
package main

import (
	"fmt"
	"os"

	"github.com/sebnyberg/flagtags"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Port             int    `value:"3001" usage:"Server port"`
	DisableAuth      bool   `usage:"Disable authentication"`
	JWTSignKey       string `usage:"JWT signing key"`
	DatabaseHost     string `name:"pghost" env:"PGHOST" usage:"Postgres hostname"`
	DatabasePassword string `name:"pgpassword" env:"PGPASSWORD" usage:"Postgres password"`
}

func main() {
	var config Config
	
	opts := flagtags.Options{
		EnvPrefix: "MYAPP_",
	}
	flags := flagtags.MustParseFlagsWithOpts(&config, opts)

	app := &cli.App{
		Name:     "server",
		Usage:    "run the server",
		Flags:    flags,
		Action: func(c *cli.Context) error {
			fmt.Printf("Config: %+v\n", config)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
```

**Example CLI output:**

```bash
$ ./server --help
NAME:
   server - run the server

USAGE:
   server [global options]

GLOBAL OPTIONS:
   --port value          Server port (default: 3001) [$MYAPP_PORT]
   --disable-auth        Disable authentication [$MYAPP_DISABLE_AUTH]
   --jwt-sign-key value  JWT signing key [$MYAPP_JWT_SIGN_KEY]
   --pghost value        Postgres hostname [$PGHOST]
   --pgpassword value    Postgres password [$PGPASSWORD]
   --help, -h            show help
```

For convenience, use `MustParseFlags()` or `MustParseFlagsWithOpts()` to panic on errors during application startup.

## API Reference

### Functions

- `ParseFlags(s interface{}) ([]cli.Flag, error)` - Parse flags with default options
- `ParseFlagsWithOpts(s interface{}, opts Options) ([]cli.Flag, error)` - Parse flags with custom options
- `MustParseFlags(s interface{}) []cli.Flag` - Parse flags, panic on error
- `MustParseFlagsWithOpts(s interface{}, opts Options) []cli.Flag` - Parse flags with options, panic on error

### Types

```go
type Options struct {
    EnvPrefix  string // Prefix for environment variables
    FlagPrefix string // Prefix for CLI flags
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
