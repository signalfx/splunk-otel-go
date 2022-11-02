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

// tokenize.go generator.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
)

// To update Elasticsearch paths and operations:
//  - Download the latest schema.json file and put it in this directory.
//  - Prune schema.json for all `type` definitions (only `_info` and
//    `endpoints` are needed for the generation).
//  - Update the `filepath` and `schemaURL` const below.
//  - Run `go run main.go`
//  - Copy printed output into ../tokenize.go, replacing existing definitions.

const (
	filepath  = "./schema.json"
	schemaURL = "https://github.com/elastic/elasticsearch-specification/blob/60aa3a276e4c617ca7944816a6b4979c2384c675/output/schema/schema.json"
)

type schema struct {
	Info      info       `json:"_info"`
	Endpoints []endpoint `json:"endpoints"`

	Discard []byte `json:"-"`
}

type info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Hash    string `json:"hash"`

	Discard []byte `json:"-"`
}

type endpoint struct {
	Name       string `json:"name"`
	URLs       []url  `json:"urls"`
	Visibility string `json:"visibility"`

	Discard []byte `json:"-"`
}

func (e *endpoint) public() bool {
	return e.Visibility == "public"
}

type url struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`

	Discard []byte `json:"-"`
}

func printOrigin(s *schema) {
	fmt.Printf("// Generated from the %s\n", s.Info.Title)
	fmt.Printf("// Version: %s (hash: %s)\n", s.Info.Version, s.Info.Hash)
	fmt.Printf("// %s\n", schemaURL)
}

func printPaths(s *schema) {
	paths := make([]string, 0, len(s.Endpoints))
	for _, ep := range s.Endpoints {
		if ep.public() {
			continue
		}
		for _, u := range ep.URLs {
			paths = append(paths, u.Path)
		}
	}
	paths = unique(paths)

	printOrigin(s)
	fmt.Println("var paths = []string{")
	for _, p := range paths {
		fmt.Printf("\t%q,\n", p)
	}
	fmt.Println("}")
}

var urlType = `
type url struct {
	method, path string
}
`

func printOperationsMap(s *schema) {
	operations := make([]string, 0, len(s.Endpoints))
	for _, ep := range s.Endpoints {
		if ep.public() {
			continue
		}
		for _, u := range ep.URLs {
			for _, m := range u.Methods {
				op := fmt.Sprintf("\t{path: %q, method: %q}: %q,", u.Path, m, ep.Name)
				operations = append(operations, op)
			}
		}
	}
	operations = unique(operations)

	fmt.Println(urlType)
	printOrigin(s)
	fmt.Println("var operations = map[url]string{")
	for _, o := range operations {
		fmt.Println(o)
	}
	fmt.Println("}")
}

// unique returns a sorted slice of strings with all duplicates removed.
func unique(s []string) []string {
	if len(s) == 0 {
		return s
	}

	strS := sort.StringSlice(s)

	sort.Sort(strS)

	var k int
	for i := 1; i < strS.Len(); i++ {
		if strS.Less(k, i) {
			k++
			strS.Swap(k, i)
		}
	}

	return s[:k+1]
}

func main() {
	schemaData, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatalf("failed to read schema file: %v", err)
	}

	s := &schema{}
	err = json.Unmarshal(schemaData, s)
	if err != nil {
		log.Fatalf("failed to parse schema file: %v", err)
	}

	printPaths(s)
	fmt.Println()
	printOperationsMap(s)
}
