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

package splunkelastic

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var metricRegexp = regexp.MustCompile(`/_nodes/{metric}`)

var result string

func testPaths() map[string]int {
	// Use a map here to radomize benchmark.
	tp := make(map[string]int, len(paths))
	for i, p := range paths {
		// {metric} and {node_id} need to be distinguished. Replace {metric}
		// with all the known values it can be.
		metricR := []byte("/_nodes/process")
		testPath := metricRegexp.ReplaceAll([]byte(p), metricR)

		tokenR := []byte(fmt.Sprintf("token-%d", i))
		testPath = tokenRegexp.ReplaceAll(testPath, tokenR)
		tp[string(testPath)] = i
	}
	return tp
}

func BenchmarkTokenize(b *testing.B) {
	tp := testPaths()
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for p, _ := range tp {
			result = tokenize(p)
		}
	}
}

func TestTokenizeValidPaths(t *testing.T) {
	for p, i := range testPaths() {
		assert.Equalf(t, paths[i], tokenize(p), "tokenize(%q)", p)
	}
}

func TestTokenizeInvalidPaths(t *testing.T) {
	paths := []string{
		"",
		"/not/a/valid/path",
	}
	for _, p := range paths {
		assert.Equalf(t, "", tokenize(p), "invalid path %q should be empty", p)
	}
}

func TestSegment(t *testing.T) {
	path := "/_security/service/{namespace}/{service}/credential/token/{name}/_clear_cache"
	segments := []struct {
		part       string
		start, end int
	}{
		{
			part:  "/_security",
			start: 0,
			end:   10,
		},
		{
			part:  "/service",
			start: 10,
			end:   18,
		},
		{
			part:  "/{namespace}",
			start: 18,
			end:   30,
		},
		{
			part:  "/{service}",
			start: 30,
			end:   40,
		},
		{
			part:  "/credential",
			start: 40,
			end:   51,
		},
		{
			part:  "/token",
			start: 51,
			end:   57,
		},
		{
			part:  "/{name}",
			start: 57,
			end:   64,
		},
		{
			part:  "/_clear_cache",
			start: 64,
			end:   -1,
		},
	}
	for _, s := range segments {
		part, end := segment(path, s.start)
		assert.Equal(t, part, s.part)
		assert.Equal(t, end, s.end)
	}
}
