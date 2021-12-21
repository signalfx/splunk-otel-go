// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package distro

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
)

// Environment variable keys that set values of the configuration.
const (
	// Access token added to exported data.
	accessTokenKey = "SPLUNK_ACCESS_TOKEN"

	// OpenTelemetry TextMapPropagator to set as global.
	otelPropagatorsKey = "OTEL_PROPAGATORS"

	// FIXME: support OTEL_SPAN_LINK_COUNT_LIMIT
	// FIXME: support OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT
	// FIXME: support OTEL_TRACES_EXPORTER
)

// config is the configuration used to create and operate an SDK.
type config struct {
	AccessToken string
	Endpoint    string
	Propagator  propagation.TextMapPropagator
}

// newConfig returns a validated config with Splunk defaults.
func newConfig(opts ...Option) (*config, error) {
	c := &config{
		AccessToken: envOr(accessTokenKey, ""),
	}

	for _, o := range opts {
		o.apply(c)
	}

	// Apply default field values if they were not set.
	if c.Propagator == nil {
		c.Propagator = loadPropagator(
			envOr(otelPropagatorsKey, "tracecontext,baggage"),
		)
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// Validate ensures c is valid or a, otherwise returning an appropriate error.
func (c *config) Validate() error {
	var errs []string

	if c.Endpoint != "" {
		if _, err := url.Parse(c.Endpoint); err != nil {
			errs = append(errs, "invalid endpoint: %s", err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %v", errs)
	}
	return nil
}

type nonePropagatorType struct{ propagation.TextMapPropagator }

var nonePropagator = nonePropagatorType{}

// propagators maps environment variable values to TextMapPropagator creation
// functions.
var propagators = map[string]func() propagation.TextMapPropagator{
	// W3C Trace Context.
	"tracecontext": func() propagation.TextMapPropagator {
		return propagation.TraceContext{}
	},
	// W3C Baggage
	"baggage": func() propagation.TextMapPropagator {
		return propagation.Baggage{}
	},
	// B3 Single
	"b3": func() propagation.TextMapPropagator {
		return b3.New(b3.WithInjectEncoding(b3.B3SingleHeader))
	},
	// B3 Multi
	"b3multi": func() propagation.TextMapPropagator {
		return b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	},
	// Jaeger
	"jaeger": func() propagation.TextMapPropagator {
		return jaeger.Jaeger{}
	},
	// AWS X-Ray.
	"xray": func() propagation.TextMapPropagator {
		return xray.Propagator{}
	},
	// OpenTracing Trace
	"ottrace": func() propagation.TextMapPropagator {
		return ot.OT{}
	},
	// None, explicitly do not set a global propagator.
	"none": func() propagation.TextMapPropagator {
		return nonePropagator
	},
}

func loadPropagator(name string) propagation.TextMapPropagator {
	var props []propagation.TextMapPropagator
	for _, part := range strings.Split(name, ",") {
		factory, ok := propagators[part]
		if !ok {
			// Skip invalid data.
			continue
		}
		props = append(props, factory())
	}

	switch len(props) {
	case 0:
		// Default to "tracecontext,baggage".
		return propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	case 1:
		return props[0]
	default:
		return propagation.NewCompositeTextMapPropagator(props...)
	}
}

// envOr returns the environment variable value associated with key if it
// exists, otherwise it returns alt.
func envOr(key, alt string) string {
	v, ok := os.LookupEnv(key)
	if ok {
		return v
	}
	return alt
}

// Option sets a config setting value.
type Option interface {
	apply(*config)
}

// optionFunc is a functional option implementation for Option interface.
type optionFunc func(*config)

func (fn optionFunc) apply(c *config) {
	fn(c)
}

// WithEndpoint configures the endpoint telemetry is sent to.
// Setting an empty string results in no operation.
func WithEndpoint(endpoint string) Option {
	return optionFunc(func(c *config) {
		c.Endpoint = endpoint
	})
}

// WithAccessToken configures the authentication token
// allowing exporters to send data directly to a Splunk back-end.
// Setting an empty string results in no operation.
func WithAccessToken(accessToken string) Option {
	return optionFunc(func(c *config) {
		c.AccessToken = accessToken
	})
}

// WithPropagator configures the OpenTelemetry TextMapPropagator set as the
// global TextMapPropagator. Passing nil will set a TextMapPropagator that
// propagates nothing.
//
// The OTEL_PROPAGATORS environment variable value is used if this Option is
// not provided.
//
// By default, a tracecontext and baggage TextMapPropagator is set as the
// global TextMapPropagator if this is not provided or the OTEL_PROPAGATORS
// environment variable is not set.
func WithPropagator(p propagation.TextMapPropagator) Option {
	return optionFunc(func(c *config) {
		if p == nil {
			p = nonePropagator
		}
		c.Propagator = p
	})
}
