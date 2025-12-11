module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.24.0

require (
	github.com/miekg/dns v1.1.69
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.29.0
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
