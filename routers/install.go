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

package routers

import (
	"errors"
	"net/mail"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"github.com/go-xorm/xorm"
	log "gopkg.in/clog.v1"
	"gopkg.in/ini.v1"
	"gopkg.in/macaron.v1"

	"fmt"

	"github.com/Aiicy/AiicyDS/models"
	"github.com/Aiicy/AiicyDS/modules/auth"
	"github.com/Aiicy/AiicyDS/modules/base"
	"github.com/Aiicy/AiicyDS/modules/context"
	"github.com/Aiicy/AiicyDS/modules/mailer"
	"github.com/Aiicy/AiicyDS/modules/setting"
)

const (
	INSTALL base.TplName = "install"
)

func checkRunMode() {
	if setting.ProdMode {
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
	}
	log.Info("Run Mode: %s", strings.Title(macaron.Env))
}

func NewServices() {
	setting.NewServices()
	mailer.NewContext()
}

// GlobalInit is for global configuration reload-able.
func GlobalInit() {
	setting.NewContext()
	log.Trace("Custom path: %s", setting.CustomPath)
	log.Trace("Log path: %s", setting.LogRootPath)
	models.LoadConfigs()
	NewServices()

	if setting.InstallLock {
		models.HasEngine = true
	}
	if models.EnableSQLite3 {
		log.Info("SQLite3 Supported")
	}
	if setting.SupportMiniWinService {
		log.Info("Builtin Windows Service Supported")
	}
	checkRunMode()
}

func InstallInit(ctx *context.Context) {
	if setting.InstallLock {
		ctx.Handle(404, "Install", errors.New("Installation is prohibited"))
		return
	}

	ctx.Data["Title"] = ctx.Tr("install.install")
	ctx.Data["PageIsInstall"] = true

	dbOpts := []string{"MySQL", "PostgreSQL", "MSSQL"}
	if models.EnableSQLite3 {
		dbOpts = append(dbOpts, "SQLite3")
	}
	ctx.Data["DbOptions"] = dbOpts
}

func Install(ctx *context.Context) {
	form := auth.InstallForm{}
	form.AdminName = "admin"
	// Database settings
	form.DbHost = models.DbCfg.Host
	form.DbUser = models.DbCfg.User
	form.DbName = models.DbCfg.Name
	form.DbPath = models.DbCfg.Path

	ctx.Data["CurDbOption"] = "MySQL"
	switch models.DbCfg.Type {
	case "postgres":
		ctx.Data["CurDbOption"] = "PostgreSQL"
	case "mssql":
		ctx.Data["CurDbOption"] = "MSSQL"
	case "sqlite3":
		if models.EnableSQLite3 {
			ctx.Data["CurDbOption"] = "SQLite3"
		}
	}

	// Application general settings
	form.AppName = setting.AppName

	form.RunUser = setting.RunUser

	form.Domain = setting.Domain
	form.HTTPPort = setting.HTTPPort
	form.AppUrl = setting.AppUrl
	form.LogRootPath = setting.LogRootPath

	// E-mail service settings
	if setting.MailService != nil {
		form.SMTPHost = setting.MailService.Host
		form.SMTPFrom = setting.MailService.From
		form.SMTPUser = setting.MailService.User
	}
	form.RegisterConfirm = setting.Service.RegisterEmailConfirm
	form.MailNotify = setting.Service.EnableNotifyMail

	// Server and other services settings
	form.OfflineMode = setting.OfflineMode
	form.DisableGravatar = setting.DisableGravatar
	form.EnableFederatedAvatar = setting.EnableFederatedAvatar
	form.DisableRegistration = setting.Service.DisableRegistration
	form.EnableCaptcha = setting.Service.EnableCaptcha
	form.RequireSignInView = setting.Service.RequireSignInView

	auth.AssignForm(form, ctx.Data)
	ctx.HTML(200, INSTALL)
}

