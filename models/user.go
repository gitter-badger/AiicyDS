// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/pbkdf2"

	"github.com/Aiicy/AiicyDS/modules/base"
	"github.com/polaris1119/goutils"
)

type UserType int

const (
	USER_TYPE_INDIVIDUAL UserType = iota // Historic reason to make it starts at 0.
	USER_TYPE_ORGANIZATION
)

var (
	ErrUserNotKeyOwner       = errors.New("User does not the owner of public key")
	ErrEmailNotExist         = errors.New("E-mail does not exist")
	ErrEmailNotActivated     = errors.New("E-mail address has not been activated")
	ErrUserNameIllegal       = errors.New("User name contains illegal characters")
	ErrLoginSourceNotActived = errors.New("Login source is not actived")
	ErrUnsupportedLoginType  = errors.New("Login source is unknown")
)

// 用户登录信息
type UserLogin struct {
	Uid       int       `json:"uid" xorm:"pk"`
	Username  string    `json:"username"`
	Passwd    string    `json:"passwd"`
	Email     string    `json:"email"`
	LoginTime time.Time `json:"login_time" xorm:"<-"`
	Passcode  string    `json:"passcode"` // 加密随机串
}

func (this *UserLogin) TableName() string {
	return "user_login"
}

// 生成加密密码
func (this *UserLogin) GenMd5Passwd() error {
	if this.Passwd == "" {
		return errors.New("password is empty!")
	}
	this.Passcode = fmt.Sprintf("%x", rand.Int31())
	// 密码经过md5(passwd+passcode)加密保存
	this.Passwd = goutils.Md5(this.Passwd + this.Passcode)
	return nil
}

const (
	UserStatusNoAudit = iota
	UserStatusAudit   // 已激活
	UserStatusRefuse
	UserStatusFreeze // 冻结
	UserStatusOutage // 停用
)

// 用户基本信息
type User struct {
	ID          int64     `xorm:"pk autoincr"`
	Uid         int       `json:"uid" xorm:"pk autoincr"`
	Username    string    `json:"username" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	LowerName   string    `xorm:"UNIQUE NOT NULL"`
	Passwd      string    `xorm:"NOT NULL"`
	Rands       string    `xorm:"VARCHAR(10)"`
	Salt        string    `xorm:"VARCHAR(10)"`
	Email       string    `json:"email"`
	Open        int       `json:"open"`
	Name        string    `json:"name"`
	City        string    `json:"city"`
	Company     string    `json:"company"`
	Github      string    `json:"github"`
	Weibo       string    `json:"weibo"`
	Website     string    `json:"website"`
	Monlog      string    `json:"monlog"`
	Introduce   string    `json:"introduce"`
	Unsubscribe int       `json:"unsubscribe"`
	Status      int       `json:"status"`
	IsRoot      bool      `json:"is_root"`
	Ctime       OftenTime `json:"ctime" xorm:"created"`
	Mtime       time.Time `json:"mtime" xorm:"<-"`

	// Avatar
	Avatar      string `xorm:"VARCHAR(2048) NOT NULL"`
	AvatarEmail string `xorm:"NOT NULL"`

	// 非用户表中的信息，为了方便放在这里
	Roleids   []int    `xorm:"-"`
	Rolenames []string `xorm:"-"`

	// Permissions
	IsActive      bool // Activate primary email
	IsAdmin       bool
	ProhibitLogin bool
}

func (this *User) TableName() string {
	return "user_info"
}

func (this *User) String() string {
	buffer := goutils.NewBuffer()
	buffer.Append(this.Username).Append(this.Email).Append(this.Uid).Append(this.Mtime)

	return buffer.String()
}

// Me 代表当前用户
type Me struct {
	Uid      int    `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
	MsgNum   int    `json:"msgnum"`
	IsAdmin  bool   `json:"isadmin"`
	IsRoot   bool   `json:"is_root"`
}

