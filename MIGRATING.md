# Migrate from the SignalFx Tracing Library for Go

The [Splunk Distribution of OpenTelemetry for Go] replaces the [SignalFx Tracing
Library for Go].

Use the following instructions to migrate help migrate to the [Splunk
Distribution of OpenTelemetry for Go].

## Requirements

Go version 1.16 or greater is required by the [Splunk Distribution of
OpenTelemetry for Go].

## Migration Steps

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

### Replace all Instrumentation Libraries

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
