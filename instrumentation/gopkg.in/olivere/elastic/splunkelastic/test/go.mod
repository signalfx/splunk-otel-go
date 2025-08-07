module github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic/test

go 1.23.0

require (
	github.com/olivere/elastic/v7 v7.0.32
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic v1.27.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/sdk v1.37.0
	go.opentelemetry.io/otel/trace v1.37.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.27.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

replace (
	github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic => ../
	github.com/signalfx/splunk-otel-go/instrumentation/internal => ../../../../../internal
)
