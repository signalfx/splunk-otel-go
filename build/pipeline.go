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
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goyek/goyek"
)

// Config is used to configure the registered pipeline.
type Config struct {
	RepoPackagePrefix string // used in "fmt" and "test" tasks
}

// Pipeline contains the tasks and parameters
// registered via Flow function.
type Pipeline struct {
	Tasks struct {
		Clean        goyek.RegisteredTask
		Fmt          goyek.RegisteredTask
		Markdownlint goyek.RegisteredTask
		Misspell     goyek.RegisteredTask
		GolangciLint goyek.RegisteredTask
		Test         goyek.RegisteredTask
		ModTidy      goyek.RegisteredTask
		Diff         goyek.RegisteredTask
		Lint         goyek.RegisteredTask
		All          goyek.RegisteredTask
	}
	Params struct {
		CI         goyek.RegisteredBoolParam
		SkipDocker goyek.RegisteredBoolParam
		TestShrot  goyek.RegisteredBoolParam
	}
}

// Register registers common Go tasks.
func Register(flow *goyek.Flow, cfg Config) Pipeline {
	result := Pipeline{}

	// parameters
	result.Params.CI = flow.RegisterBoolParam(goyek.BoolParam{
		Name:  "ci",
		Usage: "Whether CI is calling the build script",
	})

	result.Params.SkipDocker = flow.RegisterBoolParam(goyek.BoolParam{
		Name:  "skip-docker",
		Usage: "Skip tasks using Docker",
	})

	result.Params.TestShrot = flow.RegisterBoolParam(goyek.BoolParam{
		Name:  "test-short",
		Usage: "Do not run long-running tests (using Docker)",
	})

	// tasks
	result.Tasks.Clean = flow.Register(taskClean())
	result.Tasks.Fmt = flow.Register(taskFmt(cfg.RepoPackagePrefix))
	result.Tasks.Markdownlint = flow.Register(taskMarkdownLint(result.Params.SkipDocker))
	result.Tasks.Misspell = flow.Register(taskMisspell())
	result.Tasks.GolangciLint = flow.Register(taskGolangciLint())
	result.Tasks.Test = flow.Register(taskTest(cfg.RepoPackagePrefix, flow.VerboseParam(), result.Params.TestShrot))
	result.Tasks.ModTidy = flow.Register(taskModTidy())
	result.Tasks.Diff = flow.Register(taskDiff(result.Params.CI))

	// pipelines
	result.Tasks.Lint = flow.Register(taskLint(goyek.Deps{
		result.Tasks.Misspell,
		result.Tasks.Markdownlint,
		result.Tasks.GolangciLint,
	}))
	result.Tasks.All = flow.Register(taskAll(goyek.Deps{
		result.Tasks.ModTidy,
		result.Tasks.Fmt,
		result.Tasks.Lint,
		result.Tasks.Test,
		result.Tasks.Diff,
	}))
	flow.DefaultTask = result.Tasks.All

	return result
}

const buildDir = "build"

