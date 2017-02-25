// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package routers

import (
	"github.com/Aiicy/AiicyDS/modules/base"

	"github.com/Aiicy/AiicyDS/modules/context"
	//"github.com/Aiicy/AiicyDS/modules/setting"
)

const (
	INDEX base.TplName = "home/index"
)

func Index(ctx *context.Context) {

	ctx.HTML(200, INDEX)
}
