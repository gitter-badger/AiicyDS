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

package base

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"html/template"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/Aiicy/AiicyDS/modules/setting"
	"github.com/Unknwon/com"
	"github.com/Unknwon/i18n"
	log "gopkg.in/clog.v1"
)

const DOC_URL = "https://github.com/Aiicy/AiicyDS/wiki"

type (
	TplName string
)

// EncodeMD5 encodes string to md5 hex value.
func EncodeMD5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}

func ShortSha(sha1 string) string {
	if len(sha1) > 10 {
		return sha1[:10]
	}
	return sha1
}

// GetRandomString generate random string by specify chars.
func GetRandomString(n int) (string, error) {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))

	for i := 0; i < n; i++ {
		index, err := randomInt(max)
		if err != nil {
			return "", err
		}

		buffer[i] = alphanum[index]
	}

	return string(buffer), nil
}

func randomInt(max *big.Int) (int, error) {
	rand, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return int(rand.Int64()), nil
}

// verify time limit code
func VerifyTimeLimitCode(data string, minutes int, code string) bool {
	if len(code) <= 18 {
		return false
	}

	// split code
	start := code[:12]
	lives := code[12:18]
	if d, err := com.StrTo(lives).Int(); err == nil {
		minutes = d
	}

	// right active code
	retCode := CreateTimeLimitCode(data, minutes, start)
	if retCode == code && minutes > 0 {
		// check time is expired or not
		before, _ := time.ParseInLocation("200601021504", start, time.Local)
		now := time.Now()
		if before.Add(time.Minute*time.Duration(minutes)).Unix() > now.Unix() {
			return true
		}
	}

	return false
}

const TimeLimitCodeLength = 12 + 6 + 40

// create a time limit code
// code format: 12 length date time string + 6 minutes string + 40 sha1 encoded string
func CreateTimeLimitCode(data string, minutes int, startInf interface{}) string {
	format := "200601021504"

	var start, end time.Time
	var startStr, endStr string

	if startInf == nil {
		// Use now time create code
		start = time.Now()
		startStr = start.Format(format)
	} else {
		// use start string create code
		startStr = startInf.(string)
		start, _ = time.ParseInLocation(format, startStr, time.Local)
		startStr = start.Format(format)
	}

	end = start.Add(time.Minute * time.Duration(minutes))
	endStr = end.Format(format)

	// create sha1 encode string
	sh := sha1.New()
	sh.Write([]byte(data + setting.SecretKey + startStr + endStr + com.ToStr(minutes)))
	encoded := hex.EncodeToString(sh.Sum(nil))

	code := fmt.Sprintf("%s%06d%s", startStr, minutes, encoded)
	return code
}

// HashEmail hashes email address to MD5 string.
// https://en.gravatar.com/site/implement/hash/
func HashEmail(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	h := md5.New()
	h.Write([]byte(email))
	return hex.EncodeToString(h.Sum(nil))
}

// AvatarLink returns relative avatar link to the site domain by given email,
// which includes app sub-url as prefix. However, it is possible
// to return full URL if user enables Gravatar-like service.
func AvatarLink(email string) (url string) {
	if setting.EnableFederatedAvatar && setting.LibravatarService != nil &&
		strings.Contains(email, "@") {
		var err error
		url, err = setting.LibravatarService.FromEmail(email)
		if err != nil {
			log.Error(2, "LibravatarService.FromEmail [%s]: %v", email, err)
		}
	}
	if len(url) == 0 && !setting.DisableGravatar {
		url = setting.GravatarSource + HashEmail(email)
	}
	if len(url) == 0 {
		url = setting.AppSubUrl + "/img/avatar_default.png"
	}
	return url
}

// Seconds-based time units
const (
	Minute = 60
	Hour   = 60 * Minute
	Day    = 24 * Hour
	Week   = 7 * Day
	Month  = 30 * Day
	Year   = 12 * Month
)

