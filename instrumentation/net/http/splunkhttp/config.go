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

package splunkhttp

import (
	"os"
	"strings"
)

// Environmental variables used for configuration.
const (
	envVarTraceResponseHeaderEnabled = "SPLUNK_TRACE_RESPONSE_HEADER_ENABLED" // Adds `Server-Timing` header to HTTP responses
)

// config represents the available configuration options.
type config struct {
	TraceResponseHeaderEnabled bool
}

// newConfig creates a new config struct.
func newConfig() *config {
	traceResponseHeaderEnabled := true
	if v := os.Getenv(envVarTraceResponseHeaderEnabled); strings.EqualFold(v, "false") {
		traceResponseHeaderEnabled = false
	}

	return &config{
		TraceResponseHeaderEnabled: traceResponseHeaderEnabled,
	}
}
