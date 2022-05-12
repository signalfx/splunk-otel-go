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
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

var errTest = errors.New("testing error")

type mockConn struct {
	err error

	prepareN        int
	closeN          int
	beginN          int
	pingN           int
	execN           int
	execContextN    int
	queryN          int
	queryContextN   int
	prepareContextN int
	beginTxN        int
	resetSessionN   int
}

var (
	_ driver.Pinger             = (*mockConn)(nil)
	_ driver.Execer             = (*mockConn)(nil) // nolint: staticcheck // Ensure backwards support of deprecated interface.
	_ driver.ExecerContext      = (*mockConn)(nil)
	_ driver.Queryer            = (*mockConn)(nil) // nolint: staticcheck // Ensure backwards support of deprecated interface.
	_ driver.QueryerContext     = (*mockConn)(nil)
	_ driver.Conn               = (*mockConn)(nil)
	_ driver.ConnPrepareContext = (*mockConn)(nil)
	_ driver.ConnBeginTx        = (*mockConn)(nil)
	_ driver.SessionResetter    = (*mockConn)(nil)
)

func newMockConn(err error) *mockConn {
	return &mockConn{err: err}
}

func (c *mockConn) Prepare(string) (driver.Stmt, error) {
	c.prepareN++
	return newMockStmt(c.err), c.err
}

func (c *mockConn) Close() error {
	c.closeN++
	return c.err
}

func (c *mockConn) Begin() (driver.Tx, error) {
	c.beginN++
	return newMockTx(c.err), c.err
}

func (c *mockConn) Ping(context.Context) error {
	c.pingN++
	return c.err
}

func (c *mockConn) Exec(string, []driver.Value) (driver.Result, error) {
	c.execN++
	return nil, c.err
}

func (c *mockConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	c.execContextN++
	return nil, c.err
}

func (c *mockConn) Query(string, []driver.Value) (driver.Rows, error) {
	c.queryN++
	return nil, c.err
}

func (c *mockConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	c.queryContextN++
	return nil, c.err
}

func (c *mockConn) PrepareContext(context.Context, string) (driver.Stmt, error) {
	c.prepareContextN++
	return newMockStmt(c.err), c.err
}

func (c *mockConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.beginTxN++
	return newMockTx(c.err), c.err
}

func (c *mockConn) ResetSession(context.Context) error {
	c.resetSessionN++
	return c.err
}

type ConnSuite struct {
	suite.Suite

	MockConn *mockConn
	OTelConn *otelConn
}

func (s *ConnSuite) SetupTest() {
	s.MockConn = newMockConn(nil)
	s.OTelConn = newConn(s.MockConn, newTraceConfig())
}

func (s *ConnSuite) TestPrepareCallsWrapped() {
	_, err := s.OTelConn.Prepare("")
	s.NoError(err)
	s.Equal(1, s.MockConn.prepareN)
}

func (s *ConnSuite) TestPrepareReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.Prepare("")
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.prepareN)
}

func (s *ConnSuite) TestCloseCallsWrapped() {
	s.NoError(s.OTelConn.Close())
	s.Equal(1, s.MockConn.closeN)
}

func (s *ConnSuite) TestCloseReturnsWrappedError() {
	s.MockConn.err = errTest
	s.ErrorIs(s.OTelConn.Close(), errTest)
	s.Equal(1, s.MockConn.closeN)
}

func (s *ConnSuite) TestBeginCallsWrapped() {
	_, err := s.OTelConn.Begin()
	s.NoError(err)
	s.Equal(1, s.MockConn.beginN)
}

func (s *ConnSuite) TestBeginReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.Begin()
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.beginN)
}

func (s *ConnSuite) TestBeginTxCallsWrapped() {
	_, err := s.OTelConn.BeginTx(context.Background(), driver.TxOptions{})
	s.NoError(err)
	s.Equal(1, s.MockConn.beginTxN)
}

func (s *ConnSuite) TestBeginTxReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.BeginTx(context.Background(), driver.TxOptions{})
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.beginTxN)
}

func (s *ConnSuite) TestBeginTxFallsbackToExec() {
	s.OTelConn = newConn(struct {
		driver.Conn
	}{s.MockConn}, newTraceConfig())

	_, err := s.OTelConn.BeginTx(context.Background(), driver.TxOptions{})
	s.NoError(err)
	s.Equal(0, s.MockConn.beginTxN)
	s.Equal(1, s.MockConn.beginN)
}

func (s *ConnSuite) TestPingCallsWrapped() {
	s.NoError(s.OTelConn.Ping(context.Background()))
	s.Equal(1, s.MockConn.pingN)
}

func (s *ConnSuite) TestPingReturnsWrappedError() {
	s.MockConn.err = errTest
	s.ErrorIs(s.OTelConn.Ping(context.Background()), errTest)
	s.Equal(1, s.MockConn.pingN)
}

func (s *ConnSuite) TestPingReturnsErrSkipIfNotImplemented() {
	s.OTelConn = newConn(struct{ driver.Conn }{s.MockConn}, newTraceConfig())
	s.ErrorIs(s.OTelConn.Ping(context.Background()), driver.ErrSkip)
	s.Equal(0, s.MockConn.pingN)
}

func (s *ConnSuite) TestExecCallsWrapped() {
	_, err := s.OTelConn.Exec("", nil)
	s.NoError(err)
	s.Equal(1, s.MockConn.execN)
}

func (s *ConnSuite) TestExecReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.Exec("", nil)
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.execN)
}

