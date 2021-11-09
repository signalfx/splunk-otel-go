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
	return tr.cfg.withSpan("Commit", func(context.Context) error {
		return tr.Transaction.Commit()
	})
}

// Get gets the value for the given key. It returns ErrNotFound if the
// DB does not contains the key.
//
// The returned slice is its own copy, it is safe to modify the contents
// of the returned slice.
// It is safe to modify the contents of the argument after Get returns.
func (tr *Transaction) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = tr.cfg.withSpan("Get", func(context.Context) error {
		value, err = tr.Transaction.Get(key, ro)
		return err
	})
	return
}

// Has returns true if the DB does contains the given key.
//
// It is safe to modify the contents of the argument after Has returns.
func (tr *Transaction) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = tr.cfg.withSpan("Has", func(context.Context) error {
		ret, err = tr.Transaction.Has(key, ro)
		return err
	})
	return
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
