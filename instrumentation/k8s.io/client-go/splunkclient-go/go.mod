module github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go

go 1.15

require (
	github.com/signalfx/splunk-otel-go v0.6.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/trace v1.2.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
)

replace github.com/signalfx/splunk-otel-go => ../../../../
