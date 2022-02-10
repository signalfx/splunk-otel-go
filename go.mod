module github.com/signalfx/splunk-otel-go

go 1.16

require (
	github.com/go-logr/logr v1.2.2
	github.com/go-logr/zapr v1.2.2
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib/propagators/aws v1.3.0
	go.opentelemetry.io/contrib/propagators/b3 v1.3.0
	go.opentelemetry.io/contrib/propagators/jaeger v1.3.0
	go.opentelemetry.io/contrib/propagators/ot v1.3.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/exporters/jaeger v1.3.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.3.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
	go.opentelemetry.io/proto/otlp v0.12.0
	go.uber.org/goleak v1.1.12
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.44.0
)
