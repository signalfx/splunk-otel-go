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
package splunkbuntdb

import (
	"context"
	"time"

	"github.com/tidwall/buntdb"
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
		cfg: newConfig(),
		// cfg: newConfig(opts), // TODO: fixme
	}
}

// Begin calls the underlying DB.Begin and traces the transaction.
func (db *DB) Begin(writable bool) (*Tx, error) {
	tx, err := db.DB.Begin(writable)
	if err != nil {
		return nil, err
	}
	return WrapTx(tx, db.cfg), nil
}

// Update calls the underlying DB.Update and traces the transaction.
func (db *DB) Update(fn func(tx *Tx) error) error {
	return db.DB.Update(func(tx *buntdb.Tx) error {
		return fn(WrapTx(tx, db.cfg))
	})
}

// View calls the underlying DB.View and traces the transaction.
func (db *DB) View(fn func(tx *Tx) error) error {
	return db.DB.View(func(tx *buntdb.Tx) error {
		return fn(WrapTx(tx, db.cfg))
	})
}

// WithContext sets the context for the DB.
func (db *DB) WithContext(ctx context.Context) *DB {
	newdb := WrapDB(db.DB, optionFunc(func(c *config) {
		// FIXME: add a cfg.copy method that will make a deep copy.
		// specifically the options slice.
		newCopy := copyConfig(db.cfg)
		*c = *newCopy
	}))
	return newdb
}

// outside context
// db.WIthContext(newContext).Update(...)
// returns

// A Tx wraps a buntdb.Tx, automatically tracing any queries.
type Tx struct {
	*buntdb.Tx
	cfg *config
}

// WrapTx wraps a buntdb.Tx so it can be traced.
func WrapTx(tx *buntdb.Tx, cfg *config) *Tx {
	return &Tx{
		Tx:  tx,
		cfg: cfg,
	}
}

// WithContext sets the context for the Tx.
func (tx *Tx) WithContext(ctx context.Context) *Tx {
	newcfg := *tx.cfg
	newcfg.ctx = ctx
	return &Tx{
		Tx:  tx.Tx,
		cfg: &newcfg,
	}
}

// Ascend calls the underlying Tx.Ascend and traces the query.
func (tx *Tx) Ascend(index string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("Ascend", func() error {
		return tx.Tx.Ascend(index, iterator)
	})
}

// AscendEqual calls the underlying Tx.AscendEqual and traces the query.
func (tx *Tx) AscendEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("AscendEqual", func() error {
		return tx.Tx.AscendEqual(index, pivot, iterator)
	})
}

// AscendGreaterOrEqual calls the underlying Tx.AscendGreaterOrEqual and traces the query.
func (tx *Tx) AscendGreaterOrEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("AscendGreaterOrEqual", func() error {
		return tx.Tx.AscendGreaterOrEqual(index, pivot, iterator)
	})
}

// AscendKeys calls the underlying Tx.AscendKeys and traces the query.
func (tx *Tx) AscendKeys(pattern string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("AscendKeys", func() error {
		return tx.Tx.AscendKeys(pattern, iterator)
	})
}

// AscendLessThan calls the underlying Tx.AscendLessThan and traces the query.
func (tx *Tx) AscendLessThan(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("AscendLessThan", func() error {
		return tx.Tx.AscendLessThan(index, pivot, iterator)
	})
}

// AscendRange calls the underlying Tx.AscendRange and traces the query.
func (tx *Tx) AscendRange(index, greaterOrEqual, lessThan string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("AscendRange", func() error {
		return tx.Tx.AscendRange(index, greaterOrEqual, lessThan, iterator)
	})
}

// CreateIndex calls the underlying Tx.CreateIndex and traces the query.
func (tx *Tx) CreateIndex(name, pattern string, less ...func(a, b string) bool) error {
	return tx.cfg.withSpan("CreateIndex", func() error {
		return tx.Tx.CreateIndex(name, pattern, less...)
	})
}

// CreateIndexOptions calls the underlying Tx.CreateIndexOptions and traces the query.
func (tx *Tx) CreateIndexOptions(name, pattern string, opts *buntdb.IndexOptions, less ...func(a, b string) bool) error {
	return tx.cfg.withSpan("CreateIndexOptions", func() error {
		return tx.Tx.CreateIndexOptions(name, pattern, opts, less...)
	})
}

// CreateSpatialIndex calls the underlying Tx.CreateSpatialIndex and traces the query.
func (tx *Tx) CreateSpatialIndex(name, pattern string, rect func(item string) (min, max []float64)) error {
	return tx.cfg.withSpan("CreateSpatialIndex", func() error {
		return tx.Tx.CreateSpatialIndex(name, pattern, rect)
	})
}

// CreateSpatialIndexOptions calls the underlying Tx.CreateSpatialIndexOptions and traces the query.
func (tx *Tx) CreateSpatialIndexOptions(name, pattern string, opts *buntdb.IndexOptions, rect func(item string) (min, max []float64)) error {
	return tx.cfg.withSpan("CreateSpatialIndexOptions", func() error {
		return tx.Tx.CreateSpatialIndexOptions(name, pattern, opts, rect)
	})
}

