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
	"testing"
)

var result string

func testPaths() map[string]int {
	// Use a map here to radomize benchmark.
	tp := make(map[string]int, len(paths))
	for i, p := range paths {
		metricR := []byte("/_nodes/process")
		tokenR := []byte(fmt.Sprintf("token-%d", i))

		testPath := metricRegexp.ReplaceAll([]byte(p), metricR)
		testPath = tokenRegexp.ReplaceAll(testPath, tokenR)
		tp[string(testPath)] = i
	}
	return tp
}

func BenchmarkTokenizeFromSlice(b *testing.B) {
	tp := testPaths()
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for p, i := range tp {
			result = tokenizeFromSlice(p)
			if want := paths[i]; want != result {
				b.Errorf("[%d:%q] %q != %q", i, p, want, result)
			}
		}
	}
}

func BenchmarkTokenizeFromTrie(b *testing.B) {
	tp := testPaths()
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for p, i := range tp {
			result = tokenizeFromTrie(p)
			if want := paths[i]; want != result {
				b.Errorf("[%d:%q] %q != %q", i, p, want, result)
			}
		}
	}
}

func BenchmarkTokenizeFromNoAllocTrie(b *testing.B) {
	tp := testPaths()
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for p, i := range tp {
			result = tokenizeFromNoAllocTrie(p)
			if want := paths[i]; want != result {
				b.Errorf("[%d:%q] %q != %q", i, p, want, result)
			}
		}
	}
}
