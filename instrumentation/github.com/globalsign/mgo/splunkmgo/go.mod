module github.com/signalfx/splunk-otel-go/instrumentation/github.com/globalsign/mgo/splunkmgo

go 1.16

replace github.com/signalfx/splunk-otel-go => ../../../../../

require (
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/signalfx/splunk-otel-go v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
