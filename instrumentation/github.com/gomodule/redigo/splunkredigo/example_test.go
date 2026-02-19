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

package splunkredigo_test

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	//nolint:staticcheck // Deprecated package, but still used here.
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/option"
	//nolint:staticcheck // Deprecated package, but still used here.
	splunkredis "github.com/signalfx/splunk-otel-go/instrumentation/github.com/gomodule/redigo/splunkredigo/redis"
)

func Example() {
	// The context used should be propagated from a calling process to ensure
	// trace continuity.
	ctx := context.TODO()

	const db = 15
	// Options passed to Dial* functions can be either redis DialOptions or
	// splunkredigo options.
	conn, err := splunkredis.DialContext(ctx, "tcp", "127.0.0.1:6379",
		redis.DialDatabase(db),
		option.WithAttributes([]attribute.KeyValue{
			attribute.String("tier", "alpha"),
			semconv.DBRedisDBIndexKey.Int(db),
		}),
		redis.DialConnectTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}

	// Create spans per command by using the Redis connection as usual.
	if _, err := conn.Do("SET", "vehicle", "truck"); err != nil {
		panic(err)
	}

	tracer := otel.Tracer("my-instrumentation.name")
	// Use a context to pass information down the processing chain.
	ctx, root := tracer.Start(ctx, "parent.request")

	// When passed a context as an argument, conn.Do will create a span
	// inheriting from the active span it contains ('parent.request').
	if _, err := conn.Do("SET", "food", "cheese", ctx); err != nil {
		panic(err)
	}
	root.End()
}

func Example_dialURL() {
	conn, err := splunkredis.DialURL("redis://127.0.0.1:6379/15")
	if err != nil {
		panic(err)
	}
	if _, err := conn.Do("SET", "vehicle", "truck"); err != nil {
		panic(err)
	}
}

func Example_pool() {
	// Set your Dial function when using a redis Pool to trace all Conn.
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return splunkredis.Dial("tcp", "127.0.0.1:6379")
		},
	}

	conn := pool.Get()
	if _, err := conn.Do("SET", "vehicle", "truck"); err != nil {
		panic(err)
	}
}
