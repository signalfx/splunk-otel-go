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

package splunkelastic_test

import (
	"context"
	"net/http"

	elasticv3 "gopkg.in/olivere/elastic.v3"
	elasticv5 "gopkg.in/olivere/elastic.v5"

	"github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic"
)

func Example_v5() {
	// Wrap the HTTP Transport used by the client to communicate with the
	// Elasticsearch cluster so all requests made are included in traces.
	client, err := elasticv5.NewClient(
		elasticv5.SetURL("http://127.0.0.1:9200"),
		elasticv5.SetHttpClient(&http.Client{
			Transport: splunkelastic.WrapRoundTripper(http.DefaultTransport),
		}),
	)
	if err != nil {
		// Handle error.
		panic(err)
	}

	// Spans are emitted for all operations the client performs against the
	// Elasticsearch cluster.
	_, err = client.Index().
		Index("twitter").
		Type("tweet").
		Index("1").
		BodyString(`{"user": "test", "message": "hello"}`).
		// If a context that contains an span is passed here it will be used
		// as the parent span for the performed operations.
		Do(context.Background())
	if err != nil {
		// Handle error.
		panic(err)
	}
}

func Example_v3() {
	// Wrap the HTTP Transport used by the client to communicate with the
	// Elasticsearch cluster so all requests made are included in traces.
	client, err := elasticv3.NewClient(
		elasticv3.SetURL("http://127.0.0.1:9200"),
		elasticv3.SetHttpClient(&http.Client{
			Transport: splunkelastic.WrapRoundTripper(http.DefaultTransport),
		}),
	)
	if err != nil {
		// Handle error.
		panic(err)
	}

	// Spans are emitted for all operations the client performs against the
	// Elasticsearch cluster.
	_, err = client.Index().
		Index("twitter").
		Type("tweet").
		Index("1").
		BodyString(`{"user": "test", "message": "hello"}`).
		// Be sure to use DoC to propagate a trace with a context that
		// contains a parent span.
		DoC(context.Background())
	if err != nil {
		// Handle error.
		panic(err)
	}
}
