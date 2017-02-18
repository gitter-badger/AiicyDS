// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"net/http"

	"github.com/Aiicy/AiicyDS/models"
	"github.com/Unknwon/log"
	"github.com/go-macaron/i18n"
	"github.com/go-macaron/pongo2"
	"github.com/urfave/cli"
	"gopkg.in/macaron.v1"

	"github.com/Aiicy/AiicyDS/modules/middleware"
	"github.com/Aiicy/AiicyDS/modules/setting"
	"github.com/Aiicy/AiicyDS/routers"
)

var Web = cli.Command{
	Name:   "web",
	Usage:  "Start AiicyDS web server",
	Action: runWeb,
	Flags: []cli.Flag{
		stringFlag("config, c", "custom/app.ini", "Custom configuration file path"),
	},
}

func runWeb(ctx *cli.Context) {
	if ctx.IsSet("config") {
		setting.CustomConf = ctx.String("config")
	}
	setting.NewContext()
	models.NewContext()

	log.Info("Peach %s", setting.AppVer)

	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Statics(macaron.StaticOptions{
		SkipLogging: setting.ProdMode,
	}, "custom/public", "public", models.HTMLRoot))
	m.Use(i18n.I18n(i18n.Options{
		Files:       setting.Docs.Locales,
		DefaultLang: setting.Docs.Langs[0],
	}))
	tplDir := "templates"
	if setting.Page.UseCustomTpl {
		tplDir = "custom/templates"
	}
	m.Use(pongo2.Pongoer(pongo2.Options{
		Directory: tplDir,
	}))
	m.Use(middleware.Contexter())

	m.Get("/", routers.Home)
	m.Get("/docs", routers.Docs)
	m.Get("/docs/images/*", routers.DocsStatic)
	m.Get("/docs/*", routers.Protect, routers.Docs)
	m.Post("/hook", routers.Hook)
	m.Get("/search", routers.Search)
	m.Get("/*", routers.Pages)

	m.NotFound(routers.NotFound)

	listenAddr := fmt.Sprintf("0.0.0.0:%d", setting.HTTPPort)
	log.Info("%s Listen on %s", setting.Site.Name, listenAddr)
	log.Fatal("Fail to start AiicyDS: %v", http.ListenAndServe(listenAddr, m))
}
