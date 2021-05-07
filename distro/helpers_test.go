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

package distro

import "os"

// Setenv sets the value of the environment variable named by the key.
// It returns a function that rollbacks the setting.
func Setenv(key, val string) func() {
	valSnapshot, ok := os.LookupEnv(key)
	os.Setenv(key, val)
	return func() {
		if ok {
			os.Setenv(key, valSnapshot)
		} else {
			os.Unsetenv(key)
		}
	}
}
