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
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// DB wraps a *leveldb.DB, tracing all operations performed.
type DB struct {
	*leveldb.DB
	cfg *config
}

// Open opens or creates a traced DB for the given storage.
// The DB will be created if not exist, unless ErrorIfMissing is true.
// Also, if ErrorIfExist is true and the DB exist Open will returns
// os.ErrExist error.
//
// Open will return an error with type of ErrCorrupted if corruption
// detected in the DB. Use errors.IsCorrupted to test whether an error is
// due to corruption. Corrupted DB can be recovered with Recover function.
//
// The returned DB instance is safe for concurrent use.
// The DB must be closed after use, by calling Close method.
func Open(stor storage.Storage, o *opt.Options, opts ...Option) (*DB, error) {
	db, err := leveldb.Open(stor, o)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// OpenFile opens or creates a traced DB for the given path.
// The DB will be created if not exist, unless ErrorIfMissing is true.
// Also, if ErrorIfExist is true and the DB exist OpenFile will returns
// os.ErrExist error.
//
// OpenFile uses standard file-system backed storage implementation as
// described in the leveldb/storage package.
//
// OpenFile will return an error with type of ErrCorrupted if corruption
// detected in the DB. Use errors.IsCorrupted to test whether an error is
// due to corruption. Corrupted DB can be recovered with Recover function.
//
// The returned DB instance is safe for concurrent use.
// The DB must be closed after use, by calling Close method.
func OpenFile(path string, o *opt.Options, opts ...Option) (*DB, error) {
	db, err := leveldb.OpenFile(path, o)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// Recover recovers and opens a traced DB with missing or corrupted manifest
// files for the given storage. It will ignore any manifest files, valid or
// not. The DB must already exist or it will returns an error. Also, Recover
// will ignore ErrorIfMissing and ErrorIfExist options.
//
// The returned DB instance is safe for concurrent use.
// The DB must be closed after use, by calling Close method.
func Recover(stor storage.Storage, o *opt.Options, opts ...Option) (*DB, error) {
	db, err := leveldb.Recover(stor, o)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// RecoverFile recovers and opens a traced DB with missing or corrupted
// manifest files for the given path. It will ignore any manifest files, valid
// or not. The DB must already exist or it will returns an error. Also,
// Recover will ignore ErrorIfMissing and ErrorIfExist options.
//
// RecoverFile uses standard file-system backed storage implementation as described
// in the leveldb/storage package.
//
// The returned DB instance is safe for concurrent use.
// The DB must be closed after use, by calling Close method.
func RecoverFile(path string, o *opt.Options, opts ...Option) (*DB, error) {
	db, err := leveldb.RecoverFile(path, o)
	if err != nil {
		return nil, err
	}
	return WrapDB(db, opts...), nil
}

// WrapDB returns a traced DB that wraps a *leveldb.DB.
func WrapDB(db *leveldb.DB, opts ...Option) *DB {
	return &DB{
		DB:  db,
		cfg: newConfig(opts...),
	}
}

// WithContext returns a new DB that will use ctx. If ctx contains any active
// spans of a trace, all traced operations of the returned DB will be
// represented as child spans of that active span.
func (db *DB) WithContext(ctx context.Context) *DB {
	newcfg := *db.cfg
	newcfg.ctx = ctx
	return &DB{
		DB:  db.DB,
		cfg: &newcfg,
	}
}

// CompactRange compacts the underlying traced DB for the given key range.
// In particular, deleted and overwritten versions are discarded,
// and the data is rearranged to reduce the cost of operations
// needed to access the data. This operation should typically only
// be invoked by users who understand the underlying implementation.
//
// A nil Range.Start is treated as a key before all keys in the DB.
// And a nil Range.Limit is treated as a key after all keys in the DB.
// Therefore if both is nil then it will compact entire DB.
func (db *DB) CompactRange(r util.Range) error {
	return db.cfg.WithSpan(
		db.cfg.ctx,
		"CompactRange",
		func(context.Context) error { return db.DB.CompactRange(r) },
		trace.WithAttributes(semconv.DBOperationKey.String("CompactRange")),
	)
}

// Delete deletes the value for the given key. Delete will not returns error if
// key doesn't exist. Write merge also applies for Delete, see Write.
//
// It is safe to modify the contents of the arguments after Delete returns but
// not before.
func (db *DB) Delete(key []byte, wo *opt.WriteOptions) error {
	return db.cfg.WithSpan(
		db.cfg.ctx,
		"Delete",
		func(context.Context) error { return db.DB.Delete(key, wo) },
		trace.WithAttributes(semconv.DBOperationKey.String("Delete")),
	)
}

// Get gets the value for the given key. It returns ErrNotFound if the
// DB does not contains the key.
//
// The returned slice is its own copy, it is safe to modify the contents
// of the returned slice.
// It is safe to modify the contents of the argument after Get returns.
func (db *DB) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = db.cfg.WithSpan(
		db.cfg.ctx,
		"Get",
		func(context.Context) error {
			var e error
			value, e = db.DB.Get(key, ro)
			return e
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Get")),
	)
	return value, err
}

// GetSnapshot returns a latest snapshot of the underlying DB. A snapshot
// is a frozen snapshot of a DB state at a particular point in time. The
// content of snapshot are guaranteed to be consistent.
//
// The snapshot must be released after use, by calling Release method.
func (db *DB) GetSnapshot() (*Snapshot, error) {
	snap, err := db.DB.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return WrapSnapshot(snap, optionFunc(func(cfg *config) {
		*cfg = *db.cfg
	})), nil
}

// Has returns true if the DB does contains the given key.
//
// It is safe to modify the contents of the argument after Has returns.
func (db *DB) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = db.cfg.WithSpan(
		db.cfg.ctx,
		"Has",
		func(context.Context) error {
			var e error
			ret, e = db.DB.Has(key, ro)
			return e
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Has")),
	)
	return ret, err
}

// NewIterator returns a traced iterator for the latest snapshot of the
// underlying DB.
// The returned iterator is not safe for concurrent use, but it is safe to use
// multiple iterators concurrently, with each in a dedicated goroutine.
// It is also safe to use an iterator concurrently with modifying its
// underlying DB. The resultant key/value pairs are guaranteed to be
// consistent.
//
// Slice allows slicing the iterator to only contains keys in the given
// range. A nil Range.Start is treated as a key before all keys in the
// DB. And a nil Range.Limit is treated as a key after all keys in
// the DB.
//
// WARNING: Any slice returned by interator (e.g. slice returned by calling
// Iterator.Key() or Iterator.Key() methods), its content should not be modified
// unless noted otherwise.
//
// The iterator must be released after use, by calling Release method.
//
// Also read Iterator documentation of the leveldb/iterator package.
func (db *DB) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return WrapIterator(db.DB.NewIterator(slice, ro), optionFunc(func(cfg *config) {
		*cfg = *db.cfg
	}))
}

// OpenTransaction opens an atomic DB transaction. Only one transaction can be
// opened at a time. Subsequent call to Write and OpenTransaction will be blocked
// until in-flight transaction is committed or discarded.
// The returned transaction handle is safe for concurrent use.
//
// Transaction is expensive and can overwhelm compaction, especially if
// transaction size is small. Use with caution.
//
// The transaction must be closed once done, either by committing or discarding
// the transaction.
// Closing the DB will discard open transaction.
func (db *DB) OpenTransaction() (*Transaction, error) {
	tr, err := db.DB.OpenTransaction()
	if err != nil {
		return nil, err
	}
	return WrapTransaction(tr, optionFunc(func(cfg *config) {
		*cfg = *db.cfg
	})), nil
}

// Put sets the value for the given key. It overwrites any previous value
// for that key; a DB is not a multi-map. Write merge also applies for Put, see
// Write.
//
// It is safe to modify the contents of the arguments after Put returns but not
// before.
func (db *DB) Put(key, value []byte, wo *opt.WriteOptions) error {
	return db.cfg.WithSpan(
		db.cfg.ctx,
		"Put",
		func(context.Context) error { return db.DB.Put(key, value, wo) },
		trace.WithAttributes(semconv.DBOperationKey.String("Put")),
	)
}

// Write apply the given batch to the DB. The batch records will be applied
// sequentially. Write might be used concurrently, when used concurrently and
// batch is small enough, write will try to merge the batches. Set NoWriteMerge
// option to true to disable write merge.
//
// It is safe to modify the contents of the arguments after Write returns but
// not before. Write will not modify content of the batch.
func (db *DB) Write(batch *leveldb.Batch, wo *opt.WriteOptions) error {
	return db.cfg.WithSpan(
		db.cfg.ctx,
		"Write",
		func(context.Context) error { return db.DB.Write(batch, wo) },
		trace.WithAttributes(semconv.DBOperationKey.String("Write")),
	)
}
