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
	return newMockRows(), s.err
}

func (s *mockStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	s.queryContextN++
	return newMockRows(), s.err
}