func taskClean() goyek.Task {
	return goyek.Task{
		Name:  "clean",
		Usage: "remove git ignored files",
		Action: func(tf *goyek.TF) {
			if err := tf.Cmd("git", "clean", "-fX").Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskModTidy() goyek.Task {
	return goyek.Task{
		Name:  "mod-tidy",
		Usage: "go mod tidy",
		Action: func(tf *goyek.TF) {
			ForGoModules(tf, func(tf *goyek.TF) {
				if err := tf.Cmd("go", "mod", "tidy").Run(); err != nil {
					tf.Error(err)
				}
			})
		},
	}
}

func taskFmt(repoPrefix string) goyek.Task {
	return goyek.Task{
		Name:  "fmt",
		Usage: "gofumports",
		Action: func(tf *goyek.TF) {
			installFmt := tf.Cmd("go", "install", "mvdan.cc/gofumpt")
			installFmt.Dir = buildDir
			if err := installFmt.Run(); err != nil {
				tf.Fatal(err)
			}

			ForGoModules(tf, func(tf *goyek.TF) {
				tf.Cmd("gofumpt", "-l", "-w", ".").Run() //nolint // it is OK if it returns error
			})

			installGoImports := tf.Cmd("go", "install", "golang.org/x/tools/cmd/goimports")
			installGoImports.Dir = buildDir
			if err := installGoImports.Run(); err != nil {
				tf.Fatal(err)
			}

			ForGoModules(tf, func(tf *goyek.TF) {
				tf.Cmd("goimports", "-l", "-w", "-local", repoPrefix, ".").Run() //nolint // it is OK if it returns error
			})
		},
	}
}

func taskMarkdownLint(skipDocker goyek.RegisteredBoolParam) goyek.Task {
	return goyek.Task{
		Name:   "markdownlint",
		Usage:  "markdownlint-cli (requires docker)",
		Params: goyek.Params{skipDocker},
		Action: func(tf *goyek.TF) {
			if skipDocker.Get(tf) {
				tf.Skip("skipping as Docker is needed")
			}

			if err := tf.Cmd("docker", "run", "--rm", "-v", WorkDir(tf)+":/markdown", "06kellyjac/markdownlint-cli:0.28.1", "**/*.md").Run(); err != nil {
				tf.Error(err)
			}
		},
	}
}

func taskMisspell() goyek.Task {
	return goyek.Task{
		Name:  "misspell",
		Usage: "misspell",
		Action: func(tf *goyek.TF) {
			installFmt := tf.Cmd("go", "install", "github.com/client9/misspell/cmd/misspell")
			installFmt.Dir = buildDir
			if err := installFmt.Run(); err != nil {
				tf.Fatal(err)
			}

			lint := tf.Cmd("misspell", "-error", "-locale=US", "-i=importas", ".")
			if err := lint.Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func taskGolangciLint() goyek.Task {
	return goyek.Task{
		Name:  "golangci-lint",
		Usage: "golangci-lint",
		Action: func(tf *goyek.TF) {
			installLint := tf.Cmd("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint")
			installLint.Dir = buildDir
			if err := installLint.Run(); err != nil {
				tf.Fatal(err)
			}

			ForGoModules(tf, func(tf *goyek.TF) {
				lint := tf.Cmd("golangci-lint", "run", "--timeout", "4m0s")
				if err := lint.Run(); err != nil {
					tf.Error(err)
				}
			})
		},
	}
}

func taskTest(repoPrefix string, verbose, testShort goyek.RegisteredBoolParam) goyek.Task {
	return goyek.Task{
		Name:   "test",
		Usage:  "go test with race detector and code covarage",
		Params: goyek.Params{verbose, testShort},
		Action: func(tf *goyek.TF) {
			// prepare test-results
			curDir := WorkDir(tf)
			testResultDir := filepath.Join(curDir, "test-results")
			if err := os.RemoveAll(testResultDir); err != nil {
				tf.Fatal(err)
			}
			if err := os.Mkdir(testResultDir, 0o750); err != nil { // nolint:gomnd
				tf.Fatal(err)
			}

			// run go test race with code covarage for each Go Module
			ForGoModules(tf, func(tf *goyek.TF) {
				const fileNameLen = 12
				covOut := filepath.Join(testResultDir, RandString(tf, fileNameLen)+".out")
				if err := tf.Cmd("go", goTestArgs(repoPrefix, verbose.Get(tf), testShort.Get(tf), covOut)...).Run(); err != nil {
					tf.Error(err)
				}
			})

			// merge the coverage output files into a single coverage.out file
			installGocovmerge := tf.Cmd("go", "install", "github.com/wadey/gocovmerge")
			installGocovmerge.Dir = buildDir
			if err := installGocovmerge.Run(); err != nil {
				tf.Fatal(err)
			}

			var covFiles []string
			files, err := os.ReadDir(testResultDir)
			if err != nil {
				tf.Fatal(err)
			}
			for _, file := range files {
				covFiles = append(covFiles, file.Name())
			}

			mergedCovFile, err := os.Create(filepath.Join(testResultDir, "coverage.out"))
			if err != nil {
				tf.Fatal(err)
			}
			defer func() {
				if err := mergedCovFile.Close(); err != nil {
					tf.Fatal(err)
				}
			}()

			gocovmerge := tf.Cmd("gocovmerge", covFiles...)
			gocovmerge.Dir = testResultDir
			gocovmerge.Stdout = mergedCovFile
			if err := gocovmerge.Run(); err != nil {
				tf.Fatal(err)
			}
		},
	}
}

func goTestArgs(repoPrefix string, verbose, short bool, covOut string) []string {
	result := []string{"test", "-race", "-covermode=atomic"}
	if verbose {
		result = append(result, "-v")
	}
	if short {
		result = append(result, "-short")
	}
	if repoPrefix != "" {
		result = append(result, "-coverpkg="+repoPrefix+"/...")
	}
	result = append(result, "-coverprofile="+covOut, "./...")
	return result
}

func taskDiff(ci goyek.RegisteredBoolParam) goyek.Task {
	return goyek.Task{
		Name:   "diff",
		Usage:  "git diff",
		Params: goyek.Params{ci},
		Action: func(tf *goyek.TF) {
			if !ci.Get(tf) {
				tf.Skip("ci param is not set, skipping")
			}

			if err := tf.Cmd("git", "diff", "--exit-code").Run(); err != nil {
				tf.Error(err)
			}

			cmd := tf.Cmd("git", "status", "--porcelain")
			sb := &strings.Builder{}
			cmd.Stdout = io.MultiWriter(tf.Output(), sb)
			if err := cmd.Run(); err != nil {
				tf.Error(err)
			}
			if sb.Len() > 0 {
				tf.Error("git status --porcelain returned output")
			}
		},
	}
}

func taskLint(deps goyek.Deps) goyek.Task {
	return goyek.Task{
		Name:  "lint",
		Usage: "all linters",
		Deps:  deps,
	}
}

func taskAll(deps goyek.Deps) goyek.Task {
	return goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps:  deps,
	}
}

// ForGoModules is a helper that executes given function
// in each directory containing go.mod file.
func ForGoModules(tf *goyek.TF, fn func(tf *goyek.TF)) {
	curDir := WorkDir(tf)
	_ = filepath.WalkDir(curDir, func(path string, dir fs.DirEntry, err error) error {
		if dir.Name() != "go.mod" {
			return nil
		}

		goModDir := filepath.Dir(path)
		tf.Log("Go Module:", goModDir)
		if err := os.Chdir(goModDir); err != nil {
			tf.Fatal(err)
		}

		fn(tf) // execute function in file containing go.mod

		return nil
	})

	defer ChDir(tf, curDir)
}

// WorkDir returns current working directory.
func WorkDir(tf *goyek.TF) string {
	curDir, err := os.Getwd()
	if err != nil {
		tf.Fatal(err)
	}
	return curDir
}

// ChDir changes the working directory.
func ChDir(tf *goyek.TF, path string) {
	if err := os.Chdir(path); err != nil {
		tf.Fatal(err)
	}
}

// RandString returns securely generated hex-string.
func RandString(tf *goyek.TF, length int) string {
	n := length / 2
	if length%2 != 0 {
		n++
	}

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		tf.Fatal(err)
	}
	return hex.EncodeToString(b)[:length]
}
