# Splunk distribution of OpenTelemetry Go

[![GitHub Release](https://img.shields.io/github/v/release/signalfx/splunk-otel-go?include_prereleases)](https://github.com/signalfx/splunk-otel-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/signalfx/splunk-otel-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/signalfx/splunk-otel-go)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/signalfx/splunk-otel-go/ci)](https://github.com/signalfx/splunk-otel-go/actions?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/signalfx/splunk-otel-go)](https://goreportcard.com/report/github.com/signalfx/splunk-otel-go)

The Splunk distribution of [OpenTelemetry
Go](https://github.com/open-telemetry/opentelemetry-go) provides
multiple packages that automatically instruments your Go
application to capture and report distributed traces to Splunk APM.

> :construction: This project is currently in **BETA**.
> It is **officially supported** by Splunk.
> However, breaking changes **MAY** be introduced.

Table of Contents:

- [Getting Started](#getting-started)
  - [Basic Configuration](#basic-configuration)
- [Library Instrumentation](#library-instrumentation)
- [Manual Instrumentation](#manual-instrumentation)
- [Advanced Configuration](#advanced-configuration)
  - [Splunk Distribution Configuration](#splunk-distribution-configuration)
  - [Trace Configuration](#trace-configuration)
  - [Trace Exporter Configuration](#trace-exporter-configuration)
  - [Trace Propagation Configuration](#trace-propagation-configuration)
- [License](#license)

## Getting Started

This Splunk distribution comes with the following defaults:

- [B3 context propagation](https://github.com/openzipkin/b3-propagation).
- [Jaeger Thrift over HTTP
  exporter](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/jaeger)
  configured to send spans to a locally running [Splunk OpenTelemetry Connector](https://github.com/signalfx/splunk-otel-collector)
  (`http://localhost:14268/api/traces`).
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

### Basic Configuration

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

## Library Instrumentation

Supported libraries are listed
[here](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation).

Additional recommended Splunk specific instrumentations:

- [`splunksql`](./instrumentation/database/sql/splunksql)
- [`splunkgorm`](./instrumentation/github.com/jinzhu/gorm/splunkgorm)
- [`splunkhttp`](./instrumentation/net/http/splunkhttp)
- [`splunkmysql`](./instrumentation/github.com/go-sql-driver/mysql/splunkmysql)
- [`splunkpgx`](./instrumentation/github.com/jackc/pgx/splunkpgx)
- [`splunkpq`](./instrumentation/github.com/lib/pq/splunkpq)

## Manual Instrumentation

Documentation on how to manually instrument a Go application is available
[here](https://opentelemetry.io/docs/go/getting-started/).

## Advanced Configuration

Below you will find all the configuration options supported by this distribution.

### Splunk Distribution Configuration

<!-- markdownlint-disable MD013 -->
| Environment variable      | Option             | Default value  | Description |
| ------------------------- | -------------------| -------------- | ----------- |
| `SPLUNK_ACCESS_TOKEN`     | [`WithAccessToken`](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#WithAccessToken)  |                | The [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html). [[1](#cfg1)] |
| `OTEL_RESOURCE_ATTRIBUTES` |                    |                | Comma-separated list of resource attributes added to every reported span. |
<!-- markdownlint-enable MD013 -->

[<a name="cfg1">1</a>]: The [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html)
allows exporters sending data directly to the [Splunk Observability Cloud](https://dev.splunk.com/observability/docs/apibasics/api_list/).
To do so, the `OTEL_EXPORTER_JAEGER_ENDPOINT` must be set
or `distro.WithEndpoint` must be passed to `distro.Run`
with Splunk back-end ingest endpoint URL: `https://ingest.<REALM>.signalfx.com/v2/trace`.

### Trace Configuration

<!-- markdownlint-disable MD013 -->
| Environment variable       | Option             | Default value  | Description |
| -------------------------- | -------------------| -------------- | ----------- |
| `OTEL_RESOURCE_ATTRIBUTES` |                    |                | Comma-separated list of resource attributes added to every reported span. |
<!-- markdownlint-enable MD013 -->

### Trace Exporter Configuration

<!-- markdownlint-disable MD013 -->
| Environment variable            | Option             | Default value  | Description |
| ------------------------------- | -------------------| -------------- | ----------- |
| `OTEL_EXPORTER_JAEGER_ENDPOINT` | [`WithEndpoint`](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#WithEndpoint)     | `http://localhost:14268/api/traces` | Jaeger Thrift HTTP endpoint for sending spans. |
| `OTEL_EXPORTER_JAEGER_USER`     |                    |                | Username to be used for HTTP basic authentication. |
| `OTEL_EXPORTER_JAEGER_PASSWORD` |                    |                | Password to be used for HTTP basic authentication. |
<!-- markdownlint-enable MD013 -->

### Trace Propagation Configuration

The trace propagtor can be changed by using
[`otel.SetTextMapPropagator`](https://pkg.go.dev/go.opentelemetry.io/otel#SetTextMapPropagator)
after `distro.Run()` is invoked e.g.:

```go
distro.Run()
otel.SetTextMapPropagator(propagation.TraceContext{})
```

## License

The Splunk distribution of OpenTelemetry Go is a
distribution of the [OpenTelemetry Go
project](https://github.com/open-telemetry/opentelemetry-go). It is
released under the terms of the Apache Software License version 2.0. See [the
license file](./LICENSE) for more details.
