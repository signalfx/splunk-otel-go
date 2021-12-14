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

package splunkgraphql_test

import (
	"net/http"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/graph-gophers/graphql-go/splunkgraphql"
)

const schema = `
	schema {
		query: Query
	}
	type Query {
		hello: String!
	}
`

type resolver struct{}

func (*resolver) Hello() string { return "Hello, world!" }

func Example() {
	tracer := graphql.Tracer(splunkgraphql.NewTracer())
	s := graphql.MustParseSchema(schema, new(resolver), tracer)
	http.Handle("/query", &relay.Handler{Schema: s})

	/*
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}

		...
	*/
}
