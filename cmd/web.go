// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"strings"

	"github.com/Aiicy/AiicyDS/models"
	"github.com/Aiicy/AiicyDS/modules/auth"
	"github.com/Aiicy/AiicyDS/modules/bindata"
	"github.com/Aiicy/AiicyDS/modules/context"
	"github.com/Aiicy/AiicyDS/modules/mailer"
	"github.com/Aiicy/AiicyDS/modules/template"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/gzip"
	"github.com/go-macaron/i18n"
	"github.com/go-macaron/session"
	"github.com/go-macaron/toolbox"
	"github.com/go-xorm/xorm"
	"github.com/mcuadros/go-version"
	"github.com/urfave/cli"
	log "gopkg.in/clog.v1"
	"gopkg.in/ini.v1"
	"gopkg.in/macaron.v1"

	"github.com/Aiicy/AiicyDS/modules/setting"
	"github.com/Aiicy/AiicyDS/routers"
)

var Web = cli.Command{
	Name:  "web",
	Usage: "Start AiicyDS web server",
	Description: `AiicyDS web server is the only thing you need to run,
and it takes care of all the other things for you`,
	Action: runWeb,
	Flags: []cli.Flag{
		stringFlag("port, p", "8080", "Temporary port number to prevent conflict"),
		stringFlag("config, c", "custom/conf/app.ini", "Custom configuration file path"),
	},
}

type VerChecker struct {
	ImportPath string
	Version    func() string
	Expected   string
}

// checkVersion checks if binary matches the version of templates files.
func checkVersion() {
	// Templates.
	data, err := ioutil.ReadFile(setting.StaticRootPath + "/templates/.VERSION")
	if err != nil {
		log.Fatal(4, "Fail to read 'templates/.VERSION': %v", err)
	}
	tplVer := string(data)
	if tplVer != setting.AppVer {
		if version.Compare(tplVer, setting.AppVer, ">") {
			log.Fatal(4, "Binary version is lower than template file version, did you forget to recompile AiicyDS?")
		} else {
			log.Fatal(4, "Binary version is higher than template file version, did you forget to update template files?")
		}
	}

	// Check dependency version.
	// LEGACY [0.11]: no need to check version as we check in vendor into version control
	checkers := []VerChecker{
		{"github.com/go-xorm/xorm", func() string { return xorm.Version }, "0.6.0"},
		{"github.com/go-macaron/binding", binding.Version, "0.3.2"},
		{"github.com/go-macaron/cache", cache.Version, "0.1.2"},
		{"github.com/go-macaron/csrf", csrf.Version, "0.1.0"},
		{"github.com/go-macaron/i18n", i18n.Version, "0.3.0"},
		{"github.com/go-macaron/session", session.Version, "0.1.6"},
		{"github.com/go-macaron/toolbox", toolbox.Version, "0.1.3"},
		{"gopkg.in/ini.v1", ini.Version, "1.8.4"},
		{"gopkg.in/macaron.v1", macaron.Version, "1.1.7"},
	}
	for _, c := range checkers {
		if !version.Compare(c.Version(), c.Expected, ">=") {
			log.Fatal(4, `Dependency outdated!
Package '%s' current version (%s) is below requirement (%s),
please use following command to update this package and recompile AiicyDS:
go get -u %[1]s`, c.ImportPath, c.Version(), c.Expected)
		}
	}
}

