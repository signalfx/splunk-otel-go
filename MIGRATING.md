# Migrate from the SignalFx Tracing Library for Go

The [Splunk Distribution of OpenTelemetry for Go] replaces the [SignalFx Tracing
Library for Go].

Use the following instructions to migrate help migrate to the [Splunk
Distribution of OpenTelemetry for Go].

## Requirements

Go version 1.16 or greater is required by the [Splunk Distribution of
OpenTelemetry for Go].

## Migration Steps

The following steps identify all actions needed to migrate from [SignalFx
Tracing Library for Go] to the [Splunk Distribution of OpenTelemetry for Go].

After the migration steps are complete, all tracing telemetry will continue to
be transmitted and you will no longer have any dependency on
`github.com/signalfx/signalfx-go-tracing` packages. Be sure to verify this by
checking your `go.mod` files after they have been tidied.

### Replace [`tracing`] Setup with [`distro`] Setup

The [SignalFx Tracing Library for Go] uses the [`tracing`] package to configure
and start tracing functionality. This is replaced with the [`distro`] package
from the [Splunk Distribution of OpenTelemetry for Go] in the following way.

The [`tracing.Start`] function needs to be replaced with [`distro.Run`]. The
following replacements are used for [`tracing.StartOption`] instances.

| [`tracing.StartOption`] | Replacement |
| --- | --- |
| [`tracing.WithAccessToken`] | [`distro.WithAccessToken`] |
| [`tracing.WithEndpointURL`] | [`distro.WithEndpoint`] |
| [`tracing.WithGlobalTag`] | See [Defining a Resource](#defining-a-resource) |
| [`tracing.WithRecordedValueMaxLength`] | See [Setting Span Limits](#setting-span-limits) |
| [`tracing.WithServiceName`] | See [Defining a Resource](#defining-a-resource) |
| [`tracing.WithoutLibraryTags`] | N/A (see below) |

Note: The [`tracing.WithoutLibraryTags`] [`tracing.StartOption`] does not have
an equivalent in the [Splunk Distribution of OpenTelemetry for Go]. Metadata
about the tracing library is contained in the [`Resource`] associated with the
[`distro.SDK`]. See [Defining a Resource](#defining-a-resource) for more
information on [`Resource`]s.

Unlike the [`tracing`] package, the [`distro.SDK`] needs to shut down when your
application stops. This ensures that all spans are flushed and any held state
is released. Defer a cleanup function in your application `main` function.

```go
sdk, err := distro.Run()
if err != nil {
    panic(err)
}
defer func() {
	// A context with a deadline can be passed here instead if needed.
	if err := sdk.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}()
/* ... */
```

#### Defining a Resource

OpenTelemetry uses a [`Resource`] to describe the common metadata about the
[`distro.SDK`] that applies to all spans it produces. The [`distro.Run`]
function will create a default [`Resource`] containing all needed Splunk and
OpenTelemetry metadata for traces. However, you will need to provided a
information about your service to include in this [`Resource`].

**Importantly** you must set the service name of your service. Not doing so
will result in all trace data being unidentifiable. To do this, set the
`OTEL_SERVICE_NAME` environment variable to the name of your service.

If you would like to included additional attributes to include in the metadata
for all traces produced by the [`distro.SDK`], use the
`OTEL_RESOURCE_ATTRIBUTES` environment variable. For example.

```sh
export OTEL_RESOURCE_ATTRIBUTES="ab-test-value=red,owner=Lisa"
```

#### Setting Span Limits

OpenTelemetry includes guards to prevent erroneous code producing excess
computational resource usage. These [span limits] are set using environment
variables.

| Name | Description | Default |
| --- | --- | --- |
| `OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Maximum allowed attribute value size | 12000 |
| `OTEL_ATTRIBUTE_COUNT_LIMIT` | Maximum allowed span attribute count | unlimited |
| `OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT` | Maximum allowed span attribute count | unlimited |
| `OTEL_SPAN_EVENT_COUNT_LIMIT` | Maximum allowed span event count | unlimited |
| `OTEL_SPAN_LINK_COUNT_LIMIT` | Maximum allowed span link count | 1000 |
| `OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT` | Maximum allowed attribute per span event count | unlimited |
| `OTEL_LINK_ATTRIBUTE_COUNT_LIMIT`| Maximum allowed attribute per span link count | unlimited |

Note: Prior use of [`tracing.WithRecordedValueMaxLength`] should be replaced by
setting `OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` to the same value.

### Replace all Manual Instrumentation

All spans created with the [`tracer`] package need to be recreated with
OpenTelemetry. OpenTelemetry uses `Tracer`s to encapsulate the tracing function
of a single instrumentation library. Create a `Tracer` from the global
`TracerProvider` registered when you started the [`distro.SDK`]. To do this use
the [`otel.Tracer`] function and pass the name of your application.

```go
tracer := otel.Tracer("my-application")
```

Use this created `tracer` and its `Start` function to replace all
[`tracer.StartSpan`] invocations. The `operationName` parameter can be used as
the `name` parameter for `Start`. The following replacements are used for
[`tracer.StartSpanOption`] instances.

| [`tracer.StartSpanOption`] | Replacement |
| --- | --- |
| [`tracer.ChildOf`] | N/A (see below) |
| [`tracer.ResourceName`] | N/A (see below) |
| [`tracer.ServiceName`] | N/A (see below) |
| [`tracer.SpanType`] | [`trace.WithSpanKind`] |
| [`tracer.StartTime`] | [`trace.WithTimestamp`] |
| [`tracer.Tag`] | [`trace.WithAttributes`] |
| [`tracer.WithRecordedValueMaxLength`] | N/A (see below) |
| [`tracer.WithSpanID`] | N/A (see below) |

Notes:

- [`tracer.ChildOf`]: The parent-child relationship of spans is defined with a
  [`context.Context`]. The [`context.Context`] passed to `Start` needs to
  contain the parent span. This is automatically done if the context was
  returned from a previous call to `Start`. It can explicitly be done using the
  [`trace.ContextWithSpan`].
- [`tracer.ResourceName`]: This value is defined global with a [`Resource`]
  that applies to all spans. See [Defining a Resource](#defining-a-resource)
  for more information.
- [`tracer.ServiceName`]: This value is defined global in a [`Resource`] that
  applies to all spans. See [Defining a Resource](#defining-a-resource) for
  more information.
- [`tracer.WithRecordedValueMaxLength`]: This value is set globally. See
  [Setting Span Limits](#setting-span-limits) for more information.
- [`tracer.WithSpanID`]: Span IDs are automatically set. If custom span IDs are
  needed you will need to create a custom [`IDGenerator`].

Finally, the created span, similar to before, needs to be ended. Use the
OpenTelemetry span's `End` method to do this.

### Replace all Instrumentation Libraries

| [SignalFx Tracing Library for Go] | OpenTelemetry |
| --- | --- |
| [`aws/aws-sdk-go/aws`] | [`otelaws`] |
| [`bradfitz/gomemcache/memcache`] | [`otelmemcache`] |
| [`confluentinc/confluent-kafka-go/kafka`] | [`splunkkafka`] |
| [`database/sql`] | [`splunksql`] ([`splunkmysql`], [`splunkpgx`], [`splunkpq`]) |
| [`emicklei/go-restful`] | [`otelrestful`] |
| [`garyburd/redigo`] | N/A (See below) |
| [`gin-gonic/gin`] | [`otelgin`] |
| [`globalsign/mgo`] | N/A (See below) |
| [`go-chi/chi`] | [`splunkchi`] |
| [`go-redis/redis`] | N/A (See below) |
| [`gocql/gocql`] | [`otelgocql`] |
| [`gomodule/redigo`] | [`splunkredigo`] |
| [`google.golang.org/api`] | N/A (See below) |
| [`google.golang.org/grpc.v12`] | N/A (See below) |
| [`google.golang.org/grpc`] | [`otelgrpc`] |
| [`gorilla/mux`] | [`otelmux`] |
| [`graph-gophers/graphql-go`] | [`splunkgraphql`] |
| [`jinzhu/gorm`] | [`splunkgorm`] |
| [`jmoiron/sqlx`] | [`splunksqlx`] |
| [`julienschmidt/httprouter`] | [`splunkhttprouter`] |
| [`k8s.io/client-go/kubernetes`] | [`splunkclient-go`] |
| [`labstack/echo.v4`] | [`otelecho`] |
| [`labstack/echo`] | N/A (See below) |
| [`miekg/dns`] | [`splunkdns`]
| [`mongodb/mongo-go-driver/mongo`] | [`otelmongo`] |
| [`net/http`] | [`splunkhttp`], [`otelhttp`] |
| [`olivere/elastic`] | [`splunkelastic`] |
| [`Shopify/sarama`] | [`otelsarama`] |
| [`syndtr/goleveldb/leveldb`] | [`splunkleveldb`] |
| [`tidwall/buntdb`] | [`splunkbuntdb`] |

Note:

- [`garyburd/redigo`]: This project is archived. Use `gomodule/redigo` and
  [`splunkredigo`] instead.
- [`globalsign/mgo`]: This project is an unsported fork of an abandoned
  project. Use `mongodb/mongo-go-driver` and `otelmongo` instead.
- [`go-redis/redis`]: This package now provides native support for
  OpenTelemetry. See [this
  example](https://github.com/go-redis/redis/tree/master/example/otel) for more
  information.
- [`google.golang.org/api`]: Use either [`otelgrpc`] or [`otelhttp`] with a
  gRPC or HTTP client when calling [`cloudresourcemanager.NewService`].
- [`google.golang.org/grpc.v12`]: This version is no longer supported. Use the
  latest version along with [`otelgrpc`].
- [`labstack/echo`]: Versions prior to v4 are no longer supported. Upgrade to
  `echo@v4` and use `otelecho`.

## Troubleshooting

TODO

[SignalFx Tracing Library for Go]: https://github.com/signalfx/signalfx-go-tracing
[Splunk Distribution of OpenTelemetry for Go]: https://github.com/signalfx/splunk-otel-go
[`tracing`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing
[`distro`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro
[`tracing.Start`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#Start
[`distro.Run`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#Run
[`tracing.StartOption`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#StartOption
[`tracing.WithAccessToken`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#WithAccessToken
[`tracing.WithEndpointURL`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#WithEndpointURL
[`tracing.WithGlobalTag`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#WithGlobalTag
[`tracing.WithRecordedValueMaxLength`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#WithRecordedValueMaxLength
[`tracing.WithServiceName`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#WithServiceName
[`tracing.WithoutLibraryTags`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/tracing#WithoutLibraryTags
[`distro.WithAccessToken`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#WithAccessToken
[`distro.WithEndpoint`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#WithEndpoint
[`Resource`]: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/resource#Resource
[`distro.SDK`]: https://pkg.go.dev/github.com/signalfx/splunk-otel-go/distro#SDK
[span limits]: https://github.com/open-telemetry/opentelemetry-specification/blob/v1.9.0/specification/sdk-environment-variables.md#span-limits-
[`tracer`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer
[`otel.Tracer`]: https://pkg.go.dev/go.opentelemetry.io/otel#Tracer
[`tracer.StartSpan`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#StartSpan
[`tracer.StartSpanOption`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#StartSpanOption
[`tracer.ChildOf`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#ChildOf
[`tracer.ResourceName`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#ResourceName
[`tracer.ServiceName`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#ServiceName
[`tracer.SpanType`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#SpanType
[`tracer.StartTime`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#StartTime
[`tracer.Tag`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#Tag
[`tracer.WithRecordedValueMaxLength`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#WithRecordedValueMaxLength
[`tracer.WithSpanID`]: https://pkg.go.dev/github.com/signalfx/signalfx-go-tracing/ddtrace/tracer#WithSpanID
[`trace.WithSpanKind`]: https://pkg.go.dev/go.opentelemetry.io/otel/trace#WithSpanKind
[`trace.WithAttributes`]: https://pkg.go.dev/go.opentelemetry.io/otel/trace#WithAttributes
[`trace.WithTimestamp`]: https://pkg.go.dev/go.opentelemetry.io/otel/trace#WithTimestamp
[`IDGenerator`]: https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#IDGenerator
[`context.Context`]: https://pkg.go.dev/context#Context
[`trace.ContextWithSpan`]: https://pkg.go.dev/go.opentelemetry.io/otel/trace#ContextWithSpan
[`Shopify/sarama`]: http://github.com/signalfx/signalfx-go-tracing/contrib/Shopify/sarama
[`aws/aws-sdk-go/aws`]: http://github.com/signalfx/signalfx-go-tracing/contrib/aws/aws-sdk-go/aws
[`bradfitz/gomemcache/memcache`]: http://github.com/signalfx/signalfx-go-tracing/contrib/bradfitz/gomemcache/memcache
[`confluentinc/confluent-kafka-go/kafka`]: http://github.com/signalfx/signalfx-go-tracing/contrib/confluentinc/confluent-kafka-go/kafka
[`database/sql`]: http://github.com/signalfx/signalfx-go-tracing/contrib/database/sql
[`emicklei/go-restful`]: http://github.com/signalfx/signalfx-go-tracing/contrib/emicklei/go-restful
[`garyburd/redigo`]: http://github.com/signalfx/signalfx-go-tracing/contrib/garyburd/redigo
[`gin-gonic/gin`]: http://github.com/signalfx/signalfx-go-tracing/contrib/gin-gonic/gin
[`globalsign/mgo`]: http://github.com/signalfx/signalfx-go-tracing/contrib/globalsign/mgo
[`go-chi/chi`]: http://github.com/signalfx/signalfx-go-tracing/contrib/go-chi/chi
[`go-redis/redis`]: http://github.com/signalfx/signalfx-go-tracing/contrib/go-redis/redis
[`gocql/gocql`]: http://github.com/signalfx/signalfx-go-tracing/contrib/gocql/gocql
[`gomodule/redigo`]: http://github.com/signalfx/signalfx-go-tracing/contrib/gomodule/redigo
[`google.golang.org/api`]: http://github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/api
[`google.golang.org/grpc`]: http://github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc
[`google.golang.org/grpc.v12`]: http://github.com/signalfx/signalfx-go-tracing/contrib/google.golang.org/grpc.v12
[`gorilla/mux`]: http://github.com/signalfx/signalfx-go-tracing/contrib/gorilla/mux
[`graph-gophers/graphql-go`]: http://github.com/signalfx/signalfx-go-tracing/contrib/graph-gophers/graphql-go
[`jinzhu/gorm`]: http://github.com/signalfx/signalfx-go-tracing/contrib/jinzhu/gorm
[`jmoiron/sqlx`]: http://github.com/signalfx/signalfx-go-tracing/contrib/jmoiron/sqlx
[`julienschmidt/httprouter`]: http://github.com/signalfx/signalfx-go-tracing/contrib/julienschmidt/httprouter
[`k8s.io/client-go/kubernetes`]: http://github.com/signalfx/signalfx-go-tracing/contrib/k8s.io/client-go/kubernetes
[`labstack/echo`]: http://github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo
[`labstack/echo.v4`]: http://github.com/signalfx/signalfx-go-tracing/contrib/labstack/echo.v4
[`miekg/dns`]: http://github.com/signalfx/signalfx-go-tracing/contrib/miekg/dns
[`mongodb/mongo-go-driver/mongo`]: http://github.com/signalfx/signalfx-go-tracing/contrib/mongodb/mongo-go-driver/mongo
[`net/http`]: http://github.com/signalfx/signalfx-go-tracing/contrib/net/http
[`olivere/elastic`]: http://github.com/signalfx/signalfx-go-tracing/contrib/olivere/elastic
[`syndtr/goleveldb/leveldb`]: http://github.com/signalfx/signalfx-go-tracing/contrib/syndtr/goleveldb/leveldb
[`tidwall/buntdb`]: http://github.com/signalfx/signalfx-go-tracing/contrib/tidwall/buntdb
[`splunksql`]: github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql
[`splunkkafka`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka
[`splunkchi`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi
[`splunkmysql`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql
[`splunkredigo`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo
[`splunkgraphql`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql
[`splunkpgx`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx
[`splunkgorm`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm
[`splunksqlx`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx
[`splunkhttprouter`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter
[`splunkpq`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/lib/pq/splunkpq
[`splunkdns`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns
[`splunkleveldb`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb
[`splunkbuntdb`]: github.com/signalfx/splunk-otel-go/instrumentation/github.com/tidwall/buntdb/splunkbuntdb
[`splunkelastic`]: github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic
[`splunkclient-go`]: github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go
[`splunkhttp`]: github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp
[`otelaws`]: go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws
[`otelmemcache`]: go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache
[`otelrestful`]: go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful
[`otelgin`]: go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
[`otelgocql`]: go.opentelemetry.io/contrib/instrumentation/github.com/gocql/gocql/otelgocql
[`otelmux`]: go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux
[`otelecho`]: go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho
[`otelsarama`]: go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama
[`otelmongo`]: go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo
[`otelgrpc`]: go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc
[`otelhttp`]: go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
[`cloudresourcemanager.NewService`]: https://pkg.go.dev/google.golang.org/api@v0.70.0/cloudresourcemanager/v3#NewService
