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

package avatar

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_RandomImage(t *testing.T) {
	Convey("Generate a random avatar from email", t, func() {
		_, err := RandomImage([]byte("AiicyDS@local"))
		So(err, ShouldBeNil)

		Convey("Try to generate an image with size zero", func() {
			_, err := RandomImageSize(0, []byte("aiicyds@local"))
			So(err, ShouldNotBeNil)
		})
	})
}
