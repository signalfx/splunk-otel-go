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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
)

// To update Elasticsearch paths and operations:
//  - download the latest schema.json file and put it in this directory.
//  - Update the `filepath` and `schemaURL` const below.
//  - Run `go run main.go`
//  - Copy printed output into ../tokenize.go, replacing existing definitions.

const (
	filepath  = "./schema.json"
	schemaURL = "https://github.com/elastic/elasticsearch-specification/blob/60aa3a276e4c617ca7944816a6b4979c2384c675/output/schema/schema.json"
)

type Schema struct {
	Info      Info       `json:"_info"`
	Endpoints []Endpoint `json:"endpoints"`

	Discard []byte `json:"-"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
	Hash    string `json:"hash"`

	Discard []byte `json:"-"`
}

type Endpoint struct {
	Name       string `json:"name"`
	URLs       []URL  `json:"urls"`
	Visibility string `json:"visibility"`

	Discard []byte `json:"-"`
}

type URL struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`

	Discard []byte `json:"-"`
}

func printOrigin(schema *Schema) {
	fmt.Printf("// Generated from the %s\n", schema.Info.Title)
	fmt.Printf("// Version: %s (hash: %s)\n", schema.Info.Version, schema.Info.Hash)
	fmt.Printf("// %s\n", schemaURL)
}

func printPaths(schema *Schema) {
	paths := make([]string, 0, len(schema.Endpoints))
	for _, ep := range schema.Endpoints {
		if ep.Visibility != "public" {
			continue
		}
		for _, u := range ep.URLs {
			paths = append(paths, u.Path)
		}
	}
	paths = unique(paths)

	printOrigin(schema)
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

func printOperationsMap(schema *Schema) {
	operations := make([]string, 0, len(schema.Endpoints))
	for _, ep := range schema.Endpoints {
		if ep.Visibility != "public" {
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
	printOrigin(schema)
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
	schemaData, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalf("failed to read schema file: %v", err)
	}

	schema := &Schema{}
	err = json.Unmarshal([]byte(schemaData), schema)
	if err != nil {
		log.Fatalf("failed to parse schema file: %v", err)
	}

	printPaths(schema)
	fmt.Println()
	printOperationsMap(schema)
}
