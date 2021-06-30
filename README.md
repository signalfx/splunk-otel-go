# Splunk distribution of OpenTelemetry Go

[![GitHub Release](https://img.shields.io/github/v/release/signalfx/splunk-otel-go)](https://github.com/signalfx/splunk-otel-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go)
[![go.mod](https://img.shields.io/github/go-mod/go-version/signalfx/splunk-otel-go)](go.mod)
[![LICENSE](https://img.shields.io/github/license/signalfx/splunk-otel-go)](LICENSE)
[![Build Status](https://img.shields.io/github/workflow/status/signalfx/splunk-otel-go/test)](https://github.com/signalfx/splunk-otel-go/actions?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/signalfx/splunk-otel-go)](https://goreportcard.com/report/github.com/signalfx/splunk-otel-go)

The Splunk distribution of [OpenTelemetry
Go](https://github.com/open-telemetry/opentelemetry-go) provides
multiple installable packages that automatically instruments your Go
application to capture and report distributed traces to Splunk APM.

This Splunk distribution comes with the following defaults:

- [B3 context propagation](https://github.com/openzipkin/b3-propagation).
- [Jaeger Thrift over HTTP
  exporter](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/jaeger)
  configured to send spans to a locally running Splunk OpenTelemetry Connector](https://github.com/signalfx/splunk-otel-collector)
  (`http://localhost:14268/api/traces`).
- Unlimited default limits for [configuration options](#trace-configuration) to
  support full-fidelity traces.

> :construction: This project is currently in **BETA**. It is **officially supported** by Splunk. However, breaking changes **MAY** be introduced.

## Getting Started

### Bootstrapping

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

    /* ... */
```

### Library instrumentation

Supported libraries are listed
[here](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation).

Splunk specific instrumentations:

- [`splunkhttp`](./instrumentation/net/http/splunkhttp)

### Manual instrumentation

Documentation on how to manually instrument a Go application is available
[here](https://opentelemetry.io/docs/go/getting-started/).

## Splunk specific configuration

| Environment variable      | Option             | Default value  | Description |
| ------------------------- | -------------------| -------------- | ---------------------------------------------------------------------- |
| `SPLUNK_ACCESS_TOKEN`     | `WithAccessToken`  |                | The [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html). [[1](#cfg1)] |

[<a name="cfg1">1</a>]: The [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html)
allows exporters sending data directly to the [Splunk Observability Cloud](https://dev.splunk.com/observability/docs/apibasics/api_list/).
To do so, the `OTEL_EXPORTER_JAEGER_ENDPOINT` or `distro.WithEndpoint` must be passed to `distro.Run`
with Splunk back-end ingest endpoint URL: `https://ingest.<REALM>.signalfx.com/v2/trace`.

## License and versioning

The Splunk distribution of OpenTelemetry Go is a
distribution of the [OpenTelemetry Go
project](https://github.com/open-telemetry/opentelemetry-go). It is
released under the terms of the Apache Software License version 2.0. See [the
license file](./LICENSE) for more details.
