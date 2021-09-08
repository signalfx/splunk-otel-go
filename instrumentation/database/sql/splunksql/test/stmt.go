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

type MockStmt struct {
	query string
}

var (
	_ driver.Stmt             = (*MockStmt)(nil)
	_ driver.StmtExecContext  = (*MockStmt)(nil)
	_ driver.StmtQueryContext = (*MockStmt)(nil)
)

func NewMockStmt(query string) *MockStmt {
	return &MockStmt{query: query}
}

func (s *MockStmt) Close() error {
	return nil
}

func (s *MockStmt) NumInput() int {
	return 0
}

func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, nil
}

func (s *MockStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	return nil, nil
}

func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return NewMockRows(), nil
}

func (s *MockStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	return NewMockRows(), nil
}
