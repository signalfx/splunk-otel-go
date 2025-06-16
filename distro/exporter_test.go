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
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tonglil/buflogr"
)

const (
	noneRealm       = "none"
	invalidRealm    = "not-a-valid-realm"
	fakeEndpoint    = "some non-zero value"
	invalidProtocol = "invalid-protocol"
)

func TestOTLPTracesEndpoint(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		assert.Equal(t, "", otlpRealmTracesEndpoint())
	})

	t.Run("none realm", func(t *testing.T) {
		t.Setenv(splunkRealmKey, noneRealm)

		assert.Equal(t, "", otlpRealmTracesEndpoint())
	})

	t.Run("realm", func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)

		want := fmt.Sprintf(otlpRealmTracesEndpointFormat, invalidRealm)
		assert.Equal(t, want, otlpRealmTracesEndpoint())
	})

	t.Run(otelExporterOTLPEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterOTLPEndpointKey, fakeEndpoint)

		// SPLUNK_REALM is set, make sure it does not take precedence.
		assert.Equal(t, "", otlpRealmTracesEndpoint())
	})

	t.Run(otelExporterOTLPTracesEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterOTLPTracesEndpointKey, "some non-zero value")

		// SPLUNK_REALM is set, make sure it does not take precedence.
		assert.Equal(t, "", otlpRealmTracesEndpoint())
	})

	t.Run(otelExporterOTLPMetricsEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterOTLPMetricsEndpointKey, "some non-zero value")

		// OTEL_EXPORTER_OTLP_METRICS_ENDPOINT is ignored for traces exporter.
		want := fmt.Sprintf(otlpRealmTracesEndpointFormat, invalidRealm)
		assert.Equal(t, want, otlpRealmTracesEndpoint())
	})
}

func TestOTLPProtocol(t *testing.T) {
	var buf bytes.Buffer
	var logger = buflogr.NewWithBuffer(&buf)

	t.Run("default", func(t *testing.T) {
		// Neither specific nor general env vars are set
		assert.Equal(t, defaultOTLPProtocol, otlpProtocol(logger, otelTracesExporterOTLPProtocolKey))
		assert.NotContains(t, buf.String(), "invalid")
	})

	t.Run("only general protocol", func(t *testing.T) {
		t.Setenv(otelExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		assert.Equal(t, "http/protobuf", otlpProtocol(logger, otelTracesExporterOTLPProtocolKey))
		assert.NotContains(t, buf.String(), "invalid")
	})

	t.Run("only specific protocol", func(t *testing.T) {
		t.Setenv(otelTracesExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		assert.Equal(t, otlpProtocolHTTPProtobuf, otlpProtocol(logger, otelTracesExporterOTLPProtocolKey))
		assert.NotContains(t, buf.String(), "invalid")
	})

	t.Run("specific overrides general", func(t *testing.T) {
		t.Setenv(otelExporterOTLPProtocolKey, defaultOTLPProtocol)
		t.Setenv(otelTracesExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		assert.Equal(t, otlpProtocolHTTPProtobuf, otlpProtocol(logger, otelTracesExporterOTLPProtocolKey))
		assert.NotContains(t, buf.String(), "invalid")
	})

	t.Run("invalid specific value", func(t *testing.T) {
		t.Setenv(otelTracesExporterOTLPProtocolKey, invalidProtocol)
		assert.Equal(t, defaultOTLPProtocol, otlpProtocol(logger, otelTracesExporterOTLPProtocolKey))
		assert.Contains(t, buf.String(), fmt.Sprintf("invalid %s: %q", otelTracesExporterOTLPProtocolKey, invalidProtocol))
	})

	t.Run("invalid general value", func(t *testing.T) {
		t.Setenv(otelExporterOTLPProtocolKey, invalidProtocol)
		assert.Equal(t, defaultOTLPProtocol, otlpProtocol(logger, otelTracesExporterOTLPProtocolKey))
		assert.Contains(t, buf.String(), fmt.Sprintf("invalid %s: %q", otelExporterOTLPProtocolKey, invalidProtocol))
	})
}

func TestOTLPMetricsEndpoint(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		assert.Equal(t, "", otlpRealmMetricsEndpoint())
	})

	t.Run("none realm", func(t *testing.T) {
		t.Setenv(splunkRealmKey, noneRealm)

		assert.Equal(t, "", otlpRealmMetricsEndpoint())
	})

	t.Run("realm", func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)

		want := fmt.Sprintf(otlpRealmMetricsEndpointFormat, invalidRealm)
		assert.Equal(t, want, otlpRealmMetricsEndpoint())
	})

	t.Run(otelExporterOTLPEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterOTLPEndpointKey, fakeEndpoint)

		// SPLUNK_REALM is set, make sure it does not take precedence.
		assert.Equal(t, "", otlpRealmMetricsEndpoint())
	})

	t.Run(otelExporterOTLPMetricsEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterOTLPMetricsEndpointKey, "some non-zero value")

		// SPLUNK_REALM is set, make sure it does not take precedence.
		assert.Equal(t, "", otlpRealmMetricsEndpoint())
	})

	t.Run(otelExporterOTLPTracesEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterOTLPTracesEndpointKey, "some non-zero value")

		// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT is ignored for metrics exporter.
		want := fmt.Sprintf(otlpRealmMetricsEndpointFormat, invalidRealm)
		assert.Equal(t, want, otlpRealmMetricsEndpoint())
	})
}

func TestJaegerEndpoint(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		assert.Equal(t, jaegerDefaultEndpoint, jaegerEndpoint())
	})

	t.Run("none realm", func(t *testing.T) {
		t.Setenv(splunkRealmKey, noneRealm)

		assert.Equal(t, jaegerDefaultEndpoint, jaegerEndpoint())
	})

	t.Run("realm", func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)

		want := fmt.Sprintf(jaegerRealmEndpointFormat, invalidRealm)
		assert.Equal(t, want, jaegerEndpoint())
	})

	t.Run(otelExporterJaegerEndpointKey, func(t *testing.T) {
		t.Setenv(splunkRealmKey, invalidRealm)
		t.Setenv(otelExporterJaegerEndpointKey, fakeEndpoint)

		// SPLUNK_REALM is still set, make sure it does not take precedence.
		assert.Equal(t, "", jaegerEndpoint())
	})
}
