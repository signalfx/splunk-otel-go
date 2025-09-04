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
	"time"

	"github.com/miekg/dns"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// A Client wraps a DNS Client so that requests are traced.
type Client struct {
	*dns.Client

	cfg *internal.Config
}

// WrapClient returns a wraped DNS client.
func WrapClient(client *dns.Client, opts ...Option) *Client {
	o := append([]internal.Option{
		internal.OptionFunc(func(c *internal.Config) {
			c.Version = Version()
		}),
	}, localToInternal(opts)...)

	return &Client{
		Client: client,
		cfg: internal.NewConfig(
			instrumentationName,
			o...,
		),
	}
}

// Exchange calls the underlying Client.Exchange and traces the request.
func (c *Client) Exchange(m *dns.Msg, addr string) (*dns.Msg, time.Duration, error) {
	return c.ExchangeContext(context.Background(), m, addr)
}

// ExchangeContext calls the underlying Client.ExchangeContext and traces the
// request.
func (c *Client) ExchangeContext(ctx context.Context, m *dns.Msg, addr string) (resp *dns.Msg, rtt time.Duration, err error) {
	err = c.cfg.WithSpan(ctx, name(m), func(ctx context.Context) error {
		var sErr error
		resp, rtt, sErr = c.Client.ExchangeContext(ctx, m, addr)
		return sErr
	}, trace.WithSpanKind(trace.SpanKindClient))
	return resp, rtt, err
}
