module github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport/test

go 1.16

require (
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)

replace github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go => ../../
