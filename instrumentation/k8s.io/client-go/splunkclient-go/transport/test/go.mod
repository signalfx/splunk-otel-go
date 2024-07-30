module github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport/test

go 1.22.0

toolchain go1.22.5

require (
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go v1.18.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.18.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.30.3 // indirect
	k8s.io/client-go v0.30.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go => ../..
)