// newMacaron initializes Macaron instance.
func newMacaron() *macaron.Macaron {
	m := macaron.New()
	if !setting.DisableRouterLog {
		m.Use(macaron.Logger())
	}
	m.Use(macaron.Recovery())
	if setting.EnableGzip {
		m.Use(gzip.Gziper())
	}
	if setting.Protocol == setting.SCHEME_FCGI {
		m.SetURLPrefix(setting.AppSubUrl)
	}
	m.Use(macaron.Static(
		path.Join(setting.StaticRootPath, "public"),
		macaron.StaticOptions{
			SkipLogging: setting.DisableRouterLog,
		},
	))
	m.Use(macaron.Static(
		setting.AvatarUploadPath,
		macaron.StaticOptions{
			Prefix:      "avatars",
			SkipLogging: setting.DisableRouterLog,
		},
	))

	funcMap := template.NewFuncMap()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:         path.Join(setting.StaticRootPath, "templates"),
		AppendDirectories: []string{path.Join(setting.CustomPath, "templates")},
		Funcs:             funcMap,
		IndentJSON:        macaron.Env != macaron.PROD,
	}))
	mailer.InitMailRender(path.Join(setting.StaticRootPath, "templates/mail"),
		path.Join(setting.CustomPath, "templates/mail"), funcMap)

	localeNames, err := bindata.AssetDir("conf/locale")
	if err != nil {
		log.Fatal(4, "Fail to list locale files: %v", err)
	}
	localFiles := make(map[string][]byte)
	for _, name := range localeNames {
		localFiles[name] = bindata.MustAsset("conf/locale/" + name)
	}
	m.Use(i18n.I18n(i18n.Options{
		SubURL:          setting.AppSubUrl,
		Files:           localFiles,
		CustomDirectory: path.Join(setting.CustomPath, "conf/locale"),
		Langs:           setting.Langs,
		Names:           setting.Names,
		DefaultLang:     "en-US",
		Redirect:        true,
	}))
	m.Use(cache.Cacher(cache.Options{
		Adapter:       setting.CacheAdapter,
		AdapterConfig: setting.CacheConn,
		Interval:      setting.CacheInterval,
	}))
	m.Use(captcha.Captchaer(captcha.Options{
		SubURL: setting.AppSubUrl,
	}))
	m.Use(session.Sessioner(setting.SessionConfig))
	m.Use(csrf.Csrfer(csrf.Options{
		Secret:     setting.SecretKey,
		Cookie:     setting.CSRFCookieName,
		SetCookie:  true,
		Header:     "X-Csrf-Token",
		CookiePath: setting.AppSubUrl,
	}))
	m.Use(toolbox.Toolboxer(m, toolbox.Options{
		HealthCheckFuncs: []*toolbox.HealthCheckFuncDesc{
			&toolbox.HealthCheckFuncDesc{
				Desc: "Database connection",
				Func: models.Ping,
			},
		},
	}))
	m.Use(context.Contexter())
	return m
}

func runWeb(ctx *cli.Context) error {
	if ctx.IsSet("config") {
		setting.CustomConf = ctx.String("config")
	}
	routers.GlobalInit()
	checkVersion()

	m := newMacaron()

	ignSignIn := context.Toggle(&context.ToggleOptions{SignInRequired: setting.Service.RequireSignInView})
	bindIgnErr := binding.BindIgnErr

	// FIXME: not all routes need go through same middlewares.
	// Especially some AJAX requests, we can reduce middleware number to improve performance.
	// Routers.
	m.Get("/", ignSignIn, routers.Home)
	m.Combo("/install", routers.InstallInit).Get(routers.Install).
		Post(bindIgnErr(auth.InstallForm{}), routers.InstallPost)

	// robots.txt
	m.Get("/robots.txt", func(ctx *context.Context) {
		if setting.HasRobotsTxt {
			ctx.ServeFileContent(path.Join(setting.CustomPath, "robots.txt"))
		} else {
			ctx.Error(404)
		}
	})

	// Not found handler.
	m.NotFound(routers.NotFound)

	// Flag for port number in case first time run conflict.
	if ctx.IsSet("port") {
		setting.AppUrl = strings.Replace(setting.AppUrl, setting.HTTPPort, ctx.String("port"), 1)
		setting.HTTPPort = ctx.String("port")
	}

	var listenAddr string
	if setting.Protocol == setting.SCHEME_UNIX_SOCKET {
		listenAddr = fmt.Sprintf("%s", setting.HTTPAddr)
	} else {
		listenAddr = fmt.Sprintf("%s:%s", setting.HTTPAddr, setting.HTTPPort)
	}
	log.Info("Listen: %v://%s%s", setting.Protocol, listenAddr, setting.AppSubUrl)

	var err error
	switch setting.Protocol {
	case setting.SCHEME_HTTP:
		err = http.ListenAndServe(listenAddr, m)
	case setting.SCHEME_HTTPS:
		server := &http.Server{Addr: listenAddr, TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS10,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Required for HTTP/2 support.
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}, Handler: m}
		err = server.ListenAndServeTLS(setting.CertFile, setting.KeyFile)
	case setting.SCHEME_FCGI:
		err = fcgi.Serve(nil, m)
	case setting.SCHEME_UNIX_SOCKET:
		os.Remove(listenAddr)

		var listener *net.UnixListener
		listener, err = net.ListenUnix("unix", &net.UnixAddr{listenAddr, "unix"})
		if err != nil {
			break // Handle error after switch
		}

		// FIXME: add proper implementation of signal capture on all protocols
		// execute this on SIGTERM or SIGINT: listener.Close()
		if err = os.Chmod(listenAddr, os.FileMode(setting.UnixSocketPermission)); err != nil {
			log.Fatal(4, "Failed to set permission of unix socket: %v", err)
		}
		err = http.Serve(listener, m)
	default:
		log.Fatal(4, "Invalid protocol: %s", setting.Protocol)
	}

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}

	return nil
}
