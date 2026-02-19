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

// Package splunkbuntdb provides instrumentation for the github.com/tidwall/buntdb
// package.
//
// Deprecated: the module is not going to be released in future.
// See https://github.com/signalfx/splunk-otel-go/issues/4402 for more details.
package splunkbuntdb

import (
	"context"
	"time"

	"github.com/tidwall/buntdb"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// A DB wraps a buntdb.DB, automatically tracing any transactions.
type DB struct {
	*buntdb.DB
	cfg *config
}

// Open calls buntdb.Open and wraps the result.
func Open(path string, opts ...Option) (*DB, error) {
	db, err := buntdb.Open(path)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// WrapDB wraps a buntdb.DB so it can be traced.
func WrapDB(db *buntdb.DB, opts ...Option) *DB {
	return &DB{
		DB:  db,
		cfg: newConfig(opts...),
	}
}

// Begin calls the underlying DB.Begin and traces the transaction.
func (db *DB) Begin(writable bool) (*Tx, error) {
	tx, err := db.DB.Begin(writable)
	if err != nil {
		return nil, err
	}
	return WrapTx(tx, optionFunc(func(c *config) {
		newCopy := db.cfg.copy()
		*c = *newCopy
	})), nil
}

// Update calls the underlying DB.Update and traces the transaction.
func (db *DB) Update(fn func(tx *Tx) error) error {
	return db.DB.Update(func(tx *buntdb.Tx) error {
		return fn(WrapTx(tx, optionFunc(func(c *config) {
			newCopy := db.cfg.copy()
			*c = *newCopy
		})))
	})
}

// View calls the underlying DB.View and traces the transaction.
func (db *DB) View(fn func(tx *Tx) error) error {
	return db.DB.View(func(tx *buntdb.Tx) error {
		return fn(WrapTx(tx, optionFunc(func(c *config) {
			newCopy := db.cfg.copy()
			*c = *newCopy
		})))
	})
}

// WithContext sets the context for the DB.
func (db *DB) WithContext(ctx context.Context) *DB {
	newdb := WrapDB(db.DB, optionFunc(func(c *config) {
		newCopy := db.cfg.copy()
		newCopy.ctx = ctx
		*c = *newCopy
	}))
	return newdb
}

// A Tx wraps a buntdb.Tx, automatically tracing any queries.
type Tx struct {
	*buntdb.Tx
	cfg *config
}

// WrapTx wraps a buntdb.Tx so it can be traced.
func WrapTx(tx *buntdb.Tx, opts ...Option) *Tx {
	return &Tx{
		Tx:  tx,
		cfg: newConfig(opts...),
	}
}

// WithContext sets the context for the Tx.
func (tx *Tx) WithContext(ctx context.Context) *Tx {
	newCfg := tx.cfg.copy()
	newCfg.ctx = ctx
	return &Tx{
		Tx:  tx.Tx,
		cfg: newCfg,
	}
}

// Ascend calls the underlying Tx.Ascend and traces the query.
func (tx *Tx) Ascend(index string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "Ascend", func(context.Context) error {
		return tx.Tx.Ascend(index, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("Ascend")))
}

// AscendEqual calls the underlying Tx.AscendEqual and traces the query.
func (tx *Tx) AscendEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "AscendEqual", func(context.Context) error {
		return tx.Tx.AscendEqual(index, pivot, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("AscendEqual")))
}

// AscendGreaterOrEqual calls the underlying Tx.AscendGreaterOrEqual and traces the query.
func (tx *Tx) AscendGreaterOrEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "AscendGreaterOrEqual", func(context.Context) error {
		return tx.Tx.AscendGreaterOrEqual(index, pivot, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("AscendGreaterOrEqual")))
}

// AscendKeys calls the underlying Tx.AscendKeys and traces the query.
func (tx *Tx) AscendKeys(pattern string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "AscendKeys", func(context.Context) error {
		return tx.Tx.AscendKeys(pattern, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("AscendKeys")))
}

// AscendLessThan calls the underlying Tx.AscendLessThan and traces the query.
func (tx *Tx) AscendLessThan(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "AscendLessThan", func(context.Context) error {
		return tx.Tx.AscendLessThan(index, pivot, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("AscendLessThan")))
}

