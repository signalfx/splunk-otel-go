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
	"testing"

	"github.com/stretchr/testify/suite"
)

type mockStmt struct {
	err error

	closeN        int
	numInputN     int
	execN         int
	execContextN  int
	queryN        int
	queryContextN int
}

var (
	_ driver.Stmt             = (*mockStmt)(nil)
	_ driver.StmtExecContext  = (*mockStmt)(nil)
	_ driver.StmtQueryContext = (*mockStmt)(nil)
)

func newMockStmt(err error) *mockStmt {
	return &mockStmt{err: err}
}

func (s *mockStmt) Close() error {
	s.closeN++
	return s.err
}

func (s *mockStmt) NumInput() int {
	s.numInputN++
	return 0
}

func (s *mockStmt) Exec([]driver.Value) (driver.Result, error) {
	s.execN++
	return nil, s.err
}

func (s *mockStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	s.execContextN++
	return nil, s.err
}

func (s *mockStmt) Query([]driver.Value) (driver.Rows, error) {
	s.queryN++
	return nil, s.err
}

func (s *mockStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	s.queryContextN++
	return nil, s.err
}

type StmtSuite struct {
	suite.Suite

	MockStmt *mockStmt
	OTelStmt *otelStmt
}

func (s *StmtSuite) SetupTest() {
	s.MockStmt = newMockStmt(nil)
	s.OTelStmt = newStmt(s.MockStmt, newTraceConfig(), "")
}

func (s *StmtSuite) TestCloseCallsWrapped() {
	s.NoError(s.OTelStmt.Close())
	s.Equal(1, s.MockStmt.closeN)
}

func (s *StmtSuite) TestCloseReturnsWrappedError() {
	s.MockStmt.err = errTest
	s.ErrorIs(s.OTelStmt.Close(), errTest)
	s.Equal(1, s.MockStmt.closeN)
}

func (s *StmtSuite) TestNumInputCallsWrapped() {
	_ = s.OTelStmt.NumInput()
	s.Equal(1, s.MockStmt.numInputN)
}

func (s *StmtSuite) TestExecCallsWrapped() {
	_, err := s.OTelStmt.Exec(nil)
	s.NoError(err)
	s.Equal(1, s.MockStmt.execN)
}

func (s *StmtSuite) TestExecReturnsWrappedError() {
	s.MockStmt.err = errTest
	_, err := s.OTelStmt.Exec(nil)
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockStmt.execN)
}

func (s *StmtSuite) TestExecContextCallsWrapped() {
	_, err := s.OTelStmt.ExecContext(context.Background(), nil)
	s.NoError(err)
	s.Equal(1, s.MockStmt.execContextN)
}

func (s *StmtSuite) TestExecContextFallsbackToExec() {
	s.OTelStmt = newStmt(struct{ driver.Stmt }{s.MockStmt}, newTraceConfig(), "")

	_, err := s.OTelStmt.ExecContext(context.Background(), nil)
	s.NoError(err)
	s.Equal(0, s.MockStmt.execContextN)
	s.Equal(1, s.MockStmt.execN)
}

func (s *StmtSuite) TestExecContextReturnsWrappedError() {
	s.MockStmt.err = errTest
	_, err := s.OTelStmt.ExecContext(context.Background(), nil)
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockStmt.execContextN)
}

func (s *StmtSuite) TestQueryCallsWrapped() {
	_, err := s.OTelStmt.Query(nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.NoError(err)
	s.Equal(1, s.MockStmt.queryN)
}

func (s *StmtSuite) TestQueryReturnsWrappedError() {
	s.MockStmt.err = errTest
	_, err := s.OTelStmt.Query(nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockStmt.queryN)
}

func (s *StmtSuite) TestQueryContextCallsWrapped() {
	_, err := s.OTelStmt.QueryContext(context.Background(), nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.NoError(err)
	s.Equal(1, s.MockStmt.queryContextN)
}

func (s *StmtSuite) TestQueryContextFallsbackToQuery() {
	s.OTelStmt = newStmt(struct{ driver.Stmt }{s.MockStmt}, newTraceConfig(), "")

	_, err := s.OTelStmt.QueryContext(context.Background(), nil) // nolint: gocritic // there is no connection leak for this test structure.
	s.NoError(err)
	s.Equal(0, s.MockStmt.queryContextN)
	s.Equal(1, s.MockStmt.queryN)
}

func (s *StmtSuite) TestQueryContextReturnsWrappedError() {
	s.MockStmt.err = errTest
	_, err := s.OTelStmt.QueryContext(context.Background(), nil) // nolint: gocritic // test Query not Exec
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockStmt.queryContextN)
}

func TestStmtSuite(t *testing.T) {
	s := new(StmtSuite)
	suite.Run(t, s)
}
