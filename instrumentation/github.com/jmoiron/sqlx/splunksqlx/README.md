# Splunk Instrumentation for `github.com/jmoiron/sqlx`

[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx)

This package provides instrumentation for the `github.com/jmoiron/sqlx`
package. The instrumentation is provided by wrapping the
[`splunksql`](../../../../database/sql/splunksql) instrumentation.

## Getting Started

To start using this instrumentation, replace the use of the `sqlx.Open` and
`sqlx.Connect` kind of functions with the equivalent from this package. The
returned values can then be use as normal with the `sqlx` package.

An example of this can be found [here](./example_test.go).
