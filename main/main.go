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
