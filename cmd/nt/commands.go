package main

import (
	"fmt"
	"strconv"

	"github.com/gntouts/gomuq/pkg/stlink"
	"github.com/gntouts/gomuq/pkg/usbtool"
	"github.com/urfave/cli/v2"
)

var commands []*cli.Command = []*cli.Command{{
	Name:    "reset",
	Aliases: []string{"r"},
	Usage:   "resets the STM32 board that is currently connected",
	Action: func(cCtx *cli.Context) error {
		stlink.Reset()
		return stlink.Reset()
	},
}, {
	Name:    "flash",
	Aliases: []string{"f"},
	Usage:   "flashes given binary to the STM32 board ",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "binary",
			Aliases: []string{"b"},
			Usage:   "binary to flash",
		},
	},
	Action: func(cCtx *cli.Context) error {
		binary := cCtx.String("binary")
		return stlink.Flash(binary)
	},
}, {
	Name:    "list",
	Aliases: []string{"l", "ls"},
	Usage:   "list all connected USB devices",
	Action: func(cCtx *cli.Context) error {
		devices := usbtool.GetAllDevices()
		for i, d := range devices {
			fmt.Println(strconv.Itoa(i) + ") " + d.String())
		}
		return nil
	},
}, {
	Name:    "search",
	Aliases: []string{"s"},
	Usage:   "search connected USB devices",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "term",
			Aliases: []string{"t"},
			Usage:   "term to search for",
		},
	},
	Action: func(cCtx *cli.Context) error {
		term := cCtx.String("term")
		res, err := usbtool.SearchDevice(term)
		if err != nil {
			return err
		}
		fmt.Println(res.String())
		return nil
	},
},
}
