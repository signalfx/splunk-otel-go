module github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport/test

go 1.18

require (
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go v1.1.0
	github.com/stretchr/testify v1.8.1
	go.opentelemetry.io/otel v1.11.2
	go.opentelemetry.io/otel/sdk v1.11.2
	go.opentelemetry.io/otel/trace v1.11.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go v1.1.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.1.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.25.4 // indirect
	k8s.io/client-go v0.25.4 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/utils v0.0.0-20221101230645-61b03e2f6476 // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go => ../../
)
