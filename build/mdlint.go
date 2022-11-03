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
	"strings"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var mdlint = goyek.Define(goyek.Task{
	Name:  "mdlint",
	Usage: "markdownlint-cli (uses docker)",
	Action: func(tf *goyek.TF) {
		if *flagSkipDocker {
			tf.Skip("skipping as Docker is needed")
		}

		mdFiles := Find(tf, ".md")
		if len(mdFiles) == 0 {
			tf.Skip("no .md files")
		}

		if !cmd.Exec(tf, "docker build -t markdownlint-cli -f build/markdownlint-cli.dockerfile .") {
			return
		}
		cmd.Exec(tf, "docker run --rm -v '"+WorkDir(tf)+":/workdir' markdownlint-cli "+strings.Join(mdFiles, " "))
	},
})
