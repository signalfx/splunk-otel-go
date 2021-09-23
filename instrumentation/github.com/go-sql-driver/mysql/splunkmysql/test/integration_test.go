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

/*
Package test provides end-to-end testing of the splunkmysql instrumentation
with the default SDK.

This package is in a separate module from the instrumentation it tests to
isolate the dependency of the default SDK and not impose this as a transitive
dependency for users.
*/
package test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	user     = "testuser"
	pass     = "testuser-password"
	rootPass = "secret"
	host     = "localhost"
	port     = 3306
	dbName   = "testdb"

	createStmt = "CREATE TABLE IF NOT EXISTS squareNum ( number integer, squareNumber integer )"
	queryStmt  = "SELECT squareNumber FROM squareNum WHERE number = ?"
	insertStmt = "INSERT INTO squareNum VALUES( ?, ? )"
)

var (
	dsn          = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pass, host, port, dbName)
	dsnRoot      = fmt.Sprintf("root:%s@tcp(%s:%d)/%s", rootPass, host, port, dbName)
	dsnSanitized = fmt.Sprintf("%s@tcp(%s:%d)/%s", user, host, port, dbName)
)

func newFixtures(t *testing.T) (*tracetest.SpanRecorder, *trace.TracerProvider, *sql.DB, func(*testing.T)) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	db, err := splunksql.Open("mysql", dsn, splunksql.WithTracerProvider(tp))
	require.NoError(t, err)
	return sr, tp, db, func(t *testing.T) {
		require.NoError(t, db.Close())
		require.NoError(t, tp.Shutdown(context.Background()))
	}
}

func TestNoContextSpans(t *testing.T) {
	sr, _, db, shutdownFunc := newFixtures(t)

	require.NoError(t, db.Ping())

	_, err := db.Exec(createStmt)
	require.NoError(t, err)

	tx, err := db.Begin()
	require.NoError(t, err)
	stmtIns, err := tx.Prepare(insertStmt)
	require.NoError(t, err)
	for i := 0; i < 25; i++ {
		_, err = stmtIns.Exec(i, (i * i))
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	var sqNum int
	stmtOut, err := db.Prepare(queryStmt)
	require.NoError(t, err)
	require.NoError(t, stmtOut.QueryRow(13).Scan(&sqNum))
	assert.Equal(t, 13*13, sqNum, "failed to query square of 13")

	// Directly do the query.
	require.NoError(t, db.QueryRow(queryStmt, 1).Scan(&sqNum))
	assert.Equal(t, 1, sqNum, "failed to query square of 1")

	shutdownFunc(t)

	// How the DB ensures connections means the number of spans cannot be
	// tested for equality, but we can ensure that each of the operations
	// performed above is represented with at least one span.
	assert.GreaterOrEqual(t, len(sr.Ended()), 33)
	for _, span := range sr.Ended() {
		assertSpanBaseAttrs(t, span)
	}
}

func TestContextSpans(t *testing.T) {
	sr, tp, db, shutdownFunc := newFixtures(t)
	// The TracerProvider that created the span in the passed context will be
	// used to create all the other spans. Make sure to use the TracerProvider
	// with the registered SpanRecorder.
	ctx, parent := tp.Tracer("").Start(context.Background(), "parent")

	require.NoError(t, db.PingContext(ctx))

	_, err := db.ExecContext(ctx, createStmt)
	require.NoError(t, err)

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	stmtIns, err := tx.PrepareContext(ctx, insertStmt)
	require.NoError(t, err)
	for i := 0; i < 25; i++ {
		_, err = stmtIns.ExecContext(ctx, i, (i * i))
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	var sqNum int
	stmtOut, err := db.PrepareContext(ctx, queryStmt)
	require.NoError(t, err)
	require.NoError(t, stmtOut.QueryRowContext(ctx, 13).Scan(&sqNum))
	assert.Equal(t, 13*13, sqNum, "failed to query square of 13")

	// Directly do the query.
	require.NoError(t, db.QueryRowContext(ctx, queryStmt, 1).Scan(&sqNum))
	assert.Equal(t, 1, sqNum, "failed to query square of 1")

	shutdownFunc(t)

	// How the DB ensures connections means the number of spans cannot be
	// tested for equality, but we can ensure that each of the operations
	// performed above is represented with at least one span.
	assert.GreaterOrEqual(t, len(sr.Ended()), 33)
	for _, span := range sr.Ended() {
		// Each span should to be a child of the parent span.
		assert.Equal(t, parent.SpanContext().TraceID(), span.SpanContext().TraceID())
		assertSpanBaseAttrs(t, span)
	}
}

func assertSpanBaseAttrs(t *testing.T, span trace.ReadOnlySpan) {
	a := span.Attributes()
	assert.Contains(t, a, semconv.DBSystemMySQL)
	assert.Contains(t, a, semconv.DBNameKey.String(dbName))
	assert.Contains(t, a, semconv.DBConnectionStringKey.String(dsnSanitized))
	assert.Contains(t, a, semconv.DBUserKey.String(user))
	assert.Contains(t, a, semconv.NetPeerNameKey.String(host))
	assert.Contains(t, a, semconv.NetPeerPortKey.Int(port))
	assert.Contains(t, a, semconv.NetTransportTCP)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		fmt.Println("Skipping running heavy integration test in short mode.")
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3306/tcp": {
				{HostIP: "localhost", HostPort: "3306"},
			},
		},
		Env: []string{
			fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", rootPass),
			fmt.Sprintf("MYSQL_DATABASE=%s", dbName),
			fmt.Sprintf("MYSQL_USER=%s", user),
			fmt.Sprintf("MYSQL_PASSWORD=%s", pass),
		},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// Wait for the database to come up using an exponential-backoff retry.
	if err := pool.Retry(func() error {
		db, err := sql.Open("mysql", dsnRoot)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
