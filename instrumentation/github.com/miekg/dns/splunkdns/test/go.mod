module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns/test

go 1.16

require (
	github.com/miekg/dns v1.1.49
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns v0.9.0
	github.com/stretchr/testify v1.7.2
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
)
