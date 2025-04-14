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

package splunkbuntdb

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb"

type config struct {
	*internal.Config

	ctx context.Context
}

func newConfig(options ...Option) *config {
	c := config{
		Config: internal.NewConfig(instrumentationName, internal.OptionFunc(
			func(c *internal.Config) {
				c.Version = Version()
				c.DefaultStartOpts = []trace.SpanStartOption{
					trace.WithAttributes(
						semconv.DBSystemKey.String("buntdb"),
						semconv.NetTransportInProc,
					),
					// From the specification: span kind MUST always be CLIENT.
					trace.WithSpanKind(trace.SpanKindClient),
				}
			}),
		),
		ctx: context.Background(),
	}

	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
	}

	return &c
}

// copy returns a deep copy of c.
func (c *config) copy() *config {
	return &config{
		Config: c.Copy(),
		ctx:    c.ctx,
	}
}

// Option applies options to a configuration.
type Option interface {
	apply(*config)
}

type optionConv struct {
	iOpt internal.Option
}

func (o optionConv) apply(c *config) {
	o.iOpt.Apply(c.Config)
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optionConv{iOpt: internal.WithTracerProvider(tp)}
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return optionConv{iOpt: internal.WithAttributes(attr)}
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithContext returns an Option that sets the Context used with this
// instrumentation library by default. This is used to pass context of any
// existing trace to the instrumentation.
func WithContext(ctx context.Context) Option {
	return optionFunc(func(c *config) {
		c.ctx = ctx
	})
}
