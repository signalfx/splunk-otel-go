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
	"database/sql/driver"
)

type mockTx struct {
	err error

	commitN   int
	rollbackN int
}

var _ driver.Tx = (*mockTx)(nil)

func newMockTx(err error) *mockTx {
	return &mockTx{err: err}
}

func (t *mockTx) Commit() error {
	t.commitN++
	return t.err
}

func (t *mockTx) Rollback() error {
	t.rollbackN++
	return t.err
}