func InstallPost(ctx *context.Context, form auth.InstallForm) {
	ctx.Data["CurDbOption"] = form.DbType

	if ctx.HasError() {
		if ctx.HasValue("Err_SMTPEmail") {
			ctx.Data["Err_SMTP"] = true
		}
		if ctx.HasValue("Err_AdminName") ||
			ctx.HasValue("Err_AdminPasswd") ||
			ctx.HasValue("Err_AdminEmail") {
			ctx.Data["Err_Admin"] = true
		}

		ctx.HTML(200, INSTALL)
		return
	}

	if _, err := exec.LookPath("git"); err != nil {
		ctx.RenderWithErr(ctx.Tr("install.test_git_failed", err), INSTALL, &form)
		return
	}

	// Pass basic check, now test configuration.
	// Test database setting.
	dbTypes := map[string]string{"MySQL": "mysql", "PostgreSQL": "postgres", "MSSQL": "mssql", "SQLite3": "sqlite3", "TiDB": "tidb"}
	models.DbCfg.Type = dbTypes[form.DbType]
	models.DbCfg.Host = form.DbHost
	models.DbCfg.User = form.DbUser
	models.DbCfg.Passwd = form.DbPasswd
	models.DbCfg.Name = form.DbName
	models.DbCfg.SSLMode = form.SSLMode
	models.DbCfg.Path = form.DbPath

	if (models.DbCfg.Type == "sqlite3" || models.DbCfg.Type == "tidb") &&
		len(models.DbCfg.Path) == 0 {
		ctx.Data["Err_DbPath"] = true
		ctx.RenderWithErr(ctx.Tr("install.err_empty_db_path"), INSTALL, &form)
		return
	} else if models.DbCfg.Type == "tidb" &&
		strings.ContainsAny(path.Base(models.DbCfg.Path), ".-") {
		ctx.Data["Err_DbPath"] = true
		ctx.RenderWithErr(ctx.Tr("install.err_invalid_tidb_name"), INSTALL, &form)
		return
	}

	// Set test engine.
	var x *xorm.Engine
	if err := models.NewTestEngine(x); err != nil {
		if strings.Contains(err.Error(), `Unknown database type: sqlite3`) {
			ctx.Data["Err_DbType"] = true
			ctx.RenderWithErr(ctx.Tr("install.sqlite3_not_available", "https://aiicyds.io/docs/installation/install_from_binary.html"), INSTALL, &form)
		} else {
			ctx.Data["Err_DbSetting"] = true
			ctx.RenderWithErr(ctx.Tr("install.invalid_db_setting", err), INSTALL, &form)
		}
		fmt.Println("[AiicyDS]", err)
		return
	}

	// Test log root path.
	form.LogRootPath = strings.Replace(form.LogRootPath, "\\", "/", -1)
	if err := os.MkdirAll(form.LogRootPath, os.ModePerm); err != nil {
		ctx.Data["Err_LogRootPath"] = true
		ctx.RenderWithErr(ctx.Tr("install.invalid_log_root_path", err), INSTALL, &form)
		return
	}

	currentUser, match := setting.IsRunUserMatchCurrentUser(form.RunUser)
	if !match {
		ctx.Data["Err_RunUser"] = true
		ctx.RenderWithErr(ctx.Tr("install.run_user_not_match", form.RunUser, currentUser), INSTALL, &form)
		return
	}

	// Make sure FROM field is valid
	if len(form.SMTPFrom) > 0 {
		_, err := mail.ParseAddress(form.SMTPFrom)
		if err != nil {
			ctx.Data["Err_SMTP"] = true
			ctx.Data["Err_SMTPFrom"] = true
			ctx.RenderWithErr(ctx.Tr("install.invalid_smtp_from", err), INSTALL, &form)
			return
		}
	}

	// Check logic loophole between disable self-registration and no admin account.
	if form.DisableRegistration && len(form.AdminName) == 0 {
		ctx.Data["Err_Services"] = true
		ctx.Data["Err_Admin"] = true
		ctx.RenderWithErr(ctx.Tr("install.no_admin_and_disable_registration"), INSTALL, form)
		return
	}

	// Check admin password.
	if len(form.AdminName) > 0 && len(form.AdminPasswd) == 0 {
		ctx.Data["Err_Admin"] = true
		ctx.Data["Err_AdminPasswd"] = true
		ctx.RenderWithErr(ctx.Tr("install.err_empty_admin_password"), INSTALL, form)
		return
	}
	if form.AdminPasswd != form.AdminConfirmPasswd {
		ctx.Data["Err_Admin"] = true
		ctx.Data["Err_AdminPasswd"] = true
		ctx.RenderWithErr(ctx.Tr("form.password_not_match"), INSTALL, form)
		return
	}

	if form.AppUrl[len(form.AppUrl)-1] != '/' {
		form.AppUrl += "/"
	}

	// Save settings.
	cfg := ini.Empty()
	if com.IsFile(setting.CustomConf) {
		// Keeps custom settings if there is already something.
		if err := cfg.Append(setting.CustomConf); err != nil {
			log.Error(4, "Fail to load custom conf '%s': %v", setting.CustomConf, err)
		}
	}
	cfg.Section("database").Key("DB_TYPE").SetValue(models.DbCfg.Type)
	cfg.Section("database").Key("HOST").SetValue(models.DbCfg.Host)
	cfg.Section("database").Key("NAME").SetValue(models.DbCfg.Name)
	cfg.Section("database").Key("USER").SetValue(models.DbCfg.User)
	cfg.Section("database").Key("PASSWD").SetValue(models.DbCfg.Passwd)
	cfg.Section("database").Key("SSL_MODE").SetValue(models.DbCfg.SSLMode)
	cfg.Section("database").Key("PATH").SetValue(models.DbCfg.Path)

	cfg.Section("").Key("APP_NAME").SetValue(form.AppName)
	cfg.Section("").Key("RUN_USER").SetValue(form.RunUser)
	cfg.Section("server").Key("DOMAIN").SetValue(form.Domain)
	cfg.Section("server").Key("HTTP_PORT").SetValue(form.HTTPPort)
	cfg.Section("server").Key("ROOT_URL").SetValue(form.AppUrl)

	if len(strings.TrimSpace(form.SMTPHost)) > 0 {
		cfg.Section("mailer").Key("ENABLED").SetValue("true")
		cfg.Section("mailer").Key("HOST").SetValue(form.SMTPHost)
		cfg.Section("mailer").Key("FROM").SetValue(form.SMTPFrom)
		cfg.Section("mailer").Key("USER").SetValue(form.SMTPUser)
		cfg.Section("mailer").Key("PASSWD").SetValue(form.SMTPPasswd)
	} else {
		cfg.Section("mailer").Key("ENABLED").SetValue("false")
	}
	cfg.Section("service").Key("REGISTER_EMAIL_CONFIRM").SetValue(com.ToStr(form.RegisterConfirm))
	cfg.Section("service").Key("ENABLE_NOTIFY_MAIL").SetValue(com.ToStr(form.MailNotify))

	cfg.Section("server").Key("OFFLINE_MODE").SetValue(com.ToStr(form.OfflineMode))
	cfg.Section("picture").Key("DISABLE_GRAVATAR").SetValue(com.ToStr(form.DisableGravatar))
	cfg.Section("picture").Key("ENABLE_FEDERATED_AVATAR").SetValue(com.ToStr(form.EnableFederatedAvatar))
	cfg.Section("service").Key("DISABLE_REGISTRATION").SetValue(com.ToStr(form.DisableRegistration))
	cfg.Section("service").Key("ENABLE_CAPTCHA").SetValue(com.ToStr(form.EnableCaptcha))
	cfg.Section("service").Key("REQUIRE_SIGNIN_VIEW").SetValue(com.ToStr(form.RequireSignInView))

	cfg.Section("").Key("RUN_MODE").SetValue("prod")

	cfg.Section("session").Key("PROVIDER").SetValue("file")

	cfg.Section("log").Key("MODE").SetValue("file")
	cfg.Section("log").Key("LEVEL").SetValue("Info")
	cfg.Section("log").Key("ROOT_PATH").SetValue(form.LogRootPath)

	cfg.Section("security").Key("INSTALL_LOCK").SetValue("true")
	secretKey, err := base.GetRandomString(15)
	if err != nil {
		ctx.RenderWithErr(ctx.Tr("install.secret_key_failed", err), INSTALL, &form)
		return
	}
	cfg.Section("security").Key("SECRET_KEY").SetValue(secretKey)

	os.MkdirAll(filepath.Dir(setting.CustomConf), os.ModePerm)
	if err := cfg.SaveTo(setting.CustomConf); err != nil {
		ctx.RenderWithErr(ctx.Tr("install.save_config_failed", err), INSTALL, &form)
		return
	}

	GlobalInit()

	// Create admin account
	if len(form.AdminName) > 0 {
		u := &models.User{
			Name:     form.AdminName,
			Email:    form.AdminEmail,
			Passwd:   form.AdminPasswd,
			IsAdmin:  true,
			IsActive: true,
		}
		if err := models.CreateUser(u); err != nil {
			if !models.IsErrUserAlreadyExist(err) {
				setting.InstallLock = false
				ctx.Data["Err_AdminName"] = true
				ctx.Data["Err_AdminEmail"] = true
				ctx.RenderWithErr(ctx.Tr("install.invalid_admin_setting", err), INSTALL, &form)
				return
			}
			log.Info("Admin account already exist")
			u, _ = models.GetUserByName(u.Name)
		}

		// Auto-login for admin
		ctx.Session.Set("uid", u.ID)
		ctx.Session.Set("uname", u.Name)
	}

	log.Info("First-time run install finished!")
	ctx.Flash.Success(ctx.Tr("install.install_success"))
	ctx.Redirect(form.AppUrl + "user/login")
}