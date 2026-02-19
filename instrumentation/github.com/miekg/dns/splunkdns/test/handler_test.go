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
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	traceapi "go.opentelemetry.io/otel/trace"

	//nolint:staticcheck // Deprecated module, but still used in this test.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns"
)

var defaultServerAttrs = []attribute.KeyValue{
	attribute.String("default attribute", "server"),
}

func TestHandler(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	opts := []splunkdns.Option{
		splunkdns.WithTracerProvider(tp),
		splunkdns.WithAttributes(defaultServerAttrs),
	}

	// Let the system determine the port.
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		_ = w.WriteMsg(m)
	})
	errCode := dns.RcodeFormatError
	mux.HandleFunc("error.", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetRcode(r, errCode)
		_ = w.WriteMsg(m)
	})
	handler := splunkdns.WrapHandler(mux, opts...)

	server := &dns.Server{
		Addr:       pc.LocalAddr().String(),
		Net:        pc.LocalAddr().Network(),
		PacketConn: pc,
		Handler:    handler,
	}
	go func() {
		_ = server.ActivateAndServe()
	}()
	// serverUp will make a request to the server that will generate a span.
	require.NoError(t, serverUp(pc.LocalAddr().String(), time.Second*10))

	// Generate an errored span.
	msg := new(dns.Msg)
	msg.SetQuestion("error.", dns.TypeMX)
	_, err = dns.Exchange(msg, pc.LocalAddr().String())
	require.NoError(t, err)

	// Ensure all requests are handled and all expected spans are ended.
	require.NoError(t, server.Shutdown())

	spans := sr.Ended()
	require.Len(t, spans, 2)
	for _, span := range spans {
		assert.Equal(t, "DNS QUERY", span.Name())
		assert.Equal(t, traceapi.SpanKindServer, span.SpanKind())
		attrs := span.Attributes()
		for _, a := range defaultServerAttrs {
			assert.Contains(t, attrs, a)
		}
	}

	errSpan := spans[1]
	assert.Equal(t, codes.Error, errSpan.Status().Code)
	assert.Equal(t, dns.RcodeToString[errCode], errSpan.Status().Description)
}
