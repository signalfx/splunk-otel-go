module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.16

require (
	github.com/miekg/dns v1.1.47
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.8.0
	go.opentelemetry.io/otel v1.6.1
	go.opentelemetry.io/otel/trace v1.6.1
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
