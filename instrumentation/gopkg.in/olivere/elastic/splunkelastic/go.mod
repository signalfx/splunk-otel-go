module github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/olivere/elastic/v7 v7.0.31
	github.com/signalfx/splunk-otel-go v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/olivere/elastic.v3 v3.0.75
)

replace github.com/signalfx/splunk-otel-go => ../../../../../
