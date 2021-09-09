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

type mockRows struct {
	err error

	columnsN int
	closeN   int
	nextN    int
}

var _ driver.Rows = (*mockRows)(nil)

func newMockRows(err error) *mockRows {
	return &mockRows{err: err}
}

func (r *mockRows) Columns() []string {
	r.columnsN++
	return nil
}

func (r *mockRows) Close() error {
	r.closeN++
	return r.err
}

func (r *mockRows) Next([]driver.Value) error {
	r.nextN++
	return r.err
}

type RowsSuite struct {
	suite.Suite

	MockRows *mockRows
	OTelRows *otelRows
}

func (s *RowsSuite) SetupTest() {
	s.MockRows = newMockRows(nil)
	s.OTelRows = newRows(context.Background(), s.MockRows, newTraceConfig())
}

func (s *RowsSuite) TestColumnsCallsWrapped() {
	_ = s.OTelRows.Columns()
	s.Equal(1, s.MockRows.columnsN)
}

func (s *RowsSuite) TestCloseCallsWrapped() {
	s.NoError(s.OTelRows.Close())
	s.Equal(1, s.MockRows.closeN)
}

func (s *RowsSuite) TestCloseReturnsWrappedError() {
	s.MockRows.err = errTest
	s.ErrorIs(s.OTelRows.Close(), errTest)
	s.Equal(1, s.MockRows.closeN)
}

func (s *RowsSuite) TestNextCallsWrapped() {
	s.NoError(s.OTelRows.Next(nil))
	s.Equal(1, s.MockRows.nextN)
}

func (s *RowsSuite) TestNextReturnsWrappedError() {
	s.MockRows.err = errTest
	s.ErrorIs(s.OTelRows.Next(nil), errTest)
	s.Equal(1, s.MockRows.nextN)
}

func TestRowsSuite(t *testing.T) {
	s := new(RowsSuite)
	suite.Run(t, s)
}
