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

package splunkleveldb

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb"

// config contains tracing configuration options.
type config struct {
	*internal.Config

	ctx context.Context
}

func newConfig(options ...Option) *config {
	c := config{
		Config: internal.NewConfig(instrumentationName, internal.OptionFunc(
			func(c *internal.Config) {
				c.DefaultStartOpts = []trace.SpanStartOption{
					trace.WithAttributes(
						semconv.DBSystemKey.String("leveldb"),
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

// Option applies options to a configuration.
type Option interface {
	apply(*config)
}

type optConv struct {
	iOpt internal.Option
}

func (o optConv) apply(c *config) {
	o.iOpt.Apply(c.Config)
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optConv{iOpt: internal.WithTracerProvider(tp)}
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return optConv{iOpt: internal.WithAttributes(attr)}
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
