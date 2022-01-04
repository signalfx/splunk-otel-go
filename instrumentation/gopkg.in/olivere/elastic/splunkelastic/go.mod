module github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic

go 1.16

require (
	github.com/signalfx/splunk-otel-go v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)

replace github.com/signalfx/splunk-otel-go => ../../../../../
