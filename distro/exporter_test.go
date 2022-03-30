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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOTLPEndpoint(t *testing.T) {
	t.Run("configured", func(t *testing.T) {
		e := "configured endpoint"
		assert.Equal(t, e, otlpEndpoint(e))
	})

	t.Run("default", func(t *testing.T) {
		assert.Equal(t, "", otlpEndpoint(""))
	})

	realm := "not-a-valid-realm"
	t.Cleanup(Setenv(splunkRealmKey, realm))
	t.Run("realm", func(t *testing.T) {
		want := fmt.Sprintf(otlpRealmEndpointFormat, realm)
		assert.Equal(t, want, otlpEndpoint(""))
	})

	t.Run(otelExporterOTLPEndpointKey, func(t *testing.T) {
		t.Cleanup(Setenv(otelExporterOTLPEndpointKey, "some non-zero value"))
		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", otlpEndpoint(""))
	})

	t.Run(otelExporterOTLPTracesEndpointKey, func(t *testing.T) {
		t.Cleanup(Setenv(otelExporterOTLPTracesEndpointKey, "some non-zero value"))
		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", otlpEndpoint(""))
	})
}

func TestJaegerEndpoint(t *testing.T) {
	t.Run("configured", func(t *testing.T) {
		e := "configured endpoint"
		assert.Equal(t, e, jaegerEndpoint(e))
	})

	t.Run("default", func(t *testing.T) {
		assert.Equal(t, defaultJaegerEndpoint, jaegerEndpoint(""))
	})

	realm := "not-a-valid-realm"
	t.Cleanup(Setenv(splunkRealmKey, realm))
	t.Run("realm", func(t *testing.T) {
		want := fmt.Sprintf(realmEndpointFormat, realm)
		assert.Equal(t, want, jaegerEndpoint(""))
	})

	t.Run(otelExporterJaegerEndpointKey, func(t *testing.T) {
		t.Cleanup(Setenv(otelExporterJaegerEndpointKey, "some non-zero value"))
		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", jaegerEndpoint(""))
	})
}
