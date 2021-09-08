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

package test

import (
	"context"
	"database/sql/driver"
)

type mockStmt struct {
	query string
}

var (
	_ driver.Stmt             = (*mockStmt)(nil)
	_ driver.StmtExecContext  = (*mockStmt)(nil)
	_ driver.StmtQueryContext = (*mockStmt)(nil)
)

func newMockStmt(query string) *mockStmt {
	return &mockStmt{query: query}
}

func (s *mockStmt) Close() error {
	return nil
}

func (s *mockStmt) NumInput() int {
	return 0
}

func (s *mockStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *mockStmt) ExecContext(context.Context, []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (s *mockStmt) Query([]driver.Value) (driver.Rows, error) {
	return newMockRows(), nil
}

func (s *mockStmt) QueryContext(context.Context, []driver.NamedValue) (driver.Rows, error) {
	return newMockRows(), nil
}
