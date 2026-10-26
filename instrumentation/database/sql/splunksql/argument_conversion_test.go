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
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	conversionConnRoute   = "conn"
	conversionStmtRoute   = "stmt"
	conversionColumnRoute = "column"
	conversionOption      = "option"
)

type (
	conversionValue     struct{}
	conversionCheckFunc func(*driver.NamedValue) error
)

func (f conversionCheckFunc) CheckNamedValue(v *driver.NamedValue) error { return f(v) }

type conversionValueFunc func(any) (driver.Value, error)

func (f conversionValueFunc) ConvertValue(v any) (driver.Value, error) { return f(v) }

type conversionColumn struct{ driver.ValueConverter }

func (c conversionColumn) ColumnConverter(int) driver.ValueConverter { return c.ValueConverter }

type conversionStmt struct {
	driver.Stmt
	args *[]driver.NamedValue
}

func (*conversionStmt) Close() error  { return nil }
func (*conversionStmt) NumInput() int { return 1 }
func (s *conversionStmt) ExecContext(_ context.Context, args []driver.NamedValue) (driver.Result, error) {
	*s.args = args
	return driver.RowsAffected(1), nil
}

func (s *conversionStmt) QueryContext(_ context.Context, args []driver.NamedValue) (driver.Rows, error) {
	*s.args = args
	return conversionRows{}, nil
}

type conversionRows struct{}

func (conversionRows) Columns() []string         { return nil }
func (conversionRows) Close() error              { return nil }
func (conversionRows) Next([]driver.Value) error { return io.EOF }

type conversionConn struct {
	driver.Conn
	stmt driver.Stmt
	args *[]driver.NamedValue
}

func (c *conversionConn) Prepare(string) (driver.Stmt, error) { return c.stmt, nil }
func (*conversionConn) Close() error                          { return nil }
func (c *conversionConn) ExecContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Result, error) {
	*c.args = args
	return driver.RowsAffected(1), nil
}

func (c *conversionConn) QueryContext(_ context.Context, _ string, args []driver.NamedValue) (driver.Rows, error) {
	*c.args = args
	return conversionRows{}, nil
}

type conversionConnector struct{ conn driver.Conn }

func (c conversionConnector) Connect(context.Context) (driver.Conn, error) { return c.conn, nil }
func (conversionConnector) Driver() driver.Driver                          { return newMockDriver(nil) }

func TestArgumentConversion(t *testing.T) {
	const (
		accepted  = "accepted"
		converted = "converted"
	)
	accept := conversionCheckFunc(func(v *driver.NamedValue) error { v.Value = accepted; return nil })
	skip := conversionCheckFunc(func(*driver.NamedValue) error { return driver.ErrSkip })
	reject := conversionCheckFunc(func(*driver.NamedValue) error { return errTest })
	remove := conversionCheckFunc(func(v *driver.NamedValue) error {
		if v.Value == conversionOption {
			return driver.ErrRemoveArgument
		}
		return accept(v)
	})
	convert := conversionValueFunc(func(any) (driver.Value, error) { return converted, nil })

	tests := []struct {
		name      string
		connCheck conversionCheckFunc
		stmtCheck conversionCheckFunc
		converter driver.ValueConverter
		args      []any
		value     any
		route     []string
		err       error
	}{
		{name: "connection checker", connCheck: accept, args: []any{conversionValue{}}, value: accepted, route: []string{conversionConnRoute}},
		{name: "statement checker precedence", connCheck: reject, stmtCheck: accept, args: []any{conversionValue{}}, value: accepted, route: []string{conversionStmtRoute}},
		{name: "column converter", converter: convert, args: []any{conversionValue{}}, value: converted, route: []string{conversionColumnRoute}},
		{name: "connection skip to column", connCheck: skip, converter: convert, args: []any{conversionValue{}}, value: converted, route: []string{conversionConnRoute, conversionColumnRoute}},
		{name: "statement skip to column", connCheck: reject, stmtCheck: skip, converter: convert, args: []any{conversionValue{}}, value: converted, route: []string{conversionStmtRoute, conversionColumnRoute}},
		{name: "statement skip to default", connCheck: reject, stmtCheck: skip, args: []any{42}, value: int64(42), route: []string{conversionStmtRoute}},
		{name: "connection skip to default", connCheck: skip, args: []any{42}, value: int64(42), route: []string{conversionConnRoute}},
		{name: "default conversion", args: []any{42}, value: int64(42)},
		{name: "remove argument", connCheck: remove, args: []any{conversionOption, conversionValue{}}, value: accepted, route: []string{conversionConnRoute, conversionConnRoute}},
		{name: "connection error", connCheck: reject, args: []any{42}, route: []string{conversionConnRoute}, err: errTest},
		{name: "statement error", connCheck: accept, stmtCheck: reject, args: []any{42}, route: []string{conversionStmtRoute}, err: errTest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, query := range []bool{false, true} {
				for _, wrapped := range []bool{false, true} {
					var got []driver.NamedValue
					var route []string
					stmt := newConversionStmt(tt.stmtCheck, tt.converter, &got, &route)
					conn := newConversionConn(stmt, tt.connCheck, &got, &route)
					db := conversionDB(conn, wrapped)
					t.Cleanup(func() { assert.NoError(t, db.Close()) })
					prepared, err := db.PrepareContext(context.Background(), "test")
					require.NoError(t, err)
					t.Cleanup(func() { assert.NoError(t, prepared.Close()) })
					err = runConversionStatement(t, prepared, query, tt.args)
					if tt.err != nil {
						assert.ErrorIs(t, err, tt.err)
						assert.Nil(t, got)
					} else {
						require.NoError(t, err, "wrapped=%t, query=%t", wrapped, query)
						assert.Equal(t, []driver.NamedValue{{Ordinal: 1, Value: tt.value}}, got)
					}
					assert.Equal(t, tt.route, route, "wrapped=%t, query=%t", wrapped, query)
				}
			}
		})
	}
}

