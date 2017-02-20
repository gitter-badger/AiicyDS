// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/Aiicy/AiicyDS/models"
	"github.com/Aiicy/AiicyDS/modules/setting"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
)

type Context struct {
	*macaron.Context
	Cache   cache.Cache
	csrf    csrf.CSRF
	Flash   *session.Flash
	Session session.Store

	User        *models.User
	IsSigned    bool
	IsBasicAuth bool
}

func Contexter() macaron.Handler {
	return func(c *macaron.Context) {
		ctx := &Context{
			Context: c,
		}
		c.Map(ctx)

		ctx.Data["Link"] = strings.TrimSuffix(ctx.Req.URL.Path, ".html")
		ctx.Data["AppVer"] = setting.AppVer
		ctx.Data["Site"] = setting.Site
		ctx.Data["Page"] = setting.Page
		ctx.Data["Navbar"] = setting.Navbar
		ctx.Data["Asset"] = setting.Asset
		ctx.Data["Extension"] = setting.Extension
	}
}