// AscendRange calls the underlying Tx.AscendRange and traces the query.
func (tx *Tx) AscendRange(index, greaterOrEqual, lessThan string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "AscendRange", func(context.Context) error {
		return tx.Tx.AscendRange(index, greaterOrEqual, lessThan, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("AscendRange")))
}

// CreateIndex calls the underlying Tx.CreateIndex and traces the query.
func (tx *Tx) CreateIndex(name, pattern string, less ...func(a, b string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "CreateIndex", func(context.Context) error {
		return tx.Tx.CreateIndex(name, pattern, less...)
	}, trace.WithAttributes(semconv.DBOperationKey.String("CreateIndex")))
}

// CreateIndexOptions calls the underlying Tx.CreateIndexOptions and traces the query.
func (tx *Tx) CreateIndexOptions(name, pattern string, opts *buntdb.IndexOptions, less ...func(a, b string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "CreateIndexOptions", func(context.Context) error {
		return tx.Tx.CreateIndexOptions(name, pattern, opts, less...)
	}, trace.WithAttributes(semconv.DBOperationKey.String("CreateIndexOptions")))
}

// CreateSpatialIndex calls the underlying Tx.CreateSpatialIndex and traces the query.
func (tx *Tx) CreateSpatialIndex(name, pattern string, rect func(item string) (minimum, maximum []float64)) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "CreateSpatialIndex", func(context.Context) error {
		return tx.Tx.CreateSpatialIndex(name, pattern, rect)
	}, trace.WithAttributes(semconv.DBOperationKey.String("CreateSpatialIndex")))
}

// CreateSpatialIndexOptions calls the underlying Tx.CreateSpatialIndexOptions and traces the query.
func (tx *Tx) CreateSpatialIndexOptions(name, pattern string, opts *buntdb.IndexOptions, rect func(item string) (minimum, maximum []float64)) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "CreateSpatialIndexOptions", func(context.Context) error {
		return tx.Tx.CreateSpatialIndexOptions(name, pattern, opts, rect)
	}, trace.WithAttributes(semconv.DBOperationKey.String("CreateSpatialIndexOptions")))
}

// Delete calls the underlying Tx.Delete and traces the query.
func (tx *Tx) Delete(key string) (val string, err error) {
	err = tx.cfg.WithSpan(tx.cfg.ctx, "Delete", func(context.Context) error {
		var iErr error
		val, iErr = tx.Tx.Delete(key)
		return iErr
	}, trace.WithAttributes(semconv.DBOperationKey.String("Delete")))
	return val, err
}

// DeleteAll calls the underlying Tx.DeleteAll and traces the query.
func (tx *Tx) DeleteAll() error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DeleteAll", func(context.Context) error {
		return tx.Tx.DeleteAll()
	}, trace.WithAttributes(semconv.DBOperationKey.String("DeleteAll")))
}

// Descend calls the underlying Tx.Descend and traces the query.
func (tx *Tx) Descend(index string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "Descend", func(context.Context) error {
		return tx.Tx.Descend(index, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("Descend")))
}

// DescendEqual calls the underlying Tx.DescendEqual and traces the query.
func (tx *Tx) DescendEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DescendEqual", func(context.Context) error {
		return tx.Tx.DescendEqual(index, pivot, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("DescendEqual")))
}

// DescendGreaterThan calls the underlying Tx.DescendGreaterThan and traces the query.
func (tx *Tx) DescendGreaterThan(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DescendGreaterThan", func(context.Context) error {
		return tx.Tx.DescendGreaterThan(index, pivot, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("DescendGreaterThan")))
}

// DescendKeys calls the underlying Tx.DescendKeys and traces the query.
func (tx *Tx) DescendKeys(pattern string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DescendKeys", func(context.Context) error {
		return tx.Tx.DescendKeys(pattern, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("DescendKeys")))
}

// DescendLessOrEqual calls the underlying Tx.DescendLessOrEqual and traces the query.
func (tx *Tx) DescendLessOrEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DescendLessOrEqual", func(context.Context) error {
		return tx.Tx.DescendLessOrEqual(index, pivot, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("DescendLessOrEqual")))
}