// Delete calls the underlying Tx.Delete and traces the query.
func (tx *Tx) Delete(key string) (val string, err error) {
	err = tx.cfg.withSpan("CreateSpatialIndexOptions", func() error {
		var iErr error
		val, iErr = tx.Tx.Delete(key)
		return iErr
	})
	return
}

// DeleteAll calls the underlying Tx.DeleteAll and traces the query.
func (tx *Tx) DeleteAll() error {
	return tx.cfg.withSpan("DeleteAll", func() error {
		return tx.Tx.DeleteAll()
	})
}

// Descend calls the underlying Tx.Descend and traces the query.
func (tx *Tx) Descend(index string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("Descend", func() error {
		return tx.Tx.Descend(index, iterator)
	})
}

// DescendEqual calls the underlying Tx.DescendEqual and traces the query.
func (tx *Tx) DescendEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("DescendEqual", func() error {
		return tx.Tx.DescendEqual(index, pivot, iterator)
	})
}

// DescendGreaterThan calls the underlying Tx.DescendGreaterThan and traces the query.
func (tx *Tx) DescendGreaterThan(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("DescendGreaterThan", func() error {
		return tx.Tx.DescendGreaterThan(index, pivot, iterator)
	})
}

// DescendKeys calls the underlying Tx.DescendKeys and traces the query.
func (tx *Tx) DescendKeys(pattern string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("DescendKeys", func() error {
		return tx.Tx.DescendKeys(pattern, iterator)
	})
}

// DescendLessOrEqual calls the underlying Tx.DescendLessOrEqual and traces the query.
func (tx *Tx) DescendLessOrEqual(index, pivot string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("DescendLessOrEqual", func() error {
		return tx.Tx.DescendLessOrEqual(index, pivot, iterator)
	})
}

// DescendRange calls the underlying Tx.DescendRange and traces the query.
func (tx *Tx) DescendRange(index, lessOrEqual, greaterThan string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("DescendRange", func() error {
		return tx.Tx.DescendRange(index, lessOrEqual, greaterThan, iterator)
	})
}

// DropIndex calls the underlying Tx.DropIndex and traces the query.
func (tx *Tx) DropIndex(name string) error {
	return tx.cfg.withSpan("DropIndex", func() error {
		return tx.Tx.DropIndex(name)
	})
}

// Get calls the underlying Tx.Get and traces the query.
func (tx *Tx) Get(key string, ignoreExpired ...bool) (val string, err error) {
	err = tx.cfg.withSpan("Get", func() error {
		var iErr error
		val, iErr = tx.Tx.Get(key, ignoreExpired...)
		return iErr
	})
	return
}

// Indexes calls the underlying Tx.Indexes and traces the query.
func (tx *Tx) Indexes() (indexes []string, err error) {
	err = tx.cfg.withSpan("Indexes", func() error {
		var iErr error
		indexes, iErr = tx.Tx.Indexes()
		return iErr
	})
	return
}

// Intersects calls the underlying Tx.Intersects and traces the query.
func (tx *Tx) Intersects(index, bounds string, iterator func(key, value string) bool) error {
	return tx.cfg.withSpan("Intersects", func() error {
		return tx.Tx.Intersects(index, bounds, iterator)
	})
}

// Len calls the underlying Tx.Len and traces the query.
func (tx *Tx) Len() (n int, err error) {
	err = tx.cfg.withSpan("Len", func() error {
		var iErr error
		n, iErr = tx.Tx.Len()
		return iErr
	})
	return
}

// Nearby calls the underlying Tx.Nearby and traces the query.
func (tx *Tx) Nearby(index, bounds string, iterator func(key, value string, dist float64) bool) error {
	return tx.cfg.withSpan("Nearby", func() error {
		return tx.Tx.Nearby(index, bounds, iterator)
	})
}

// Set calls the underlying Tx.Set and traces the query.
func (tx *Tx) Set(key, value string, opts *buntdb.SetOptions) (previousValue string, replaced bool, err error) {
	err = tx.cfg.withSpan("Set", func() error {
		var iErr error
		previousValue, replaced, iErr = tx.Tx.Set(key, value, opts)
		return iErr
	})
	return
}

// TTL calls the underlying Tx.TTL and traces the query.
func (tx *Tx) TTL(key string) (duration time.Duration, err error) {
	err = tx.cfg.withSpan("TTL", func() error {
		var iErr error
		duration, iErr = tx.Tx.TTL(key)
		return iErr
	})
	return
}

// Commit calls the underlying Tx.Commit and traces the query.
func (tx *Tx) Commit() error {
	return tx.cfg.withSpan("Commit", func() error {
		return tx.Tx.Commit()
	})
}

// Rollback calls the underlying Tx.Rollback and traces the query.
func (tx *Tx) Rollback() {
	tx.cfg.withSpan("Rollback", func() error {
		tx.Tx.Rollback()
		return nil
	})
}
