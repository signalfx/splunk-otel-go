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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"

	//nolint:staticcheck // Deprecated module, but still used in this test.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb"
)

var expectedValue = []byte("world")

func TestDBOperations(t *testing.T) {
	t.Run("CompactRange", testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
		assert.NoError(t, db.CompactRange(util.Range{}))
	}, "CompactRange"))

	t.Run("Delete", testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
		assert.NoError(t, db.Delete([]byte("hello"), nil))
	}, "Delete"))

	t.Run("Put/Has", testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
		assert.NoError(t, db.Put([]byte("hello"), expectedValue, nil))

		ok, err := db.Has([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.True(t, ok, "should contain key 'hello'")
	}, "Put", "Has"))

	t.Run("Put/Get", testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
		assert.NoError(t, db.Put([]byte("hello"), expectedValue, nil))

		v, err := db.Get([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, v)
	}, "Put", "Get"))

	t.Run("Write", testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
		var batch leveldb.Batch
		batch.Put([]byte("hello"), []byte("world"))
		assert.NoError(t, db.Write(&batch, nil))
	}, "Write"))
}

func TestSnapshotOperations(t *testing.T) {
	t.Run("Has", testSnapshotOp(func(t *testing.T, snapshot *splunkleveldb.Snapshot) {
		ok, err := snapshot.Has([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.True(t, ok, "should contain key 'hello'")
	}, "Has"))

	t.Run("Get", testSnapshotOp(func(t *testing.T, snapshot *splunkleveldb.Snapshot) {
		v, err := snapshot.Get([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, v)
	}, "Get"))
}

func TestTransactionOperations(t *testing.T) {
	t.Run("Commit", testTransactionOp(func(t *testing.T, transaction *splunkleveldb.Transaction) {
		assert.NoError(t, transaction.Commit())
	}, "Commit"))

	t.Run("Discard", testTransactionOp(func(_ *testing.T, transaction *splunkleveldb.Transaction) {
		transaction.Discard()
	}, "Discard"))

	t.Run("Delete", testTransactionOp(func(t *testing.T, transaction *splunkleveldb.Transaction) {
		assert.NoError(t, transaction.Delete([]byte("hello"), nil))
	}, "Delete"))

	t.Run("Put/Has", testTransactionOp(func(t *testing.T, transaction *splunkleveldb.Transaction) {
		assert.NoError(t, transaction.Put([]byte("hello"), expectedValue, nil))

		ok, err := transaction.Has([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.True(t, ok, "should contain key 'hello'")
	}, "Put", "Has"))

	t.Run("Put/Get", testTransactionOp(func(t *testing.T, transaction *splunkleveldb.Transaction) {
		assert.NoError(t, transaction.Put([]byte("hello"), expectedValue, nil))

		v, err := transaction.Get([]byte("hello"), nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, v)
	}, "Put", "Get"))

	t.Run("Write", testTransactionOp(func(t *testing.T, transaction *splunkleveldb.Transaction) {
		var batch leveldb.Batch
		batch.Put([]byte("hello"), []byte("world"))
		assert.NoError(t, transaction.Write(&batch, nil))
	}, "Write"))
}

func TestIteratorOperation(t *testing.T) {
	testDBOp(func(_ *testing.T, db *splunkleveldb.DB) {
		db.NewIterator(nil, nil).Release()
	}, "Iterator")(t)
}

func withTestingDeadline(ctx context.Context, t *testing.T) context.Context {
	d, ok := t.Deadline()
	if !ok {
		d = time.Now().Add(10 * time.Second)
	} else {
		d = d.Add(-time.Millisecond)
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithDeadline(ctx, d)
	t.Cleanup(cancel)
	return ctx
}

func assertSpansFunc(parent string, traceID traceapi.TraceID, names ...string) func(*testing.T, []trace.ReadOnlySpan) {
	expected := make(map[string]struct{}, len(names))
	for _, n := range names {
		expected[n] = struct{}{}
	}
	got := make(map[string]struct{}, len(names))

	return func(t *testing.T, spans []trace.ReadOnlySpan) {
		for _, span := range spans {
			name := span.Name()
			got[name] = struct{}{}

			if name == parent {
				continue
			}

			if _, ok := expected[name]; !ok {
				t.Errorf("unexpected span %q created", name)
				continue
			}

			assert.Equal(t, traceapi.SpanKindClient, span.SpanKind())
			assert.Equal(t, traceID, span.SpanContext().TraceID())
			assert.Equal(t, splunkleveldb.Version(), span.InstrumentationScope().Version)
			assert.Contains(t, span.Attributes(), semconv.DBSystemKey.String("leveldb"))
			assert.Contains(t, span.Attributes(), semconv.NetTransportInProc)
			assert.Contains(t, span.Attributes(), semconv.DBOperationKey.String(name))
		}

		for k := range expected {
			if _, ok := got[k]; !ok {
				t.Errorf("expected span %q, none created", k)
			}
		}
	}
}

func testDBOp(f func(*testing.T, *splunkleveldb.DB), spanNames ...string) func(*testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	pname := "Parent Span"
	ctx, parent := tp.Tracer("testDBOp").Start(context.Background(), pname)

	assertSpans := assertSpansFunc(pname, parent.SpanContext().TraceID(), spanNames...)

	return func(t *testing.T) {
		ctx = withTestingDeadline(ctx, t)

		db, err := splunkleveldb.Open(storage.NewMemStorage(), &opt.Options{})
		require.NoError(t, err)
		db = db.WithContext(ctx)

		f(t, db)

		parent.End()
		require.NoError(t, tp.Shutdown(ctx))

		assertSpans(t, sr.Ended())
	}
}

func testSnapshotOp(f func(*testing.T, *splunkleveldb.Snapshot), spanNames ...string) func(*testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	pname := "Parent Snapshot Span"
	ctx, parent := tp.Tracer("testSnapshotOp").Start(context.Background(), pname)

	assertSpans := assertSpansFunc(pname, parent.SpanContext().TraceID(), spanNames...)
	return func(t *testing.T) {
		ctx = withTestingDeadline(ctx, t)
		testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
			require.NoError(t, db.Put([]byte("hello"), expectedValue, nil))

			snapshot, err := db.GetSnapshot()
			require.NoError(t, err)

			// This should not affect the snapshot.
			require.NoError(t, db.Delete([]byte("hello"), nil))

			// Reset the context to use the TracerProvider from this tests'
			// parent span.
			snapshot = snapshot.WithContext(ctx)
			f(t, snapshot)

			snapshot.Release()
			parent.End()
			require.NoError(t, tp.Shutdown(ctx))

			assertSpans(t, sr.Ended())
		}, "Put", "Delete")(t)
	}
}

func testTransactionOp(f func(*testing.T, *splunkleveldb.Transaction), spanNames ...string) func(*testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	pname := "Parent Transaction Span"
	ctx, parent := tp.Tracer("testTransactionOp").Start(context.Background(), pname)

	assertSpans := assertSpansFunc(pname, parent.SpanContext().TraceID(), spanNames...)
	return func(t *testing.T) {
		ctx = withTestingDeadline(ctx, t)
		testDBOp(func(t *testing.T, db *splunkleveldb.DB) {
			transaction, err := db.OpenTransaction()
			require.NoError(t, err)

			// Reset the context to use the TracerProvider from this tests'
			// parent span.
			transaction = transaction.WithContext(ctx)
			f(t, transaction)

			parent.End()
			require.NoError(t, tp.Shutdown(ctx))

			assertSpans(t, sr.Ended())
		})(t)
	}
}

type errIterator struct {
	iterator.Iterator
}

var errExpected = errors.New("expected error")

func (errIterator) Error() error {
	return errExpected
}

func TestIteratorErrors(t *testing.T) {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	errIter := errIterator{db.NewIterator(nil, nil)}
	i := splunkleveldb.WrapIterator(errIter, splunkleveldb.WithTracerProvider(tp))
	i.Release()

	ctx := withTestingDeadline(context.Background(), t)
	require.NoError(t, tp.Shutdown(ctx))

	spans := sr.Ended()
	require.Len(t, spans, 1)
	span := spans[0]
	assert.Equal(t, codes.Error, span.Status().Code)
	assert.Equal(t, errExpected.Error(), span.Status().Description)
}
