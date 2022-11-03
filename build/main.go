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

// Build is the build pipeline for this repository.
package main

import (
	"flag"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/boot"
)

const (
	dirBuild          = "build"
	repoPackagePrefix = "github.com/signalfx/splunk-otel-go"
)

var flagSkipDocker = flag.Bool("skip-docker", false, "skip tasks and tests using Docker")

func main() {
	goyek.SetDefault(all)
	boot.Main()
}
