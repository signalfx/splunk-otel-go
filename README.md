# Splunk Distribution of OpenTelemetry Go

[![Splunk GDI Specification](https://img.shields.io/badge/GDI-1.2.0-blue)](https://github.com/signalfx/gdi-specification/releases/tag/v1.2.0)
[![GitHub Release](https://img.shields.io/github/v/release/signalfx/splunk-otel-go?include_prereleases)](https://github.com/signalfx/splunk-otel-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/signalfx/splunk-otel-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/signalfx/splunk-otel-go)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/signalfx/splunk-otel-go/ci)](https://github.com/signalfx/splunk-otel-go/actions?query=branch%3Amain)
[![codecov](https://codecov.io/gh/signalfx/splunk-otel-go/branch/main/graph/badge.svg)](https://codecov.io/gh/signalfx/splunk-otel-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/signalfx/splunk-otel-go)](https://goreportcard.com/report/github.com/signalfx/splunk-otel-go)

The Splunk distribution of [OpenTelemetry
Go](https://github.com/open-telemetry/opentelemetry-go) provides
multiple packages that automatically instruments your Go
application to capture and report distributed traces to Splunk APM.

Read the official documentation for this distribution in the
[Splunk Docs site](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/get-started.html).

> :construction: This project is currently in **BETA**.
> It is **officially supported** by Splunk.
> However, breaking changes **MAY** be introduced.

If you're currently using the [SignalFx Tracing Library for Go](https://github.com/signalfx/signalfx-go-tracing)
and want to migrate to the
Splunk Distribution of OpenTelemetry Go, see [Migrate from the SignalFx Go
Agent](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/troubleshooting/migrate-signalfx-go-to-otel.html).

Table of Contents:

- [Getting started](#getting-started)
  - [Basic configuration](#basic-configuration)
- [Library instrumentation](#library-instrumentation)
- [Manual instrumentation](#manual-instrumentation)
- [Advanced configuration](#advanced-configuration)
  - [Splunk distribution configuration](#splunk-distribution-configuration)
  - [Trace configuration](#trace-configuration)
  - [Trace exporter configuration](#trace-exporter-configuration)
  - [Trace propagation configuration](#trace-propagation-configuration)
- [License](#license)

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

Configure OpenTelemetry using the [`distro`](./distro) package:

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

For advanced configuration options, refer to the [`distro` package documentation](./distro/README.md#Configuration).

## Advanced configuration

See [Advanced settings](./docs/advanced-config.md).

## Correlate traces and logs

You can add trace metadata to logs using the OpenTelemetry trace API. Trace
metadata lets you explore logs in Splunk Observability Cloud.

See [Correlating traces and logs](./docs/correlating-traces-and-logs.md) for more
information.

## Library instrumentation

Supported libraries are listed
[here](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation).

Additional recommended Splunk specific instrumentations:

- [`splunkbuntdb`](./instrumentation/github.com/tidwall/buntdb/splunkbuntdb)
- [`splunkchi`](./instrumentation/github.com/go-chi/chi/splunkchi)
- [`splunkclient-go`](./instrumentation/k8s.io/client-go/splunkclient-go)
- [`splunkdns`](./instrumentation/github.com/miekg/dns/splunkdns)
- [`splunkelastic`](./instrumentation/gopkg.in/olivere/elastic/splunkelastic)
- [`splunkgorm`](./instrumentation/github.com/jinzhu/gorm/splunkgorm)
- [`splunkgraphql`](./instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql)
- [`splunkhttp`](./instrumentation/net/http/splunkhttp)
- [`splunkhttprouter`](./instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter)
- [`splunkkafka`](./instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka)
- [`splunkleveldb`](./instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb)
- [`splunkmysql`](./instrumentation/github.com/go-sql-driver/mysql/splunkmysql)
- [`splunkpgx`](./instrumentation/github.com/jackc/pgx/splunkpgx)
- [`splunkpq`](./instrumentation/github.com/lib/pq/splunkpq)
- [`splunkredigo`](./instrumentation/github.com/gomodule/redigo/splunkredigo)
- [`splunksql`](./instrumentation/database/sql/splunksql)
- [`splunksqlx`](./instrumentation/github.com/jmoiron/sqlx/splunksqlx)

## Manual instrumentation

Documentation on how to manually instrument a Go application is available
[here](https://opentelemetry.io/docs/go/getting-started/).

## Troubleshooting

For troubleshooting information, see the
[Troubleshooting](./docs/troubleshooting.md) documentation.

## License

The Splunk distribution of OpenTelemetry Go is a
distribution of the [OpenTelemetry Go
project](https://github.com/open-telemetry/opentelemetry-go). It is
released under the terms of the Apache Software License version 2.0. See [the
license file](./LICENSE) for more details.

>ℹ️&nbsp;&nbsp;SignalFx was acquired by Splunk in October 2019. See [Splunk
SignalFx](https://www.splunk.com/en_us/investor-relations/acquisitions/signalfx.html)
for more information.
