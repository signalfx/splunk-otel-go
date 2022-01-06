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

//go:build cgo && linux
// +build cgo,linux

package test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/ory/dockertest"
	"github.com/signalfx/splunk-otel-go/instrumentation/gopkg.in/olivere/elastic/splunkelastic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	apitrace "go.opentelemetry.io/otel/trace"
)

var addr string

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		fmt.Println("Skipping running heavy integration test in short mode.")
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	resource, err := pool.Run("elasticsearch", "7.16.2", []string{"discovery.type=single-node"})
	if err != nil {
		log.Fatalf("Could not create elasticsearch container: %v", err)
	}

	// If run with docker-machine the hostname needs to be set.
	u, err := url.Parse(pool.Client.Endpoint())
	if err != nil {
		log.Fatalf("Could not parse endpoint: %s", pool.Client.Endpoint())
	}
	hostname := u.Hostname()
	if hostname == "" {
		hostname = "127.0.0.1"
	}

	target := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(hostname, resource.GetPort("9200/tcp")),
	}
	addr = target.String()

	// Wait for the Elasticsearch to come up using an exponential-backoff
	// retry.
	if err = pool.Retry(func() error {
		client, e := elastic.NewClient(elastic.SetURL(addr), elastic.SetSniff(false))
		if e != nil {
			return e
		}
		_, code, e := client.Ping(addr).Do(context.Background())
		if e != nil {
			return e
		}
		if code != 200 {
			return fmt.Errorf("elasticsearch server not ready: %d", code)
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to elasticsearch server: %s", err)
	}

	code := m.Run()

	// Run sequentially becauase os.Exit will skip a defer.
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

type Tweet struct {
	User     string                `json:"user"`
	Message  string                `json:"message"`
	Retweets int                   `json:"retweets"`
	Image    string                `json:"image,omitempty"`
	Created  time.Time             `json:"created,omitempty"`
	Tags     []string              `json:"tags,omitempty"`
	Location string                `json:"location,omitempty"`
	Suggest  *elastic.SuggestField `json:"suggest_field,omitempty"`
}

func run(ctx context.Context, t *testing.T, client *elastic.Client) {
	// Ping the Elasticsearch server to get e.g. the version number
	info, code, err := client.Ping(addr).Do(ctx)
	require.NoError(t, err)
	t.Logf("Elasticsearch returned with code %d and version %s", code, info.Version.Number)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("twitter").Do(ctx)
	require.NoError(t, err)
	if !exists {
		// Create a new index.
		mapping := `
{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
		"properties":{
			"user":{
				"type":"keyword"
			},
			"message":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"retweets":{
				"type":"long"
			},
			"tags":{
				"type":"keyword"
			},
			"location":{
				"type":"geo_point"
			},
			"suggest_field":{
				"type":"completion"
			}
		}
	}
}
`
		createIndex, err := client.CreateIndex("twitter").Body(mapping).Do(ctx)
		require.NoError(t, err)
		require.True(t, createIndex.Acknowledged, "createIndex unacknowledged")
	}

	// Index a tweet (using JSON serialization)
	tweet1 := Tweet{User: "olivere", Message: "Take Five", Retweets: 0}
	put1, err := client.Index().
		Index("twitter").
		Id("1").
		BodyJson(tweet1).
		Do(ctx)
	require.NoError(t, err)
	t.Logf("Indexed tweet %s to index %s, type %s", put1.Id, put1.Index, put1.Type)

	// Index a second tweet (by string)
	tweet2 := `{"user" : "olivere", "message" : "It's a Raggy Waltz"}`
	put2, err := client.Index().
		Index("twitter").
		Id("2").
		BodyString(tweet2).
		Do(ctx)
	require.NoError(t, err)
	t.Logf("Indexed tweet %s to index %s, type %s", put2.Id, put2.Index, put2.Type)

	// Get tweet with specified ID
	get1, err := client.Get().
		Index("twitter").
		Id("1").
		Do(ctx)
	require.NoError(t, err)
	t.Logf("Got document %s in version %d from index %s, type %s", get1.Id, get1.Version, get1.Index, get1.Type)

	// Refresh to make sure the documents are searchable.
	_, err = client.Refresh().Index("twitter").Do(ctx)
	require.NoError(t, err)

	// Search with a term query
	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := client.Search().
		Index("twitter").   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute
	require.NoError(t, err)

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	t.Logf("Query took %d milliseconds", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Tweet
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		tweet := item.(Tweet)
		t.Logf("Tweet by %s: %s", tweet.User, tweet.Message)
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	t.Logf("Found a total of %d tweets", searchResult.TotalHits())

	// Here's how you iterate through results with full control over each step.
	if searchResult.TotalHits() > 0 {
		t.Logf("Found a total of %d tweets", searchResult.TotalHits())

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var tweet Tweet
			err := json.Unmarshal(hit.Source, &tweet)
			if err != nil {
				// Deserialization failed
			}

			// Work with tweet
			t.Logf("Tweet by %s: %s", tweet.User, tweet.Message)
		}
	} else {
		// No hits
		t.Log("Found no tweets")
	}

	// Update a tweet by the update API of Elasticsearch.
	// We just increment the number of retweets.
	script := elastic.NewScript("ctx._source.retweets += params.num").Param("num", 1)
	update, err := client.Update().Index("twitter").Id("1").
		Script(script).
		Upsert(map[string]interface{}{"retweets": 0}).
		Do(ctx)
	require.NoError(t, err)
	t.Logf("New version of tweet %q is now %d", update.Id, update.Version)

	// Delete an index.
	deleteIndex, err := client.DeleteIndex("twitter").Do(ctx)
	require.NoError(t, err)
	require.True(t, deleteIndex.Acknowledged, "deleteIndex unacknowledged")
}

type testLogger struct {
	t *testing.T
}

func (l testLogger) Printf(format string, v ...interface{}) {
	l.t.Logf(format, v...)
}

func TestSpans(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	client, err := elastic.NewClient(
		elastic.SetURL(addr),
		elastic.SetSniff(false),
		elastic.SetErrorLog(testLogger{t}),
		elastic.SetHttpClient(&http.Client{
			Transport: splunkelastic.WrapRoundTripper(
				http.DefaultTransport,
				splunkelastic.WithTracerProvider(tp),
			),
		}),
	)
	require.NoError(t, err)

	const name = "parent"
	ctx, parent := tp.Tracer("TestSpans").Start(context.Background(), name)
	run(ctx, t, client)
	parent.End()

	require.NoError(t, tp.Shutdown(context.Background()))

	spans := sr.Ended()
	require.Greater(t, len(spans), 0, "no spans created")

	parentRO := spans[len(spans)-1]
	require.Equal(t, name, parentRO.Name())
	spans = spans[:len(spans)-1]

	// Client creation spans are not children of parent, test them
	// independently.
	expectedNames := []string{"HTTP HEAD"}
	require.GreaterOrEqual(t, len(spans), len(expectedNames), "no client creation spans created")
	createClientSpans := spans[:len(expectedNames)]
	for i, span := range createClientSpans {
		assertSpan(t, expectedNames[i], span)
	}

	spans = spans[len(expectedNames):]
	expectedNames = []string{
		"HTTP GET /",                      // Ping.
		"HTTP HEAD /{index}",              // IndexExists.
		"HTTP PUT /{index}",               // Create a new index.
		"HTTP PUT /{index}/_doc/{id}",     // Index a tweet (using JSON serialization).
		"HTTP PUT /{index}/_doc/{id}",     // Index a second tweet (by string).
		"HTTP GET /{index}/_doc/{id}",     // Get tweet with specified ID.
		"HTTP POST /{index}/_refresh",     // Refresh to make sure the documents are searchable.
		"HTTP POST /{index}/_search",      // Search with a term qauery.
		"HTTP POST /{index}/_update/{id}", // Update a tweet.
		"HTTP DELETE /{index}",            // Delete an index.
	}
	require.Len(t, spans, len(expectedNames))
	traceid := parentRO.SpanContext().TraceID()
	for i, span := range spans {
		assert.Equal(t, traceid, span.SpanContext().TraceID(), span.Name())
		assertSpan(t, expectedNames[i], span)
	}
}

func assertSpan(t *testing.T, name string, span trace.ReadOnlySpan) {
	assert.Equal(t, name, span.Name())
	assert.Equal(t, apitrace.SpanKindClient, span.SpanKind())
}
