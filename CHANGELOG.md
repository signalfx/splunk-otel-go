# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.29.0] - 2025-12-09

This release upgrades [OpenTelemetry Go to v1.39.0/v0.61.0/v0.15.0/v0.0.14][otel-v1.39.0]
and [OpenTelemetry Go Contrib to v1.39.0/v2.1.0/v0.64.0/v0.33.0/v0.19.0/v0.14.0/v0.12.0/v0.11.0][contrib-v1.39.0].

The release requires at least [Go 1.24].

### Changed

- The following environment variable values are now case-insensitive. (#4197)
  - OTEL_EXPORTER_OTLP_PROTOCOL
  - OTEL_EXPORTER_OTLP_TRACES_PROTOCOL
  - OTEL_EXPORTER_OTLP_METRICS_PROTOCOL
  - OTEL_EXPORTER_OTLP_LOGS_PROTOCOL
  - OTEL_TRACES_EXPORTER
  - OTEL_METRICS_EXPORTER
  - OTEL_LOGS_EXPORTER

### Removed

- Drop support for [Go 1.23]. (#4093)

## [1.28.0] - 2025-09-03

This release upgrades [OpenTelemetry Go to v1.38.0/v0.60.0/v0.14.0/v0.0.13][otel-v1.38.0]
and [OpenTelemetry Go Contrib to v1.38.0/v2.0.0/v0.63.0/v0.32.0/v0.18.0/v0.13.0/v0.11.0/v0.10.0][contrib-v1.38.0].

This minor release is the last to support [Go 1.23].
The next minor release will require at least [Go 1.24].

## [1.27.0] - 2025-07-03

This release upgrades [OpenTelemetry Go to 1.37.0/0.59.0/0.13.0][otel-v1.37.0]
and [OpenTelemetry Go Contrib to 1.37.0/0.62.0/0.31.0/0.17.0/0.12.0/0.10.0/0.9.0][contrib-v1.37.0].

### Added

- Support for OTLP protocol configuration via environment variables:
  - `OTEL_EXPORTER_OTLP_PROTOCOL` to specify the protocol used by OTLP exporters.
    - Allowed values:
      - `grpc` - gRPC protocol (default)
      - `http/protobuf` - HTTP with Protobuf encoding
  - `OTEL_EXPORTER_OTLP_TRACES_PROTOCOL` to override the protocol specifically
  for traces export.
  - `OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` to override the protocol specifically
  for metrics export.
  - `OTEL_EXPORTER_OTLP_LOGS_PROTOCOL` to override the protocol specifically
  for logs export.

### Changed

- Because of changes in `go.opentelemetry.io/contrib/instrumentation/runtime`,
  `github.com/signalfx/splunk-otel-go/distro` now produces the new metrics by
  default. Set `OTEL_GO_X_DEPRECATED_RUNTIME_METRICS=true` environment variable
  to additionally produce the deprecated metrics. See
  <https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime> for
  more information. We advise migrating to the new metrics, as the deprecated
  ones will be removed in the future.

## [1.26.0] - 2025-05-23

This release upgrades [OpenTelemetry Go to v1.36.0/v0.58.0][otel-v1.36.0]/[v0.12.2][otel-log-v0.12.2]
and [OpenTelemetry Go Contrib to v1.36.0/v0.61.0/v0.30.0/v0.16.0/v0.11.0/v0.9.0/v0.8.0][contrib-v1.36.0].

### Added

- Add `telemetry.distro.name` resource attribute with the value `splunk-otel-go`.
- Add `telemetry.distro.version` resource attribute with the value of the current
  release version.

### Deprecated

- `splunk.distro.version` resource attribute is deprecated and may be removed
  in a future release.

## [1.25.0] - 2025-03-20

This release adds experimental logs support. Set `OTEL_LOGS_EXPORTER=otlp` to
enable logs support. However, take notice that, as of now, the OpenTelemetry Go
Logs API is not stable.

The release requires at least [Go 1.23].

This release upgrades [OpenTelemetry Go to v1.35.0/v0.57.0/v0.11.0][otel-v1.35.0]
and [OpenTelemetry Go Contrib to v1.35.0/v0.60.0/v0.29.0/v0.15.0/v0.10.0/v0.8.0/v0.7.0][contrib-v1.35.0].

### Added

- Add logs support and `OTEL_LOGS_EXPORTER` environment variable.
  `OTEL_LOGS_EXPORTER` accepts:
  - `none` - logs disabled,
  - `otlp` - OTLP gRPC exporter.
  
  Currently, `OTEL_LOGS_EXPORTER` defaults to `none` as the OpenTelemetry Go
  Logs API and SDK are not stable yet. Set `OTEL_LOGS_EXPORTER=otlp` to enable
  logs support. (#3673)

### Removed

- Drop support for [Go 1.22]. (#3721)

## [1.24.0] - 2025-01-22

This release upgrades [OpenTelemetry Go to v1.34.0/v0.56.0/v0.10.0][otel-v1.34.0]
and [OpenTelemetry Go Contrib to v1.34.0/v0.59.0/v0.28.0/v0.14.0/v0.9.0/v0.7.0/v0.6.0][contrib-v1.34.0].

## [1.23.0] - 2024-12-13

This release upgrades [OpenTelemetry Go to v1.33.0/v0.55.0/v0.9.0/v0.0.12][otel-v1.33.0]
and [OpenTelemetry Go Contrib to v1.33.0/v0.58.0/v0.27.0/v0.13.0/v0.8.0/v0.6.0/v0.5.0][contrib-v1.33.0].

## [1.22.0] - 2024-11-14

This release upgrades [OpenTelemetry Go to v1.32.0/v0.54.0/v0.8.0/v0.0.11][otel-v1.32.0]
and [OpenTelemetry Go Contrib to v1.32.0/v0.57.0/v0.26.0/v0.12.0/v0.7.0/v0.5.0/v0.4.0][contrib-v1.32.0].

## [1.21.0] - 2024-10-16

This release upgrades [OpenTelemetry Go to v1.31.0/v0.53.0/v0.7.0/v0.0.10][otel-v1.31.0]
and [OpenTelemetry Go Contrib to v1.31.0/v0.56.0/v0.25.0/v0.11.0/v0.6.0/v0.4.0/v0.3.0][contrib-v1.31.0].

## [1.20.0] - 2024-09-13

This release upgrades [OpenTelemetry Go to v1.30.0/v0.52.0/v0.6.0/v0.0.9][otel-v1.30.0]
and [OpenTelemetry Go Contrib to v1.30.0/v0.55.0/v0.24.0/v0.10.0/v0.5.0/v0.3.0/v0.2.0][contrib-v1.30.0].
The release requires at least [Go 1.22].

### Removed

- Drop support for [Go 1.21]. (#3422)

## [1.19.0] - 2024-09-03

This release upgrades [OpenTelemetry Go to v1.29.0/v0.51.0/v0.5.0][otel-v1.29.0]
and [OpenTelemetry Go Contrib to v1.29.0/v0.54.0/v0.23.0/v0.9.0/v0.4.0/v0.2.0/v0.1.0][contrib-v1.29.0].

### Added

- Add container attributes to resource if available. (#3374)

## [1.18.0] - 2024-07-11

This release upgrades [OpenTelemetry Go to v1.28.0/v0.50.0/v0.4.0][otel-v1.28.0]
and [OpenTelemetry Go Contrib to v1.28.0/v0.53.0/v0.22.0/v0.8.0/v0.3.0/v0.1.0][contrib-v1.28.0].

## [1.17.0] - 2024-05-22

This release upgrades [OpenTelemetry Go to v1.27.0/v0.49.0/v0.3.0][otel-v1.27.0]
and [OpenTelemetry Go Contrib to v1.27.0/v0.52.0/v0.21.0/v0.7.0/v0.2.0][contrib-v1.27.0].

## [1.16.0] - 2024-04-26

This release upgrades [OpenTelemetry Go to v1.26.0/v0.48.0/v0.2.0-alpha][otel-v1.26.0]
and [OpenTelemetry Go Contrib to v1.26.0/v0.51.0/v0.20.0/v0.6.0/v0.1.0][contrib-v1.26.0].

## [1.15.0] - 2024-04-11

This release upgrades [OpenTelemetry Go to v1.25.0/v0.47.0/v0.0.8/v0.1.0-alpha][otel-v1.25.0]
and [OpenTelemetry Go Contrib to v1.25.0/v0.50.0/v0.19.0/v0.5.0/v0.0.1][contrib-v1.25.0].

### Removed

- Drop support for [Go 1.20]. (#3002)

## [1.14.0] - 2024-02-26

This release upgrades [OpenTelemetry Go to v1.24.0/v0.46.0][otel-v1.24.0]
and [OpenTelemetry Go Contrib to v1.24.0/v0.49.0/v0.18.0/v0.4.0][contrib-v1.24.0].
This minor release is the last to support [Go 1.20].
The next minor release will require at least [Go 1.21].

### Fixed

- Drop errant build tags in `github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go`,
  allowing import with Go 1.20+. (#2891)

## [1.13.0] - 2024-02-09

This release upgrades [OpenTelemetry Go to v1.23.1/v0.45.2][otel-v1.23.1]
and [OpenTelemetry Go Contrib to v1.23.0/v0.48.0/v0.17.0/v0.3.0][contrib-v1.23.0].

### Fixed

- Allow bumping OpenTelemetry Go (`go.opentelemetry.io/otel`)
  without bumping the Splunk Distribution (`github.com/signalfx/splunk-otel-go`).
  It fixes a merge resource runtime error, which could occur when
  the application uses a version of OpenTelemetry Go that is newer
  than the one which the Splunk Distribution is depending on. (#2759)

## [1.12.0] - 2024-01-18

This release deprecates `jaeger-thrift-splunk` option support for `OTEL_TRACES_EXPORTER`
environment variable.

This release upgrades [OpenTelemetry Go to v1.22.0/v0.45.0][otel-v1.22.0]
and [OpenTelemetry Go Contrib to 1.22.0/0.47.0/0.16.0/0.2.0][contrib-v1.22.0].

### Deprecated

- `jaeger-thrift-splunk` trace exporter is deprecated and may be removed
  in a future release. Use the default OTLP exporter instead,
  or set the `SPLUNK_REALM` and `SPLUNK_ACCESS_TOKEN` environment variables
  to send telemetry directly to Splunk Observability Cloud. (#2690)

## [1.11.0] - 2023-11-16

The release adds support for sending metrics directly to Splunk Observability Cloud.

This release upgrades [OpenTelemetry Go to v1.21.0/v0.44.0][otel-v1.21.0]
and [OpenTelemetry Go Contrib to v1.21.1/v0.46.1/v0.15.1/v0.1.1][contrib-v1.21.1].

### Add

- Add the `WithIDGenerator` option to
  `github.com/signalfx/splunk-otel-go/distro`. (#2634)
- Metrics are sent directly to Splunk Observability Cloud when `SPLUNK_REALM` is
  set. (#2637)

## [1.10.0] - 2023-11-10

This release upgrades [OpenTelemetry Go to v1.20.0/v0.43.0][otel-v1.20.0]
and [OpenTelemetry Go Contrib to v1.21.0/v0.46.0/v0.15.0/v0.1.0][contrib-v1.21.0].

## [1.9.0] - 2023-09-29

This release upgrades [OpenTelemetry Go to v1.19.0/v0.42.0/v0.0.7][otel-v1.19.0]
and [OpenTelemetry Go Contrib to v1.20.0/v0.45.0/v0.14.0][contrib-v1.20.0].

## [1.8.1] - 2023-09-15

This release is built on top of [OpenTelemetry Go to v1.18.0/v0.41.0/v0.0.6][otel-v1.18.0]
and [OpenTelemetry Go Contrib to v1.19.0/v0.44.0/v0.13.0][contrib-v1.19.0].

### Removed

- Drop support for Go 1.19 as `go.opentelemetry.io/otel` did the same in
  [v1.18.0][otel-v1.18.0]. (#2492)

## [1.8.0] - 2023-09-14

This upgrades [OpenTelemetry Go to v1.18.0/v0.41.0/v0.0.6][otel-v1.18.0] and
[OpenTelemetry Go Contrib to v1.19.0/v0.44.0/v0.13.0][contrib-v1.19.0].

### Added

- Add the `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/v5/splunkpgx`
  instrumentation for the `github.com/jackc/pgx/v5` package. (#2406)

## [1.7.0] - 2023-07-17

This release is built on top of [OpenTelemetry Go v1.16.0/v0.39.0][otel-v1.16.0]
and [OpenTelemetry Go Contrib v1.17.0/v0.42.0/v0.11.0][contrib-v1.17.0].

### Added

- Add the `github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka`
  instrumentation for the `github.com/confluentinc/confluent-kafka-go/v2/kafka`
  package. (#2301)

## [1.6.0] - 2023-05-25

The release enables metrics support by default
as OpenTelemetry Go metrics API is stable
([`v1.16.0`][otel-v1.16.0]).

This upgrades [OpenTelemetry Go to v1.16.0/v0.39.0][otel-v1.16.0] and
[OpenTelemetry Go Contrib to v1.17.0/v0.42.0/v0.11.0][contrib-v1.17.0].

### Added

- Add `Version` function to the following Go modules. (#1992)
  - `github.com/signalfx/splunk-otel-go/distro`
  - `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb`
  - `github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic`
  - `github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go`
- Add `WithMeterProvider` function in `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`.
  (#2258)

### Changed

- `OTEL_METRICS_EXPORTER` defaults to `otlp`.
  Therefore, metrics support are enabled by default. (#2259)

### Fixed

- Fix `telemetry.sdk.version` resource attribute
  to properly return the `github.com/signalfx/splunk-otel-go/distro` Go module version.
  (#1990)
- Fix instrumentation scope version value
  to properly return the instrumentation library versions. (#1990)

### Removed

- Drop support for Go 1.18 as `go.opentelemetry.io/otel` did the same in
  [`v1.15.0`][otel-v1.15.0]. (#2095)

## [1.5.0] - 2023-03-30

### Added

- Add metrics in the following database instrumentation libraries. (#1973)
  - `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq`

### Changed

- Update attributes in
  `github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka`
  to align `go.opentelemetry.io/otel/semconv` standard in `v1.17.0`. (#1839)
  - Add `semconv.MessagingOperationPublish` in producer span
  - Change to `semconv.MessagingSourceKindTopic` in consumer span

## [1.4.0] - 2023-03-02

The release adds metrics support.
Set `OTEL_METRICS_EXPORTER=otlp` to enable metrics support.
However, take notice that as for now
the OpenTelemetry Go metrics API is not stable
([`v0.37.0`][otel-v1.14.0]).

This upgrades [OpenTelemetry Go to v1.14.0/v0.37.0/v0.0.4][otel-v1.14.0] and
[OpenTelemetry Go Contrib to v1.15.0/v0.40.0/v0.9.0][contrib-v1.15.0].

### Added

- Add metrics support and `OTEL_METRICS_EXPORTER` environment variable.
  `OTEL_METRICS_EXPORTER` accepts:
  `none` - metrics disabled,
  `otlp` - OTLP gRPC exporter.
  Currently `OTEL_METRICS_EXPORTER` defaults to `none`
  as OpenTelemetry Go metrics API and SDK are not stable yet.
  Set `OTEL_METRICS_EXPORTER=otlp` to enable
  metrics support.
- Add process and Go runtime attributes to resource.
- Add runtime metrics instrumentation.

## [1.3.1] - 2023-02-08

This upgrades [OpenTelemetry Go to v1.13.0/v0.36.0][otel-v1.13.0] and
[OpenTelemetry Go Contrib to v1.14.0/v0.39.0/v0.8.0][contrib-v1.14.0].

## [1.3.0] - 2023-02-01

This upgrades [OpenTelemetry Go to v1.12.0/v0.35.0][otel-v1.12.0] and
[OpenTelemetry Go Contrib to v1.13.0/v0.38.0/v0.7.0][contrib-v1.13.0].

### Fixed

- The goroutine created by the `Open` function in
  `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`
  is no longer orphaned. (#1682)

### Added

- The `NetSockFamily` type and related variables to be use in the
  `ConnectionConfig` from
  `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`.
  (#1749)

### Changed

- Add the `NetSockFamily` field to the `ConnectionConfig` in
  `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`.
  This is used to define the protocol address family used for communication with
  the database. (#1749)
- Update `go.opentelemetry.io/otel/semconv` to `v1.17.0` in the following
  packages. (#1749)
  - `github.com/signalfx/splunk-otel-go/distro`
  - `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb`
  - `github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb`
  - `github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic`
  - `github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go`

### Deprecated

- The `NetTransportIP` and `NetTransportUnix` variables from
  `github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql`
  are deprecated as they are no longer available in `go.opentelemetry.io/otel/semconv/v1.17.0`.
  Use an appropriate `NetSockFamily*` variable instead. (#1749)

## [1.2.0] - 2023-01-11

This upgrades [OpenTelemetry Go to v1.11.2/v0.34.0][otel-v1.11.2] and
[OpenTelemetry Go Contrib to v1.12.0/v0.37.0/v0.6.0][contrib-v1.12.0].

### Added

- `OTEL_LOG_LEVEL` environment variable accepts case insensitive values. (#1374)

### Removed

- Drop support for Go 1.17 as `go.opentelemetry.io/otel` did the same in
  [`v1.11.0`][otel-v1.11.0]. (#1570)

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

- Drop support for Go 1.16 as `go.opentelemetry.io/otel` did the same in
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
to [`v1.3.0`/`v0.28.0`][contrib-v1.3.0].

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

- Drop support for Go 1.14 as [`go.opentelemetry.io/otel`][otel-v1.0.0-RC1]
  did the same.

## [0.3.0] - 2021-05-18

The primary changes of this release is added support
for sending data directly to Splunk Observability Cloud.

### Added

- Add support for setting the [Splunk's organization access token](https://help.splunk.com/en/splunk-observability-cloud/administer/authentication-and-security/authentication-tokens/org-access-tokens)
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

[Unreleased]: https://github.com/signalfx/splunk-otel-go/compare/v1.29.0...HEAD
[1.29.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.29.0
[1.28.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.28.0
[1.27.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.27.0
[1.26.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.26.0
[1.25.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.25.0
[1.24.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.24.0
[1.23.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.23.0
[1.22.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.22.0
[1.21.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.21.0
[1.20.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.20.0
[1.19.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.19.0
[1.18.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.18.0
[1.17.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.17.0
[1.16.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.16.0
[1.15.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.15.0
[1.14.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.14.0
[1.13.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.13.0
[1.12.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.12.0
[1.11.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.11.0
[1.10.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.10.0
[1.9.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.9.0
[1.8.1]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.8.1
[1.8.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.8.0
[1.7.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.7.0
[1.6.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.6.0
[1.5.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.5.0
[1.4.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.4.0
[1.3.1]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.3.1
[1.3.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.3.0
[1.2.0]: https://github.com/signalfx/splunk-otel-go/releases/tag/v1.2.0
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

[otel-v1.39.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.39.0
[otel-v1.38.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.38.0
[otel-v1.37.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.37.0
[otel-log-v0.12.2]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/log/v0.12.2
[otel-v1.36.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.36.0
[otel-v1.35.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.35.0
[otel-v1.34.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.34.0
[otel-v1.33.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.33.0
[otel-v1.32.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.32.0
[otel-v1.31.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.31.0
[otel-v1.30.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.30.0
[otel-v1.29.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.29.0
[otel-v1.28.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.28.0
[otel-v1.27.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.27.0
[otel-v1.26.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.26.0
[otel-v1.25.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.25.0
[otel-v1.24.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.24.0
[otel-v1.23.1]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.23.1
[otel-v1.22.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.22.0
[otel-v1.21.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.21.0
[otel-v1.20.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.20.0
[otel-v1.19.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.19.0
[otel-v1.18.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.18.0
[otel-v1.16.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.16.0
[otel-v1.15.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.15.0
[otel-v1.14.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.14.0
[otel-v1.13.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.13.0
[otel-v1.12.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.12.0
[otel-v1.11.2]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.11.2
[otel-v1.11.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.11.0
[otel-v1.8.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.8.0
[otel-v1.7.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.7.0
[otel-v1.6.1]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.6.1
[otel-v1.3.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.3.0
[otel-v1.0.0-RC3]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC3
[otel-v1.0.0-RC2]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC2
[otel-v1.0.0-RC1]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v1.0.0-RC1
[otel-v0.20.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v0.20.0
[otel-v0.19.0]: https://github.com/open-telemetry/opentelemetry-go/releases/tag/v0.19.0

[contrib-v1.39.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.39.0
[contrib-v1.38.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.38.0
[contrib-v1.37.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.37.0
[contrib-v1.36.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.36.0
[contrib-v1.35.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.35.0
[contrib-v1.34.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.34.0
[contrib-v1.33.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.33.0
[contrib-v1.32.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.32.0
[contrib-v1.31.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.31.0
[contrib-v1.30.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.30.0
[contrib-v1.29.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.29.0
[contrib-v1.28.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.28.0
[contrib-v1.27.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.27.0
[contrib-v1.26.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.26.0
[contrib-v1.25.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.25.0
[contrib-v1.24.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.24.0
[contrib-v1.23.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.23.0
[contrib-v1.22.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.22.0
[contrib-v1.21.1]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.21.1
[contrib-v1.21.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.21.0
[contrib-v1.20.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.20.0
[contrib-v1.19.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.19.0
[contrib-v1.17.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.17.0
[contrib-v1.15.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.15.0
[contrib-v1.14.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.14.0
[contrib-v1.13.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.13.0
[contrib-v1.12.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.12.0
[contrib-v1.8.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.8.0
[contrib-v1.7.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.7.0
[contrib-v1.6.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.6.0
[contrib-v1.3.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v1.3.0
[contrib-v0.23.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.23.0
[contrib-v0.22.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.22.0
[contrib-v0.21.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.21.0
[contrib-v0.20.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.20.0
[contrib-v0.19.0]: https://github.com/open-telemetry/opentelemetry-go-contrib/releases/tag/v0.19.0

[Go 1.24]: https://go.dev/doc/go1.24
[Go 1.23]: https://go.dev/doc/go1.23
[Go 1.22]: https://go.dev/doc/go1.22
[Go 1.21]: https://go.dev/doc/go1.21
[Go 1.20]: https://go.dev/doc/go1.20
