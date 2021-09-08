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
	"database/sql/driver"
)

type mockRows struct{}

var _ driver.Rows = (*mockRows)(nil)

func newMockRows() *mockRows {
	return &mockRows{}
}

func (r *mockRows) Columns() []string         { return nil }
func (r *mockRows) Close() error              { return nil }
func (r *mockRows) Next([]driver.Value) error { return nil }
