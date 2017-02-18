// Copyright 2017 The Aiicy Team. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"os"

	"github.com/Unknwon/com"
	"github.com/Unknwon/log"

	"github.com/peachdocs/peach/modules/setting"
)

func NewContext() {
	if com.IsExist(HTMLRoot) {
		if err := os.RemoveAll(HTMLRoot); err != nil {
			log.Fatal("Fail to clean up HTMLRoot: %v", err)
		}
	}

	if err := ReloadDocs(); err != nil {
		log.Fatal("Fail to init docs: %v", err)
	}
}