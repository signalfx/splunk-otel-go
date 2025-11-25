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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/goyek/goyek/v3"
)

// ForGoModules is a helper that executes given function
// in each directory containing go.mod file.
func ForGoModules(a *goyek.A, fn func(a *goyek.A), ignoredPaths ...string) {
	a.Helper()

	var goModDirs []string
	_ = filepath.WalkDir(".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		path = filepath.ToSlash(path)
		for _, ignored := range ignoredPaths {
			if path == ignored {
				return filepath.SkipDir
			}
		}

		if dir.Name() != "go.mod" {
			return nil
		}
		goModDir := filepath.ToSlash(filepath.Dir(path))
		goModDirs = append(goModDirs, goModDir)
		return nil
	})

	for _, goModDir := range goModDirs {
		func() {
			a.Helper()
			curDir := WorkDir(a)
			defer ChDir(a, curDir)

			a.Log("Go Module: ", goModDir)
			ChDir(a, goModDir)

			fn(a) // execute function in file containing go.mod
		}()
	}
}

// WorkDir returns current working directory.
func WorkDir(a *goyek.A) string {
	a.Helper()

	curDir, err := os.Getwd()
	if err != nil {
		a.Fatal(err)
	}
	return curDir
}

// ChDir changes the working directory.
func ChDir(a *goyek.A, path string) {
	a.Helper()

	if err := os.Chdir(path); err != nil {
		a.Fatal(err)
	}
}

// Find returns all files with given extension.
func Find(a *goyek.A, ext string) []string {
	a.Helper()

	var files []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ext {
			files = append(files, filepath.ToSlash(path))
		}
		return nil
	})
	if err != nil {
		a.Fatal(err)
	}
	return files
}

// RandString returns securely generated hex-string.
func RandString(a *goyek.A, length int) string {
	a.Helper()

	if length < 1 {
		a.Fatal("length must be greater than 0")
	}

	n := length / 2
	if length%2 != 0 {
		n++
	}

	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		a.Fatal(err)
	}
	return hex.EncodeToString(b)[:length]
}