// 活跃用户信息
// 活跃度规则：
//	1、注册成功后 +2
//	2、登录一次 +1
//	3、修改资料 +1
//	4、发帖子 + 10
//	5、评论 +5
//	6、创建Wiki页 +10
type UserActive struct {
	Uid      int       `json:"uid" xorm:"pk"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
	Weight   int       `json:"weight"`
	Mtime    time.Time `json:"mtime" xorm:"<-"`
}

// 用户角色信息
type UserRole struct {
	Uid    int    `json:"uid"`
	Roleid int    `json:"roleid"`
	ctime  string `xorm:"-"`
}

// EncodePasswd encodes password to safe format.
func (u *User) EncodePasswd() {
	newPasswd := pbkdf2.Key([]byte(u.Passwd), []byte(u.Salt), 10000, 50, sha256.New)
	u.Passwd = fmt.Sprintf("%x", newPasswd)
}

// IsUserExist checks if given user name exist,
// the user name should be noncased unique.
// If uid is presented, then check will rule out that one,
// it is used when update a user name in settings page.
func IsUserExist(uid int64, name string) (bool, error) {
	if len(name) == 0 {
		return false, nil
	}
	return x.Where("id!=?", uid).Get(&User{LowerName: strings.ToLower(name)})
}

// GetUserSalt returns a ramdom user salt token.
func GetUserSalt() (string, error) {
	return base.GetRandomString(10)
}

var (
	reservedUsernames    = []string{"assets", "css", "img", "js", "less", "plugins", "debug", "raw", "install", "api", "avatar", "user", "org", "help", "stars", "issues", "pulls", "commits", "repo", "template", "admin", "new", ".", ".."}
	reservedUserPatterns = []string{"*.keys"}
)

// isUsableName checks if name is reserved or pattern of name is not allowed
// based on given reserved names and patterns.
// Names are exact match, patterns can be prefix or suffix match with placeholder '*'.
func isUsableName(names, patterns []string, name string) error {
	name = strings.TrimSpace(strings.ToLower(name))
	if utf8.RuneCountInString(name) == 0 {
		return ErrNameEmpty
	}

	for i := range names {
		if name == names[i] {
			return ErrNameReserved{name}
		}
	}

	for _, pat := range patterns {
		if pat[0] == '*' && strings.HasSuffix(name, pat[1:]) ||
			(pat[len(pat)-1] == '*' && strings.HasPrefix(name, pat[:len(pat)-1])) {
			return ErrNamePatternNotAllowed{pat}
		}
	}

	return nil
}

func IsUsableUsername(name string) error {
	return isUsableName(reservedUsernames, reservedUserPatterns, name)
}

// CreateUser creates record of a new user.
func CreateUser(u *User) (err error) {
	if err = IsUsableUsername(u.Name); err != nil {
		return err
	}

	isExist, err := IsUserExist(0, u.Name)
	if err != nil {
		return err
	} else if isExist {
		return ErrUserAlreadyExist{u.Name}
	}

	u.Email = strings.ToLower(u.Email)
	isExist, err = IsEmailUsed(u.Email)
	if err != nil {
		return err
	} else if isExist {
		return ErrEmailAlreadyUsed{u.Email}
	}

	u.LowerName = strings.ToLower(u.Name)
	u.AvatarEmail = u.Email
	u.Avatar = base.HashEmail(u.AvatarEmail)
	if u.Rands, err = GetUserSalt(); err != nil {
		return err
	}
	if u.Salt, err = GetUserSalt(); err != nil {
		return err
	}
	u.EncodePasswd()

	sess := x.NewSession()
	defer sessionRelease(sess)
	if err = sess.Begin(); err != nil {
		return err
	}

	if _, err = sess.Insert(u); err != nil {
		return err
	} else if err = os.MkdirAll(UserPath(u.Name), os.ModePerm); err != nil {
		return err
	}

	return sess.Commit()
}

// UserPath returns the path absolute path of user repositories.
func UserPath(userName string) string {
	return filepath.Join(strings.ToLower(userName))
}

func GetUserByKeyID(keyID int64) (*User, error) {
	user := new(User)
	has, err := x.Sql("SELECT a.* FROM `user` AS a, public_key AS b WHERE a.id = b.owner_id AND b.id=?", keyID).Get(user)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrUserNotKeyOwner
	}
	return user, nil
}

func getUserByID(e Engine, id int64) (*User, error) {
	u := new(User)
	has, err := e.Id(id).Get(u)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrUserNotExist{id, ""}
	}
	return u, nil
}

// GetUserByID returns the user object by given ID if exists.
func GetUserByID(id int64) (*User, error) {
	return getUserByID(x, id)
}

// GetUserByName returns user by given name.
func GetUserByName(name string) (*User, error) {
	if len(name) == 0 {
		return nil, ErrUserNotExist{0, name}
	}
	u := &User{LowerName: strings.ToLower(name)}
	has, err := x.Get(u)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrUserNotExist{0, name}
	}
	return u, nil
}
