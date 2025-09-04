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

// Snapshot wraps a leveldb.Snapshot, tracing all operations performed.
type Snapshot struct {
	*leveldb.Snapshot
	cfg *config
}

// WrapSnapshot returns a traced *Snapshot that wraps a *leveldb.Snapshot.
func WrapSnapshot(snap *leveldb.Snapshot, opts ...Option) *Snapshot {
	return &Snapshot{
		Snapshot: snap,
		cfg:      newConfig(opts...),
	}
}

// WithContext returns a new Snapshot that will use ctx. If ctx contains any
// active spans of a trace, all traced operations of the returned DB will be
// represented as child spans of that active span.
func (snap *Snapshot) WithContext(ctx context.Context) *Snapshot {
	newcfg := *snap.cfg
	newcfg.ctx = ctx
	return &Snapshot{
		Snapshot: snap.Snapshot,
		cfg:      &newcfg,
	}
}

// Get gets the value for the given key. It returns ErrNotFound if
// the DB does not contains the key.
//
// The caller should not modify the contents of the returned slice, but
// it is safe to modify the contents of the argument after Get returns.
func (snap *Snapshot) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = snap.cfg.WithSpan(
		snap.cfg.ctx,
		"Get",
		func(context.Context) error {
			var e error
			value, e = snap.Snapshot.Get(key, ro)
			return e
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Get")),
	)
	return value, err
}

// Has returns true if the DB does contains the given key.
//
// It is safe to modify the contents of the argument after Get returns.
func (snap *Snapshot) Has(key []byte, ro *opt.ReadOptions) (ret bool, err error) {
	err = snap.cfg.WithSpan(
		snap.cfg.ctx,
		"Has",
		func(context.Context) error {
			var e error
			ret, e = snap.Snapshot.Has(key, ro)
			return e
		},
		trace.WithAttributes(semconv.DBOperationKey.String("Has")),
	)
	return ret, err
}

// NewIterator returns a traced iterator for the snapshot of the underlying
// DB. The returned iterator is not safe for concurrent use, but it is safe to
// use multiple iterators concurrently, with each in a dedicated goroutine. It
// is also safe to use an iterator concurrently with modifying its underlying
// DB. The resultant key/value pairs are guaranteed to be
// consistent.
//
// Slice allows slicing the iterator to only contains keys in the given
// range. A nil Range.Start is treated as a key before all keys in the
// DB. And a nil Range.Limit is treated as a key after all keys in
// the DB.
//
// WARNING: Any slice returned by interator (e.g. slice returned by calling
// Iterator.Key() or Iterator.Value() methods), its content should not be
// modified unless noted otherwise.
//
// The iterator must be released after use, by calling Release method.
// Releasing the snapshot doesn't mean releasing the iterator too, the
// iterator would be still valid until released.
//
// Also read Iterator documentation of the leveldb/iterator package.
func (snap *Snapshot) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	return WrapIterator(snap.Snapshot.NewIterator(slice, ro), optionFunc(func(cfg *config) {
		*cfg = *snap.cfg
	}))
}
