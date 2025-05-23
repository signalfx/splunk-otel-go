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
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	traceapi "go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
)

// SplunkSQLSuite is the tracing test suite for the splunksql package.
type SplunkSQLSuite struct {
	suite.Suite

	SpanRecorder   *tracetest.SpanRecorder
	BaseAttributes []attribute.KeyValue
	TracerProvider *trace.TracerProvider
	DB             *sql.DB

	ConnImplementsPinger         bool
	ConnImplementsExecer         bool
	ConnImplementsExecerContext  bool
	ConnImplementsQueryer        bool
	ConnImplementsQueryerContext bool
}

// NewSplunkSQLSuite returns an initialized SplunkSQLSuite.
func NewSplunkSQLSuite(dName string, d driver.Driver) (*SplunkSQLSuite, error) {
	s := new(SplunkSQLSuite)
	s.SpanRecorder = tracetest.NewSpanRecorder()
	s.TracerProvider = trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(s.SpanRecorder),
	)

	dbSys := splunksql.DBSystemOtherSQL
	connCfg := splunksql.ConnectionConfig{
		// Do not set the Name field so monikers are used to identify
		// spans.
		ConnectionString: "mockDB://bob@localhost:8080/testDB",
		User:             "bob",
		Host:             "localhost",
		Port:             8080,
		NetTransport:     splunksql.NetTransportOther,
	}
	s.BaseAttributes, _ = connCfg.Attributes()
	s.BaseAttributes = append(s.BaseAttributes, dbSys.Attribute())

	sql.Register(dName, d)
	splunksql.Register(dName, splunksql.InstrumentationConfig{
		DBSystem:  dbSys,
		DSNParser: func(string) (splunksql.ConnectionConfig, error) { return connCfg, nil },
	})
	db, err := splunksql.Open(dName, "mockDB", splunksql.WithTracerProvider(s.TracerProvider))
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
	_, ok = c.(driver.Execer) //nolint: staticcheck // Ensure backwards support of deprecated interface.
	s.ConnImplementsExecer = ok
	_, ok = c.(driver.ExecerContext)
	s.ConnImplementsExecerContext = ok
	_, ok = c.(driver.Queryer) //nolint: staticcheck // Ensure backwards support of deprecated interface.
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

func (s *SplunkSQLSuite) TearDownSuite() {
	s.NoError(s.DB.Close())
}

func (s *SplunkSQLSuite) TestDBPing() {
	err := s.DB.Ping() //nolint:noctx  // Test Ping operation.
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
	_, err := s.DB.Exec("test") //nolint:noctx  // Test Exec operation.
	s.Require().NoError(err)
	if s.ConnImplementsExecer {
		s.assertSpan(moniker.Exec, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	} else {
		s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	}
}

func (s *SplunkSQLSuite) TestDBExecContext() {
	_, err := s.DB.ExecContext(context.Background(), "test")
	s.Require().NoError(err)
	if s.ConnImplementsExecer {
		s.assertSpan(moniker.Exec, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	} else {
		s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	}
}

func (s *SplunkSQLSuite) TestDBQuery() {
	rows, err := s.DB.Query("test") //nolint:noctx  // Test Query operation.
	s.T().Cleanup(func() { s.Assert().NoError(rows.Err()) })
	s.T().Cleanup(func() { s.Assert().NoError(rows.Close()) })
	s.Require().NoError(err)
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	} else {
		s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	}
}

func (s *SplunkSQLSuite) TestDBQueryContext() {
	rows, err := s.DB.QueryContext(context.Background(), "test")
	s.T().Cleanup(func() { s.Assert().NoError(rows.Close()) })
	s.Require().NoError(err)
	s.Require().NoError(rows.Err())
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	} else {
		s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	}
}

func (s *SplunkSQLSuite) TestDBQueryRow() {
	r := s.DB.QueryRow("test") //nolint:noctx  // Test QueryRow operation.
	s.Require().NoError(r.Err())
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	} else {
		s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	}
}

func (s *SplunkSQLSuite) TestDBQueryRowContext() {
	r := s.DB.QueryRowContext(context.Background(), "test")
	s.Require().NoError(r.Err())
	if s.ConnImplementsQueryer {
		s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	} else {
		s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
	}
}

func (s *SplunkSQLSuite) TestDBPrepare() {
	stmt, err := s.DB.Prepare("test") //nolint:noctx  // Test Prepare operation.
	s.T().Cleanup(func() { s.Assert().NoError(stmt.Close()) })
	s.Require().NoError(err)
	s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
}

func (s *SplunkSQLSuite) TestDBPrepareContext() {
	// If the database does not support PrepareContext, the instrumentation
	// should fallback to wrapping Prepare directly.
	stmt, err := s.DB.PrepareContext(context.Background(), "test")
	s.T().Cleanup(func() { s.Assert().NoError(stmt.Close()) })
	s.Require().NoError(err)
	s.assertSpan(moniker.Prepare, traceapi.WithAttributes(semconv.DBStatementKey.String("test")))
}

