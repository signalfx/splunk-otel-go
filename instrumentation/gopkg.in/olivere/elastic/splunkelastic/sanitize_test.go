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
		for p, i := range tp {
			result = tokenize(p)
			if want := paths[i]; want != result {
				b.Errorf("[%d:%q] %q != %q", i, p, want, result)
			}
		}
	}
}
