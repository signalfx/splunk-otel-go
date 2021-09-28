# Splunk Instrumentation for the `database/sql` Package

[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql)

This package instruments the [`database/sql`](https://pkg.go.dev/database/sql)
package.

## Getting Started

This package is designed to be used in conjunction with existing database
drivers that have been instrumented so OpenTelemetry semantic conventions can
be discovered. An example of this type of use can be found
[here](../../../github.com/go-sql-driver/mysql/splunkmysql).

This package can be used directly as well. It becomes the users responsibility
to ensure accurate and complete information about the database system is passed
as attributes to ensure OpenTelemetry semantic conventions are satisfied. An
example of this can be found [here](./example_test.go).
