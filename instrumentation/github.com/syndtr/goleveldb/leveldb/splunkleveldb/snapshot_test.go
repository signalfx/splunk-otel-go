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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestSnapshotNewIterator(t *testing.T) {
	db, err := Open(storage.NewMemStorage(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	snap, err := db.GetSnapshot()
	require.NoError(t, err)
	t.Cleanup(snap.Release)

	i := snap.NewIterator(nil, nil)
	assert.IsType(t, &iter{}, i)
	assert.NotNil(t, i.(*iter).span)
	i.Release()
}
