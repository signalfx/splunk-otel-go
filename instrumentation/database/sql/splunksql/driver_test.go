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

type mockDriver struct {
	err error

	openN          int
	openConnectorN int
}

var (
	_ driver.Driver        = (*mockDriver)(nil)
	_ driver.DriverContext = (*mockDriver)(nil)
)

func newMockDriver(err error) *mockDriver {
	return &mockDriver{err: err}
}

func (d *mockDriver) Open(string) (driver.Conn, error) {
	d.openN++
	return newMockConn(d.err), d.err
}

func (d *mockDriver) OpenConnector(string) (driver.Connector, error) {
	d.openConnectorN++
	return newMockConnector(d), d.err
}

type DriverSuite struct {
	suite.Suite

	MockDriver *mockDriver
	OTelDriver *otelDriver
}

func (s *DriverSuite) SetupTest() {
	s.MockDriver = newMockDriver(nil)
	s.OTelDriver = newDriver(s.MockDriver, newTraceConfig()).(*otelDriver)
}

func (s *DriverSuite) TestNewDriverImplementation() {
	fullImpl := newDriver(s.MockDriver, newTraceConfig())
	s.Implements((*driver.Driver)(nil), fullImpl)
	s.Implements((*driver.DriverContext)(nil), fullImpl)

	partImpl := newDriver(struct{ driver.Driver }{s.MockDriver}, newTraceConfig())
	s.Implements((*driver.Driver)(nil), partImpl)
	if _, ok := partImpl.(driver.DriverContext); ok {
		s.Fail("wrapped driver does not implement DriverContext")
	}
}

func (s *DriverSuite) TestOpenCallsWrapped() {
	c, err := s.OTelDriver.Open("")
	s.NoError(err)
	s.Equal(1, s.MockDriver.openN)
	if _, ok := c.(*otelConn); !ok {
		s.FailNow("driver did not return instrumented conn")
	}
}

func (s *DriverSuite) TestOpenReturnsWrappedError() {
	s.MockDriver.err = errTest
	_, err := s.OTelDriver.Open("")
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockDriver.openN)
}

func (s *DriverSuite) TestOpenConnectorCallsWrapped() {
	c, err := s.OTelDriver.OpenConnector("")
	s.NoError(err)
	s.Equal(1, s.MockDriver.openConnectorN)

	oc, ok := c.(*otelConnector)
	if !ok {
		s.FailNow("driver did not return instrumented connector")
	}
	mc, ok := oc.Connector.(*mockConnector)
	if !ok {
		s.FailNow("mock driver did not return mock connector")
	}
	_, err = oc.Connect(context.Background())
	s.NoError(err)
	s.Equal(1, mc.connectN)

	s.Same(s.OTelDriver, oc.Driver())
}

func (s *DriverSuite) TestOpenConnectorReturnsWrappedError() {
	s.MockDriver.err = errTest
	_, err := s.OTelDriver.OpenConnector("")
	s.ErrorIs(err, errTest)
	s.Equal(1, s.MockDriver.openConnectorN)
}

func TestDriverSuite(t *testing.T) {
	s := new(DriverSuite)
	suite.Run(t, s)
}
