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

const (
	noneRealm    = "none"
	invalidRealm = "not-a-valid-realm"
	fakeEndpoint = "some non-zero value"
)

func TestOTLPEndpoint(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		assert.Equal(t, "", otlpEndpoint())
	})

	t.Cleanup(Setenv(splunkRealmKey, noneRealm))
	t.Run("none realm", func(t *testing.T) {
		// Revert to default.
		assert.Equal(t, "", otlpEndpoint())
	})

	t.Cleanup(Setenv(splunkRealmKey, invalidRealm))
	t.Run("realm", func(t *testing.T) {
		want := fmt.Sprintf(otlpRealmEndpointFormat, invalidRealm)
		assert.Equal(t, want, otlpEndpoint())
	})

	t.Run(otelExporterOTLPEndpointKey, func(t *testing.T) {
		t.Cleanup(Setenv(otelExporterOTLPEndpointKey, fakeEndpoint))
		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", otlpEndpoint())
	})

	t.Run(otelExporterOTLPTracesEndpointKey, func(t *testing.T) {
		t.Cleanup(Setenv(otelExporterOTLPTracesEndpointKey, "some non-zero value"))
		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", otlpEndpoint())
	})
}

func TestJaegerEndpoint(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		assert.Equal(t, defaultJaegerEndpoint, jaegerEndpoint())
	})

	t.Cleanup(Setenv(splunkRealmKey, noneRealm))
	t.Run("none realm", func(t *testing.T) {
		// Revert to default.
		assert.Equal(t, defaultJaegerEndpoint, jaegerEndpoint())
	})

	t.Cleanup(Setenv(splunkRealmKey, invalidRealm))
	t.Run("realm", func(t *testing.T) {
		want := fmt.Sprintf(realmEndpointFormat, invalidRealm)
		assert.Equal(t, want, jaegerEndpoint())
	})

	t.Run(otelExporterJaegerEndpointKey, func(t *testing.T) {
		t.Cleanup(Setenv(otelExporterJaegerEndpointKey, fakeEndpoint))
		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", jaegerEndpoint())
	})
}

func TestIsLocalhost(t *testing.T) {
	local := []string{
		"localhost",
		"http://localhost",
		"https://localhost:8080",
		"127.0.0.1:8080",
		"127.1.2.3",
		"::1",
		"0000:0000:0000:0000:0000:0000:0000:0001",
	}

	for _, addr := range local {
		assert.Truef(t, isLocalhost(addr), "%q is a local address", addr)
	}

	remote := []string{
		"google.com",
		"8.8.8.8",
		"1:2:3",
		"not valid",
	}

	for _, addr := range remote {
		assert.Falsef(t, isLocalhost(addr), "%q is a remote address", addr)
	}
}
