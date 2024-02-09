module github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic

go 1.20

require (
	github.com/olivere/elastic/v7 v7.0.32
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.13.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
	gopkg.in/olivere/elastic.v3 v3.0.75
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/otel/metric v1.23.1 // indirect
	golang.org/x/sync v0.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../internal
