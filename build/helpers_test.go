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
	"strconv"
	"testing"

	"github.com/goyek/goyek/v3"
)

func TestRandString(t *testing.T) {
	testCases := []int{
		1,
		2,
		5,
		16,
	}
	for _, length := range testCases {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			action := func(a *goyek.A) {
				if got := RandString(a, length); len(got) != length {
					t.Errorf("got length %v, want %v", len(got), length) // for invalid input this line is not even executed as RandString calls a.Fatal
				}
			}

			result := goyek.NewRunner(action)(goyek.Input{})
			if result.Status != goyek.StatusPassed {
				t.Errorf("want StatusPassed but was: %v", result.Status)
			}
		})
	}
}

func TestRandString_invalid(t *testing.T) {
	testCases := []int{
		-1,
		0,
	}
	for _, length := range testCases {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			action := func(a *goyek.A) {
				RandString(a, length)
				t.Error("should not return for invalid input")
			}

			result := goyek.NewRunner(action)(goyek.Input{})
			if result.Status != goyek.StatusFailed {
				t.Errorf("want StatusFailed but was: %v", result.Status)
			}
		})
	}
}
