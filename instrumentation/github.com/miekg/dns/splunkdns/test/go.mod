module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns/test

go 1.15

replace github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns => ../

require (
	github.com/miekg/dns v1.1.43
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/sdk v1.1.0
	go.opentelemetry.io/otel/trace v1.2.0
)
