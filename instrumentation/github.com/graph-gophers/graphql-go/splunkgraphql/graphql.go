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

// Package splunkgraphql provides OpenTelemetry instrumentation for the
// github.com/graph-gophers/graphql-go module.
//
// Deprecated: the module is not going to be released in future.
// See https://github.com/signalfx/splunk-otel-go/issues/4398 for more details.
package splunkgraphql

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/trace/tracer"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	gql "github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql/internal"
	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql"

// otelTracer implements the graphql-go/trace.Tracer interface using
// OpenTelemetry.
type otelTracer struct {
	cfg internal.Config
}

var (
	_ tracer.Tracer           = (*otelTracer)(nil)
	_ tracer.ValidationTracer = (*otelTracer)(nil)
)

// NewTracer returns a new graphql Tracer backed by OpenTelemetry.
func NewTracer(opts ...Option) tracer.Tracer {
	o := append([]internal.Option{
		internal.OptionFunc(func(c *internal.Config) {
			c.Version = Version()
		}),
	}, localToInternal(opts)...)

	cfg := internal.NewConfig(instrumentationName, o...)
	return &otelTracer{cfg: *cfg}
}

func traceQueryFinishFunc(span oteltrace.Span) tracer.ValidationFinishFunc {
	return func(errs []*errors.QueryError) {
		for _, err := range errs {
			span.RecordError(err)
		}
		switch n := len(errs); n {
		case 0:
			// Nothing to do.
		case 1:
			span.SetStatus(codes.Error, errs[0].Error())
		default:
			msg := fmt.Sprintf("%s (and %d more errors)", errs[0], n-1)
			span.SetStatus(codes.Error, msg)
		}
		span.End()
	}
}

// TraceQuery traces a GraphQL query.
func (t *otelTracer) TraceQuery(ctx context.Context, queryString, _ string, _ map[string]interface{}, _ map[string]*introspection.Type) (context.Context, tracer.QueryFinishFunc) { //nolint: gocritic  // un-named returned values.
	spanCtx, span := t.cfg.ResolveTracer(ctx).Start(
		ctx,
		"GraphQL request",
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(gql.GraphQLQueryKey.String(queryString)),
	)

	return spanCtx, traceQueryFinishFunc(span)
}

// TraceField traces a GraphQL field access.
func (t *otelTracer) TraceField(ctx context.Context, _, typeName, fieldName string, trivial bool, _ map[string]interface{}) (context.Context, tracer.FieldFinishFunc) { //nolint: gocritic  // un-named returned values.
	if trivial {
		return ctx, func(*errors.QueryError) {}
	}

	spanCtx, span := t.cfg.ResolveTracer(ctx).Start(
		ctx,
		"GraphQL field",
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(
			gql.GraphQLFieldKey.String(fieldName),
			gql.GraphQLTypeKey.String(typeName),
		),
	)

	return spanCtx, func(err *errors.QueryError) {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
}

// TraceValidation traces the schema validation step preceding an operation.
func (t *otelTracer) TraceValidation(ctx context.Context) tracer.ValidationFinishFunc {
	_, span := t.cfg.ResolveTracer(ctx).Start(
		ctx,
		"GraphQL validation",
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
	)
	return traceQueryFinishFunc(span)
}
