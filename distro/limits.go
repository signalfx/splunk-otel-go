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

	attributeCountKey         = "OTEL_ATTRIBUTE_COUNT_LIMIT"
	spanAttributeCountKey     = "OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT"
	spanAttributeCountDefault = -1 // Unlimited.

	spanEventCountKey     = "OTEL_SPAN_EVENT_COUNT_LIMIT"
	spanEventCountDefault = -1 // Unlimited.

	spanEventAttributeCountKey     = "OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT"
	spanEventAttributeCountDefault = -1 // Unlimited.

	spanLinkCountKey     = "OTEL_SPAN_LINK_COUNT_LIMIT"
	spanLinkCountDefault = 1000

	spanLinkAttributeCountKey     = "OTEL_LINK_ATTRIBUTE_COUNT_LIMIT"
	spanLinkAttributeCountDefault = -1 // Unlimited.
)

// newSpanLimits returns new span limits that use Splunk defaults (the link
// count is limited to 1000, the attribute value length is limited to 12000,
// and all other limts are set to be unlimited) or the corresponding OTel
// environment variable value if it is set.
func newSpanLimits() *trace.SpanLimits {
	// Use trace.NewSpanLimits here to ensure any future additions are not set
	// to zero, which would happen if we delared with &trace.SpanLimits{...}.
	limits := trace.NewSpanLimits()

	// limits will use OTel defaults or the applicable environment variable if
	// they are set. The Splunk defaults need to be applied only if the
	// environment variables were unset.
	limits.AttributeValueLengthLimit = limitValue(limits.AttributeValueLengthLimit, spanAttributeValueLengthDefault, attributeValueLengthKey, spanAttributeValueLengthKey)
	limits.AttributeCountLimit = limitValue(limits.AttributeCountLimit, spanAttributeCountDefault, attributeCountKey, spanAttributeCountKey)
	limits.EventCountLimit = limitValue(limits.EventCountLimit, spanEventCountDefault, spanEventCountKey)
	limits.LinkCountLimit = limitValue(limits.LinkCountLimit, spanLinkCountDefault, spanLinkCountKey)
	limits.AttributePerEventCountLimit = limitValue(limits.AttributePerEventCountLimit, spanEventAttributeCountDefault, spanEventAttributeCountKey)
	limits.AttributePerLinkCountLimit = limitValue(limits.AttributePerLinkCountLimit, spanLinkAttributeCountDefault, spanLinkAttributeCountKey)

	return &limits
}

// limitValue returns the current limit value if it was set because one of
// envs is defined, otherwise it returns the limit zero value.
func limitValue(current, zero int, envs ...string) int {
	for _, env := range envs {
		if _, ok := os.LookupEnv(env); ok {
			return current
		}
	}
	return zero
}
