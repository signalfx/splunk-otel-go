module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.18

require (
	github.com/miekg/dns v1.1.53
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.5.0
	go.opentelemetry.io/otel v1.15.1
	go.opentelemetry.io/otel/trace v1.15.1
)

require (
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel/metric v0.37.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
