module github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic

go 1.16

require (
	github.com/olivere/elastic/v7 v7.0.32
	github.com/signalfx/splunk-otel-go/instrumentation/internal v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/otel v1.5.0
	go.opentelemetry.io/otel/trace v1.5.0
	gopkg.in/olivere/elastic.v3 v3.0.75
)

replace (
	github.com/signalfx/splunk-otel-go => ../../../../../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal/
)