func timeSince(then time.Time, lang string) string {
	now := time.Now()

	lbl := i18n.Tr(lang, "tool.ago")
	diff := now.Unix() - then.Unix()
	if then.After(now) {
		lbl = i18n.Tr(lang, "tool.from_now")
		diff = then.Unix() - now.Unix()
	}

	switch {
	case diff <= 0:
		return i18n.Tr(lang, "tool.now")
	case diff <= 2:
		return i18n.Tr(lang, "tool.1s", lbl)
	case diff < 1*Minute:
		return i18n.Tr(lang, "tool.seconds", diff, lbl)

	case diff < 2*Minute:
		return i18n.Tr(lang, "tool.1m", lbl)
	case diff < 1*Hour:
		return i18n.Tr(lang, "tool.minutes", diff/Minute, lbl)

	case diff < 2*Hour:
		return i18n.Tr(lang, "tool.1h", lbl)
	case diff < 1*Day:
		return i18n.Tr(lang, "tool.hours", diff/Hour, lbl)

	case diff < 2*Day:
		return i18n.Tr(lang, "tool.1d", lbl)
	case diff < 1*Week:
		return i18n.Tr(lang, "tool.days", diff/Day, lbl)

	case diff < 2*Week:
		return i18n.Tr(lang, "tool.1w", lbl)
	case diff < 1*Month:
		return i18n.Tr(lang, "tool.weeks", diff/Week, lbl)

	case diff < 2*Month:
		return i18n.Tr(lang, "tool.1mon", lbl)
	case diff < 1*Year:
		return i18n.Tr(lang, "tool.months", diff/Month, lbl)

	case diff < 2*Year:
		return i18n.Tr(lang, "tool.1y", lbl)
	default:
		return i18n.Tr(lang, "tool.years", diff/Year, lbl)
	}
}

func RawTimeSince(t time.Time, lang string) string {
	return timeSince(t, lang)
}

// TimeSince calculates the time interval and generate user-friendly string.
func TimeSince(t time.Time, lang string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span class="time-since" title="%s">%s</span>`, t.Format(setting.TimeFormat), timeSince(t, lang)))
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func humanateBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}
	e := math.Floor(logn(float64(s), base))
	suffix := sizes[int(e)]
	val := float64(s) / math.Pow(base, math.Floor(e))
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}

	return fmt.Sprintf(f+"%s", val, suffix)
}

// FileSize calculates the file size and generate user-friendly string.
func FileSize(s int64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	return humanateBytes(uint64(s), 1024, sizes)
}

// Subtract deals with subtraction of all types of number.
func Subtract(left interface{}, right interface{}) interface{} {
	var rleft, rright int64
	var fleft, fright float64
	var isInt bool = true
	switch left.(type) {
	case int:
		rleft = int64(left.(int))
	case int8:
		rleft = int64(left.(int8))
	case int16:
		rleft = int64(left.(int16))
	case int32:
		rleft = int64(left.(int32))
	case int64:
		rleft = left.(int64)
	case float32:
		fleft = float64(left.(float32))
		isInt = false
	case float64:
		fleft = left.(float64)
		isInt = false
	}

	switch right.(type) {
	case int:
		rright = int64(right.(int))
	case int8:
		rright = int64(right.(int8))
	case int16:
		rright = int64(right.(int16))
	case int32:
		rright = int64(right.(int32))
	case int64:
		rright = right.(int64)
	case float32:
		fright = float64(left.(float32))
		isInt = false
	case float64:
		fleft = left.(float64)
		isInt = false
	}

	if isInt {
		return rleft - rright
	} else {
		return fleft + float64(rleft) - (fright + float64(rright))
	}
}

// EllipsisString returns a truncated short string,
// it appends '...' in the end of the length of string is too large.
func EllipsisString(str string, length int) string {
	if len(str) < length {
		return str
	}
	return str[:length-3] + "..."
}

// TruncateString returns a truncated string with given limit,
// it returns input string if length is not reached limit.
func TruncateString(str string, limit int) string {
	if len(str) < limit {
		return str
	}
	return str[:limit]
}
