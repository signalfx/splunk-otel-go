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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tidwall/buntdb"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb"
)

func TestAscend(t *testing.T) {
	testView(t, "Ascend", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.Ascend("test-index", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:a", "1",
			"regular:b", "2",
			"regular:c", "3",
			"regular:d", "4",
			"regular:e", "5",
		}, arr)
		return nil
	})
}

func TestAscendEqual(t *testing.T) {
	testView(t, "AscendEqual", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.AscendEqual("test-index", "3", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"regular:c", "3"}, arr)
		return nil
	})
}

func TestAscendGreaterOrEqual(t *testing.T) {
	testView(t, "AscendGreaterOrEqual", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.AscendGreaterOrEqual("test-index", "3", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:c", "3",
			"regular:d", "4",
			"regular:e", "5",
		}, arr)
		return nil
	})
}

func TestAscendKeys(t *testing.T) {
	testView(t, "AscendKeys", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.AscendKeys("regular:*", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:a", "1",
			"regular:b", "2",
			"regular:c", "3",
			"regular:d", "4",
			"regular:e", "5",
		}, arr)
		return nil
	})
}

func TestAscendLessThan(t *testing.T) {
	testView(t, "AscendLessThan", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.AscendLessThan("test-index", "3", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:a", "1",
			"regular:b", "2",
		}, arr)
		return nil
	})
}

func TestAscendRange(t *testing.T) {
	testView(t, "AscendRange", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.AscendRange("test-index", "2", "4", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:b", "2",
			"regular:c", "3",
		}, arr)
		return nil
	})
}

func TestCreateIndex(t *testing.T) {
	testUpdate(t, "CreateIndex", func(tx *splunkbuntdb.Tx) error {
		err := tx.CreateIndex("test-create-index", "*")
		assert.NoError(t, err)
		return nil
	})
}

func TestCreateIndexOptions(t *testing.T) {
	testUpdate(t, "CreateIndexOptions", func(tx *splunkbuntdb.Tx) error {
		err := tx.CreateIndexOptions("test-create-index", "*", nil)
		assert.NoError(t, err)
		return nil
	})
}

func TestCreateSpatialIndex(t *testing.T) {
	testUpdate(t, "CreateSpatialIndex", func(tx *splunkbuntdb.Tx) error {
		err := tx.CreateSpatialIndex("test-create-index", "*", buntdb.IndexRect)
		assert.NoError(t, err)
		return nil
	})
}

func TestCreateSpatialIndexOptions(t *testing.T) {
	testUpdate(t, "CreateSpatialIndexOptions", func(tx *splunkbuntdb.Tx) error {
		err := tx.CreateSpatialIndexOptions("test-create-index", "*", nil, buntdb.IndexRect)
		assert.NoError(t, err)
		return nil
	})
}

func TestDelete(t *testing.T) {
	testUpdate(t, "Delete", func(tx *splunkbuntdb.Tx) error {
		val, err := tx.Delete("regular:a")
		assert.NoError(t, err)
		assert.Equal(t, "1", val)
		return nil
	})
}

func TestDeleteAll(t *testing.T) {
	testUpdate(t, "DeleteAll", func(tx *splunkbuntdb.Tx) error {
		err := tx.DeleteAll()
		assert.NoError(t, err)
		return nil
	})
}

func TestDescend(t *testing.T) {
	testView(t, "Descend", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.Descend("test-index", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:e", "5",
			"regular:d", "4",
			"regular:c", "3",
			"regular:b", "2",
			"regular:a", "1",
		}, arr)
		return nil
	})
}

func TestDescendEqual(t *testing.T) {
	testView(t, "DescendEqual", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.DescendEqual("test-index", "3", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"regular:c", "3"}, arr)
		return nil
	})
}

func TestDescendGreaterThan(t *testing.T) {
	testView(t, "DescendGreaterThan", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.DescendGreaterThan("test-index", "3", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:e", "5",
			"regular:d", "4",
		}, arr)
		return nil
	})
}

