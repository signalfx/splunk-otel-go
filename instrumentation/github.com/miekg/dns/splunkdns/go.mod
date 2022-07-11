module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.16

require (
	github.com/miekg/dns v1.1.50
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.0.0
	go.opentelemetry.io/otel v1.8.0
	go.opentelemetry.io/otel/trace v1.8.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
