# Splunk Instrumentation for `github.com/jinzhu/gorm`

[![Go Reference](https://pkg.go.dev/badge/github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm.svg)](https://pkg.go.dev/github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm)

This package provides instrumentation for the `github.com/jinzhu/gorm`
package. The instrumentation is provided by wrapping the
[`splunksql`](../../../../database/sql/splunksql) instrumentation.

## Getting Started

To start using this instrumentation, replace the use of the `gorm.Open`
function `Open` from this package. The returned `*gorm.DB` can then be use as
normal with the `gorm` package.

An example of this can be found [here](./example_test.go).
