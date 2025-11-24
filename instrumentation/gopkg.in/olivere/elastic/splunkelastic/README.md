# Splunk instrumentation for `gopkg.in/olivere/elastic`

This package provides OpenTelemetry instrumentation for the `*/olivere/elastic`
packages:

| Elastic version | Package URL                                                          |
|-----------------|----------------------------------------------------------------------|
| 7.0             | [`github.com/olivere/elastic/v7`](https://github.com/olivere/elastic)|
| 6.0             | [`github.com/olivere/elastic`](https://github.com/olivere/elastic)   |
| 5.0             | [`gopkg.in/olivere/elastic.v5`](https://gopkg.in/olivere/elastic.v5) |
| 3.0             | [`gopkg.in/olivere/elastic.v3`](https://gopkg.in/olivere/elastic.v3) |

## Getting Started

This package provides an `http.Transport` that can be used with
`gopkg.in/olivere/elastic` to instrument requests that package makes. See
[example_test.go](./example_test.go) for more information.
