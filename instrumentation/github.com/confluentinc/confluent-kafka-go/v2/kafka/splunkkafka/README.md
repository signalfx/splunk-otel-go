# Splunk instrumentation for `github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka`

This instrumentation is for the
[github.com/confluentinc/confluent-kafka-go/v2/kafka](https://pkg.go.dev/github.com/confluentinc/confluent-kafka-go/v2/kafka)
package.

## Compatibility

The Producer will end spans when a delivery report is returned. Setting
`"go.delivery.reports"` to `false` will disable the delivery reports and can
result in an build up of un-ended spans. If delivery reports are disabled, an
un-instrumented Producer should be used instead.

Similar to the
instrumented package, librdkafka 2.0.0+ is required. This means you will need
to use an environment that supports the [pre-built
binaries](https://github.com/confluentinc/confluent-kafka-go#librdkafka), or
[install](https://github.com/confluentinc/confluent-kafka-go#installing-librdkafka)
the library manually. Important to note, similar to the instrumented package,
**cgo is required** and **this instrumentation does not support the Windows
operating system**.

## Getting started

The `NewConsumer` and `NewProducer` functions are provided as drop-in
replacements of the equivalent from the `kafka` package. See [these
examples](./example_test.go) for how to use these functions.