func (s *SplunkSQLSuite) TestDBBegin() {
	tx, err := s.DB.Begin()
	s.Require().NoError(err)
	s.assertSpan(moniker.Begin)
	_ = tx.Rollback()
}

func (s *SplunkSQLSuite) TestDBBeginTx() {
	// If the database does not support BeginTx, the instrumentation should
	// fallback to wrapping Begin directly.
	tx, err := s.DB.BeginTx(context.Background(), nil)
	s.Require().NoError(err)
	s.assertSpan(moniker.Begin)
	_ = tx.Rollback()
}

func (s *SplunkSQLSuite) newStmt() *sql.Stmt {
	stmt, err := s.DB.Prepare("test query") //nolint:noctx  // Testing lib.
	s.T().Cleanup(func() { s.Assert().NoError(stmt.Close()) })
	s.Require().NoError(err)
	return stmt
}

func (s *SplunkSQLSuite) TestStmtExec() {
	_, err := s.newStmt().Exec() //nolint:sqlclosecheck // newStmt() adds the (*sql.Stmt).Close() to test cleanup
	s.Require().NoError(err)
	s.assertSpan(moniker.Exec, traceapi.WithAttributes(semconv.DBStatementKey.String("test query")))
}

func (s *SplunkSQLSuite) TestStmtExecContext() {
	_, err := s.newStmt().ExecContext(context.Background()) //nolint:sqlclosecheck // newStmt() adds the (*sql.Stmt).Close() to test cleanup
	s.Require().NoError(err)
	s.assertSpan(moniker.Exec, traceapi.WithAttributes(semconv.DBStatementKey.String("test query")))
}

func (s *SplunkSQLSuite) TestStmtQuery() {
	rows, err := s.newStmt().Query() //nolint:sqlclosecheck // newStmt() adds the (*sql.Stmt).Close() to test cleanup
	s.T().Cleanup(func() { s.Assert().NoError(rows.Close()) })
	s.Require().NoError(err)
	s.Require().NoError(rows.Err())
	s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test query")))
}

func (s *SplunkSQLSuite) TestStmtQueryContext() {
	rows, err := s.newStmt().QueryContext(context.Background()) //nolint:sqlclosecheck // newStmt() adds the (*sql.Stmt).Close() to test cleanup
	s.T().Cleanup(func() { s.Assert().NoError(rows.Close()) })
	s.Require().NoError(err)
	s.Require().NoError(rows.Err())
	s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test query")))
}

func (s *SplunkSQLSuite) TestStmtQueryRow() {
	r := s.newStmt().QueryRow() //nolint:sqlclosecheck // newStmt adds the (*sql.Stmt).Close() to test cleanup
	s.Require().NoError(r.Err())
	s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test query")))
}

func (s *SplunkSQLSuite) TestStmtQueryRowContext() {
	r := s.newStmt().QueryRowContext(context.Background()) //nolint:sqlclosecheck // newStmt() adds the (*sql.Stmt).Close() to test cleanup
	s.Require().NoError(r.Err())
	s.assertSpan(moniker.Query, traceapi.WithAttributes(semconv.DBStatementKey.String("test query")))
}

func (s *SplunkSQLSuite) TestTxCommit() {
	tx, err := s.DB.Begin()
	s.Require().NoError(err)
	s.Require().NoError(tx.Commit())
	s.assertSpan(moniker.Commit)
}

func (s *SplunkSQLSuite) TestTxRollback() {
	tx, err := s.DB.Begin()
	s.Require().NoError(err)
	s.Require().NoError(tx.Rollback())
	s.assertSpan(moniker.Rollback)
}

func (s *SplunkSQLSuite) assertSpan(name moniker.Span, opt ...traceapi.SpanStartOption) {
	c := traceapi.NewSpanStartConfig(opt...)
	s.assertSpans(name, 1, c)
}

func (s *SplunkSQLSuite) assertSpans(name moniker.Span, count int, c traceapi.SpanConfig) { //nolint: gocritic // passing c by value is fine.
	attrs := make([]attribute.KeyValue, 0, len(c.Attributes())+len(s.BaseAttributes))
	attrs = append(attrs, s.BaseAttributes...)
	attrs = append(attrs, c.Attributes()...)

	var n int
	for _, roSpan := range s.SpanRecorder.Ended() {
		if roSpan.Name() == name.String() {
			n++
			s.ElementsMatchf(attrs, roSpan.Attributes(), "span: %s", roSpan.Name())
			s.ElementsMatch(c.Links(), roSpan.Links())
			if c.NewRoot() && roSpan.Parent().IsValid() {
				s.Failf("non-root span", "span %s should not have a parent", name)
			}
		}
		s.Equalf(traceapi.SpanKindClient, roSpan.SpanKind(), "span %q is not a client span", name)
		s.Equal(splunksql.Version(), roSpan.InstrumentationLibrary().Version, "version should match")
	}
	s.Equalf(count, n, "wrong number of %s spans", name)
}
