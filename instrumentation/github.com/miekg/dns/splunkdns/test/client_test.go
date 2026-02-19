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

package test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	traceapi "go.opentelemetry.io/otel/trace"

	//nolint:staticcheck // Deprecated module, but still used in this test.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns"
)

var defaultClientAttrs = []attribute.KeyValue{
	attribute.String("default attribute", "client"),
}

func startServer(t *testing.T) *dns.Server {
	// Let the system determine the port.
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		_ = w.WriteMsg(m)
	})

	server := &dns.Server{
		Addr:       pc.LocalAddr().String(),
		Net:        pc.LocalAddr().Network(),
		PacketConn: pc,
		Handler:    mux,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ActivateAndServe()
	}()
	t.Cleanup(func() {
		assert.NoError(t, <-errCh, "failed to start server")
	})
	require.NoError(t, serverUp(pc.LocalAddr().String(), time.Second*10))

	return server
}

func serverUp(addr string, timeout time.Duration) error {
	end := time.NewTimer(timeout)
	defer end.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	msg := new(dns.Msg)
	msg.SetQuestion("miek.nl.", dns.TypeMX)

	for {
		select {
		case <-end.C:
			return errors.New("timed out")
		case <-ticker.C:
			if _, err := dns.Exchange(msg, addr); err == nil {
				return nil
			}
		}
	}
}

func newFixtures(t *testing.T) (*dns.Server, *tracetest.SpanRecorder, []splunkdns.Option, *dns.Msg) {
	server := startServer(t)
	t.Cleanup(func() {
		assert.NoError(t, server.Shutdown())
	})

	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	opts := []splunkdns.Option{
		splunkdns.WithTracerProvider(tp),
		splunkdns.WithAttributes(defaultClientAttrs),
	}

	msg := new(dns.Msg)
	msg.SetQuestion("miek.nl.", dns.TypeMX)

	return server, sr, opts, msg
}

func assertClientSpan(t *testing.T, span trace.ReadOnlySpan) {
	assert.Equal(t, "DNS QUERY", span.Name())
	assert.Equal(t, traceapi.SpanKindClient, span.SpanKind())
	assert.Equal(t, splunkdns.Version(), span.InstrumentationScope().Version)
	attrs := span.Attributes()
	for _, a := range defaultClientAttrs {
		assert.Contains(t, attrs, a)
	}
}

func TestExchange(t *testing.T) {
	server, sr, opts, msg := newFixtures(t)

	_, err := splunkdns.Exchange(msg, server.Addr, opts...)
	assert.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assertClientSpan(t, spans[0])
}

func TestExchangeContext(t *testing.T) {
	server, sr, opts, msg := newFixtures(t)

	_, err := splunkdns.ExchangeContext(context.Background(), msg, server.Addr, opts...)
	assert.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assertClientSpan(t, spans[0])
}

func TestClientExchange(t *testing.T) {
	server, sr, opts, msg := newFixtures(t)

	client := splunkdns.WrapClient(&dns.Client{Net: "udp"}, opts...)
	_, _, err := client.Exchange(msg, server.Addr)
	assert.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assertClientSpan(t, spans[0])
}

func TestClientExchangeContext(t *testing.T) {
	server, sr, opts, msg := newFixtures(t)

	client := splunkdns.WrapClient(&dns.Client{Net: "udp"}, opts...)
	_, _, err := client.ExchangeContext(context.Background(), msg, server.Addr)
	assert.NoError(t, err)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	assertClientSpan(t, spans[0])
}
