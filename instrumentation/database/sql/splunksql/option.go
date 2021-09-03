package splunksql

import (
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/config"
	"go.opentelemetry.io/otel/trace"
)

func newConfig(options ...Option) config.Config {
	c := config.NewConfig()
	for _, o := range options {
		o.apply(&c)
	}
	return c
}

type Option interface {
	apply(*config.Config)
}

type tracerProviderOption struct {
	tp trace.TracerProvider
}

func (o tracerProviderOption) apply(c *config.Config) {
	c.TracerProvider = o.tp
}

func WithTracerProvider(tp trace.TracerProvider) Option {
	return tracerProviderOption{tp: tp}
}
