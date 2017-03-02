// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"container/list"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"mime"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Aiicy/AiicyDS/modules/markdown"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/editorconfig/editorconfig-core-go.v1"

	"github.com/Aiicy/AiicyDS/modules/base"
	"github.com/Aiicy/AiicyDS/modules/setting"
)

func NewFuncMap() []template.FuncMap {
	return []template.FuncMap{map[string]interface{}{
		"GoVer": func() string {
			return strings.Title(runtime.Version())
		},
		"UseHTTPS": func() bool {
			return strings.HasPrefix(setting.AppUrl, "https")
		},
		"AppName": func() string {
			return setting.AppName
		},
		"AppSubUrl": func() string {
			return setting.AppSubUrl
		},
		"AppUrl": func() string {
			return setting.AppUrl
		},
		"AppVer": func() string {
			return setting.AppVer
		},
		"AppDomain": func() string {
			return setting.Domain
		},
		"DisableGravatar": func() bool {
			return setting.DisableGravatar
		},
		"ShowFooterTemplateLoadTime": func() bool {
			return setting.ShowFooterTemplateLoadTime
		},
		"LoadTimes": func(startTime time.Time) string {
			return fmt.Sprint(time.Since(startTime).Nanoseconds()/1e6) + "ms"
		},
		"AvatarLink":   base.AvatarLink,
		"Safe":         Safe,
		"Sanitize":     bluemonday.UGCPolicy().Sanitize,
		"Str2html":     Str2html,
		"TimeSince":    base.TimeSince,
		"RawTimeSince": base.RawTimeSince,
		"FileSize":     base.FileSize,
		"Subtract":     base.Subtract,
		"add": func(nums ...interface{}) int {
			total := 0
			for _, num := range nums {
				if n, ok := num.(int); ok {
					total += n
				}
			}
			return total
		},
		"DateFmtLong": func(t time.Time) string {
			return t.Format(time.RFC1123Z)
		},
		"DateFmtShort": func(t time.Time) string {
			return t.Format("Jan 02, 2006")
		},
		"List": List,
		"substring": func(str string, start, length int) string {
			if len(str) == 0 {
				return ""
			}
			end := start + length
			if length == -1 {
				end = len(str)
			}
			if len(str) < end {
				return str
			}
			return str[start:end]
		},
		"Join":        strings.Join,
		"MD5":         base.EncodeMD5,
		"EscapePound": EscapePound,
		"ThemeColorMetaTag": func() string {
			return setting.UI.ThemeColorMetaTag
		},
		"FilenameIsImage": func(filename string) bool {
			mimeType := mime.TypeByExtension(filepath.Ext(filename))
			return strings.HasPrefix(mimeType, "image/")
		},
		"TabSizeClass": func(ec *editorconfig.Editorconfig, filename string) string {
			if ec != nil {
				def := ec.GetDefinitionForFilename(filename)
				if def.TabWidth > 0 {
					return fmt.Sprintf("tab-size-%d", def.TabWidth)
				}
			}
			return "tab-size-8"
		},
	}}
}

func Safe(raw string) template.HTML {
	return template.HTML(raw)
}

func Str2html(raw string) template.HTML {
	return template.HTML(markdown.Sanitizer.Sanitize(raw))
}

func List(l *list.List) chan interface{} {
	e := l.Front()
	c := make(chan interface{})
	go func() {
		for e != nil {
			c <- e.Value
			e = e.Next()
		}
		close(c)
	}()
	return c
}

func EscapePound(str string) string {
	return strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F").Replace(str)
}

const qiniuDomain = "http://studygolang.qiniudn.com"

// 获取头像
func Gravatar(avatar string, emailI interface{}, size uint16) string {
	if avatar != "" {
		return fmt.Sprintf("%s/avatar/%s?imageView2/2/w/%d", qiniuDomain, avatar, size)
	}

	email, ok := emailI.(string)
	if !ok {
		return fmt.Sprintf("%s/avatar/gopher28.png?imageView2/2/w/%d", qiniuDomain, size)
	}
	return fmt.Sprintf("http://gravatar.duoshuo.com/avatar/%s?s=%d", Md5(email), size)
}

func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}
