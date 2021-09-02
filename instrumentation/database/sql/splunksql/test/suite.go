package test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

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
	DB             *sql.DB

	ConnImplementsPinger         bool
	ConnImplementsExecer         bool
	ConnImplementsExecerContext  bool
	ConnImplementsQueryer        bool
	ConnImplementsQueryerContext bool
}

func NewSplunkSQLSuite(dName string, d driver.Driver) (*SplunkSQLSuite, error) {
	s := new(SplunkSQLSuite)
	s.SpanRecorder = tracetest.NewSpanRecorder()
	s.TracerProvider = trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(s.SpanRecorder),
	)

	splunksql.Register(dName, d, splunksql.WithTracerProvider(s.TracerProvider))
	db, err := splunksql.Open(dName, "mockDB")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	s.DB = db

	c, err := d.Open("test implementation")
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	_, ok := c.(driver.Pinger)
	s.ConnImplementsPinger = ok
	_, ok = c.(driver.Execer)
	s.ConnImplementsExecer = ok
	_, ok = c.(driver.ExecerContext)
	s.ConnImplementsExecerContext = ok
	_, ok = c.(driver.Queryer)
	s.ConnImplementsQueryer = ok
	_, ok = c.(driver.QueryerContext)
	s.ConnImplementsQueryerContext = ok

	return s, nil
}

func (s *SplunkSQLSuite) SetupTest() {
	// Reset the SpanRecorder.
	s.TracerProvider.UnregisterSpanProcessor(s.SpanRecorder)
	s.SpanRecorder = tracetest.NewSpanRecorder()
	s.TracerProvider.RegisterSpanProcessor(s.SpanRecorder)
}

func (s *SplunkSQLSuite) TestDBPing() {
	err := s.DB.Ping()
	if s.ConnImplementsPinger {
		s.Require().NoError(err)
		s.assertSpan(moniker.Ping)
	} else {
		s.ErrorIs(err, driver.ErrSkip)
	}
}

func (s *SplunkSQLSuite) TestDBPingContext() {
	err := s.DB.PingContext(context.Background())
	if s.ConnImplementsPinger {
		s.Require().NoError(err)
		s.assertSpan(moniker.Ping)
	} else {
		s.ErrorIs(err, driver.ErrSkip)
	}
}

func (s *SplunkSQLSuite) TestDBExec() {
	_, err := s.DB.Exec("test")
	s.Require().NoError(err)
	if s.ConnImplementsExecer {
		s.assertSpan(moniker.Exec)
	} else {
		s.assertSpan(moniker.Prepare)
	}
}

func (s *SplunkSQLSuite) TestDBExecContext() {
	_, err := s.DB.ExecContext(context.Background(), "test")
	s.Require().NoError(err)
	if s.ConnImplementsExecer {
		s.assertSpan(moniker.Exec)
	} else {
		s.assertSpan(moniker.Prepare)
	}
}

func (s *SplunkSQLSuite) TestDBQuery() {
	_, err := s.DB.Query("test")
	s.Require().NoError(err)
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query)
	} else {
		s.assertSpan(moniker.Prepare)
	}
}

func (s *SplunkSQLSuite) TestDBQueryContext() {
	_, err := s.DB.QueryContext(context.Background(), "test")
	s.Require().NoError(err)
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query)
	} else {
		s.assertSpan(moniker.Prepare)
	}
}

func (s *SplunkSQLSuite) TestDBQueryRow() {
	r := s.DB.QueryRow("test")
	s.Require().NoError(r.Err())
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query)
	} else {
		s.assertSpan(moniker.Prepare)
	}
}

func (s *SplunkSQLSuite) TestDBQueryRowContext() {
	r := s.DB.QueryRowContext(context.Background(), "test")
	s.Require().NoError(r.Err())
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query)
	} else {
		s.assertSpan(moniker.Prepare)
	}
}

func (s *SplunkSQLSuite) TestDBPrepare() {
	_, err := s.DB.Prepare("test")
	s.Require().NoError(err)
	s.assertSpan(moniker.Prepare)
}

func (s *SplunkSQLSuite) TestDBPrepareContext() {
	// If the database does not support PrepareContext, the instrumentation
	// should fallback to wrapping Prepare directly.
	_, err := s.DB.PrepareContext(context.Background(), "test")
	s.Require().NoError(err)
	s.assertSpan(moniker.Prepare)
}

func (s *SplunkSQLSuite) TestDBBegin() {
	_, err := s.DB.Begin()
	s.Require().NoError(err)
	s.assertSpan(moniker.Begin)
}

func (s *SplunkSQLSuite) TestDBBeginTx() {
	// If the database does not support BeginTx, the instrumentation should
	// fallback to wrapping Begin directly.
	_, err := s.DB.BeginTx(context.Background(), nil)
	s.Require().NoError(err)
	s.assertSpan(moniker.Begin)
}

func (s *SplunkSQLSuite) assertSpan(name moniker.Span) {
	s.assertSpans(name, 1)
}

func (s *SplunkSQLSuite) assertSpans(name moniker.Span, count int) {
	var n int
	for _, roSpan := range s.SpanRecorder.Ended() {
		if roSpan.Name() == name.String() {
			n++
		}
	}
	s.Equalf(count, n, "wrong number of %s spans", name)
}
