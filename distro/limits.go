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
	attributeValueLengthKey     = "OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT"
	attributeCountKey           = "OTEL_ATTRIBUTE_COUNT_LIMIT"
	spanAttributeValueLengthKey = "OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT"
	spanAttributeCountKey       = "OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT"
	spanEventCountKey           = "OTEL_SPAN_EVENT_COUNT_LIMIT"
	spanEventAttributeCountKey  = "OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT"
	spanLinkCountKey            = "OTEL_SPAN_LINK_COUNT_LIMIT"
	spanLinkAttributeCountKey   = "OTEL_LINK_ATTRIBUTE_COUNT_LIMIT"
)

// sources are span limit sources with Splunk defaults.
var sources = spanLimitSources{
	// Maximum allowed attribute value size for a span.
	{
		envs: []string{attributeValueLengthKey, spanAttributeValueLengthKey},
		setDefault: func(limits trace.SpanLimits) trace.SpanLimits {
			limits.AttributeValueLengthLimit = 12000
			return limits
		},
	},

	// Maximum allowed attribute count for a span.
	{
		envs: []string{attributeCountKey, spanAttributeCountKey},
		setDefault: func(limits trace.SpanLimits) trace.SpanLimits {
			// Unlimited.
			limits.AttributeCountLimit = -1
			return limits
		},
	},

	// Maximum allowed span event count.
	{
		envs: []string{spanEventCountKey},
		setDefault: func(limits trace.SpanLimits) trace.SpanLimits {
			// Unlimited.
			limits.EventCountLimit = -1
			return limits
		},
	},

	// Maximum allowed attributes per span event.
	{
		envs: []string{spanEventAttributeCountKey},
		setDefault: func(limits trace.SpanLimits) trace.SpanLimits {
			// Unlimited.
			limits.AttributePerEventCountLimit = -1
			return limits
		},
	},

	// Maximum allowed span link count.
	{
		envs: []string{spanLinkCountKey},
		setDefault: func(limits trace.SpanLimits) trace.SpanLimits {
			limits.LinkCountLimit = 1000
			return limits
		},
	},

	// Maximum allowed attributes per span link.
	{
		envs: []string{spanLinkAttributeCountKey},
		setDefault: func(limits trace.SpanLimits) trace.SpanLimits {
			// Unlimited.
			limits.AttributePerLinkCountLimit = -1
			return limits
		},
	},
}

// spanLimitSources is a collection of all the sources of configuration for
// a group of spans limits.
type spanLimitSources []spanLimitSource

// applyDefaults will apply the defaults to all limits that do not have the
// associated limit's environment variables set.
func (s spanLimitSources) applyDefaults(limits trace.SpanLimits) trace.SpanLimits {
	for _, source := range s {
		limits = source.applyDefault(limits)
	}
	return limits
}

// spanLimitSource are all the sources a span limit is configured from and the
// default value used if none of those sources are used.
type spanLimitSource struct {
	envs       []string
	setDefault func(limits trace.SpanLimits) trace.SpanLimits
}

// envsSet returns if any environment variable for s is set.
func (s spanLimitSource) envsSet() bool {
	for _, env := range s.envs {
		if _, ok := os.LookupEnv(env); ok {
			return true
		}
	}
	return false
}

// applyDefaults will apply the default to the corresponding limit in limits
// if the associated environment variables are not set.
func (s spanLimitSource) applyDefault(limits trace.SpanLimits) trace.SpanLimits {
	if !s.envsSet() {
		// Panic if setDefault is nil. It never should be, and if it is alert
		// the developer that made the change as soon as possible.
		limits = s.setDefault(limits)
	}
	return limits
}

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
	limits = sources.applyDefaults(limits)

	return &limits
}
