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
	"os"
	"path/filepath"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

var test = goyek.Define(goyek.Task{
	Name:  "test",
	Usage: "go test",
	Action: func(a *goyek.A) {
		// prepare test-results
		curDir := WorkDir(a)
		testResultDir := filepath.Join(curDir, "test-results")
		if err := os.RemoveAll(testResultDir); err != nil {
			a.Fatal(err)
		}
		if err := os.Mkdir(testResultDir, 0o750); err != nil { //nolint:gomnd // dir permissions
			a.Fatal(err)
		}

		short := ""
		if *flagSkipDocker {
			short = "-short "
		}

		// run go test race with code covarage for each Go Module
		ForGoModules(a, func(a *goyek.A) {
			const fileNameLen = 12
			covOut := filepath.Join(testResultDir, RandString(a, fileNameLen)+".out")
			cmd.Exec(a, "go test "+short+"-v -race -covermode=atomic -coverprofile='"+covOut+"' -coverpkg="+repoPackagePrefix+"/... ./...")
		})

		// merge the coverage output files into a single coverage.out file
		installGocovmerge := a.Cmd("go", "install", "github.com/wadey/gocovmerge")
		installGocovmerge.Dir = dirBuild
		if err := installGocovmerge.Run(); err != nil {
			a.Fatal(err)
		}

		var covFiles []string
		files, err := os.ReadDir(testResultDir)
		if err != nil {
			a.Fatal(err)
		}
		for _, file := range files {
			covFiles = append(covFiles, file.Name())
		}

		mergedCovFile, err := os.Create(filepath.Join(testResultDir, "coverage.out"))
		if err != nil {
			a.Fatal(err)
		}
		defer func() {
			if err := mergedCovFile.Close(); err != nil {
				a.Fatal(err)
			}
		}()

		gocovmerge := a.Cmd("gocovmerge", covFiles...)
		gocovmerge.Dir = testResultDir
		gocovmerge.Stdout = mergedCovFile
		if err := gocovmerge.Run(); err != nil {
			a.Fatal(err)
		}
	},
})