func (s *ConnSuite) TestExecReturnsErrSkipIfNotImplemented() {
	s.OTelConn = newConn(struct{ driver.Conn }{s.MockConn}, newTraceConfig())
	_, err := s.OTelConn.Exec("", nil)
	s.ErrorIs(err, driver.ErrSkip)
	s.Equal(0, s.MockConn.execN)
}

func (s *ConnSuite) TestExecContextCallsWrapped() {
	_, err := s.OTelConn.ExecContext(context.Background(), "", nil)
	s.NoError(err)
	s.Equal(1, s.MockConn.execContextN)
}

type connExecer interface {
	driver.Conn
	driver.Execer // nolint: staticcheck // Ensure backwards support of deprecated interface.
}

func (s *ConnSuite) TestExecContextFallsbackToExec() {
	s.OTelConn = newConn(struct {
		connExecer
	}{s.MockConn}, newTraceConfig())

	_, err := s.OTelConn.ExecContext(context.Background(), "", nil)
	s.NoError(err)
	s.Equal(0, s.MockConn.execContextN)
	s.Equal(1, s.MockConn.execN)
}

func (s *ConnSuite) TestExecContextReturnsErrSkipIfNotImplemented() {
	s.OTelConn = newConn(struct{ driver.Conn }{s.MockConn}, newTraceConfig())
	_, err := s.OTelConn.ExecContext(context.Background(), "", nil)
	s.ErrorIs(err, driver.ErrSkip)
	s.Equal(0, s.MockConn.execContextN)
}

func (s *ConnSuite) TestExecContextReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.ExecContext(context.Background(), "", nil)
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.execContextN)
}

func (s *ConnSuite) TestQueryCallsWrapped() {
	_, err := s.OTelConn.Query("", nil) // nolint: gocritic // Test Query not Exec
	s.NoError(err)
	s.Equal(1, s.MockConn.queryN)
}

func (s *ConnSuite) TestQueryReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.Query("", nil) // nolint: gocritic // Test Query not Exec
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.queryN)
}

func (s *ConnSuite) TestQueryReturnsErrSkipIfNotImplemented() {
	s.OTelConn = newConn(struct{ driver.Conn }{s.MockConn}, newTraceConfig())
	_, err := s.OTelConn.Query("", nil) // nolint: gocritic // Test Query not Exec
	s.ErrorIs(err, driver.ErrSkip)
	s.Equal(0, s.MockConn.queryN)
}

func (s *ConnSuite) TestQueryContextCallsWrapped() {
	_, err := s.OTelConn.QueryContext(context.Background(), "", nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.NoError(err)
	s.Equal(1, s.MockConn.queryContextN)
}

func (s *ConnSuite) TestQueryContextReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.QueryContext(context.Background(), "", nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.queryContextN)
}

type connQuery interface {
	driver.Conn
	driver.Queryer // nolint: staticcheck // Ensure backwards support of deprecated interface.
}

func (s *ConnSuite) TestQueryContextFallsbackToExec() {
	s.OTelConn = newConn(struct {
		connQuery
	}{s.MockConn}, newTraceConfig())

	_, err := s.OTelConn.QueryContext(context.Background(), "", nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.NoError(err)
	s.Equal(0, s.MockConn.queryContextN)
	s.Equal(1, s.MockConn.queryN)
}

func (s *ConnSuite) TestQueryContextReturnsErrSkipIfNotImplemented() {
	s.OTelConn = newConn(struct{ driver.Conn }{s.MockConn}, newTraceConfig())
	_, err := s.OTelConn.QueryContext(context.Background(), "", nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.ErrorIs(err, driver.ErrSkip)
	s.Equal(0, s.MockConn.queryContextN)
}

func (s *ConnSuite) TestPrepareContextCallsWrapped() {
	r, err := s.OTelConn.PrepareContext(context.Background(), "")
	s.NoError(err)
	s.Equal(1, s.MockConn.prepareContextN)
	_ = r.Close()
}

func (s *ConnSuite) TestPrepareContextReturnsWrappedError() {
	s.MockConn.err = errTest
	_, err := s.OTelConn.PrepareContext(context.Background(), "")
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockConn.prepareContextN)
}

func (s *ConnSuite) TestPrepareContextFallsbackToExec() {
	s.OTelConn = newConn(struct {
		driver.Conn
	}{s.MockConn}, newTraceConfig())

	_, err := s.OTelConn.PrepareContext(context.Background(), "")
	s.NoError(err)
	s.Equal(0, s.MockConn.prepareContextN)
	s.Equal(1, s.MockConn.prepareN)
}

func (s *ConnSuite) TestResetSessionCallsWrapped() {
	s.NoError(s.OTelConn.ResetSession(context.Background()))
	s.Equal(1, s.MockConn.resetSessionN)
}

func (s *ConnSuite) TestResetSessionReturnsWrappedError() {
	s.MockConn.err = errTest
	s.ErrorIs(s.OTelConn.ResetSession(context.Background()), errTest)
	s.Equal(1, s.MockConn.resetSessionN)
}

func (s *ConnSuite) TestResetSessionReturnsErrSkipIfNotImplemented() {
	s.OTelConn = newConn(struct{ driver.Conn }{s.MockConn}, newTraceConfig())
	s.ErrorIs(s.OTelConn.ResetSession(context.Background()), driver.ErrSkip)
	s.Equal(0, s.MockConn.resetSessionN)
}

func TestConnSuite(t *testing.T) {
	s := new(ConnSuite)
	suite.Run(t, s)
}
