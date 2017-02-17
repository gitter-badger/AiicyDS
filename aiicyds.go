// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import "os"
import "github.com/urfave/cli"
import "./src/cmd"

const APP_VER = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "AiicyDS"
	app.Usage = "AiicyDS: A distributed website system"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		cmd.CmdWeb,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
