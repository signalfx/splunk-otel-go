# Splunk Instrumentation for the MySQL Driver Package

[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql)

This package instruments the
[`github.com/go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql)
package using the [`splunksql`](../../../../database/sql/splunksql) package.

## Getting Started

This package is design to be a drop-in replacement for the existing use of the
`mysql` package. The blank identified imports of that package can be replaced
with this package, and the standard library `sql.Open` function can be replaced
with the equivalent `Open` from `splunksql`.

An example can be found [here](./example_test.go).
