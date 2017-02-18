// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package setting

import (
	"github.com/Unknwon/com"
	"github.com/Unknwon/log"
	"gopkg.in/ini.v1"
	"gopkg.in/macaron.v1"

	"github.com/Aiicy/AiicyDS/modules/bindata"
)

type NavbarItem struct {
	Icon         string
	Locale, Link string
	Blank        bool
}

const (
	LOCAL  = "local"
	REMOTE = "remote"
)

type DocType string

func (t DocType) IsLocal() bool {
	return t == LOCAL
}

func (t DocType) IsRemote() bool {
	return t == REMOTE
}

var (
	CustomConf = "custom/app.ini"

	AppVer         string
	AppUrl         string
	AppSubUrl      string
	AppSubUrlDepth int // Number of slashes
	ProdMode       bool
	HTTPPort       int

	Site struct {
		Name   string
		Desc   string
		UseCDN bool
		URL    string
	}

	Page struct {
		HasLandingPage bool
		DocsBaseURL    string

		UseCustomTpl   bool
		NavbarTplPath  string
		HomeTplPath    string
		DocsTplPath    string
		FooterTplPath  string
		DisqusTplPath  string
		DuoShuoTplPath string
	}

	Navbar struct {
		Items []*NavbarItem
	}

	Asset struct {
		CustomCSS string
	}

	Docs struct {
		Type   DocType
		Target string
		Secret string
		Langs  []string

		// Only used for languages are not en-US or zh-CN to bypass error panic.
		Locales map[string][]byte
	}
	// Security settings
	InstallLock bool

	// Markdown sttings
	Markdown struct {
		EnableHardLineBreak bool
		CustomURLSchemes    []string `ini:"CUSTOM_URL_SCHEMES"`
		FileExtensions      []string
	}

	Extension struct {
		EnableEditPage       bool
		EditPageLinkFormat   string
		EnableDisqus         bool
		DisqusShortName      string
		EnableDuoShuo        bool
		DuoShuoShortName     string
		HighlightJSCustomCSS string
		EnableSearch         bool
		GABlock              string
	}

	Cfg *ini.File
)

var Service struct {
	ActiveCodeLives                int
	ResetPwdCodeLives              int
	RegisterEmailConfirm           bool
	DisableRegistration            bool
	ShowRegistrationButton         bool
	RequireSignInView              bool
	EnableNotifyMail               bool
	EnableReverseProxyAuth         bool
	EnableReverseProxyAutoRegister bool
	EnableCaptcha                  bool
}

// Mailer represents mail service.
type Mailer struct {
	QueueLength           int
	Name                  string
	Host                  string
	From                  string
	FromEmail             string
	User, Passwd          string
	DisableHelo           bool
	HeloHostname          string
	SkipVerify            bool
	UseCertificate        bool
	CertFile, KeyFile     string
	EnableHTMLAlternative bool
}

var (
	MailService *Mailer
)

func NewContext() {
	log.Prefix = "[Peach]"

	if !com.IsFile(CustomConf) {
		log.Fatal("No custom configuration found: 'custom/app.ini'")
	}
	sources := []interface{}{bindata.MustAsset("conf/app.ini"), CustomConf}

	var err error
	Cfg, err = macaron.SetConfig(sources[0], sources[1:]...)
	if err != nil {
		log.Fatal("Fail to load config: %v", err)
	}

	sec := Cfg.Section("")
	if sec.Key("RUN_MODE").String() == "prod" {
		ProdMode = true
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
	}

	HTTPPort = sec.Key("HTTP_PORT").MustInt(5555)

	sec = Cfg.Section("site")
	Site.Name = sec.Key("NAME").MustString("Peach Server")
	Site.Desc = sec.Key("DESC").String()
	Site.UseCDN = sec.Key("USE_CDN").MustBool()
	Site.URL = sec.Key("URL").String()

	sec = Cfg.Section("page")
	Page.HasLandingPage = sec.Key("HAS_LANDING_PAGE").MustBool()
	Page.DocsBaseURL = sec.Key("DOCS_BASE_URL").Validate(func(in string) string {
		if len(in) == 0 {
			return "/docs"
		} else if in[0] != '/' {
			return "/" + in
		}
		return in
	})

	Page.UseCustomTpl = sec.Key("USE_CUSTOM_TPL").MustBool()
	Page.NavbarTplPath = "navbar.html"
	Page.HomeTplPath = "home.html"
	Page.DocsTplPath = "docs.html"
	Page.FooterTplPath = "footer.html"
	Page.DisqusTplPath = "disqus.html"
	Page.DuoShuoTplPath = "duoshuo.html"

	sec = Cfg.Section("navbar")
	list := sec.KeyStrings()
	Navbar.Items = make([]*NavbarItem, len(list))
	for i, name := range list {
		secName := "navbar." + sec.Key(name).String()
		Navbar.Items[i] = &NavbarItem{
			Icon:   Cfg.Section(secName).Key("ICON").String(),
			Locale: Cfg.Section(secName).Key("LOCALE").MustString(secName),
			Link:   Cfg.Section(secName).Key("LINK").MustString("/"),
			Blank:  Cfg.Section(secName).Key("BLANK").MustBool(),
		}
	}

	sec = Cfg.Section("asset")
	Asset.CustomCSS = sec.Key("CUSTOM_CSS").String()

	sec = Cfg.Section("docs")
	Docs.Type = DocType(sec.Key("TYPE").In("local", []string{LOCAL, REMOTE}))
	Docs.Target = sec.Key("TARGET").String()
	Docs.Secret = sec.Key("SECRET").String()
	Docs.Langs = Cfg.Section("i18n").Key("LANGS").Strings(",")
	Docs.Locales = make(map[string][]byte)
	for _, lang := range Docs.Langs {
		if lang == "en-US" || lang == "zh-CN" {
			Docs.Locales["locale_"+lang+".ini"] = bindata.MustAsset("conf/locale/locale_" + lang + ".ini")
		} else {
			Docs.Locales["locale_"+lang+".ini"] = []byte("")
		}
	}

	sec = Cfg.Section("extension")
	Extension.EnableEditPage = sec.Key("ENABLE_EDIT_PAGE").MustBool()
	Extension.EditPageLinkFormat = sec.Key("EDIT_PAGE_LINK_FORMAT").String()
	Extension.EnableDisqus = sec.Key("ENABLE_DISQUS").MustBool()
	Extension.DisqusShortName = sec.Key("DISQUS_SHORT_NAME").String()
	Extension.EnableDuoShuo = sec.Key("ENABLE_DUOSHUO").MustBool()
	Extension.DuoShuoShortName = sec.Key("DUOSHUO_SHORT_NAME").String()
	Extension.HighlightJSCustomCSS = sec.Key("HIGHLIGHTJS_CUSTOM_CSS").String()
	Extension.EnableSearch = sec.Key("ENABLE_SEARCH").MustBool()
	Extension.GABlock = sec.Key("GA_BLOCK").String()
}
