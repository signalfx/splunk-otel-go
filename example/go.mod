module github.com/signalfx/splunk-otel-go/example

go 1.20

require (
	github.com/signalfx/splunk-otel-go/distro v1.14.0
	github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp v1.14.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0
	golang.org/x/sync v0.6.0
)

require (
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.49.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.49.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.24.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.24.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.24.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.24.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.24.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240228224816-df926f6c8641 // indirect
	google.golang.org/grpc v1.62.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/distro => ../distro

replace github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp => ../instrumentation/net/http/splunkhttp
