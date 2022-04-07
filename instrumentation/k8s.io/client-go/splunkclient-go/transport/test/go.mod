module github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport/test

go 1.17

require (
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go v0.8.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.6.2
	go.opentelemetry.io/otel/sdk v1.6.1
	go.opentelemetry.io/otel/trace v1.6.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go v0.8.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.8.0 // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f // indirect
	golang.org/x/sys v0.0.0-20210831042530-f4d43177bf5e // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apimachinery v0.23.5 // indirect
	k8s.io/client-go v0.23.5 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal/
	github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go => ../../
)
