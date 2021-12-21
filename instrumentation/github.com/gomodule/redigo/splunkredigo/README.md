# Splunk instrumentation for `github.com/gomodule/redigo`

This package provides OpenTelemetry instrumentation for the
[github.com/gomodule/redigo](https://github.com/gomodule/redigo) module.

## Getting Started

This package is designed to instrument an existing `redis.Conn` so all
communication it handles is traced. See [example_test.go](./example_test.go)
for more information.
