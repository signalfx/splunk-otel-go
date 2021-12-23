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
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"
	"github.com/signalfx/splunk-otel-go/instrumentation/internal"
)

// instrumentationName is the instrumentation library identifier for a Tracer.
const instrumentationName = "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"

// traceConfig contains tracing configuration options.
type traceConfig struct {
	*internal.Config

	DBName string
}

func newTraceConfig(options ...Option) traceConfig {
	c := traceConfig{
		Config: internal.NewConfig(instrumentationName, internal.OptionFunc(
			func(c *internal.Config) {
				c.DefaultStartOpts = []trace.SpanStartOption{
					// From the specification: span kind MUST always be CLIENT.
					trace.WithSpanKind(trace.SpanKindClient),
				}
			}),
		),
	}

	for _, o := range options {
		if o != nil {
			o.apply(&c)
		}
	}

	return c
}

// withSpan wraps the function f with a span.
func (c traceConfig) withSpan(ctx context.Context, m moniker.Span, f func(context.Context) error, opts ...trace.SpanStartOption) error {
	return c.WithSpan(ctx, c.spanName(m), f, opts...)
}

// spanName returns the OpenTelemetry compliant span name.
func (c traceConfig) spanName(m moniker.Span) string {
	// From the OpenTelemetry semantic conventions
	// (https://github.com/open-telemetry/opentelemetry-specification/blob/v1.6.1/specification/trace/semantic_conventions/database.md):
	//
	// > The **span name** SHOULD be set to a low cardinality value representing the statement executed on the database.
	// > It MAY be a stored procedure name (without arguments), DB statement without variable arguments, operation name, etc.
	// > Since SQL statements may have very high cardinality even without arguments, SQL spans SHOULD be named the
	// > following way, unless the statement is known to be of low cardinality:
	// > `<db.operation> <db.name>.<db.sql.table>`, provided that `db.operation` and `db.sql.table` are available.
	// > If `db.sql.table` is not available due to its semantics, the span SHOULD be named `<db.operation> <db.name>`.
	// > It is not recommended to attempt any client-side parsing of `db.statement` just to get these properties,
	// > they should only be used if the library being instrumented already provides them.
	// > When it's otherwise impossible to get any meaningful span name, `db.name` or the tech-specific database name MAY be used.
	//
	// The database/sql package does not provide the database operation nor
	// the SQL table the operation is being performed on during a call. It
	// would require client-side parsing of the statement to determine these
	// properties. Therefore, the database name is used if it is known.
	if c.DBName != "" {
		return c.DBName
	}

	// The database name is not known. Fallback to the known client-side
	// operation being performed. This will comply with the low cardinality
	// recommendation of the specification.
	return m.String()
}

// Option applies options to a tracing configuration.
type Option interface {
	apply(*traceConfig)
}

type optionConv struct {
	iOpt internal.Option
}

func (o optionConv) apply(c *traceConfig) {
	o.iOpt.Apply(c.Config)
}

type optionFunc func(*traceConfig)

func (o optionFunc) apply(c *traceConfig) {
	o(c)
}

// WithTracerProvider returns an Option that sets the TracerProvider used with
// this instrumentation library.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return optionConv{iOpt: internal.WithTracerProvider(tp)}
}

// WithAttributes returns an Option that appends attr to the attributes set
// for every span created with this instrumentation library.
func WithAttributes(attr []attribute.KeyValue) Option {
	return optionConv{iOpt: internal.WithAttributes(attr)}
}

// withRegistrationConfig returns an Option that sets database attributes
// required and recommended by the OpenTelemetry semantic conventions based on
// the information instrumentation registered.
func withRegistrationConfig(regCfg InstrumentationConfig, dsn string) Option {
	var connCfg ConnectionConfig
	if regCfg.DSNParser != nil {
		var err error
		connCfg, err = regCfg.DSNParser(dsn)
		if err != nil {
			otel.Handle(err)
		}
	} else {
		// Fallback. This is a best effort attempt if we do not know how to
		// explicitly parse the DSN.
		connCfg, _ = urlDSNParse(dsn)
	}

	attrs, err := connCfg.Attributes()
	if err != nil {
		otel.Handle(err)
	}
	attrs = append(attrs, regCfg.DBSystem.Attribute())

	return optionFunc(func(c *traceConfig) {
		c.DBName = connCfg.Name
		c.DefaultStartOpts = append(c.DefaultStartOpts, trace.WithAttributes(attrs...))
	})
}

// ConnectionConfig are the relevant settings parsed from a database
// connection.
type ConnectionConfig struct {
	// Name of the database being accessed.
	Name string
	// ConnectionString is the sanitized connection string (all credentials
	// have been redacted) used to connect to the database.
	ConnectionString string
	// User is the username used to access the database.
	User string
	// Host is the IP or hostname of the database.
	Host string
	// Port is the port the database is lisening on.
	Port int
	// NetTransport is the transport protocol used to connect to the database.
	NetTransport NetTransport
}

// Attributes returns the connection settings as attributes compliant with
// OpenTelemetry semantic coventions. If the settings do not conform to
// OpenTelemetry requirements an error is returned with a partial list of
// attributes that do conform.
func (c ConnectionConfig) Attributes() ([]attribute.KeyValue, error) { // nolint: gocritic // This is short lived, pass the type.
	var attrs []attribute.KeyValue
	var errs []string
	if c.Name != "" {
		attrs = append(attrs, semconv.DBNameKey.String(c.Name))
	}
	if c.ConnectionString != "" {
		attrs = append(attrs, semconv.DBConnectionStringKey.String(c.ConnectionString))
	}
	if c.User != "" {
		attrs = append(attrs, semconv.DBUserKey.String(c.User))
	}
	if c.Host != "" {
		if ip := net.ParseIP(c.Host); ip != nil {
			attrs = append(attrs, semconv.NetPeerIPKey.String(ip.String()))
		} else {
			attrs = append(attrs, semconv.NetPeerNameKey.String(c.Host))
		}
	} else {
		errs = append(errs, "missing required peer IP or hostname")
	}
	if c.Port > 0 {
		attrs = append(attrs, semconv.NetPeerPortKey.Int(c.Port))
	}
	attrs = append(attrs, c.NetTransport.Attribute())

	var err error
	if len(errs) > 0 {
		err = fmt.Errorf("invalid connection config: %s", strings.Join(errs, ", "))
	}
	return attrs, err
}

func urlDSNParse(dataSourceName string) (ConnectionConfig, error) {
	var connCfg ConnectionConfig
	u, err := url.Parse(dataSourceName)
	if err != nil {
		return connCfg, err
	}

	connCfg.Host = u.Hostname()
	if p, err := strconv.Atoi(u.Port()); err == nil {
		connCfg.Port = p
	}

	if u.User != nil {
		connCfg.User = u.User.Username()
		if _, ok := u.User.Password(); ok {
			// Redact password.
			u.User = url.User(u.User.Username())
		}
	}

	connCfg.ConnectionString = u.String()

	return connCfg, nil
}

// DSNParser processes a driver-specific data source name into
// connection-level attributes conforming with the OpenTelemetry semantic
// conventions.
type DSNParser func(dataSourceName string) (ConnectionConfig, error)

// InstrumentationConfig is the setup configuration for the instrumentation of
// a database driver.
type InstrumentationConfig struct {
	// DBSystem is the database system being registered.
	DBSystem  DBSystem
	DSNParser DSNParser
}
