# Splunk Instrumentation for the PostgreSQL Driver Package pgx

[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go/instrumentation/github.com/pgx/v5/jackc/pgx/splunkpgx.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/v5/pgx/splunkpgx)

This package instruments the
[`github.com/jackc/pgx`](https://github.com/jackc/pgx) package using the
[`splunksql`](../../../../../database/sql/splunksql) package.

## Getting Started

This package is design to be a drop-in replacement for the existing use of the
`pgx` package when it is used in conjunction with the `database/sql` package.
The blank identified import of `github.com/jackc/pgx/v5/stdlib` can be replaced
with this package, and the standard library `sql.Open` function can be replaced
with the equivalent `Open` from `splunksql`. An example can be found
[here](example_test.go).
