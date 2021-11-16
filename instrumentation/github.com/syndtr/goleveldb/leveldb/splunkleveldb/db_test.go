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
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestOpenFile(t *testing.T) {
	p := path.Join(t.TempDir(), "test.db")
	db, err := OpenFile(p, nil)
	assert.NoError(t, err)
	assert.IsType(t, db, &DB{})
	require.NoError(t, db.Close())

	// Ensure errors are forwarded.
	_, err = OpenFile(p, &opt.Options{ErrorIfExist: true})
	assert.ErrorIs(t, err, os.ErrExist)
}

func TestOpen(t *testing.T) {
	memstore := storage.NewMemStorage()
	db, err := Open(memstore, nil)
	assert.NoError(t, err)
	assert.IsType(t, db, &DB{})
	require.NoError(t, db.Close())

	// Ensure errors are forwarded.
	_, err = Open(memstore, &opt.Options{ErrorIfExist: true})
	assert.ErrorIs(t, err, os.ErrExist)
}

func TestRecoverFile(t *testing.T) {
	// Create a DB to recover
	p := path.Join(t.TempDir(), "test.db")
	db, err := OpenFile(p, nil)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	db, err = RecoverFile(p, nil)
	assert.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestRecover(t *testing.T) {
	memstore := storage.NewMemStorage()
	db, err := Open(memstore, nil)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	db, err = Recover(memstore, nil)
	assert.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestDBOpenTransaction(t *testing.T) {
	db, err := Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	// Closing the DB will discard this transaction.
	transaction, err := db.OpenTransaction()
	assert.NoError(t, err)
	assert.IsType(t, &Transaction{}, transaction)
	assert.Equal(t, db.cfg, transaction.cfg)
}

func TestDBOpenTransactionForwardsError(t *testing.T) {
	db, err := Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	// Close right away so the OpenTransaction call will fail.
	require.NoError(t, db.Close())

	// Closing the DB will discard this transaction.
	_, err = db.OpenTransaction()
	assert.ErrorIs(t, err, leveldb.ErrClosed)
}

func TestDBNewIterator(t *testing.T) {
	db, err := Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	i := db.NewIterator(nil, nil)
	assert.IsType(t, &iter{}, i)
	assert.NotNil(t, i.(*iter).span)
	i.Release()
}

func TestDBGetSnapshot(t *testing.T) {
	db, err := Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	snap, err := db.GetSnapshot()
	assert.NoError(t, err)
	assert.IsType(t, &Snapshot{}, snap)
	assert.Equal(t, db.cfg, snap.cfg)
	snap.Release()
}

func TestDBGetSnapshotForwardsError(t *testing.T) {
	db, err := Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	// Close right away so the OpenTransaction call will fail.
	require.NoError(t, db.Close())

	_, err = db.GetSnapshot()
	assert.ErrorIs(t, err, leveldb.ErrClosed)
}
