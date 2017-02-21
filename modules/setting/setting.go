// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package setting

import (
	"net/mail"
	"os"
	"path"
	"strings"

	"github.com/Aiicy/AiicyDS/modules/user"
	"github.com/Unknwon/com"
	"github.com/go-macaron/session"
	log "gopkg.in/clog.v1"
	"gopkg.in/ini.v1"
	"gopkg.in/macaron.v1"

	"github.com/Aiicy/AiicyDS/modules/bindata"
	"github.com/gogits/go-libravatar"
)

type Scheme string

const (
	SCHEME_HTTP        Scheme = "http"
	SCHEME_HTTPS       Scheme = "https"
	SCHEME_FCGI        Scheme = "fcgi"
	SCHEME_UNIX_SOCKET Scheme = "unix"
)

type LandingPage string

const (
	LANDING_PAGE_HOME    LandingPage = "/"
	LANDING_PAGE_EXPLORE LandingPage = "/explore"
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
	// Build information should only be set by -ldflags.
	BuildTime string

	// Picture settings
	AvatarUploadPath      string
	GravatarSource        string
	DisableGravatar       bool
	EnableFederatedAvatar bool
	LibravatarService     *libravatar.Libravatar

	// Log settings
	LogRootPath string
	LogModes    []string
	LogConfigs  []interface{}

	// Time settings
	TimeFormat string

	// Cache settings
	CacheAdapter  string
	CacheInterval int
	CacheConn     string

	// Session settings
	SessionConfig  session.Options
	CSRFCookieName = "_csrf"

	AppVer         string
	AppName        string
	AppUrl         string
	AppSubUrl      string
	AppSubUrlDepth int // Number of slashes

	// Server settings
	Protocol           Scheme
	Domain             string
	HTTPAddr, HTTPPort string
	OfflineMode        bool
	DisableRouterLog   bool
	StaticRootPath     string
	EnableGzip         bool
	LandingPageURL     LandingPage

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
	InstallLock    bool
	SecretKey      string
	CookieUserName string

	// Database settings
	UseSQLite3    bool
	UseMySQL      bool
	UsePostgreSQL bool
	UseMSSQL      bool

	// Webhook settings
	Webhook struct {
		QueueLength    int
		DeliverTimeout int
		SkipTLSVerify  bool
		Types          []string
		PagingNum      int
	}

	// Markdown sttings
	Markdown struct {
		EnableHardLineBreak bool
		CustomURLSchemes    []string `ini:"CUSTOM_URL_SCHEMES"`
		FileExtensions      []string
	}

	// UI settings
	UI struct {
		ExplorePagingNum   int
		IssuePagingNum     int
		FeedMaxCommitNum   int
		ThemeColorMetaTag  string
		MaxDisplayFileSize int64

		Admin struct {
			UserPagingNum   int
			RepoPagingNum   int
			NoticePagingNum int
			OrgPagingNum    int
		} `ini:"ui.admin"`
		User struct {
			RepoPagingNum int
		} `ini:"ui.user"`
	}

	// I18n settings
	Langs, Names []string

	// Other settings
	SupportMiniWinService bool

	// template seeting
	ShowFooterTemplateLoadTime bool

	// Global setting objects
	CustomPath string // Custom directory path
	ProdMode   bool
	RunUser    string
	IsWindows  bool

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

// IsRunUserMatchCurrentUser returns false if configured run user does not match
// actual user that runs the app. The first return value is the actual user name.
// This check is ignored under Windows since SSH remote login is not the main
// method to login on Windows.
func IsRunUserMatchCurrentUser(runUser string) (string, bool) {
	if IsWindows {
		return "", true
	}

	currentUser := user.CurrentUsername()
	return currentUser, runUser == currentUser
}

func NewContext() {

	if !com.IsFile(CustomConf) {
		log.Fatal(4, "No custom configuration found: 'custom/app.ini'")
	}
	sources := []interface{}{bindata.MustAsset("conf/app.ini"), CustomConf}

	var err error
	Cfg, err = macaron.SetConfig(sources[0], sources[1:]...)
	if err != nil {
		log.Fatal(4, "Fail to load config: %v", err)
	}

	sec := Cfg.Section("")
	if sec.Key("RUN_MODE").String() == "prod" {
		ProdMode = true
		macaron.Env = macaron.PROD
		macaron.ColorLog = false
	}

	HTTPPort = sec.Key("HTTP_PORT").MustString("3000")
	OfflineMode = sec.Key("OFFLINE_MODE").MustBool()

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

func newService() {
	sec := Cfg.Section("service")
	Service.ActiveCodeLives = sec.Key("ACTIVE_CODE_LIVE_MINUTES").MustInt(180)
	Service.ResetPwdCodeLives = sec.Key("RESET_PASSWD_CODE_LIVE_MINUTES").MustInt(180)
	Service.DisableRegistration = sec.Key("DISABLE_REGISTRATION").MustBool()
	Service.ShowRegistrationButton = sec.Key("SHOW_REGISTRATION_BUTTON").MustBool(!Service.DisableRegistration)
	Service.RequireSignInView = sec.Key("REQUIRE_SIGNIN_VIEW").MustBool()
	Service.EnableReverseProxyAuth = sec.Key("ENABLE_REVERSE_PROXY_AUTHENTICATION").MustBool()
	Service.EnableReverseProxyAutoRegister = sec.Key("ENABLE_REVERSE_PROXY_AUTO_REGISTRATION").MustBool()
	Service.EnableCaptcha = sec.Key("ENABLE_CAPTCHA").MustBool()
}

func newLogService() {
	if len(BuildTime) > 0 {
		log.Trace("Build Time: %s", BuildTime)
	}

	// Because we always create a console logger as primary logger before all settings are loaded,
	// thus if user doesn't set console logger, we should remove it after other loggers are created.
	hasConsole := false

	// Get and check log modes.
	LogModes = strings.Split(Cfg.Section("log").Key("MODE").MustString("console"), ",")
	LogConfigs = make([]interface{}, len(LogModes))
	levelNames := map[string]log.LEVEL{
		"trace": log.TRACE,
		"info":  log.INFO,
		"warn":  log.WARN,
		"error": log.ERROR,
		"fatal": log.FATAL,
	}
	for i, mode := range LogModes {
		mode = strings.ToLower(strings.TrimSpace(mode))
		sec, err := Cfg.GetSection("log." + mode)
		if err != nil {
			log.Fatal(4, "Unknown logger mode: %s", mode)
		}

		validLevels := []string{"trace", "info", "warn", "error", "fatal"}
		name := Cfg.Section("log." + mode).Key("LEVEL").Validate(func(v string) string {
			v = strings.ToLower(v)
			if com.IsSliceContainsStr(validLevels, v) {
				return v
			}
			return "trace"
		})
		level := levelNames[name]

		// Generate log configuration.
		switch log.MODE(mode) {
		case log.CONSOLE:
			hasConsole = true
			LogConfigs[i] = log.ConsoleConfig{
				Level:      level,
				BufferSize: Cfg.Section("log").Key("BUFFER_LEN").MustInt64(100),
			}

		case log.FILE:
			logPath := path.Join(LogRootPath, "gogs.log")
			if err = os.MkdirAll(path.Dir(logPath), os.ModePerm); err != nil {
				log.Fatal(4, "Fail to create log directory '%s': %v", path.Dir(logPath), err)
			}

			LogConfigs[i] = log.FileConfig{
				Level:      level,
				BufferSize: Cfg.Section("log").Key("BUFFER_LEN").MustInt64(100),
				Filename:   logPath,
				FileRotationConfig: log.FileRotationConfig{
					Rotate:   sec.Key("LOG_ROTATE").MustBool(true),
					Daily:    sec.Key("DAILY_ROTATE").MustBool(true),
					MaxSize:  1 << uint(sec.Key("MAX_SIZE_SHIFT").MustInt(28)),
					MaxLines: sec.Key("MAX_LINES").MustInt64(1000000),
					MaxDays:  sec.Key("MAX_DAYS").MustInt64(7),
				},
			}

		case log.SLACK:
			LogConfigs[i] = log.SlackConfig{
				Level:      level,
				BufferSize: Cfg.Section("log").Key("BUFFER_LEN").MustInt64(100),
				URL:        sec.Key("URL").String(),
			}
		}

		log.New(log.MODE(mode), LogConfigs[i])
		log.Trace("Log Mode: %s (%s)", strings.Title(mode), strings.Title(name))
	}

	// Make sure everyone gets version info printed.
	log.Info("%s %s", AppName, AppVer)
	if !hasConsole {
		log.Delete(log.CONSOLE)
	}
}

func newCacheService() {
	CacheAdapter = Cfg.Section("cache").Key("ADAPTER").In("memory", []string{"memory", "redis", "memcache"})
	switch CacheAdapter {
	case "memory":
		CacheInterval = Cfg.Section("cache").Key("INTERVAL").MustInt(60)
	case "redis", "memcache":
		CacheConn = strings.Trim(Cfg.Section("cache").Key("HOST").String(), "\" ")
	default:
		log.Fatal(4, "Unknown cache adapter: %s", CacheAdapter)
	}

	log.Info("Cache Service Enabled")
}

func newSessionService() {
	SessionConfig.Provider = Cfg.Section("session").Key("PROVIDER").In("memory",
		[]string{"memory", "file", "redis", "mysql"})
	SessionConfig.ProviderConfig = strings.Trim(Cfg.Section("session").Key("PROVIDER_CONFIG").String(), "\" ")
	SessionConfig.CookieName = Cfg.Section("session").Key("COOKIE_NAME").MustString("i_like_gogits")
	SessionConfig.CookiePath = AppSubUrl
	SessionConfig.Secure = Cfg.Section("session").Key("COOKIE_SECURE").MustBool()
	SessionConfig.Gclifetime = Cfg.Section("session").Key("GC_INTERVAL_TIME").MustInt64(86400)
	SessionConfig.Maxlifetime = Cfg.Section("session").Key("SESSION_LIFE_TIME").MustInt64(86400)

	log.Info("Session Service Enabled")
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

func newMailService() {
	sec := Cfg.Section("mailer")
	// Check mailer setting.
	if !sec.Key("ENABLED").MustBool() {
		return
	}

	MailService = &Mailer{
		QueueLength:           sec.Key("SEND_BUFFER_LEN").MustInt(100),
		Name:                  sec.Key("NAME").MustString(AppName),
		Host:                  sec.Key("HOST").String(),
		User:                  sec.Key("USER").String(),
		Passwd:                sec.Key("PASSWD").String(),
		DisableHelo:           sec.Key("DISABLE_HELO").MustBool(),
		HeloHostname:          sec.Key("HELO_HOSTNAME").String(),
		SkipVerify:            sec.Key("SKIP_VERIFY").MustBool(),
		UseCertificate:        sec.Key("USE_CERTIFICATE").MustBool(),
		CertFile:              sec.Key("CERT_FILE").String(),
		KeyFile:               sec.Key("KEY_FILE").String(),
		EnableHTMLAlternative: sec.Key("ENABLE_HTML_ALTERNATIVE").MustBool(),
	}
	MailService.From = sec.Key("FROM").MustString(MailService.User)

	parsed, err := mail.ParseAddress(MailService.From)
	if err != nil {
		log.Fatal(4, "Invalid mailer.FROM (%s): %v", MailService.From, err)
	}
	MailService.FromEmail = parsed.Address

	log.Info("Mail Service Enabled")
}

func newRegisterMailService() {
	if !Cfg.Section("service").Key("REGISTER_EMAIL_CONFIRM").MustBool() {
		return
	} else if MailService == nil {
		log.Warn("Register Mail Service: Mail Service is not enabled")
		return
	}
	Service.RegisterEmailConfirm = true
	log.Info("Register Mail Service Enabled")
}

func newNotifyMailService() {
	if !Cfg.Section("service").Key("ENABLE_NOTIFY_MAIL").MustBool() {
		return
	} else if MailService == nil {
		log.Warn("Notify Mail Service: Mail Service is not enabled")
		return
	}
	Service.EnableNotifyMail = true
	log.Info("Notify Mail Service Enabled")
}

func newWebhookService() {
	sec := Cfg.Section("webhook")
	Webhook.QueueLength = sec.Key("QUEUE_LENGTH").MustInt(1000)
	Webhook.DeliverTimeout = sec.Key("DELIVER_TIMEOUT").MustInt(5)
	Webhook.SkipTLSVerify = sec.Key("SKIP_TLS_VERIFY").MustBool()
	Webhook.Types = []string{"gogs", "slack"}
	Webhook.PagingNum = sec.Key("PAGING_NUM").MustInt(10)
}

func NewService() {
	newService()
}

func NewServices() {
	newService()
	newLogService()
	newCacheService()
	newSessionService()
	newMailService()
	newRegisterMailService()
	newNotifyMailService()
	newWebhookService()
}