// DescendRange calls the underlying Tx.DescendRange and traces the query.
func (tx *Tx) DescendRange(index, lessOrEqual, greaterThan string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DescendRange", func(context.Context) error {
		return tx.Tx.DescendRange(index, lessOrEqual, greaterThan, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("DescendRange")))
}

// DropIndex calls the underlying Tx.DropIndex and traces the query.
func (tx *Tx) DropIndex(name string) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "DropIndex", func(context.Context) error {
		return tx.Tx.DropIndex(name)
	}, trace.WithAttributes(semconv.DBOperationKey.String("DropIndex")))
}

// Get calls the underlying Tx.Get and traces the query.
func (tx *Tx) Get(key string, ignoreExpired ...bool) (val string, err error) {
	err = tx.cfg.WithSpan(tx.cfg.ctx, "Get", func(context.Context) error {
		var iErr error
		val, iErr = tx.Tx.Get(key, ignoreExpired...)
		return iErr
	}, trace.WithAttributes(semconv.DBOperationKey.String("Get")))
	return val, err
}

// Indexes calls the underlying Tx.Indexes and traces the query.
func (tx *Tx) Indexes() (indexes []string, err error) {
	err = tx.cfg.WithSpan(tx.cfg.ctx, "Indexes", func(context.Context) error {
		var iErr error
		indexes, iErr = tx.Tx.Indexes()
		return iErr
	}, trace.WithAttributes(semconv.DBOperationKey.String("Indexes")))
	return indexes, err
}

// Intersects calls the underlying Tx.Intersects and traces the query.
func (tx *Tx) Intersects(index, bounds string, iterator func(key, value string) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "Intersects", func(context.Context) error {
		return tx.Tx.Intersects(index, bounds, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("Intersects")))
}

// Len calls the underlying Tx.Len and traces the query.
func (tx *Tx) Len() (n int, err error) {
	err = tx.cfg.WithSpan(tx.cfg.ctx, "Len", func(context.Context) error {
		var iErr error
		n, iErr = tx.Tx.Len()
		return iErr
	}, trace.WithAttributes(semconv.DBOperationKey.String("Len")))
	return n, err
}

// Nearby calls the underlying Tx.Nearby and traces the query.
func (tx *Tx) Nearby(index, bounds string, iterator func(key, value string, dist float64) bool) error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "Nearby", func(context.Context) error {
		return tx.Tx.Nearby(index, bounds, iterator)
	}, trace.WithAttributes(semconv.DBOperationKey.String("Nearby")))
}

// Set calls the underlying Tx.Set and traces the query.
func (tx *Tx) Set(key, value string, opts *buntdb.SetOptions) (previousValue string, replaced bool, err error) {
	err = tx.cfg.WithSpan(tx.cfg.ctx, "Set", func(context.Context) error {
		var iErr error
		previousValue, replaced, iErr = tx.Tx.Set(key, value, opts)
		return iErr
	}, trace.WithAttributes(semconv.DBOperationKey.String("Set")))
	return previousValue, replaced, err
}

// TTL calls the underlying Tx.TTL and traces the query.
func (tx *Tx) TTL(key string) (duration time.Duration, err error) {
	err = tx.cfg.WithSpan(tx.cfg.ctx, "TTL", func(context.Context) error {
		var iErr error
		duration, iErr = tx.Tx.TTL(key)
		return iErr
	}, trace.WithAttributes(semconv.DBOperationKey.String("TTL")))
	return duration, err
}

// Commit calls the underlying Tx.Commit and traces the query.
func (tx *Tx) Commit() error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "Commit", func(context.Context) error {
		return tx.Tx.Commit()
	}, trace.WithAttributes(semconv.DBOperationKey.String("Commit")))
}

// Rollback calls the underlying Tx.Rollback and traces the query.
func (tx *Tx) Rollback() error {
	return tx.cfg.WithSpan(tx.cfg.ctx, "Rollback", func(context.Context) error {
		return tx.Tx.Rollback()
	}, trace.WithAttributes(semconv.DBOperationKey.String("Rollback")))
}
