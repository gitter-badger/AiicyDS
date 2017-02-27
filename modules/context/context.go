// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/Aiicy/AiicyDS/models"
	"github.com/Aiicy/AiicyDS/modules/auth"
	"github.com/Aiicy/AiicyDS/modules/base"
	"github.com/Aiicy/AiicyDS/modules/setting"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	log "gopkg.in/clog.v1"
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
	query       url.Values
	request     *http.Request
}

// Conver string to int.
func MustInt(s string, defaultVal ...int) int {
	getDefault := func() int {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}

	if s == "" {
		return getDefault()
	}

	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		msg := "context MustInt strconv.Atoi error:" + err.Error()
		_, callerFile, line, ok := runtime.Caller(1)
		if ok {
			msg += fmt.Sprintf("file:%s,line:%d", callerFile, line)
		}
		log.Trace("context MustInt: %s", msg)
		return getDefault()
	}

	return i
}

// QueryParam returns the query param for the provided name.
func (c *Context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

// QueryParams returns the query parameters as `url.Values`.
func (c *Context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

// HasError returns true if error occurs in form validation.
func (ctx *Context) HasError() bool {
	hasErr, ok := ctx.Data["HasError"]
	if !ok {
		return false
	}
	ctx.Flash.ErrorMsg = ctx.Data["ErrorMsg"].(string)
	ctx.Data["Flash"] = ctx.Flash
	return hasErr.(bool)
}

// HasValue returns true if value of given name exists.
func (ctx *Context) HasValue(name string) bool {
	_, ok := ctx.Data[name]
	return ok
}

// HTML calls Context.HTML and converts template name to string.
func (ctx *Context) HTML(status int, name base.TplName) {
	log.Trace("Template: %s", name)
	ctx.Context.HTML(status, string(name))
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (ctx *Context) RenderWithErr(msg string, tpl base.TplName, form interface{}) {
	if form != nil {
		auth.AssignForm(form, ctx.Data)
	}
	ctx.Flash.ErrorMsg = msg
	ctx.Data["Flash"] = ctx.Flash
	ctx.HTML(200, tpl)
}

// Handle handles and logs error by given status.
func (ctx *Context) Handle(status int, title string, err error) {
	switch status {
	case 404:
		ctx.Data["Title"] = "Page Not Found"
	case 500:
		ctx.Data["Title"] = "Internal Server Error"
		log.Error(4, "%s: %v", title, err)
		if !setting.ProdMode || (ctx.IsSigned && ctx.User.IsAdmin) {
			ctx.Data["ErrorMsg"] = err
		}
	}
	ctx.HTML(status, base.TplName(fmt.Sprintf("status/%d", status)))
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
