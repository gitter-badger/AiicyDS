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

package models

import "github.com/Aiicy/AiicyDS/modules/mailer"

// mailerUser is a wrapper for satisfying mailer.User interface.
type mailerUser struct {
	user *User
}

func (this mailerUser) ID() int64 {
	return this.user.ID
}

func (this mailerUser) DisplayName() string {
	return this.user.DisplayName()
}

func (this mailerUser) Email() string {
	return this.user.Email
}

func (this mailerUser) GenerateActivateCode() string {
	return this.user.GenerateActivateCode()
}

func (this mailerUser) GenerateEmailActivateCode(email string) string {
	return this.user.GenerateEmailActivateCode(email)
}

func NewMailerUser(u *User) mailer.User {
	return mailerUser{u}
}