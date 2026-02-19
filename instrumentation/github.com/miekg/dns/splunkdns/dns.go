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

// Package splunkdns provides OpenTelemetry instrumentation for the
// github.com/miekg/dns package.
//
// Deprecated: this module is no longer supported.
// See https://github.com/signalfx/splunk-otel-go/issues/4400 for more details.
package splunkdns

import (
	"context"

	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// ListenAndServe wraps the passed handler so all requests it servers are
// traced and starts a server on addr and network to handle incoming queries.
func ListenAndServe(addr, network string, handler dns.Handler, opts ...Option) error {
	return dns.ListenAndServe(addr, network, WrapHandler(handler, opts...))
}

// ListenAndServeTLS wraps the passed handler so all requests it servers are
// traced and starts a server on addr and network to handle incoming queries
// using the passed TLS certFile and keyFile.
func ListenAndServeTLS(addr, certFile, keyFile string, handler dns.Handler, opts ...Option) error {
	return dns.ListenAndServeTLS(addr, certFile, keyFile, WrapHandler(handler, opts...))
}

// Exchange performs a traced synchronous UDP query. It sends the message m to
// addr and waits for a reply. Exchange does not retry a failed query, nor
// will it fall back to TCP in case of truncation.
func Exchange(m *dns.Msg, addr string, opts ...Option) (*dns.Msg, error) {
	return ExchangeContext(context.Background(), m, addr, opts...)
}

// ExchangeContext performs a traced synchronous UDP query, like Exchange. It
// additionally obeys deadlines from the passed Context.
func ExchangeContext(ctx context.Context, m *dns.Msg, addr string, opts ...Option) (r *dns.Msg, err error) {
	o := append([]internal.Option{
		internal.OptionFunc(func(c *internal.Config) {
			c.Version = Version()
		}),
	}, localToInternal(opts)...)

	cfg := internal.NewConfig(instrumentationName, o...)
	err = cfg.WithSpan(ctx, name(m), func(c context.Context) error {
		var sErr error
		r, sErr = dns.ExchangeContext(c, m, addr)
		return sErr
	}, trace.WithSpanKind(trace.SpanKindClient))
	return r, err
}
