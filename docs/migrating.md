# Migrate from the SignalFx Tracing Library for Go

> The official Splunk documentation for this page is
[Migrate from the SignalFx Tracing Library for Go](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/troubleshooting/migrate-signalfx-go-to-otel.html).
For instructions on how to contribute to the docs, see
[CONTRIBUTING.md](../CONTRIBUTING.md#documentation).

The [Splunk Distribution of OpenTelemetry Go] replaces the [SignalFx Tracing
Library for Go].

The following steps describe all actions required to migrate from the [SignalFx
Tracing Library for Go] to the [Splunk Distribution of OpenTelemetry Go].

## Requirements

Go version 1.16 or higher.

## Migrate to the Splunk Distribution of OpenTelemetry Go

The following steps identify all actions needed to migrate from [SignalFx
Tracing Library for Go] to the [Splunk Distribution of OpenTelemetry Go].

After the migration is complete, all tracing data will continue to be
transmitted to Splunk Observability Cloud, without any dependency on
`github.com/signalfx/signalfx-go-tracing` packages. Make sure to verify this by
checking your `go.mod` files after cleaning them up.

### Replace [`tracing`] setup with [`distro`] setup

The [SignalFx Tracing Library for Go] uses the [`tracing`] package to configure
and start tracing functionality. This is replaced by the [`distro`] package
from the [Splunk Distribution of OpenTelemetry Go] in the following way.

Replace the [`tracing.Start`] function with [`distro.Run`]. The following
replacements are used for [`tracing.StartOption`] instances.

| [`tracing.StartOption`] | Replacement |
| --- | --- |
| [`tracing.WithAccessToken`] | Use `SPLUNK_ACCESS_TOKEN` environment variable. |
| [`tracing.WithEndpointURL`] | Use `SPLUNK_REALM` or other environment variables to configure the exporter. |
| [`tracing.WithGlobalTag`] | See [Defining a Resource](#defining-a-resource) |
| [`tracing.WithRecordedValueMaxLength`] | See [Setting Span Limits](#setting-span-limits) |
| [`tracing.WithServiceName`] | See [Defining a Resource](#defining-a-resource) |
| [`tracing.WithoutLibraryTags`] | N/A. The [`tracing.WithoutLibraryTags`] option does not have an equivalent in the Splunk Distribution of OpenTelemetry Go. Metadata about the tracing library is available in the [`Resource`] associated with the [`distro.SDK`]. See [Defining a Resource](#defining-a-resource) for more information on resources. |

Unlike the [`tracing`] package, the [`distro.SDK`] must shut down when your
application stops. This ensures that all spans are flushed and any held state
is released. Defer a cleanup function in your application `main` function.

The following example shows how to shut down the SDK:

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

#### Defining a resource

OpenTelemetry uses a [`Resource`] to describe the common metadata about the
[`distro.SDK`] that applies to all spans it produces. The [`distro.Run`]
function creates a default [`Resource`] containing all the required Splunk and
OpenTelemetry metadata for traces. To provided metadata about your service, you
must include it in the [`Resource`].

**Note:** You must set the service name of your service. Not doing so results
in all trace data being unidentifiable. To do this, set the `OTEL_SERVICE_NAME`
environment variable to the name of your service.

To include additional attributes in the metadata for all traces produced by the
[`distro.SDK`], use the `OTEL_RESOURCE_ATTRIBUTES` environment variable.  For
example:

```sh
export OTEL_RESOURCE_ATTRIBUTES="ab-test-value=red,owner=Lisa"
```

#### Setting span limits

OpenTelemetry includes guards to prevent code from producing excessive usage of
computational resources. These [span limits] are set using the following
environment variables:

| Name | Description | Default |
| --- | --- | --- |
| `OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` | Maximum allowed attribute value size | 12000 |
| `OTEL_ATTRIBUTE_COUNT_LIMIT` | Maximum allowed span attribute count | unlimited |
| `OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT` | Maximum allowed span attribute count | unlimited |
| `OTEL_SPAN_EVENT_COUNT_LIMIT` | Maximum allowed span event count | unlimited |
| `OTEL_SPAN_LINK_COUNT_LIMIT` | Maximum allowed span link count | 1000 |
| `OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT` | Maximum allowed attribute per span event count | unlimited |
| `OTEL_LINK_ATTRIBUTE_COUNT_LIMIT`| Maximum allowed attribute per span link count | unlimited |

Replace any instance of [`tracing.WithRecordedValueMaxLength`] by setting
`OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` to the same value.

### Rewrite all manual instrumentation

All spans created with the [`tracer`] package need to be recreated using
OpenTelemetry. To best understand this, consider the following function
instrumented with the [`tracer`] package.

```go
func BusinessOperation(ctx context.Context, client string) {
	opts := []tracer.StartSpanOption{
		tracer.Tag("client", client),
		tracer.SpanType("internal"),
	}
	if parent, ok := tracer.SpanFromContext(ctx); ok {
		opts = append(opts, tracer.ChildOf(parent.Context()))
	}
	span := tracer.StartSpan("BusinessOperation", opts...)
	defer span.Finish()
	/* ... */
}
```

Recreating all spans using OpenTelemetry, that function becomes:

```go
func BusinessOperation(ctx context.Context, client string) {
	tracer := otel.Tracer("app-name")
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attribute.String("client", client)),
		trace.WithSpanKind(trace.SpanKindInternal),
	}
	ctx, span := tracer.Start(ctx, "BusinessOperation", opts...)
	defer span.End()
	/* ... */
}
```

Here's how this recreation is broken down. First, OpenTelemetry uses `Tracer`s
to encapsulate the tracing function of a single instrumentation library. Create
a `Tracer` from the global `TracerProvider` registered when you started the
[`distro.SDK`]. To do this, use the [`otel.Tracer`] function and pass the name
of your application. For example:

```go
tracer := otel.Tracer("app-name")
```

Use the newly created `Tracer` and its `Start` function to replace all
[`tracer.StartSpan`] invocations.

```go
ctx, span := tracer.Start(ctx, "BusinessOperation", /* options ... */)
```

Use the `operationName` parameter from [`tracer.StartSpan`] as the `name`
parameter for `Start`. The following replacements are used for
[`tracer.StartSpanOption`] instances:

| [`tracer.StartSpanOption`] | Replacement |
| --- | --- |
| [`tracer.ChildOf`] | N/A. The parent-child relationship of spans is defined with a [`context.Context`]. The [`context.Context`] passed to `Start` needs to contain the parent span. This is automatically done if the context was returned from a previous call to `Start`. It can explicitly be done using the [`trace.ContextWithSpan`]. |
| [`tracer.ResourceName`] | N/A. This value is defined globally with a [`Resource`] that applies to all spans. See [Defining a Resource](#defining-a-resource) for more information. |
| [`tracer.ServiceName`] | N/A. This value is defined globally in a [`Resource`] that applies to all spans. See [Defining a Resource](#defining-a-resource) for more information. |
| [`tracer.SpanType`] | [`trace.WithSpanKind`] |
| [`tracer.StartTime`] | [`trace.WithTimestamp`] |
| [`tracer.Tag`] | [`trace.WithAttributes`] |
| [`tracer.WithRecordedValueMaxLength`] | N/A. This value is set globally. See [Setting Span Limits](#setting-span-limits) for more information. |
| [`tracer.WithSpanID`] | N/A. Span IDs are automatically set. If custom span IDs are needed you will need to create a custom [`IDGenerator`]. |

Finally, the created span, similar to before, needs to be ended. Use the
OpenTelemetry span's `End` method to do this.

```go
defer span.End()
```

### Replace all Instrumentation Libraries

| [SignalFx Tracing Library for Go] | OpenTelemetry |
| --- | --- |
| [`aws/aws-sdk-go/aws`] | [`otelaws`] |
| [`bradfitz/gomemcache/memcache`] | [`otelmemcache`] |
| [`confluentinc/confluent-kafka-go/kafka`] | [`splunkkafka`] |
| [`database/sql`] | [`splunksql`] ([`splunkmysql`], [`splunkpgx`], [`splunkpq`]) |
| [`emicklei/go-restful`] | [`otelrestful`] |
| [`garyburd/redigo`] | This project is archived. Use `gomodule/redigo` and [`splunkredigo`] instead. |
| [`gin-gonic/gin`] | [`otelgin`] |
| [`globalsign/mgo`] | This project is an unsupported fork of an abandoned project. Use `mongodb/mongo-go-driver` and `otelmongo` instead. |
| [`go-chi/chi`] | [`splunkchi`] |
| [`go-redis/redis`] | This package now provides native support for OpenTelemetry. See [this example](https://github.com/go-redis/redis/tree/master/example/otel) for more information. |
| [`gocql/gocql`] | [`otelgocql`] |
| [`gomodule/redigo`] | [`splunkredigo`] |
| [`google.golang.org/api`] | Use either [`otelgrpc`] or [`otelhttp`] with a gRPC or HTTP client when calling [`cloudresourcemanager.NewService`]. |
| [`google.golang.org/grpc.v12`] | This version is no longer supported. Use the latest version along with [`otelgrpc`]. |
| [`google.golang.org/grpc`] | [`otelgrpc`] |
| [`gorilla/mux`] | [`otelmux`] |
| [`graph-gophers/graphql-go`] | [`splunkgraphql`] |
| [`jinzhu/gorm`] | [`splunkgorm`] |
| [`jmoiron/sqlx`] | [`splunksqlx`] |
| [`julienschmidt/httprouter`] | [`splunkhttprouter`] |
| [`k8s.io/client-go/kubernetes`] | [`splunkclient-go`] |
| [`labstack/echo.v4`] | [`otelecho`] |
| [`labstack/echo`] | Versions prior to `v4` are no longer supported. Upgrade to `echo@v4` and use `otelecho`. |
| [`miekg/dns`] | [`splunkdns`]
| [`mongodb/mongo-go-driver/mongo`] | [`otelmongo`] |
| [`net/http`] | [`splunkhttp`], [`otelhttp`] |
| [`olivere/elastic`] | [`splunkelastic`] |
| [`Shopify/sarama`] | [`otelsarama`] |
| [`syndtr/goleveldb/leveldb`] | [`splunkleveldb`] |
| [`tidwall/buntdb`] | [`splunkbuntdb`] |

## Troubleshooting

See [Troubleshooting](./docs/troubleshooting.md) for help resolving any issues
encountered.

[SignalFx Tracing Library for Go]: https://github.com/signalfx/signalfx-go-tracing
[Splunk Distribution of OpenTelemetry Go]: https://github.com/signalfx/splunk-otel-go
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
[`Shopify/sarama`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/Shopify/sarama
[`aws/aws-sdk-go/aws`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/aws/aws-sdk-go/aws
[`bradfitz/gomemcache/memcache`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/bradfitz/gomemcache/memcache
[`confluentinc/confluent-kafka-go/kafka`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/confluentinc/confluent-kafka-go/kafka
[`database/sql`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/database/sql
[`emicklei/go-restful`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/emicklei/go-restful
[`garyburd/redigo`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/garyburd/redigo
[`gin-gonic/gin`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/gin-gonic/gin
[`globalsign/mgo`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/globalsign/mgo
[`go-chi/chi`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/go-chi/chi
[`go-redis/redis`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/go-redis/redis
[`gocql/gocql`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/gocql/gocql
[`gomodule/redigo`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/gomodule/redigo
[`google.golang.org/api`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/google.golang.org/api
[`google.golang.org/grpc`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/google.golang.org/grpc
[`google.golang.org/grpc.v12`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/google.golang.org/grpc.v12
[`gorilla/mux`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/gorilla/mux
[`graph-gophers/graphql-go`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/graph-gophers/graphql-go
[`jinzhu/gorm`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/jinzhu/gorm
[`jmoiron/sqlx`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/jmoiron/sqlx
[`julienschmidt/httprouter`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/julienschmidt/httprouter
[`k8s.io/client-go/kubernetes`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/k8s.io/client-go/kubernetes
[`labstack/echo`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/labstack/echo
[`labstack/echo.v4`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/labstack/echo.v4
[`miekg/dns`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/miekg/dns
[`mongodb/mongo-go-driver/mongo`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/mongodb/mongo-go-driver/mongo
[`net/http`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/net/http
[`olivere/elastic`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/olivere/elastic
[`syndtr/goleveldb/leveldb`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/syndtr/goleveldb/leveldb
[`tidwall/buntdb`]: https://github.com/signalfx/signalfx-go-tracing/tree/v1.12.0/contrib/tidwall/buntdb
[`splunksql`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/database/sql/splunksql
[`splunkkafka`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka
[`splunkchi`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/go-chi/chi/splunkchi
[`splunkmysql`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/go-sql-driver/mysql/splunkmysql
[`splunkredigo`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/gomodule/redigo/splunkredigo
[`splunkgraphql`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql
[`splunkpgx`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/jackc/pgx/splunkpgx
[`splunkgorm`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/jinzhu/gorm/splunkgorm
[`splunksqlx`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/jmoiron/sqlx/splunksqlx
[`splunkhttprouter`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter
[`splunkpq`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/lib/pq/splunkpq
[`splunkdns`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/miekg/dns/splunkdns
[`splunkleveldb`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb
[`splunkbuntdb`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/github.com/tidwall/buntdb/splunkbuntdb
[`splunkelastic`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/gopkg.in/olivere/elastic/splunkelastic
[`splunkclient-go`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/k8s.io/client-go/splunkclient-go
[`splunkhttp`]: https://github.com/signalfx/splunk-otel-go/tree/main/instrumentation/net/http/splunkhttp
[`otelaws`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws
[`otelmemcache`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache
[`otelrestful`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/emicklei/go-restful/otelrestful
[`otelgin`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
[`otelgocql`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gocql/gocql/otelgocql
[`otelmux`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux
[`otelecho`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho
[`otelsarama`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama
[`otelmongo`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo
[`otelgrpc`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc
[`otelhttp`]: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
[`cloudresourcemanager.NewService`]: https://pkg.go.dev/google.golang.org/api@v0.70.0/cloudresourcemanager/v3#NewService
