package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	rootFlags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug,d ",
			Usage: "debug mode",
		},
		&cli.StringFlag{
			Name:  "commit, c",
			Usage: "commit hash to generate coverage report for",
		},
		&cli.BoolFlag{
			Name:  "graph,g ",
			Usage: "generate html graphs",
		},
		&cli.StringFlag{
			Name:     "program, p",
			Usage:    "program name [k3s|rke2]",
			Required: true,
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "test-coverage"
	app.Usage = "Generate coverage report for E2E/Integration tests"
	app.Flags = rootFlags
	app.Action = coverage
	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}

}
