module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.20

require (
	github.com/miekg/dns v1.1.58
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.12.0
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
)

require (
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	golang.org/x/mod v0.15.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
