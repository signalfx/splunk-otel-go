// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io"
	"strings"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var diff = goyek.Define(goyek.Task{
	Name:  "diff",
	Usage: "git diff",
	Action: func(a *goyek.A) {
		cmd.Exec(a, "git diff --exit-code")

		sb := &strings.Builder{}
		out := io.MultiWriter(a.Output(), sb)
		cmd.Exec(a, "git status --porcelain", cmd.Stdout(out), cmd.Stderr(out))
		if sb.Len() > 0 {
			a.Error("git status --porcelain returned output")
		}
	},
})
