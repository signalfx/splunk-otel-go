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

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Transaction wraps a *leveldb.Transaction, tracing all operations performed.
type Transaction struct {
	*leveldb.Transaction
	cfg *config
}

// WrapTransaction returns a traced Transaction that wraps a
// *leveldb.Transaction.
func WrapTransaction(tr *leveldb.Transaction, opts ...Option) *Transaction {
	return &Transaction{
		Transaction: tr,
		cfg:         newConfig(opts...),
	}
}

// WithContext returns a new Transaction that will use ctx. If ctx contains
// any active spans of a trace, all traced operations of the returned DB will
// be represented as child spans of that active span.
func (tr *Transaction) WithContext(ctx context.Context) *Transaction {
	newcfg := *tr.cfg
	newcfg.ctx = ctx
	return &Transaction{
		Transaction: tr.Transaction,
		cfg:         &newcfg,
	}
}

// Commit commits the transaction. If error is not nil, then the transaction is
// not committed, it can then either be retried or discarded.
//
// Other methods should not be called after transaction has been committed.
func (tr *Transaction) Commit() error {
	return tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Commit",
		func(context.Context) error { return tr.Transaction.Commit() },
		trace.WithAttributes(semconv.DBOperationKey.String("Commit")),
	)
}

// Get gets the value for the given key. It returns ErrNotFound if the
// DB does not contains the key.
//
// The returned slice is its own copy, it is safe to modify the contents
// of the returned slice.
// It is safe to modify the contents of the argument after Get returns.
func (tr *Transaction) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Get",
		func(context.Context) error {
			var e error
			value, e = tr.Transaction.Get(key, ro)
			return e
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Get")),
	)
	return value, err
}

// Has returns true if the DB does contains the given key.
//
// It is safe to modify the contents of the argument after Has returns.
func (tr *Transaction) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Has",
		func(context.Context) error {
			var e error
			ret, e = tr.Transaction.Has(key, ro)
			return e
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Has")),
	)
	return ret, err
}

// NewIterator returns a traced iterator for the latest snapshot of the
// transaction. The returned iterator is not safe for concurrent use, but it
// is safe to use multiple iterators concurrently, with each in a dedicated
// goroutine. It is also safe to use an iterator concurrently while writes to
// the transaction. The resultant key/value pairs are guaranteed to be
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
func (tr *Transaction) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return WrapIterator(tr.Transaction.NewIterator(slice, ro), optionFunc(func(cfg *config) {
		*cfg = *tr.cfg
	}))
}

// Delete deletes the value for the given key.
// Please note that the transaction is not compacted until committed, so if you
// writes 10 same keys, then those 10 same keys are in the transaction.
//
// It is safe to modify the contents of the arguments after Delete returns.
func (tr *Transaction) Delete(key []byte, wo *opt.WriteOptions) error {
	return tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Delete",
		func(context.Context) error { return tr.Transaction.Delete(key, wo) },
		trace.WithAttributes(semconv.DBOperationKey.String("Delete")),
	)
}

// Discard discards the transaction.
//
// Other methods should not be called after transaction has been discarded.
func (tr *Transaction) Discard() {
	_ = tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Discard",
		func(context.Context) error {
			tr.Transaction.Discard()
			return nil
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Discard")),
	)
}

// Put sets the value for the given key. It overwrites any previous value
// for that key; a DB is not a multi-map.
// Please note that the transaction is not compacted until committed, so if you
// writes 10 same keys, then those 10 same keys are in the transaction.
//
// It is safe to modify the contents of the arguments after Put returns.
func (tr *Transaction) Put(key, value []byte, wo *opt.WriteOptions) error {
	return tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Put",
		func(context.Context) error { return tr.Transaction.Put(key, value, wo) },
		trace.WithAttributes(semconv.DBOperationKey.String("Put")),
	)
}

// Write applies the given batch to the transaction. The batch will be applied
// sequentially.
// Please note that the transaction is not compacted until committed, so if you
// writes 10 same keys, then those 10 same keys are in the transaction.
//
// It is safe to modify the contents of the arguments after Write returns.
func (tr *Transaction) Write(b *leveldb.Batch, wo *opt.WriteOptions) error {
	return tr.cfg.WithSpan(
		tr.cfg.ctx,
		"Write",
		func(context.Context) error { return tr.Transaction.Write(b, wo) },
		trace.WithAttributes(semconv.DBOperationKey.String("Write")),
	)
}
