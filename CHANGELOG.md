# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2022-07-14

This release uses [OpenTelemetry Go v1.8.0][otel-v1.8.0] and
[OpenTelemetry Go Contrib v1.8.0/v0.33.0][contrib-v1.8.0].

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v1.7.0`][otel-v1.7.0]
  to [`v1.8.0`][otel-v1.8.0]. (#1169)
- Update `go.opentelemetry.io/contrib*` dependencies from
  [`v1.7.0`/`v0.32.0`][contrib-v1.7.0] to [`v1.8.0`/`v0.33.0`][contrib-v1.8.0].
  (#1169)

### Removed

- Drop support for Go 1.16 as `go.opentelemetry.io` did the same in
  [`v1.7.0`][otel-v1.7.0]. (#1172)

## [1.0.0] - 2022-06-09

This is the first stable release of the Splunk Distribution of OpenTelemetry Go
which is compliant with [Splunk's GDI Specification v1.3.0](https://github.com/signalfx/gdi-specification/tree/v1.3.0).

Please note that although the distribution is marked as stable,
some of its dependent components (e.g. `otelhttp` instrumentation library)
are still experimental.

This release uses [OpenTelemetry Go v1.7.0][otel-v1.7.0] and
[OpenTelemetry Go Contrib v1.7.0/v0.32.0][contrib-v1.7.0].

## [0.9.0] - 2022-05-26

This release contains configuration fixes and simplifies the API before
a stable release is published.

`go.opentelemetry.io/otel*` dependencies are updated to [`v1.7.0`][otel-v1.7.0]
and `go.opentelemetry.io/contrib*` dependencies are updated to [`v1.7.0`/`v0.32.0`][contrib-v1.7.0].

### Changed

- The `NewTracer` function from
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/graphql-gophers/graphql-go/splunkgraphql`
  now returns a `tracer.Tracer` instead of the deprecated `trace.Tracer` from
  `github.com/graph-gophers/graphql-go`. (#855)
- The `TraceQuery` method of the `Tracer` from
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/graphql-gophers/graphql-go/splunkgraphql`
  now returns a `tracer.QueryFinishFunc` instead of the deprecated
  `trace.TraceQueryFinishFunc` from `github.com/graph-gophers/graphql-go`.
  (#855)
- The `TraceField` method of the `Tracer` from
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/graphql-gophers/graphql-go/splunkgraphql`
  now returns a `tracer.FieldFinishFunc` instead of the deprecated
  `trace.TraceFieldFinishFunc` from `github.com/graph-gophers/graphql-go`.
  (#855)
- The `TraceValidation` method of the `Tracer` from
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/graphql-gophers/graphql-go/splunkgraphql`
  now returns a `tracer.ValidationFinishFunc` instead of the deprecated
  `trace.TraceValidationFinishFunc` from `github.com/graph-gophers/graphql-go`.
  (#855)
- Configure TLS using the system CA for OTLP gRPC exporter connections when
  configured to connect to external endpoints. (#792)
- Remove `opts ...Option` parameter from `NewHandler` function
  from `github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp`
  package. (#947)
- Update `go.opentelemetry.io/otel*` dependencies from [`v1.6.1`][otel-v1.6.1]
  to [`v1.7.0`][otel-v1.7.0]. (#926)
- Update `go.opentelemetry.io/contrib*` dependencies from
  [`v1.6.0`/`v0.31.0`][contrib-v1.6.0] to [`v1.7.0`/`v0.32.0`][contrib-v1.7.0].
  (#926)
  
### Removed

- Minimize `github.com/signalfx/splunk-otel-go/distro` package to
  contain only necessary option functions. (#941)
  - Remove `WithAccessToken` function,
    use `SPLUNK_ACCESS_TOKEN` environment variable instead.
  - Remove `WithEndpoint` function,
    use one of the
    `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`, `OTEL_EXPORTER_JAEGER_ENDPOINT`
    environment variables instead.
  - Remove `WithPropagator` function,
    use `OTEL_PROPAGATORS` environment variable instead.
  - Remove `WithTraceExporter` function,
    use `OTEL_TRACES_EXPORTER` environment variable instead.
- Minimize `github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp`
  package to contain only necessary functions and types. (#947)
  - Remove `WithTraceResponseHeader` function,
    use `SPLUNK_TRACE_RESPONSE_HEADER_ENABLED` environment variable instead.
  - Remove `TraceResponseHeaderMiddleware` function,
    use `NewHandler` function instead.
  - Remove `Option` type.

### Fixed

- Use the correct Splunk Observability Cloud OTLP over gRPC endpoint
  when `SPLUNK_REALM` is set. (#791)

## [0.8.0] - 2022-04-05

### Added

- Add the `WithPropagator` option to
  `github.com/signalfx/splunk-otel-go/distro` along with parsing of the
  `OTEL_PROPAGATORS` environment variable to set the global OpenTelemetry
  `TextMapPropagator`. (#295)
- Add the `WithTraceExporter` and `WithTLSConfig` options to
  `github.com/signalfx/splunk-otel-go/distro` along with parsing of the
  `OTEL_TRACES_EXPORTER` environment variable to set the global OpenTelemetry
  `SpanExporter` used by the `SDK` to export traces. (#300)
- Add the `splunk.distro.version` attribute to the SDK resource. (#332)
- Add the `WithLogger` option to `github.com/signalfx/splunk-otel-go/distro`
  along with parsing of the `OTEL_LOG_LEVEL` environment variable to set the
  logging level of the default logger used. (#336)
- The `github.com/signalfx/splunk-otel-go/distro` package from
  `github.com/signalfx/splunk-otel-go` has been made into its own module.
  (#492)
- The `SPLUNK_REALM` environment variable is now supported. If set, the
  exporter will use the corresponding Splunk ingest endpoint. (#725)

### Changed

- The `SDK` from `github.com/signalfx/splunk-otel-go/distro` now uses an OTLP
  exporter by default. The previous default Jaeger
  thrift exporter can still be used by setting the `OTEL_TRACES_EXPORTER`
  environment variable to `jaeger-thrift-splunk`, or by directly passing the
  user configured exporter with a `WithTraceExporter` option. (#300)
- Update `go.opentelemetry.io/otel*` dependencies from [`v1.3.0`][otel-v1.3.0]
  to [`v1.6.1`][otel-v1.6.1]. (#406, #471, #656, #721)
- Update `go.opentelemetry.io/contrib*` dependencies from
  [`v1.3.0`/`v0.28.0`][contrib-v1.3.0] to [`v1.6.0`/`v0.31.0`][contrib-v1.6.0].
  (#406, #721)
- The `OTEL_TRACES_SAMPLER` environment variable is now honored instead of only
  defaulting to an always-on sampler. (#724)
- Set span limits to the Splunk defaults (the link count is limited to 1000,
  the attribute value length is limited to 12000, and all other limits are set
  to be unlimited) if they are not set by the user with environment variables.
  (#723)

### Fixed

- Consistently import `github.com/jackc/pgx/v4`, instead of
  `github.com/jackc/pgx`, in the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx`
  instrumentation. (#478)

## [0.7.0] - 2022-01-13

### Added

- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`
  instrumentation for the `database/sql` package. (#88)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql`
  instrumentation for the `github.com/go-sql-driver/mysql` package. (#90)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx`
  instrumentation for the `github.com/jackc/pgx` package. (#91)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq`
  instrumentation for the `github.com/lib/pq` package. (#92)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx`
  instrumentation for the `github.com/jmoiron/sqlx` package. (#93)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm`
  instrumentation for the `github.com/jinzhu/gorm` package. (#98)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka`
  instrumentation for the `github.com/confluentinc/confluent-kafka-go/kafka`
  package. (#100)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns`
  instrumentation for the `github.com/miekg/dns`
  package. (#155)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb`
  instrumentation for the `github.com/syndtr/goleveldb/leveldb`
  package. (#186)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb`
  instrumentation for the `github.com/tidwall/buntdb`
  package. (#228)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go`
  instrumentation for the `k8s.io/client-go` package. (#224)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi`
  instrumentation for the `github.com/go-chi/chi` package. (#227)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/graphql-gophers/graphql-go/splunkgraphql`
  instrumentation for the `github.com/graph-gophers/graphql-go` module. (#232)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter`
  instrumentation for the `github.com/julienschmidt/httprouter` package. (#273)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo`
  instrumentation for the `github.com/gomodule/redigo` package. (#288)
- Add the
  `github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic`
  instrumentation for the `gopkg.in/olivere/elastic` package. (#311)

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v1.0.0-RC3`][otel-v1.0.0-RC3]
  to [`v1.3.0`][otel-v1.3.0].
- Update `go.opentelemetry.io/contrib*` dependencies from [`v0.23.0`][contrib-v0.23.0]
to [`v0.28.0`][contrib-v0.28.0].

## [0.6.0] - 2021-09-13

The primary change of this release is updating `go.opentelemetry.io/otel*`
dependencies to [`v1.0.0-RC3`][otel-v1.0.0-RC3] and
`go.opentelemetry.io/contrib*` dependencies to [`v0.23.0`][contrib-v0.23.0].

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v1.0.0-RC2`][otel-v1.0.0-RC2]
  to [`v1.0.0-RC3`][otel-v1.0.0-RC3].
- Update `go.opentelemetry.io/contrib*` dependencies from [`v0.22.0`][contrib-v0.22.0]
to [`v0.23.0`][contrib-v0.23.0].

## [0.5.0] - 2021-07-27

The primary change of this release is updating `go.opentelemetry.io/otel*`
dependencies to [`v1.0.0-RC2`][otel-v1.0.0-RC2] and
`go.opentelemetry.io/contrib*` dependencies to [`v0.22.0`][contrib-v0.22.0].

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v1.0.0-RC1`][otel-v1.0.0-RC1]
  to [`v1.0.0-RC2`][otel-v1.0.0-RC2].
- Update `go.opentelemetry.io/contrib*` dependencies from [`v0.21.0`][contrib-v0.21.0]
  to [`v0.22.0`][contrib-v0.22.0].

## [0.4.0] - 2021-06-30

The primary change of this release is updating the dependency of `go.opentelemetry.io/otel*`
packages from [`v0.20.0`][otel-v0.20.0] to [`v1.0.0-RC1`][otel-v1.0.0-RC1] and
`go.opentelemetry.io/contrib*` packages from [`v0.20.0`][contrib-v0.20.0] to [`v0.21.0`][contrib-v0.21.0].

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v0.20.0`][otel-v0.20.0]
  to [`v1.0.0-RC1`][otel-v1.0.0-RC1].
- Update `go.opentelemetry.io/contrib*` dependencies from [`v0.20.0`][contrib-v0.20.0]
  to [`v0.21.0`][contrib-v0.21.0].

### Remove

- Drop support for Go 1.14 as [`go.opentelemetry.io@v1.0.0-RC1`](https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC1)
  did the same.

## [0.3.0] - 2021-05-18

The primary changes of this release is added support
for sending data directly to Splunk Observability Cloud.

### Added

- Add support for setting the [Splunk's organization access token](https://docs.splunk.com/observability/admin/authentication-tokens/org-tokens.html)
  using the `SPLUNK_ACCESS_TOKEN` environmental variable or `distro.WithAccessToken`
  option. It allows exporters sending data directly to the Splunk Observability Cloud.
  To do so, the `OTEL_EXPORTER_JAEGER_ENDPOINT` or `distro.WithEndpoint` must be
  set with Splunk back-end ingest endpoint URL: `https://ingest.<REALM>.signalfx.com/v2/trace`.

### Changed

- The default Jaeger exporter URL is now loaded by
  [`go.opentelemetry.io/otel/exporters/trace/jaeger`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/trace/jaeger)
  module.
- Applying `distro.WithEndpoint("")` results in no operation.

### Removed

- Remove `SIGNALFX_ENDPOINT_URL` environmental variable, use `OTEL_EXPORTER_JAEGER_ENDPOINT`
  instead.

## [0.2.0] - 2021-04-27

The primary change of this release is updating the dependency of `go.opentelemetry.io/otel*`
packages from [`v0.19.0`][otel-v0.19.0] to [`v0.20.0`][otel-v0.20.0] and similarly
`go.opentelemetry.io/contrib*` packages from [`v0.19.0`][contrib-v0.19.0] to [`v0.20.0`][contrib-v0.20.0].
This includes [a fix](https://github.com/open-telemetry/opentelemetry-go/pull/1830)
in the Jaeger exporter.
This fix removes the duplicate batching that the exporter implemented.
Now the `BatchSpanProcessor` that `distro` configures by default will not experience
an impedance mismatch with this duplicate batching.

### Changed

- Update `go.opentelemetry.io/otel*` dependencies from [`v0.19.0`][otel-v0.19.0]
  to [`v0.20.0`][otel-v0.20.0].
- Update `go.opentelemetry.io/contrib*` dependencies from [`v0.19.0`][contrib-v0.19.0]
  to [`v0.20.0`][contrib-v0.20.0].

## [0.1.0] - 2021-04-08

### Added

- Add [`distro`](./distro) package providing functionality to quickly setup
  the OpenTelemetry Go implementation with useful Splunk defaults.
- Add [`splunkhttp`](./instrumentation/net/http/splunkhttp) module providing
  additional Splunk specific instrumentation for `net/http`.

[Unreleased]: https://github.com/signalfx/splunk-otel-go/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.1.0
[1.0.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.0.0
[0.9.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.9.0
[0.8.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.8.0
[0.7.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.7.0
[0.6.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.6.0
[0.4.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.4.0
[0.3.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.3.0
[0.2.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.2.0
[0.1.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v0.1.0

[otel-v1.8.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.8.0
[otel-v1.7.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.7.0
[otel-v1.6.1]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.6.1
[otel-v1.6.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.6.0
[otel-v1.3.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.3.0
[otel-v1.0.0-RC3]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC3
[otel-v1.0.0-RC2]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC2
[otel-v1.0.0-RC1]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC1
[otel-v0.20.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v0.20.0
[otel-v0.19.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v0.19.0

[contrib-v1.8.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.8.0
[contrib-v1.7.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.7.0
[contrib-v1.6.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.6.0
[contrib-v1.3.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.3.0
[contrib-v0.28.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.28.0
[contrib-v0.23.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.23.0
[contrib-v0.22.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.22.0
[contrib-v0.21.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.21.0
[contrib-v0.20.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.20.0
[contrib-v0.19.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.19.0
