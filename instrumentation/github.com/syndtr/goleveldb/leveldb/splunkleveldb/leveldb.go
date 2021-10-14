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

// Package splunkleveldb provides OpenTelemetry instrumentation for the
// github.com/syndtr/goleveldb/leveldb package.
package splunkleveldb

import (
	"context"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// A DB wraps a leveldb.DB and traces all queries.
type DB struct {
	*leveldb.DB
	cfg *config
}

// Open calls leveldb.Open and wraps the resulting DB.
func Open(stor storage.Storage, o *opt.Options, opts ...Option) (*DB, error) {
	db, err := leveldb.Open(stor, o)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// OpenFile calls leveldb.OpenFile and wraps the resulting DB.
func OpenFile(path string, o *opt.Options, opts ...Option) (*DB, error) {
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// WrapDB wraps a leveldb.DB so that queries are traced.
func WrapDB(db *leveldb.DB, opts ...Option) *DB {
	return &DB{
		DB:  db,
		cfg: newConfig(opts...),
	}
}

// WithContext returns a new DB with the context set to ctx.
func (db *DB) WithContext(ctx context.Context) *DB {
	newcfg := *db.cfg
	newcfg.ctx = ctx
	return &DB{
		DB:  db.DB,
		cfg: &newcfg,
	}
}

// CompactRange calls DB.CompactRange and traces the result.
func (db *DB) CompactRange(r util.Range) error {
	return db.cfg.withSpan("CompactRange", func(context.Context) error {
		return db.DB.CompactRange(r)
	})
}

// Delete calls DB.Delete and traces the result.
func (db *DB) Delete(key []byte, wo *opt.WriteOptions) error {
	return db.cfg.withSpan("Delete", func(context.Context) error {
		return db.DB.Delete(key, wo)
	})
}

// Get calls DB.Get and traces the result.
func (db *DB) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = db.cfg.withSpan("Get", func(context.Context) error {
		value, err = db.DB.Get(key, ro)
		return err
	})
	return
}

// GetSnapshot calls DB.GetSnapshot and returns a wrapped Snapshot.
func (db *DB) GetSnapshot() (*Snapshot, error) {
	snap, err := db.DB.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return WrapSnapshot(snap, optionFunc(func(cfg *config) {
		*cfg = *db.cfg
	})), nil
}

// Has calls DB.Has and traces the result.
func (db *DB) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = db.cfg.withSpan("Has", func(context.Context) error {
		ret, err = db.DB.Has(key, ro)
		return err
	})
	return
}

// NewIterator calls DB.NewIterator and returns a wrapped Iterator.
func (db *DB) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return WrapIterator(db.DB.NewIterator(slice, ro), optionFunc(func(cfg *config) {
		*cfg = *db.cfg
	}))
}

// OpenTransaction calls DB.OpenTransaction and returns a wrapped Transaction.
func (db *DB) OpenTransaction() (*Transaction, error) {
	tr, err := db.DB.OpenTransaction()
	if err != nil {
		return nil, err
	}
	return WrapTransaction(tr, optionFunc(func(cfg *config) {
		*cfg = *db.cfg
	})), nil
}

// Put calls DB.Put and traces the result.
func (db *DB) Put(key, value []byte, wo *opt.WriteOptions) error {
	return db.cfg.withSpan("Put", func(context.Context) error {
		return db.DB.Put(key, value, wo)
	})
}

// Write calls DB.Write and traces the result.
func (db *DB) Write(batch *leveldb.Batch, wo *opt.WriteOptions) error {
	return db.cfg.withSpan("Write", func(context.Context) error {
		return db.DB.Write(batch, wo)
	})
}

// A Snapshot wraps a leveldb.Snapshot and traces all queries.
type Snapshot struct {
	*leveldb.Snapshot
	cfg *config
}

// WrapSnapshot wraps a leveldb.Snapshot so that queries are traced.
func WrapSnapshot(snap *leveldb.Snapshot, opts ...Option) *Snapshot {
	return &Snapshot{
		Snapshot: snap,
		cfg:      newConfig(opts...),
	}
}

// WithContext returns a new Snapshot with the context set to ctx.
func (snap *Snapshot) WithContext(ctx context.Context) *Snapshot {
	newcfg := *snap.cfg
	newcfg.ctx = ctx
	return &Snapshot{
		Snapshot: snap.Snapshot,
		cfg:      &newcfg,
	}
}

// Get calls Snapshot.Get and traces the result.
func (snap *Snapshot) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = snap.cfg.withSpan("Get", func(context.Context) error {
		value, err = snap.Snapshot.Get(key, ro)
		return err
	})
	return
}

// Has calls Snapshot.Has and traces the result.
func (snap *Snapshot) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = snap.cfg.withSpan("Has", func(context.Context) error {
		ret, err = snap.Snapshot.Has(key, ro)
		return err
	})
	return
}

// NewIterator calls Snapshot.NewIterator and returns a wrapped Iterator.
func (snap *Snapshot) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return WrapIterator(snap.Snapshot.NewIterator(slice, ro), optionFunc(func(cfg *config) {
		*cfg = *snap.cfg
	}))
}

// A Transaction wraps a leveldb.Transaction and traces all queries.
type Transaction struct {
	*leveldb.Transaction
	cfg *config
}

// WrapTransaction wraps a leveldb.Transaction so that queries are traced.
func WrapTransaction(tr *leveldb.Transaction, opts ...Option) *Transaction {
	return &Transaction{
		Transaction: tr,
		cfg:         newConfig(opts...),
	}
}

// WithContext returns a new Transaction with the context set to ctx.
func (tr *Transaction) WithContext(ctx context.Context) *Transaction {
	newcfg := *tr.cfg
	newcfg.ctx = ctx
	return &Transaction{
		Transaction: tr.Transaction,
		cfg:         &newcfg,
	}
}

// Commit calls Transaction.Commit and traces the result.
func (tr *Transaction) Commit() error {
	return tr.cfg.withSpan("Commit", func(context.Context) error {
		return tr.Transaction.Commit()
	})
}

// Get calls Transaction.Get and traces the result.
func (tr *Transaction) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = tr.cfg.withSpan("Get", func(context.Context) error {
		value, err = tr.Transaction.Get(key, ro)
		return err
	})
	return
}

// Has calls Transaction.Has and traces the result.
func (tr *Transaction) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = tr.cfg.withSpan("Has", func(context.Context) error {
		ret, err = tr.Transaction.Has(key, ro)
		return err
	})
	return
}

// NewIterator calls Transaction.NewIterator and returns a wrapped Iterator.
func (tr *Transaction) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return WrapIterator(tr.Transaction.NewIterator(slice, ro), optionFunc(func(cfg *config) {
		*cfg = *tr.cfg
	}))
}

// An Iterator wraps a leveldb.Iterator and traces until Release is called.
type Iterator struct {
	iterator.Iterator
	span trace.Span
}

// WrapIterator wraps a leveldb.Iterator so that queries are traced.
func WrapIterator(it iterator.Iterator, opts ...Option) *Iterator {
	c := newConfig(opts...)
	_, span := c.resolveTracer().Start(c.ctx, "Iterator")
	return &Iterator{
		Iterator: it,
		span:     span,
	}
}

// Release calls Iterator.Release and traces the result.
func (it *Iterator) Release() {
	if err := it.Error(); err != nil {
		it.span.RecordError(err)
		it.span.SetStatus(codes.Error, err.Error())
	}
	it.Iterator.Release()
	it.span.End()
}
