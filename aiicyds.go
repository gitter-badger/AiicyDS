// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"os"
	"runtime"

	"github.com/urfave/cli"

	"github.com/Aiicy/AiicyDS/cmd"
	"github.com/Aiicy/AiicyDS/modules/setting"
)

const APP_VER = "0.0.1"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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
		cmd.New,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
