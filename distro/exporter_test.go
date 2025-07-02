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

	"github.com/go-logr/logr"
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
	newTestLogger := func() (*bytes.Buffer, logr.Logger) {
		var buf bytes.Buffer
		return &buf, buflogr.NewWithBuffer(&buf)
	}

	t.Run("default", func(t *testing.T) {
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, defaultOTLPProtocol, got)
		assert.Empty(t, buf.String())
	})

	t.Run("only general protocol", func(t *testing.T) {
		t.Setenv(otelExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, otlpProtocolHTTPProtobuf, got)
		assert.Empty(t, buf.String())
	})

	t.Run("only specific protocol", func(t *testing.T) {
		t.Setenv(otelTracesExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, otlpProtocolHTTPProtobuf, got)
		assert.Empty(t, buf.String())
	})

	t.Run("specific overrides general", func(t *testing.T) {
		t.Setenv(otelExporterOTLPProtocolKey, defaultOTLPProtocol)
		t.Setenv(otelTracesExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, otlpProtocolHTTPProtobuf, got)
		assert.Empty(t, buf.String())
	})

	t.Run("invalid specific value", func(t *testing.T) {
		t.Setenv(otelTracesExporterOTLPProtocolKey, invalidProtocol)
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, defaultOTLPProtocol, got)
		assert.Contains(t, buf.String(), fmt.Sprintf("invalid %s: %q", otelTracesExporterOTLPProtocolKey, invalidProtocol))
		assert.Contains(t, buf.String(), "falling back to")
		assert.Contains(t, buf.String(), otelExporterOTLPProtocolKey)
	})

	t.Run("invalid general value", func(t *testing.T) {
		t.Setenv(otelExporterOTLPProtocolKey, invalidProtocol)
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, defaultOTLPProtocol, got)
		assert.Contains(t, buf.String(), fmt.Sprintf("invalid %s: %q", otelExporterOTLPProtocolKey, invalidProtocol))
		assert.Contains(t, buf.String(), "using default")
	})

	t.Run("invalid specific, valid general", func(t *testing.T) {
		t.Setenv(otelTracesExporterOTLPProtocolKey, invalidProtocol)
		t.Setenv(otelExporterOTLPProtocolKey, otlpProtocolHTTPProtobuf)
		buf, logger := newTestLogger()

		got := otlpProtocol(logger, otelTracesExporterOTLPProtocolKey)

		assert.Equal(t, otlpProtocolHTTPProtobuf, got)
		assert.Contains(t, buf.String(), fmt.Sprintf("invalid %s: %q", otelTracesExporterOTLPProtocolKey, invalidProtocol))
		assert.Contains(t, buf.String(), "falling back to")
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
