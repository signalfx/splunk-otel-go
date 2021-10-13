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

// Package splunkdns provides instrumentation for the github.com/miekg/dns
// package.
package splunkdns

import (
	"context"

	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/trace"
)

// ListenAndServe calls dns.ListenAndServe with a wrapped Handler.
func ListenAndServe(addr string, network string, handler dns.Handler, opts ...Option) error {
	return dns.ListenAndServe(addr, network, WrapHandler(handler, opts...))
}

// ListenAndServeTLS calls dns.ListenAndServeTLS with a wrapped Handler.
func ListenAndServeTLS(addr, certFile, keyFile string, handler dns.Handler, opts ...Option) error {
	return dns.ListenAndServeTLS(addr, certFile, keyFile, WrapHandler(handler, opts...))
}

// Exchange calls dns.Exchange and traces the request.
func Exchange(m *dns.Msg, addr string, opts ...Option) (*dns.Msg, error) {
	return ExchangeContext(context.Background(), m, addr, opts...)
}

// ExchangeContext calls dns.ExchangeContext and traces the request.
func ExchangeContext(ctx context.Context, m *dns.Msg, addr string, opts ...Option) (r *dns.Msg, err error) {
	newConfig(opts...).withSpan(ctx, m, func(c context.Context) error {
		r, err = dns.ExchangeContext(c, m, addr)
		return err
	}, trace.WithSpanKind(trace.SpanKindClient))
	return
}
