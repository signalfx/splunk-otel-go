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
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// iter wraps a leveldb.Iterator, tracing all operations performed.
type iter struct {
	iterator.Iterator
	span trace.Span
}

// WrapIterator returns a traced Iterator that wraps a leveldb
// iterator.Iterator with a span. The span is ended when Iterator.Release is
// called.
func WrapIterator(it iterator.Iterator, opts ...Option) iterator.Iterator {
	c := newConfig(opts...)
	sso := c.MergedSpanStartOptions(
		trace.WithAttributes(semconv.DBOperationKey.String("Iterator")),
	)
	_, span := c.ResolveTracer(c.ctx).Start(c.ctx, "Iterator", sso...)
	return &iter{
		Iterator: it,
		span:     span,
	}
}

// Release releases associated resources and ends any active span.
func (it *iter) Release() {
	if err := it.Error(); err != nil {
		it.span.RecordError(err)
		it.span.SetStatus(codes.Error, err.Error())
	}
	it.Iterator.Release()
	it.span.End()
}
