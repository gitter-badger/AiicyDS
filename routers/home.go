// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routers

import (
	"github.com/Aiicy/AiicyDS/modules/base"

	"github.com/Aiicy/AiicyDS/modules/context"
	"github.com/Aiicy/AiicyDS/modules/setting"
)

const (
	HOME base.TplName = "home"
)

func Home(ctx *context.Context) {
	if !setting.Page.HasLandingPage {
		ctx.Redirect(setting.Page.DocsBaseURL)
		return
	}

	ctx.HTML(200, "home")
}

func NotFound(ctx *context.Context) {
	ctx.Data["Title"] = "404"
	ctx.HTML(404, "404")
}
