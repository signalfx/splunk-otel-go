package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSplunkSQLFullImplementation(t *testing.T) {
	driverName := "splunktest-full"
	driver := NewFullMockDriver()
	s, err := NewSplunkSQLSuite(driverName, driver)
	if err != nil {
		t.Fatal("failed to setup test suite", err)
	}
	suite.Run(t, s)
}

func TestSplunkSQLSimpleImplementation(t *testing.T) {
	driverName := "splunktest-simple"
	driver := NewSimpleMockDriver()
	s, err := NewSplunkSQLSuite(driverName, driver)
	if err != nil {
		t.Fatal("failed to setup test suite", err)
	}
	suite.Run(t, s)
}
