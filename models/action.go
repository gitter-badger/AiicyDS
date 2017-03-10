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

import "time"

type ActionType int

// Action represents user operation type and other information to repository,
// it implemented interface base.Actioner so that can be used in template render.
type Action struct {
	ID           int64 `xorm:"pk autoincr"`
	UserID       int64 // Receiver user id.
	OpType       ActionType
	ActUserID    int64  // Action user id.
	ActUserName  string // Action user name.
	ActAvatar    string `xorm:"-"`
	RepoID       int64
	RepoUserName string
	RepoName     string
	RefName      string
	IsPrivate    bool      `xorm:"NOT NULL DEFAULT false"`
	Content      string    `xorm:"TEXT"`
	Created      time.Time `xorm:"-"`
	CreatedUnix  int64
}

// GetFeeds returns action list of given user in given context.
// actorID is the user who's requesting, ctxUserID is the user/org that is requested.
// actorID can be -1 when isProfile is true or to skip the permission check.
func GetFeeds(ctxUser *User, actorID, offset int64, isProfile bool) ([]*Action, error) {
	actions := make([]*Action, 0, 20)
	sess := x.Limit(20, int(offset)).Desc("id").Where("user_id = ?", ctxUser.ID)
	if isProfile {
		sess.And("is_private = ?", false).And("act_user_id = ?", ctxUser.ID)
	}

	err := sess.Find(&actions)
	return actions, err
}