func conversionDB(conn driver.Conn, wrapped bool) *sql.DB {
	connector := driver.Connector(conversionConnector{conn: conn})
	if wrapped {
		connector = newConnector(connector, &otelDriver{config: newConfig()})
	}
	return sql.OpenDB(connector)
}

func newConversionStmt(check conversionCheckFunc, converter driver.ValueConverter, args *[]driver.NamedValue, route *[]string) driver.Stmt {
	stmt := &conversionStmt{args: args}
	var checker driver.NamedValueChecker
	if check != nil {
		checker = conversionCheckFunc(func(v *driver.NamedValue) error {
			*route = append(*route, conversionStmtRoute)
			return check(v)
		})
	}
	var column *conversionColumn
	if converter != nil {
		column = &conversionColumn{conversionValueFunc(func(v any) (driver.Value, error) {
			*route = append(*route, conversionColumnRoute)
			return converter.ConvertValue(v)
		})}
	}
	switch {
	case checker != nil && column != nil:
		return struct {
			*conversionStmt
			driver.NamedValueChecker
			*conversionColumn
		}{stmt, checker, column}
	case checker != nil:
		return struct {
			*conversionStmt
			driver.NamedValueChecker
		}{stmt, checker}
	case column != nil:
		return struct {
			*conversionStmt
			*conversionColumn
		}{stmt, column}
	default:
		return stmt
	}
}

func newConversionConn(stmt driver.Stmt, check conversionCheckFunc, args *[]driver.NamedValue, route *[]string) driver.Conn {
	conn := &conversionConn{stmt: stmt, args: args}
	if check == nil {
		return conn
	}
	return struct {
		*conversionConn
		driver.NamedValueChecker
	}{conn, conversionCheckFunc(func(v *driver.NamedValue) error {
		*route = append(*route, conversionConnRoute)
		return check(v)
	})}
}

func runConversionStatement(t *testing.T, stmt *sql.Stmt, query bool, args []any) error {
	t.Helper()
	if !query {
		_, err := stmt.ExecContext(context.Background(), args...)
		return err
	}
	rows, err := stmt.QueryContext(context.Background(), args...)
	if err != nil {
		return err
	}
	defer func() { assert.NoError(t, rows.Close()) }()
	return rows.Err()
}

func TestConnectionArgumentConversion(t *testing.T) {
	for _, query := range []bool{false, true} {
		for _, wrapped := range []bool{false, true} {
			var got []driver.NamedValue
			var route []string
			check := conversionCheckFunc(func(v *driver.NamedValue) error {
				if v.Name == conversionOption {
					return driver.ErrRemoveArgument
				}
				return nil
			})
			conn := newConversionConn(nil, check, &got, &route)
			db := conversionDB(conn, wrapped)
			t.Cleanup(func() { assert.NoError(t, db.Close()) })
			args := []any{sql.Named(conversionOption, conversionValue{}), sql.Named("value", conversionValue{})}
			if query {
				require.NoError(t, queryConversionConnection(t, db, args))
			} else {
				_, err := db.ExecContext(context.Background(), "test", args...)
				require.NoError(t, err)
			}
			assert.Equal(t, []driver.NamedValue{{Name: "value", Ordinal: 1, Value: conversionValue{}}}, got)
			assert.Equal(t, []string{conversionConnRoute, conversionConnRoute}, route)
		}
	}
}

func queryConversionConnection(t *testing.T, db *sql.DB, args []any) error {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), "test", args...)
	if err != nil {
		return err
	}
	defer func() { assert.NoError(t, rows.Close()) }()
	return rows.Err()
}
