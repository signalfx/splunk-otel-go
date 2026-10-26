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
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	traceapi "go.opentelemetry.io/otel/trace"
)

const (
	attributeCountKey          = "OTEL_ATTRIBUTE_COUNT_LIMIT"
	spanAttributeCountKey      = "OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT"
	spanEventCountKey          = "OTEL_SPAN_EVENT_COUNT_LIMIT"
	spanEventAttributeCountKey = "OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT"
	spanLinkCountKey           = "OTEL_SPAN_LINK_COUNT_LIMIT"
	spanLinkAttributeCountKey  = "OTEL_LINK_ATTRIBUTE_COUNT_LIMIT"
)

func expectedSL(aLen, aN, eN, lN, aPerE, aPerL int) *trace.SpanLimits {
	return &trace.SpanLimits{
		AttributeValueLengthLimit:   aLen,
		AttributeCountLimit:         aN,
		EventCountLimit:             eN,
		LinkCountLimit:              lN,
		AttributePerEventCountLimit: aPerE,
		AttributePerLinkCountLimit:  aPerL,
	}
}

func TestNewSpanLimits(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want *trace.SpanLimits
	}{
		{
			name: "defaults",
			want: expectedSL(12000, 128, 128, 128, 128, 128),
		},
		{
			name: attributeValueLengthKey,
			envs: map[string]string{
				attributeValueLengthKey: "10",
			},
			want: expectedSL(10, 128, 128, 128, 128, 128),
		},
		{
			name: spanAttributeValueLengthKey,
			envs: map[string]string{
				spanAttributeValueLengthKey: "10",
			},
			want: expectedSL(10, 128, 128, 128, 128, 128),
		},
		{
			name: attributeCountKey,
			envs: map[string]string{
				attributeCountKey: "10",
			},
			want: expectedSL(12000, 10, 128, 128, 128, 128),
		},
		{
			name: spanAttributeCountKey,
			envs: map[string]string{
				spanAttributeCountKey: "10",
			},
			want: expectedSL(12000, 10, 128, 128, 128, 128),
		},
		{
			name: spanEventCountKey,
			envs: map[string]string{
				spanEventCountKey: "10",
			},
			want: expectedSL(12000, 128, 10, 128, 128, 128),
		},
		{
			name: spanEventAttributeCountKey,
			envs: map[string]string{
				spanEventAttributeCountKey: "10",
			},
			want: expectedSL(12000, 128, 128, 128, 10, 128),
		},
		{
			name: spanLinkCountKey,
			envs: map[string]string{
				spanLinkCountKey: "10",
			},
			want: expectedSL(12000, 128, 128, 10, 128, 128),
		},
		{
			name: spanLinkAttributeCountKey,
			envs: map[string]string{
				spanLinkAttributeCountKey: "10",
			},
			want: expectedSL(12000, 128, 128, 128, 128, 10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for key, val := range test.envs {
				t.Setenv(key, val)
			}
			assert.Equal(t, test.want, newSpanLimits())
		})
	}
}

func TestSpanLimitsMigration(t *testing.T) {
	tests := []struct {
		name       string
		envs       map[string]string
		attributes int
		events     int
		links      int
	}{
		{
			name:       "defaults",
			attributes: 128,
			events:     128,
			links:      128,
		},
		{
			name: "previous defaults",
			envs: map[string]string{
				spanAttributeCountKey:      "-1",
				spanEventCountKey:          "-1",
				spanLinkCountKey:           "1000",
				spanEventAttributeCountKey: "-1",
				spanLinkAttributeCountKey:  "-1",
			},
			attributes: 129,
			events:     129,
			links:      1000,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, key := range []string{
				attributeCountKey, spanAttributeCountKey, spanEventCountKey,
				spanLinkCountKey, spanEventAttributeCountKey, spanLinkAttributeCountKey,
			} {
				t.Setenv(key, test.envs[key])
			}

			recorder := tracetest.NewSpanRecorder()
			provider := trace.NewTracerProvider(
				trace.WithRawSpanLimits(*newSpanLimits()),
				trace.WithSampler(trace.AlwaysSample()),
				trace.WithSpanProcessor(recorder),
			)
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

			attributes := make([]attribute.KeyValue, 129)
			for i := range attributes {
				attributes[i] = attribute.Int(strconv.Itoa(i), i)
			}
			links := make([]traceapi.Link, 1001)
			for i := range links {
				links[i] = traceapi.Link{
					SpanContext: traceapi.NewSpanContext(traceapi.SpanContextConfig{
						TraceID: traceapi.TraceID{1},
						SpanID:  traceapi.SpanID{1},
					}),
					Attributes: attributes,
				}
			}

			_, span := provider.Tracer("test").Start(context.Background(), "migration",
				traceapi.WithAttributes(attributes...), traceapi.WithLinks(links...))
			for range 129 {
				span.AddEvent("event", traceapi.WithAttributes(attributes...))
			}
			span.End()

			spans := recorder.Ended()
			require.Len(t, spans, 1)
			got := spans[0]
			assert.Len(t, got.Attributes(), test.attributes)
			assert.Equal(t, 129-test.attributes, got.DroppedAttributes())
			require.Len(t, got.Events(), test.events)
			assert.Equal(t, 129-test.events, got.DroppedEvents())
			for _, event := range got.Events() {
				assert.Len(t, event.Attributes, test.attributes)
				assert.Equal(t, 129-test.attributes, event.DroppedAttributeCount)
			}
			require.Len(t, got.Links(), test.links)
			assert.Equal(t, 1001-test.links, got.DroppedLinks())
			for _, link := range got.Links() {
				assert.Len(t, link.Attributes, test.attributes)
				assert.Equal(t, 129-test.attributes, link.DroppedAttributeCount)
			}
		})
	}
}
