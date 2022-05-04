# Draft documentation

> The official documentation for this distribution can be found in the
> [Splunk Docs](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/get-started.html)
> site.
> For instructions on how to contribute to the docs, see
> [CONTRIBUTING.md](../CONTRIBUTING.md#documentation).

## Getting started

This Splunk distribution comes with the following defaults:

- [W3C tracecontext](https://www.w3.org/TR/trace-context/) and
  [W3C baggage](https://www.w3.org/TR/baggage/) context propagation.
- [OTLP over gRPC
  exporter](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp)
  configured to send spans to a locally running [Splunk OpenTelemetry
  Connector](https://github.com/signalfx/splunk-otel-collector)
  (`http://localhost:4317`).
- Unlimited default limits for configuration options to
  support full-fidelity traces.

Install the distribution:

```sh
go get github.com/signalfx/splunk-otel-go/distro
```

Configure OpenTelemetry using the [`distro`](../distro) package:

```go
package main

import (
	"context"

	"github.com/signalfx/splunk-otel-go/distro"
)

func main() {
	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}
	// Ensure all spans are flushed before the application exits.
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	// ...
```

### Basic configuration

The `service.name` resource attribute is the only configuration option that
needs to be specified using the `OTEL_RESOURCE_ATTRIBUTES` environment variable.

The `deployment.environment` and `service.version` resource attributes are not
strictly required, but recommended to be set if they are available.

It can be done in the shell:

```sh
OTEL_RESOURCE_ATTRIBUTES="service.name=my-app,service.version=1.2.3,deployment.environment=production"
```

As well as in Go code before executing `distro.Run()`:

```go
os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=my-app,service.version=1.2.3,deployment.environment=development")
```

## Advanced configuration

For advanced configuration options,
refer to the [`distro` package documentation](../distro/README.md#Configuration).

## Correlate traces and logs

You can add trace metadata to logs using the OpenTelemetry trace API. Trace
metadata lets you explore logs in Splunk Observability Cloud.

See [Correlating traces and logs](./correlating-traces-and-logs.md) for more
information.

## Library instrumentation

Supported libraries are listed
[here](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation).

Additional recommended Splunk specific instrumentations:

- [`splunkbuntdb`](../instrumentation/github.com/tidwall/buntdb/splunkbuntdb)
- [`splunkchi`](../instrumentation/github.com/go-chi/chi/splunkchi)
- [`splunkclient-go`](../instrumentation/k8s.io/client-go/splunkclient-go)
- [`splunkdns`](../instrumentation/github.com/miekg/dns/splunkdns)
- [`splunkelastic`](../instrumentation/gopkg.in/olivere/elastic/splunkelastic)
- [`splunkgorm`](../instrumentation/github.com/jinzhu/gorm/splunkgorm)
- [`splunkgraphql`](../instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql)
- [`splunkhttp`](../instrumentation/net/http/splunkhttp)
- [`splunkhttprouter`](../instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter)
- [`splunkkafka`](../instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka)
- [`splunkleveldb`](../instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb)
- [`splunkmysql`](../instrumentation/github.com/go-sql-driver/mysql/splunkmysql)
- [`splunkpgx`](../instrumentation/github.com/jackc/pgx/splunkpgx)
- [`splunkpq`](../instrumentation/github.com/lib/pq/splunkpq)
- [`splunkredigo`](../instrumentation/github.com/gomodule/redigo/splunkredigo)
- [`splunksql`](../instrumentation/database/sql/splunksql)
- [`splunksqlx`](../instrumentation/github.com/jmoiron/sqlx/splunksqlx)

## Manual instrumentation

Documentation on how to manually instrument a Go application is available
[here](https://opentelemetry.io/docs/go/getting-started/).

## Migrating

If you're currently using the [SignalFx Tracing Library for Go](https://github.com/signalfx/signalfx-go-tracing)
and want to migrate to the
Splunk Distribution of OpenTelemetry Go, see [Migrate from the SignalFx Go
Agent](./migrating.md).

## Troubleshooting

For troubleshooting information, see the
[Troubleshooting](./troubleshooting.md) documentation.