func TestDescendKeys(t *testing.T) {
	testView(t, "DescendKeys", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.DescendKeys("regular:*", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:e", "5",
			"regular:d", "4",
			"regular:c", "3",
			"regular:b", "2",
			"regular:a", "1",
		}, arr)
		return nil
	})
}

func TestDescendLessOrEqual(t *testing.T) {
	testView(t, "DescendLessOrEqual", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.DescendLessOrEqual("test-index", "3", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:c", "3",
			"regular:b", "2",
			"regular:a", "1",
		}, arr)
		return nil
	})
}

func TestDescendRange(t *testing.T) {
	testView(t, "DescendRange", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.DescendRange("test-index", "4", "2", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"regular:d", "4",
			"regular:c", "3",
		}, arr)
		return nil
	})
}

func TestDropIndex(t *testing.T) {
	testUpdate(t, "DropIndex", func(tx *splunkbuntdb.Tx) error {
		err := tx.DropIndex("test-index")
		assert.NoError(t, err)
		return nil
	})
}

func TestGet(t *testing.T) {
	testView(t, "Get", func(tx *splunkbuntdb.Tx) error {
		val, err := tx.Get("regular:a")
		assert.NoError(t, err)
		assert.Equal(t, "1", val)
		return nil
	})
}

func TestIndexes(t *testing.T) {
	testView(t, "Indexes", func(tx *splunkbuntdb.Tx) error {
		indexes, err := tx.Indexes()
		assert.NoError(t, err)
		assert.Equal(t, []string{"test-index", "test-spatial-index"}, indexes)
		return nil
	})
}

func TestIntersects(t *testing.T) {
	testView(t, "Intersects", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.Intersects("test-spatial-index", "[3 3],[4 4]", func(key, value string) bool {
			arr = append(arr, key, value)
			return true
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"spatial:c", "[3 3]",
			"spatial:d", "[4 4]",
		}, arr)
		return nil
	})
}

func TestLen(t *testing.T) {
	testView(t, "Len", func(tx *splunkbuntdb.Tx) error {
		n, err := tx.Len()
		assert.NoError(t, err)
		assert.Equal(t, 10, n)
		return nil
	})
}

func TestNearby(t *testing.T) {
	testView(t, "Nearby", func(tx *splunkbuntdb.Tx) error {
		var arr []string
		err := tx.Nearby("test-spatial-index", "[3 3]", func(key, value string, _ float64) bool {
			arr = append(arr, key, value)
			return false
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"spatial:c", "[3 3]",
		}, arr)
		return nil
	})
}

func TestSet(t *testing.T) {
	testUpdate(t, "Set", func(tx *splunkbuntdb.Tx) error {
		previousValue, replaced, err := tx.Set("regular:a", "11", nil)
		assert.NoError(t, err)
		assert.True(t, replaced)
		assert.Equal(t, "1", previousValue)
		return nil
	})
}

func TestTTL(t *testing.T) {
	testUpdate(t, "TTL", func(tx *splunkbuntdb.Tx) error {
		duration, err := tx.TTL("regular:a")
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-1), duration)
		return nil
	})
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

func assertSpan(t *testing.T, name string, span trace.ReadOnlySpan) {
	assert.Equal(t, span.SpanKind(), traceapi.SpanKindClient)
	assert.Equal(t, splunkbuntdb.Version(), span.InstrumentationScope().Version)
	assert.Contains(t, span.Attributes(), semconv.DBSystemKey.String("buntdb"))
	assert.Contains(t, span.Attributes(), semconv.DBOperationKey.String(name))
	assert.Equal(t, span.Name(), name)
}

func testUpdate(t *testing.T, name string, f func(tx *splunkbuntdb.Tx) error) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	db := getDatabase(t, splunkbuntdb.WithTracerProvider(tp))
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	ctx := withTestingDeadline(context.Background(), t)

	err := db.WithContext(ctx).Update(f)
	assert.NoError(t, err)

	require.NoError(t, tp.Shutdown(ctx))
	spans := sr.Ended()
	require.Len(t, spans, 1)

	assertSpan(t, name, spans[0])
}

