// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package main

import (
	"github.com/Aiicy/AiicyCMS/modules/db"
	"github.com/Aiicy/AiicyCMS/modules/logic"
	"github.com/Aiicy/AiicyCMS/modules/global"
	"github.com/Aiicy/AiicyCMS/modules/controller"
	"github.com/Aiicy/AiicyCMS/modules/controller/admin"

	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/polaris1119/config"

	pwm "github.com/Aiicy/AiicyCMS/modules/middleware"

	"github.com/fatih/structs"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
	"github.com/polaris1119/logger"
	thirdmw "github.com/polaris1119/middleware"
	"github.com/robfig/cron"
)

func init() {
	// 设置随机数种子
	rand.Seed(time.Now().Unix())

	structs.DefaultTagName = "json"
}

func main() {
	// 支持根据参数打印版本信息
	global.PrintVersion(os.Stdout)

	savePid()

	logger.Init(ROOT+"/log", ConfigFile.MustValue("global", "log_level", "DEBUG"))

	go ServeBackGround()

	e := echo.New()

	serveStatic(e)

	e.Use(thirdmw.EchoLogger())
	e.Use(mw.Recover())
	e.Use(pwm.Installed(filterPrefixs))
	e.Use(pwm.HTTPError())
	e.Use(pwm.AutoLogin())

	frontG := e.Group("", thirdmw.EchoCache())
	controller.RegisterRoutes(frontG)

	frontG.GET("/admin", echo.HandlerFunc(admin.AdminIndex), pwm.NeedLogin(), pwm.AdminAuth())
	adminG := e.Group("/admin", pwm.NeedLogin(), pwm.AdminAuth())
	admin.RegisterRoutes(adminG)

	std := standard.New(getAddr())
	std.SetHandler(e)

	gracefulRun(std)
}

func getAddr() string {
	host := ConfigFile.MustValue("listen", "host", "")
	if host == "" {
		global.App.Host = "localhost"
	} else {
		global.App.Host = host
	}
	global.App.Port = ConfigFile.MustValue("listen", "port", "8088")
	return host + ":" + global.App.Port
}

const (
	IfNoneMatch = "IF-NONE-MATCH"
	Etag        = "Etag"
)

func savePid() {
	pidFilename := ROOT + "/pid/" + filepath.Base(os.Args[0]) + ".pid"
	pid := os.Getpid()

	ioutil.WriteFile(pidFilename, []byte(strconv.Itoa(pid)), 0755)
}

//from server/studygolang/static.go

type staticRootConf struct {
	root   string
	isFile bool
}

var staticFileMap = map[string]staticRootConf{
	"/static/":     {"/static", false},
	"/favicon.ico": {"/static/img/go.ico", true},
	// 服务 sitemap 文件
	"/sitemap/": {"/sitemap", false},
}

var filterPrefixs = make([]string, 0, 3)

func serveStatic(e *echo.Echo) {
	for prefix, rootConf := range staticFileMap {
		filterPrefixs = append(filterPrefixs, prefix)

		if rootConf.isFile {
			e.File(prefix, ROOT+rootConf.root)
		} else {
			e.Static(prefix, ROOT+rootConf.root)
		}
	}
}

//from server/studygolang/background.go

// 后台运行的任务
func ServeBackGround() {

	if db.MasterDB == nil {
		return
	}

	// 初始化 七牛云存储
	logic.DefaultUploader.InitQiniu()

	// 常驻内存的数据
	go loadData()

	c := cron.New()

	// 每天对非活跃用户降频
	c.AddFunc("@daily", decrUserActiveWeight)

	// 两分钟刷一次浏览数（TODO：重启丢失问题？信号控制重启？）
	c.AddFunc("@every 2m", logic.Views.Flush)

	if global.OnlineEnv() {
		// 每天生成 sitemap 文件
		c.AddFunc("@daily", logic.GenSitemap)

		// 给用户发邮件，如通知网站最近的动态，每周的晨读汇总等
		c.AddFunc("0 0 4 * * 1", logic.DefaultEmail.EmailNotice)
	}

	c.Start()
}

func loadData() {
	logic.LoadAuthorities()
	logic.LoadRoles()
	logic.LoadRoleAuthorities()
	logic.LoadNodes()
	logic.LoadCategories()

	for {
		select {
		case <-global.AuthorityChan:
			logic.LoadAuthorities()
		case <-global.RoleChan:
			logic.LoadRoles()
		case <-global.RoleAuthChan:
			logic.LoadRoleAuthorities()
		}
	}
}

func decrUserActiveWeight() {
	logger.Debugln("start decr user active weight...")

	loginTime := time.Now().Add(-72 * time.Hour)
	userList, err := logic.DefaultUser.FindNotLoginUsers(loginTime)
	if err != nil {
		logger.Errorln("获取最近未登录用户失败：", err)
		return
	}

	logger.Debugln("need dealing users:", len(userList))

	for _, user := range userList {
		divide := 5

		if err == nil {
			hours := (loginTime.Sub(user.LoginTime) / 24).Hours()
			if hours < 24 {
				divide = 2
			} else if hours < 48 {
				divide = 3
			} else if hours < 72 {
				divide = 4
			}
		}

		logger.Debugln("decr user weight, username:", user.Username, "divide:", divide)
		logic.DefaultUser.DecrUserWeight("username", user.Username, divide)
	}

	logger.Debugln("end decr user active weight...")
}
