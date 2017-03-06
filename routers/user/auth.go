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

package user

import (
	"fmt"
	"net/url"

	"github.com/Aiicy/AiicyDS/models"
	"github.com/Aiicy/AiicyDS/modules/auth"
	"github.com/Aiicy/AiicyDS/modules/base"
	"github.com/Aiicy/AiicyDS/modules/context"
	"github.com/Aiicy/AiicyDS/modules/mailer"
	"github.com/Aiicy/AiicyDS/modules/setting"
	"github.com/go-macaron/captcha"
	log "gopkg.in/clog.v1"
)

const (
	SIGNIN          base.TplName = "user/auth/signin"
	SIGNUP          base.TplName = "user/auth/signup"
	ACTIVATE        base.TplName = "user/auth/activate"
	FORGOT_PASSWORD base.TplName = "user/auth/forgot_passwd"
	RESET_PASSWORD  base.TplName = "user/auth/reset_passwd"
)

// AutoSignIn reads cookie and try to auto-login.
func AutoSignIn(ctx *context.Context) (bool, error) {
	if !models.HasEngine {
		return false, nil
	}

	uname := ctx.GetCookie(setting.CookieUserName)
	if len(uname) == 0 {
		return false, nil
	}

	isSucceed := false
	defer func() {
		if !isSucceed {
			log.Trace("auto-login cookie cleared: %s", uname)
			ctx.SetCookie(setting.CookieUserName, "", -1, setting.AppSubUrl)
			ctx.SetCookie(setting.CookieRememberName, "", -1, setting.AppSubUrl)
		}
	}()

	u, err := models.GetUserByName(uname)
	if err != nil {
		if !models.IsErrUserNotExist(err) {
			return false, fmt.Errorf("GetUserByName: %v", err)
		}
		return false, nil
	}

	if val, ok := ctx.GetSuperSecureCookie(u.Rands+u.Passwd, setting.CookieRememberName); !ok || val != u.Name {
		return false, nil
	}

	isSucceed = true
	ctx.Session.Set("uid", u.ID)
	ctx.Session.Set("uname", u.Name)
	ctx.SetCookie(setting.CSRFCookieName, "", -1, setting.AppSubUrl)
	return true, nil
}

// isValidRedirect returns false if the URL does not redirect to same site.
// False: //url, http://url
// True: /url
func isValidRedirect(url string) bool {
	return len(url) >= 2 && url[0] == '/' && url[1] != '/'
}

func SignIn(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_in")

	// Check auto-login.
	isSucceed, err := AutoSignIn(ctx)
	if err != nil {
		ctx.Handle(500, "AutoSignIn", err)
		return
	}

	redirectTo := ctx.Query("redirect_to")
	if len(redirectTo) > 0 {
		ctx.SetCookie("redirect_to", redirectTo, 0, setting.AppSubUrl)
	} else {
		redirectTo, _ = url.QueryUnescape(ctx.GetCookie("redirect_to"))
	}
	ctx.SetCookie("redirect_to", "", -1, setting.AppSubUrl)

	if isSucceed {
		if isValidRedirect(redirectTo) {
			ctx.Redirect(redirectTo)
		} else {
			ctx.Redirect(setting.AppSubUrl + "/")
		}
		return
	}

	ctx.HTML(200, SIGNIN)
}

func SignInPost(ctx *context.Context, form auth.SignInForm) {
	ctx.Data["Title"] = ctx.Tr("sign_in")

	if ctx.HasError() {
		ctx.HTML(200, SIGNIN)
		return
	}

	u, err := models.UserSignIn(form.UserName, form.Password)
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.RenderWithErr(ctx.Tr("form.username_password_incorrect"), SIGNIN, &form)
		} else {
			ctx.Handle(500, "UserSignIn", err)
		}
		return
	}

	if form.Remember {
		days := 86400 * setting.LogInRememberDays
		ctx.SetCookie(setting.CookieUserName, u.Name, days, setting.AppSubUrl, "", setting.CookieSecure, true)
		ctx.SetSuperSecureCookie(u.Rands+u.Passwd, setting.CookieRememberName, u.Name, days, setting.AppSubUrl, "", setting.CookieSecure, true)
	}

	ctx.Session.Set("uid", u.ID)
	ctx.Session.Set("uname", u.Name)

	// Clear whatever CSRF has right now, force to generate a new one
	ctx.SetCookie(setting.CSRFCookieName, "", -1, setting.AppSubUrl)

	redirectTo, _ := url.QueryUnescape(ctx.GetCookie("redirect_to"))
	ctx.SetCookie("redirect_to", "", -1, setting.AppSubUrl)
	if isValidRedirect(redirectTo) {
		ctx.Redirect(redirectTo)
		return
	}

	ctx.Redirect(setting.AppSubUrl + "/")
}

