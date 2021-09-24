# Splunk Instrumentation for the Postgres Driver Package lib/pq

[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq)

This package instruments the [`github.com/lib/pq`](https://github.com/lib/pq)
package using the [`splunksql`](../../../../database/sql/splunksql) package.

## Getting Started

This package is design to be a drop-in replacement for the existing use of the
`lib/pg` package when it is used in conjunction with the `database/sql`
package.  The blank identified imports of that package can be replaced with
this package, and the standard library `sql.Open` function can be replaced with
the equivalent `Open` from `splunksql`.

An example can be found [here](./example_test.go).
