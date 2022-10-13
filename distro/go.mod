module github.com/signalfx/splunk-otel-go/distro

go 1.17

require (
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/signalfx/splunk-otel-go v1.1.0
	github.com/stretchr/testify v1.8.0
	github.com/tonglil/buflogr v0.0.0-20220114010534-d490b3990d7e
	go.opentelemetry.io/contrib/propagators/autoprop v0.36.3
	go.opentelemetry.io/otel v1.11.0
	go.opentelemetry.io/otel/exporters/jaeger v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.10.0
	go.opentelemetry.io/otel/sdk v1.11.0
	go.opentelemetry.io/proto/otlp v0.19.0
	go.uber.org/goleak v1.2.0
	go.uber.org/zap v1.23.0
	google.golang.org/grpc v1.50.0
)

require (
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.11.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.11.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.11.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.11.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.10.0 // indirect
	go.opentelemetry.io/otel/trace v1.11.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-go => ../
