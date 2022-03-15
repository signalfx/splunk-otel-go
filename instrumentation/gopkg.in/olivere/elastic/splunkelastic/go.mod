module github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic

go 1.16

require (
	github.com/olivere/elastic/v7 v7.0.31
	github.com/signalfx/splunk-otel-go v0.7.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.4.1
	go.opentelemetry.io/otel/trace v1.4.1
	gopkg.in/olivere/elastic.v3 v3.0.75
)

replace github.com/signalfx/splunk-otel-go => ../../../../../
