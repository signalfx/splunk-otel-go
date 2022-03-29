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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
)

func expected(aLen, aN, eN, lN, aPerE, aPerL int) *trace.SpanLimits {
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
			want: expected(12000, -1, -1, 1000, -1, -1),
		},
		{
			name: attributeValueLengthKey,
			envs: map[string]string{
				attributeValueLengthKey: "10",
			},
			want: expected(10, -1, -1, 1000, -1, -1),
		},
		{
			name: spanAttributeValueLengthKey,
			envs: map[string]string{
				spanAttributeValueLengthKey: "10",
			},
			want: expected(10, -1, -1, 1000, -1, -1),
		},
		{
			name: attributeCountKey,
			envs: map[string]string{
				attributeCountKey: "10",
			},
			want: expected(12000, 10, -1, 1000, -1, -1),
		},
		{
			name: spanAttributeCountKey,
			envs: map[string]string{
				spanAttributeCountKey: "10",
			},
			want: expected(12000, 10, -1, 1000, -1, -1),
		},
		{
			name: spanEventCountKey,
			envs: map[string]string{
				spanEventCountKey: "10",
			},
			want: expected(12000, -1, 10, 1000, -1, -1),
		},
		{
			name: spanEventAttributeCountKey,
			envs: map[string]string{
				spanEventAttributeCountKey: "10",
			},
			want: expected(12000, -1, -1, 1000, 10, -1),
		},
		{
			name: spanLinkCountKey,
			envs: map[string]string{
				spanLinkCountKey: "10",
			},
			want: expected(12000, -1, -1, 10, -1, -1),
		},
		{
			name: spanLinkAttributeCountKey,
			envs: map[string]string{
				spanLinkAttributeCountKey: "10",
			},
			want: expected(12000, -1, -1, 1000, -1, 10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for key, val := range test.envs {
				t.Cleanup(Setenv(key, val))
			}
			assert.Equal(t, test.want, newSpanLimits())
		})
	}
}
