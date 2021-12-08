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
package splunkgraphql

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	"github.com/graph-gophers/graphql-go/trace"
	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql"

var (
	graphqlFieldKey = attribute.Key("graphql.field")
	graphqlQueryKey = attribute.Key("graphql.query")
	graphqlTypeKey  = attribute.Key("graphql.type")
)

// otelTracer implements the graphql-go/trace.Tracer interface using
// OpenTelemetry.
type otelTracer struct {
	cfg internal.Config
}

var (
	_ trace.Tracer                  = (*otelTracer)(nil)
	_ trace.ValidationTracerContext = (*otelTracer)(nil)
)

// NewTracer returns a new trace.Tracer backed by OpenTelemetry.
func NewTracer(opts ...Option) trace.Tracer {
	cfg := internal.NewConfig(instrumentationName, localToInternal(opts)...)
	return &otelTracer{cfg: *cfg}
}

func traceQueryFinishFunc(span oteltrace.Span) trace.TraceQueryFinishFunc {
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
func (t *otelTracer) TraceQuery(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, varTypes map[string]*introspection.Type) (context.Context, trace.TraceQueryFinishFunc) {
	spanCtx, span := t.cfg.ResolveTracer(ctx).Start(
		ctx,
		"GraphQL request",
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		oteltrace.WithAttributes(graphqlQueryKey.String(queryString)),
	)

	return spanCtx, traceQueryFinishFunc(span)
}

// TraceField traces a GraphQL field access.
func (t *otelTracer) TraceField(ctx context.Context, label, typeName, fieldName string, trivial bool, args map[string]interface{}) (context.Context, trace.TraceFieldFinishFunc) {
	if trivial {
		return ctx, func(*errors.QueryError) {}
	}

	spanCtx, span := t.cfg.ResolveTracer(ctx).Start(
		ctx,
		"GraphQL field",
		oteltrace.WithAttributes(
			graphqlFieldKey.String(fieldName),
			graphqlTypeKey.String(typeName),
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
func (t *otelTracer) TraceValidation(ctx context.Context) trace.TraceValidationFinishFunc {
	_, span := t.cfg.ResolveTracer(ctx).Start(ctx, "GraphQL validation")
	return traceQueryFinishFunc(span)
}
