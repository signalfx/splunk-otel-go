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

package splunkdns

import (
	"context"
	"errors"

	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/trace"
)

// A Handler wraps a DNS Handler so that requests are traced.
type Handler struct {
	dns.Handler

	cfg *config
}

// WrapHandler creates a new, wrapped DNS handler.
func WrapHandler(handler dns.Handler, opts ...Option) *Handler {
	return &Handler{
		Handler: handler,
		cfg:     newConfig(opts...),
	}
}

// ServeDNS dispatches requests to the underlying Handler. All requests will
// be traced.
func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	_ = h.cfg.withSpan(context.Background(), r, func(context.Context) error {
		rw := &responseWriter{ResponseWriter: w}
		h.Handler.ServeDNS(rw, r)
		return rw.err
	}, trace.WithSpanKind(trace.SpanKindServer))
}

type responseWriter struct {
	dns.ResponseWriter
	err error
}

// WriteMsg writes the message to the response writer. If a non-success rcode
// is set the error in the struct will be non-nil.
func (rw *responseWriter) WriteMsg(msg *dns.Msg) error {
	if msg.Rcode != dns.RcodeSuccess {
		rw.err = errors.New(dns.RcodeToString[msg.Rcode])
	}
	return rw.ResponseWriter.WriteMsg(msg)
}
