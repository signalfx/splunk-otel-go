# Splunk instrumentation for `k8s.io/client-go`

This package provides OpenTelemetry instrumentation for the
[k8s.io/client-go](https://github.com/kubernetes/client-go) package.

## Getting Started

The `transport` package is used to wrap all requests to the Kubernetes API. See
the [example](./transport/example_test.go) provided there.
