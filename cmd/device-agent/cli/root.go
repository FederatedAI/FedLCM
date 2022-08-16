// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2" // imports as package "cli"
)

func init() {
	cli.HelpFlag = &cli.BoolFlag{Name: "help", Aliases: []string{"h"}, Usage: "Show help"}
}

func initCommandLine() *cli.App {
	app := &cli.App{

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "debug",
				Aliases:  []string{"d"},
				Value:    false,
				Usage:    "Enable debug logging",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			RegisterCommand(),
			StatusCommand(),
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	return app
}

func Run(Args []string) error {
	app := initCommandLine()
	return app.Run(Args)
}
