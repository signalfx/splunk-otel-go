package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSplunkSQL(t *testing.T) {
	suite.Run(t, new(SplunkSQLSuite))
}
