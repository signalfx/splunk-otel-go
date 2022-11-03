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

	"github.com/goyek/goyek/v2"
)

// ForGoModules is a helper that executes given function
// in each directory containing go.mod file.
func ForGoModules(tf *goyek.TF, fn func(tf *goyek.TF)) {
	tf.Helper()

	var goModDirs []string
	_ = filepath.WalkDir(".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dir.Name() != "go.mod" {
			return nil
		}
		goModDirs = append(goModDirs, filepath.Dir(path))
		return nil
	})

	for _, goModDir := range goModDirs {
		func() {
			tf.Helper()
			curDir := WorkDir(tf)
			defer ChDir(tf, curDir)

			tf.Log("Go Module: ", goModDir)
			ChDir(tf, goModDir)

			fn(tf) // execute function in file containing go.mod
		}()
	}
}

// WorkDir returns current working directory.
func WorkDir(tf *goyek.TF) string {
	tf.Helper()

	curDir, err := os.Getwd()
	if err != nil {
		tf.Fatal(err)
	}
	return curDir
}

// ChDir changes the working directory.
func ChDir(tf *goyek.TF, path string) {
	tf.Helper()

	if err := os.Chdir(path); err != nil {
		tf.Fatal(err)
	}
}

// Find returns all files with given extension.
func Find(tf *goyek.TF, ext string) []string {
	tf.Helper()

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
		tf.Fatal(err)
	}
	return files
}

// RandString returns securely generated hex-string.
func RandString(tf *goyek.TF, length int) string {
	tf.Helper()

	if length < 1 {
		tf.Fatal("length must be greater than 0")
	}

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
