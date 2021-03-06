// Copyright 2017 The Aiicy Team.
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.
package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/Aiicy/AiicyDS/cmd"
	"github.com/Aiicy/AiicyDS/modules/setting"
)

const APP_VER = "0.0.1"

func init() {
	setting.AppVer = APP_VER
}

func main() {
	app := cli.NewApp()
	app.Name = "AiicyDS"
	app.Usage = "AiicyDS: A distributed web system."
	app.Version = APP_VER
	app.Author = "Aiicy Team"
	app.Email = "admin@aiicy.com"
	app.Commands = []cli.Command{
		cmd.Web,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