func SignOut(ctx *context.Context) {
	ctx.Session.Delete("uid")
	ctx.Session.Delete("uname")
	ctx.Session.Delete("socialId")
	ctx.Session.Delete("socialName")
	ctx.Session.Delete("socialEmail")
	ctx.SetCookie(setting.CookieUserName, "", -1, setting.AppSubUrl)
	ctx.SetCookie(setting.CookieRememberName, "", -1, setting.AppSubUrl)
	ctx.SetCookie(setting.CSRFCookieName, "", -1, setting.AppSubUrl)
	ctx.Redirect(setting.AppSubUrl + "/")
}

func SignUp(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_up")

	ctx.Data["EnableCaptcha"] = setting.Service.EnableCaptcha

	if setting.Service.DisableRegistration {
		ctx.Data["DisableRegistration"] = true
		ctx.HTML(200, SIGNUP)
		return
	}

	ctx.HTML(200, SIGNUP)
}

func SignUpPost(ctx *context.Context, cpt *captcha.Captcha, form auth.RegisterForm) {
	ctx.Data["Title"] = ctx.Tr("sign_up")

	ctx.Data["EnableCaptcha"] = setting.Service.EnableCaptcha

	if setting.Service.DisableRegistration {
		ctx.Error(403)
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, SIGNUP)
		return
	}

	if setting.Service.EnableCaptcha && !cpt.VerifyReq(ctx.Req) {
		ctx.Data["Err_Captcha"] = true
		ctx.RenderWithErr(ctx.Tr("form.captcha_incorrect"), SIGNUP, &form)
		return
	}

	if form.Password != form.Retype {
		ctx.Data["Err_Password"] = true
		ctx.RenderWithErr(ctx.Tr("form.password_not_match"), SIGNUP, &form)
		return
	}

	u := &models.User{
		Name:     form.UserName,
		Email:    form.Email,
		Passwd:   form.Password,
		IsActive: !setting.Service.RegisterEmailConfirm,
	}
	if err := models.CreateUser(u); err != nil {
		switch {
		case models.IsErrUserAlreadyExist(err):
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(ctx.Tr("form.username_been_taken"), SIGNUP, &form)
		case models.IsErrEmailAlreadyUsed(err):
			ctx.Data["Err_Email"] = true
			ctx.RenderWithErr(ctx.Tr("form.email_been_used"), SIGNUP, &form)
		case models.IsErrNameReserved(err):
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), SIGNUP, &form)
		case models.IsErrNamePatternNotAllowed(err):
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), SIGNUP, &form)
		default:
			ctx.Handle(500, "CreateUser", err)
		}
		return
	}
	log.Trace("Account created: %s", u.Name)

	// Auto-set admin for the only user.
	if models.CountUsers() == 1 {
		u.IsAdmin = true
		u.IsActive = true
		if err := models.UpdateUser(u); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}
	}

	// Send confirmation email, no need for social account.
	if setting.Service.RegisterEmailConfirm && u.ID > 1 {
		mailer.SendActivateAccountMail(ctx.Context, models.NewMailerUser(u))
		ctx.Data["IsSendRegisterMail"] = true
		ctx.Data["Email"] = u.Email
		ctx.Data["Hours"] = setting.Service.ActiveCodeLives / 60
		ctx.HTML(200, ACTIVATE)

		if err := ctx.Cache.Put("MailResendLimit_"+u.LowerName, u.LowerName, 180); err != nil {
			log.Error(4, "Set cache(MailResendLimit) fail: %v", err)
		}
		return
	}

	ctx.Redirect(setting.AppSubUrl + "/user/login")
}

func ResetPasswd(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("auth.reset_password")

	code := ctx.Query("code")
	if len(code) == 0 {
		ctx.Error(404)
		return
	}
	ctx.Data["Code"] = code
	ctx.Data["IsResetForm"] = true
	ctx.HTML(200, RESET_PASSWORD)
}

func ResetPasswdPost(ctx *context.Context) {
	ctx.Data["Title"] = ctx.Tr("auth.reset_password")

	code := ctx.Query("code")
	if len(code) == 0 {
		ctx.Error(404)
		return
	}
	ctx.Data["Code"] = code

	if u := models.VerifyUserActiveCode(code); u != nil {
		// Validate password length.
		passwd := ctx.Query("password")
		if len(passwd) < 6 {
			ctx.Data["IsResetForm"] = true
			ctx.Data["Err_Password"] = true
			ctx.RenderWithErr(ctx.Tr("auth.password_too_short"), RESET_PASSWORD, nil)
			return
		}

		u.Passwd = passwd
		var err error
		if u.Rands, err = models.GetUserSalt(); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}
		if u.Salt, err = models.GetUserSalt(); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}
		u.EncodePasswd()
		if err := models.UpdateUser(u); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}

		log.Trace("User password reset: %s", u.Name)
		ctx.Redirect(setting.AppSubUrl + "/user/login")
		return
	}

	ctx.Data["IsResetFailed"] = true
	ctx.HTML(200, RESET_PASSWORD)
}
