module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.16

require (
	github.com/miekg/dns v1.1.47
	github.com/signalfx/splunk-otel-go v0.7.0
	go.opentelemetry.io/otel v1.4.1
	go.opentelemetry.io/otel/trace v1.4.1
)

replace github.com/signalfx/splunk-otel-go => ../../../../../
