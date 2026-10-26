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

package splunksql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeginTxLegacyOptions(t *testing.T) {
	for _, test := range []struct {
		name string
		opts driver.TxOptions
		err  string
	}{
		{name: "default"},
		{name: "read only", opts: driver.TxOptions{ReadOnly: true}, err: "driver does not support read-only transactions"},
		{name: "isolation", opts: driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelSerializable)}, err: "driver does not support non-default isolation level"},
		{name: "isolation and read only", opts: driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelSerializable), ReadOnly: true}, err: "driver does not support non-default isolation level"},
	} {
		t.Run(test.name, func(t *testing.T) {
			conn := newMockConn(nil)
			wrapped := newConn(struct{ driver.Conn }{conn}, newConfig())
			tx, err := wrapped.BeginTx(context.Background(), test.opts)
			if tx != nil {
				require.NoError(t, tx.Rollback())
			}
			if test.err == "" {
				require.NoError(t, err)
				require.NotNil(t, tx)
				require.Equal(t, 1, conn.beginN)
			} else {
				require.ErrorContains(t, err, test.err)
				require.Nil(t, tx)
				require.Zero(t, conn.beginN)
			}
		})
	}
}

type beginOptionsConn struct {
	*mockConn
	opts driver.TxOptions
}

func (c *beginOptionsConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	c.opts = opts
	return c.mockConn.BeginTx(ctx, opts)
}

func TestBeginTxPassesOptionsToContextDriver(t *testing.T) {
	conn := &beginOptionsConn{mockConn: newMockConn(nil)}
	opts := driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelSerializable), ReadOnly: true}
	tx, err := newConn(conn, newConfig()).BeginTx(context.Background(), opts)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.Equal(t, opts, conn.opts)
	require.Equal(t, 1, conn.beginTxN)
	require.Zero(t, conn.beginN)
}
