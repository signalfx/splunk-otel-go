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
	"flag"
	"io"
	"strings"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/x/cmd"
)

var (
	flagCommit = flag.String("commit", "", "git commit to be tagged (used by: release)")
	flagRemote = flag.String("remote", "", "git remote to be used (used by: release)")
)

var _ = goyek.Define(goyek.Task{
	Name:  "release",
	Usage: "publish Go modules",
	Action: func(a *goyek.A) {
		if *flagCommit == "" {
			a.Fatal("flag commit is required")
		}
		if *flagRemote == "" {
			a.Fatal("flag remote is required")
		}

		if !cmd.Exec(a, "go install go.opentelemetry.io/build-tools/multimod", cmd.Dir(dirBuild)) {
			return
		}

		if !cmd.Exec(a, "multimod verify") {
			return
		}

		sb := &strings.Builder{}
		out := io.MultiWriter(a.Output(), sb)
		if !cmd.Exec(a, "multimod tag -m stable-v1 --print-tags --commit-hash "+*flagCommit, cmd.Stdout(out)) {
			return
		}

		tags := strings.Split(sb.String(), "\n")
		for _, tag := range tags {
			cmd.Exec(a, "git push "+*flagRemote+" "+tag)
		}
	},
})
