module github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-chi/chi/splunkchi

go 1.16

require (
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/signalfx/splunk-otel-go v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
	golang.org/x/net v0.0.0-20220107192237-5cfca573fb4d // indirect
)

replace github.com/signalfx/splunk-otel-go => ../../../../../
