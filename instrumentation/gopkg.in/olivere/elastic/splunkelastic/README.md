# Splunk instrumentation for `gopkg.in/olivere/elastic`

This package provides OpenTelemetry instrumentation for the
[gopkg.in/olivere/elastic.v{3,5}](https://gopkg.in/olivere/elastic.v5)
packages.

## Getting Started

This package provides an `http.Client` that can used with
`gopkg.in/olivere/elastic` to instrument requests that package makes. See
[example_test.go](./example_test.go) for more information.
