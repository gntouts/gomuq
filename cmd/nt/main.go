package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "nt"
	app.Version = "0.0.1"
	app.Description = "nt is a simple CLI tool that wraps st-link functionalities and provides helper methods to search for USB devices"
	app.Commands = append(app.Commands, commands...)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