func testView(t *testing.T, name string, f func(tx *splunkbuntdb.Tx) error) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	db := getDatabase(t, splunkbuntdb.WithTracerProvider(tp))
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	ctx := withTestingDeadline(context.Background(), t)

	err := db.WithContext(ctx).View(f)
	assert.NoError(t, err)

	require.NoError(t, tp.Shutdown(ctx))
	spans := sr.Ended()
	require.Len(t, spans, 1)

	assertSpan(t, name, spans[0])
}

func TestCommit(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	db := getDatabase(t, splunkbuntdb.WithTracerProvider(tp))
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	tx, err := db.Begin(true)
	assert.NoError(t, err)

	previousValue, replaced, err := tx.Set("regular:a", "7", nil)
	assert.NoError(t, err)
	assert.True(t, replaced)
	assert.Equal(t, "1", previousValue)

	err = tx.Commit()
	assert.NoError(t, err)

	err = db.View(func(tx *splunkbuntdb.Tx) error {
		val, errIn := tx.Get("regular:a")
		assert.NoError(t, errIn)
		assert.Equal(t, "7", val)
		return nil
	})
	assert.NoError(t, err)

	ctx := withTestingDeadline(context.Background(), t)
	require.NoError(t, tp.Shutdown(ctx))
	spans := sr.Ended()
	require.Len(t, spans, 3)

	assertSpan(t, "Set", spans[0])
	assertSpan(t, "Commit", spans[1])
	assertSpan(t, "Get", spans[2])
}

func TestRollback(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	db := getDatabase(t, splunkbuntdb.WithTracerProvider(tp))
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	tx, err := db.Begin(true)
	assert.NoError(t, err)

	previousValue, replaced, err := tx.Set("regular:a", "11", nil)
	assert.NoError(t, err)
	assert.True(t, replaced)
	assert.Equal(t, "1", previousValue)

	err = tx.Rollback()
	assert.NoError(t, err)

	err = db.View(func(tx *splunkbuntdb.Tx) error {
		val, errIn := tx.Get("regular:a")
		assert.NoError(t, errIn)
		assert.Equal(t, "1", val)
		return nil
	})
	assert.NoError(t, err)

	ctx := withTestingDeadline(context.Background(), t)
	require.NoError(t, tp.Shutdown(ctx))
	spans := sr.Ended()
	require.Len(t, spans, 3)

	assertSpan(t, "Set", spans[0])
	assertSpan(t, "Rollback", spans[1])
	assertSpan(t, "Get", spans[2])
}

func getDatabase(t *testing.T, opts ...splunkbuntdb.Option) *splunkbuntdb.DB {
	bdb, err := buntdb.Open(":memory:")
	require.NoError(t, err)

	err = bdb.CreateIndex("test-index", "regular:*", buntdb.IndexBinary)
	require.NoError(t, err)

	err = bdb.CreateSpatialIndex("test-spatial-index", "spatial:*", buntdb.IndexRect)
	require.NoError(t, err)

	require.NoError(t, bdb.Update(func(tx *buntdb.Tx) error {
		_, _, _ = tx.Set("regular:a", "1", nil)
		_, _, _ = tx.Set("regular:b", "2", nil)
		_, _, _ = tx.Set("regular:c", "3", nil)
		_, _, _ = tx.Set("regular:d", "4", nil)
		_, _, _ = tx.Set("regular:e", "5", nil)

		_, _, _ = tx.Set("spatial:a", "[1 1]", nil)
		_, _, _ = tx.Set("spatial:b", "[2 2]", nil)
		_, _, _ = tx.Set("spatial:c", "[3 3]", nil)
		_, _, _ = tx.Set("spatial:d", "[4 4]", nil)
		_, _, _ = tx.Set("spatial:e", "[5 5]", nil)

		return nil
	}))

	return splunkbuntdb.WrapDB(bdb, opts...)
}
