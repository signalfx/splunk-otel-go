module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.24.0

require (
	github.com/miekg/dns v1.1.68
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.28.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
)

require (
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
