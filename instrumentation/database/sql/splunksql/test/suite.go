package test

import (
	"context"
	"database/sql"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type SplunkSQLSuite struct {
	suite.Suite

	SpanRecorder   *tracetest.SpanRecorder
	TracerProvider *trace.TracerProvider
	MockDriver     *MockDriver
	DB             *sql.DB
}

func (suite *SplunkSQLSuite) SetupSuite() {
	suite.SpanRecorder = tracetest.NewSpanRecorder()
	suite.TracerProvider = trace.NewTracerProvider(
		trace.WithSpanProcessor(suite.SpanRecorder),
		trace.WithSampler(trace.AlwaysSample()),
	)
	suite.MockDriver = NewMockDriver()
	splunksql.Register(
		"splunktest",
		suite.MockDriver,
		splunksql.WithTracerProvider(suite.TracerProvider),
	)

	db, err := splunksql.Open("splunktest", "mockDB")
	if err != nil {
		suite.FailNow("failed to open database", err)
	}
	suite.DB = db
}

func (suite *SplunkSQLSuite) SetupTest() {
	// Reset the SpanRecorder.
	suite.TracerProvider.UnregisterSpanProcessor(suite.SpanRecorder)
	suite.SpanRecorder = tracetest.NewSpanRecorder()
	suite.TracerProvider.RegisterSpanProcessor(suite.SpanRecorder)

	suite.MockDriver.Reset()
}

func (suite *SplunkSQLSuite) TestDriverOpen() {
	suite.DB.Driver().Open("name")
	suite.Equal(uint64(1), suite.MockDriver.OpenN)
}

func (suite *SplunkSQLSuite) TestDBPing() {
	suite.Require().NoError(suite.DB.Ping())
	suite.assertPing()
}

func (suite *SplunkSQLSuite) TestDBPingContext() {
	suite.Require().NoError(suite.DB.PingContext(context.Background()))
	suite.assertPing()
}

func (suite *SplunkSQLSuite) assertPing() {
	suite.assertSpans(uint64(1), suite.MockDriver.connector.conn.PingN, moniker.Ping)
}

func (suite *SplunkSQLSuite) TestDBExec() {
	_, err := suite.DB.Exec("test")
	suite.Require().NoError(err)
	suite.assertExec()
}

func (suite *SplunkSQLSuite) TestDBExecContext() {
	_, err := suite.DB.ExecContext(context.Background(), "test")
	suite.Require().NoError(err)
	suite.assertExec()
}

func (suite *SplunkSQLSuite) assertExec() {
	suite.Equal(uint64(0), suite.MockDriver.connector.conn.ExecN)
	suite.assertSpans(uint64(1), suite.MockDriver.connector.conn.ExecContextN, moniker.Exec)
}

func (suite *SplunkSQLSuite) TestDBQuery() {
	_, err := suite.DB.Query("test")
	suite.Require().NoError(err)
	suite.assertQuery()
}

func (suite *SplunkSQLSuite) TestDBQueryContext() {
	_, err := suite.DB.QueryContext(context.Background(), "test")
	suite.Require().NoError(err)
	suite.assertQuery()
}

func (suite *SplunkSQLSuite) TestDBQueryRow() {
	_ = suite.DB.QueryRow("test")
	suite.assertQuery()
}

func (suite *SplunkSQLSuite) TestDBQueryRowContext() {
	_ = suite.DB.QueryRowContext(context.Background(), "test")
	suite.assertQuery()
}

func (suite *SplunkSQLSuite) assertQuery() {
	suite.Equal(uint64(0), suite.MockDriver.connector.conn.QueryN)
	suite.assertSpans(uint64(1), suite.MockDriver.connector.conn.QueryContextN, moniker.Query)
}

func (suite *SplunkSQLSuite) TestDBPrepare() {
	_, err := suite.DB.Prepare("test")
	suite.Require().NoError(err)
	suite.assertPrepare()
}

func (suite *SplunkSQLSuite) TestDBPrepareContext() {
	_, err := suite.DB.PrepareContext(context.Background(), "test")
	suite.Require().NoError(err)
	suite.assertPrepare()
}

func (suite *SplunkSQLSuite) assertPrepare() {
	suite.Equal(uint64(0), suite.MockDriver.connector.conn.PrepareN)
	suite.assertSpans(uint64(1), suite.MockDriver.connector.conn.PrepareContextN, moniker.Prepare)
}

func (suite *SplunkSQLSuite) TestDBBegin() {
	_, err := suite.DB.Begin()
	suite.Require().NoError(err)
	suite.assertBegin()
}

func (suite *SplunkSQLSuite) TestDBBeginTx() {
	_, err := suite.DB.BeginTx(context.Background(), nil)
	suite.Require().NoError(err)
	suite.assertBegin()
}

func (suite *SplunkSQLSuite) assertBegin() {
	suite.Equal(uint64(0), suite.MockDriver.connector.conn.BeginN)
	suite.assertSpans(uint64(1), suite.MockDriver.connector.conn.BeginTxN, moniker.Begin)
}

func (suite *SplunkSQLSuite) assertSpans(expected, got uint64, m moniker.Span) {
	suite.Equal(expected, got)

	var n uint64
	for _, s := range suite.SpanRecorder.Ended() {
		switch s.Name() {
		case moniker.Reset:
			// Ignore. Reset is called by the DB based on prior state.
		case m.String():
			n++
		default:
			suite.Failf("unknown span", "%s", s.Name())
		}
	}
	suite.Equalf(expected, n, "wrong number of %s spans", m)
}
