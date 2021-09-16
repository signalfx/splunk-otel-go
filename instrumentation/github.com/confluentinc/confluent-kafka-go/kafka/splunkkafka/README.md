# Splunk instrumentation for `github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka`

This instrumentation is for the
[github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka](https://github.com/confluentinc/confluent-kafka-go)
package.

## Compatibility

This instrumentation was built to support
[v1.7.0](https://github.com/confluentinc/confluent-kafka-go/releases/tag/v1.7.0)
of github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka. Similar to the
instrumented package, librdkafka 1.6.0+ is required. This means you will need
to use an environment that supports the [pre-built
binaries](https://github.com/confluentinc/confluent-kafka-go#librdkafka), or
[install](https://github.com/confluentinc/confluent-kafka-go#installing-librdkafka)
the library manually. Important to note, similar to the instrumented package,
this instrumentation does not support the Windows operating system.

## Getting started

A consumer that traces all received messages can be created with `NewConsumer`.

```golang
import (
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	/* ... */
)

func main() {
	c, err := splunkkafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	/* ... */
}
```

This should be a drop-in replacement for the `kafka.NewConsumer` function.
Similarly, a producer that traces all messages it sends to a topic can be
created with `NewProducer`.

```golang
import (
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
    /* ... */
)

func main() {
	p, err := splunkkafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}

    /* ... */
}
```

Again, `NewProducer` should be a drop-in replacement for `kafka.NewProducer`.
