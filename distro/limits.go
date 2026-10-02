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

import (
	"os"

	"go.opentelemetry.io/otel/sdk/trace"
)

const (
	attributeValueLengthKey         = "OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT"
	spanAttributeValueLengthKey     = "OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT"
	spanAttributeValueLengthDefault = 12000
)

// newSpanLimits returns new span limits that use the Splunk default attribute
// value length and OpenTelemetry defaults for all count limits. Environment
// variable values override these defaults.
func newSpanLimits() *trace.SpanLimits {
	// Use trace.NewSpanLimits here to ensure any future additions are not set
	// to zero, which would happen if we declared with &trace.SpanLimits{...}.
	limits := trace.NewSpanLimits()

	// limits will use OpenTelemetry defaults or the applicable environment
	// variable if set. Apply the Splunk default only if neither attribute value
	// length environment variable was set.
	limits.AttributeValueLengthLimit = limitValue(limits.AttributeValueLengthLimit, spanAttributeValueLengthDefault, attributeValueLengthKey, spanAttributeValueLengthKey)

	return &limits
}

// limitValue returns the current limit value if one of envs is defined;
// otherwise, it returns defaultValue.
func limitValue(current, defaultValue int, envs ...string) int {
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			return current
		}
	}
	return defaultValue
}
