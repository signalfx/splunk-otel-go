module github.com/signalfx/splunk-otel-go/instrumentation/github.com/miekg/dns/splunkdns

go 1.22.0
<<<<<<< Updated upstream

toolchain go1.23.1
=======
>>>>>>> Stashed changes

require (
	github.com/miekg/dns v1.1.62
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.19.0
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/trace v1.30.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel/metric v1.30.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
