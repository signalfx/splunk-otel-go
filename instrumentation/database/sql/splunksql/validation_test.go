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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type poolingState struct {
	valid       atomic.Bool
	opens       atomic.Int32
	closes      atomic.Int32
	validations atomic.Int32
	resets      atomic.Int32
	rollbacks   atomic.Int32
}

type poolingConn struct{ state *poolingState }

func (c *poolingConn) Prepare(string) (driver.Stmt, error) { return newMockStmt(nil), nil }
func (c *poolingConn) Close() error {
	c.state.closes.Add(1)
	return nil
}
func (c *poolingConn) Begin() (driver.Tx, error) { return &poolingTx{state: c.state}, nil }
func (*poolingConn) Ping(context.Context) error  { return nil }

type poolingValidator struct{ state *poolingState }

func (v poolingValidator) IsValid() bool {
	v.state.validations.Add(1)
	return v.state.valid.Load()
}

type poolingResetter struct{ state *poolingState }

func (r poolingResetter) ResetSession(context.Context) error {
	r.state.resets.Add(1)
	return nil
}

type poolingTx struct{ state *poolingState }

func (*poolingTx) Commit() error { return nil }
func (tx *poolingTx) Rollback() error {
	tx.state.rollbacks.Add(1)
	return nil
}

type poolingCapabilities struct {
	validator bool
	resetter  bool
}

var poolingCapabilityCases = map[string]poolingCapabilities{
	"neither":        {},
	"validator only": {validator: true},
	"resetter only":  {resetter: true},
	"both":           {validator: true, resetter: true},
}

type poolingDriver struct {
	state *poolingState
	caps  poolingCapabilities
}

func (d *poolingDriver) Open(string) (driver.Conn, error) {
	d.state.opens.Add(1)
	c := &poolingConn{state: d.state}
	switch {
	case d.caps.validator && d.caps.resetter:
		return struct {
			*poolingConn
			driver.Validator
			driver.SessionResetter
		}{c, poolingValidator{d.state}, poolingResetter{d.state}}, nil
	case d.caps.validator:
		return struct {
			*poolingConn
			driver.Validator
		}{c, poolingValidator{d.state}}, nil
	case d.caps.resetter:
		return struct {
			*poolingConn
			driver.SessionResetter
		}{c, poolingResetter{d.state}}, nil
	default:
		return c, nil
	}
}

type poolingConnector struct{ driver driver.Driver }

const (
	poolingDriverPath    = "driver"
	poolingConnectorPath = "connector"
)

func (c poolingConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open("")
}
func (c poolingConnector) Driver() driver.Driver { return c.driver }

func newPoolingConnector(path string, caps poolingCapabilities) (driver.Connector, *poolingState) {
	state := new(poolingState)
	state.valid.Store(true)
	d := &poolingDriver{state: state, caps: caps}
	if path == poolingDriverPath {
		return poolingConnector{driver: newDriver(d, newConfig())}, state
	}
	return newConnector(poolingConnector{driver: d}, &otelDriver{driver: d, config: newConfig()}), state
}

func TestConnectionPoolingCapabilities(t *testing.T) {
	for _, path := range []string{poolingDriverPath, poolingConnectorPath} {
		for name, caps := range poolingCapabilityCases {
			t.Run(path+"/"+name, func(t *testing.T) {
				connector, state := newPoolingConnector(path, caps)
				conn, err := connector.Connect(context.Background())
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, conn.Close()) })
				validator, validates := conn.(driver.Validator)
				_, resets := conn.(driver.SessionResetter)
				assert.Equal(t, caps.validator, validates)
				assert.Equal(t, caps.resetter, resets)
				assert.Implements(t, (*driver.ConnBeginTx)(nil), conn)
				assert.Implements(t, (*driver.ConnPrepareContext)(nil), conn)
				assert.Implements(t, (*driver.ExecerContext)(nil), conn)
				assert.Implements(t, (*driver.QueryerContext)(nil), conn)
				assert.Implements(t, (*driver.Pinger)(nil), conn)
				assert.Implements(t, (*driver.Execer)(nil), conn)  //nolint:staticcheck // Ensure backwards support of deprecated interface.
				assert.Implements(t, (*driver.Queryer)(nil), conn) //nolint:staticcheck // Ensure backwards support of deprecated interface.
				if !validates {
					return
				}
				assert.True(t, validator.IsValid())
				state.valid.Store(false)
				assert.False(t, validator.IsValid())
				assert.EqualValues(t, 2, state.validations.Load())
			})
		}
	}
}

func TestInvalidConnectionDiscarded(t *testing.T) {
	for _, path := range []string{poolingDriverPath, poolingConnectorPath} {
		t.Run(path, func(t *testing.T) {
			connector, state := newPoolingConnector(path, poolingCapabilities{validator: true})
			db := sql.OpenDB(connector)
			t.Cleanup(func() { assert.NoError(t, db.Close()) })
			db.SetMaxOpenConns(1)
			conn, err := db.Conn(context.Background())
			require.NoError(t, err)
			state.valid.Store(false)
			require.NoError(t, conn.Close())
			assert.EqualValues(t, 1, state.validations.Load())
			assert.EqualValues(t, 1, state.closes.Load())
			assert.Zero(t, db.Stats().Idle)
			state.valid.Store(true)
			conn, err = db.Conn(context.Background())
			require.NoError(t, err)
			assert.EqualValues(t, 2, state.opens.Load())
			require.NoError(t, conn.Close())
			assert.Equal(t, 1, db.Stats().Idle)
		})
	}
}

func TestCanceledTransactionPooling(t *testing.T) {
	for _, path := range []string{poolingDriverPath, poolingConnectorPath} {
		for name, caps := range poolingCapabilityCases {
			t.Run(path+"/"+name, func(t *testing.T) {
				connector, state := newPoolingConnector(path, caps)
				db := sql.OpenDB(connector)
				t.Cleanup(func() { assert.NoError(t, db.Close()) })
				db.SetMaxOpenConns(1)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				tx, err := db.BeginTx(ctx, nil)
				require.NoError(t, err)
				cancel()
				require.Eventually(t, func() bool { return db.Stats().InUse == 0 }, 5*time.Second, time.Millisecond)
				assert.EqualValues(t, 1, state.rollbacks.Load())
				assert.ErrorIs(t, tx.Rollback(), sql.ErrTxDone)
				if caps.validator && caps.resetter {
					assert.Zero(t, state.closes.Load())
					assert.Equal(t, 1, db.Stats().Idle)
				} else {
					assert.EqualValues(t, 1, state.closes.Load())
					assert.Zero(t, db.Stats().Idle)
				}
				require.NoError(t, db.PingContext(context.Background()))
				if caps.validator && caps.resetter {
					assert.EqualValues(t, 1, state.opens.Load())
					assert.EqualValues(t, 1, state.resets.Load())
				} else {
					assert.EqualValues(t, 2, state.opens.Load())
				}
			})
		}
	}
}
