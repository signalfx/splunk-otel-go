module github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/redis/test

go 1.16

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/containerd/continuity v0.2.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/gomodule/redigo v1.8.8
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.0.3 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo v0.8.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.6.3
	go.opentelemetry.io/otel/sdk v1.6.1
	go.opentelemetry.io/otel/trace v1.6.3
	golang.org/x/net v0.0.0-20211215060638-4ddde0e984e9 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo => ../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../../internal/
)
