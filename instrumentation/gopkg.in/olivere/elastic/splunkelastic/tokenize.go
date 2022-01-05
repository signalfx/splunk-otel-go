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

package splunkelastic

import (
	"regexp"
	"sort"
	"strings"
)

// Taken from
// https://github.com/elastic/elasticsearch-specification/blob/60aa3a276e4c617ca7944816a6b4979c2384c675/output/schema/schema.json
var paths = []string{
	"/",
	"/_alias",
	"/_alias/{name}",
	"/_aliases",
	"/_analyze",
	"/_async_search",
	"/_async_search/status/{id}",
	"/_async_search/{id}",
	"/_autoscaling/capacity",
	"/_autoscaling/policy/{name}",
	"/_bulk",
	"/_cache/clear",
	"/_cat",
	"/_cat/aliases",
	"/_cat/aliases/{name}",
	"/_cat/allocation",
	"/_cat/allocation/{node_id}",
	"/_cat/count",
	"/_cat/count/{index}",
	"/_cat/fielddata",
	"/_cat/fielddata/{fields}",
	"/_cat/health",
	"/_cat/indices",
	"/_cat/indices/{index}",
	"/_cat/master",
	"/_cat/ml/anomaly_detectors",
	"/_cat/ml/anomaly_detectors/{job_id}",
	"/_cat/ml/data_frame/analytics",
	"/_cat/ml/data_frame/analytics/{id}",
	"/_cat/ml/datafeeds",
	"/_cat/ml/datafeeds/{datafeed_id}",
	"/_cat/ml/trained_models",
	"/_cat/ml/trained_models/{model_id}",
	"/_cat/nodeattrs",
	"/_cat/nodes",
	"/_cat/pending_tasks",
	"/_cat/plugins",
	"/_cat/recovery",
	"/_cat/recovery/{index}",
	"/_cat/repositories",
	"/_cat/segments",
	"/_cat/segments/{index}",
	"/_cat/shards",
	"/_cat/shards/{index}",
	"/_cat/snapshots",
	"/_cat/snapshots/{repository}",
	"/_cat/tasks",
	"/_cat/templates",
	"/_cat/templates/{name}",
	"/_cat/thread_pool",
	"/_cat/thread_pool/{thread_pool_patterns}",
	"/_cat/transforms",
	"/_cat/transforms/{transform_id}",
	"/_ccr/auto_follow",
	"/_ccr/auto_follow/{name}",
	"/_ccr/auto_follow/{name}/pause",
	"/_ccr/auto_follow/{name}/resume",
	"/_ccr/stats",
	"/_cluster/allocation/explain",
	"/_cluster/health",
	"/_cluster/health/{index}",
	"/_cluster/pending_tasks",
	"/_cluster/reroute",
	"/_cluster/settings",
	"/_cluster/state",
	"/_cluster/state/{metric}",
	"/_cluster/state/{metric}/{index}",
	"/_cluster/stats",
	"/_cluster/stats/nodes/{node_id}",
	"/_cluster/voting_config_exclusions",
	"/_component_template",
	"/_component_template/{name}",
	"/_count",
	"/_dangling",
	"/_dangling/{index_uuid}",
	"/_data_stream",
	"/_data_stream/_migrate/{name}",
	"/_data_stream/_modify",
	"/_data_stream/_promote/{name}",
	"/_data_stream/_stats",
	"/_data_stream/{name}",
	"/_data_stream/{name}/_stats",
	"/_delete_by_query/{task_id}/_rethrottle",
	"/_enrich/_stats",
	"/_enrich/policy",
	"/_enrich/policy/{name}",
	"/_enrich/policy/{name}/_execute",
	"/_eql/search/status/{id}",
	"/_eql/search/{id}",
	"/_features",
	"/_features/_reset",
	"/_field_caps",
	"/_fleet/_fleet_msearch",
	"/_flush",
	"/_forcemerge",
	"/_ilm/migrate_to_data_tiers",
	"/_ilm/move/{index}",
	"/_ilm/policy",
	"/_ilm/policy/{policy}",
	"/_ilm/start",
	"/_ilm/status",
	"/_ilm/stop",
	"/_index_template",
	"/_index_template/_simulate",
	"/_index_template/_simulate/{name}",
	"/_index_template/_simulate_index/{name}",
	"/_index_template/{name}",
	"/_ingest/geoip/stats",
	"/_ingest/pipeline",
	"/_ingest/pipeline/_simulate",
	"/_ingest/pipeline/{id}",
	"/_ingest/pipeline/{id}/_simulate",
	"/_ingest/processor/grok",
	"/_license",
	"/_license/basic_status",
	"/_license/start_basic",
	"/_license/start_trial",
	"/_license/trial_status",
	"/_logstash/pipeline/{id}",
	"/_mapping",
	"/_mapping/field/{fields}",
	"/_mget",
	"/_migration/deprecations",
	"/_migration/system_features",
	"/_ml/_delete_expired_data",
	"/_ml/_delete_expired_data/{job_id}",
	"/_ml/anomaly_detectors",
	"/_ml/anomaly_detectors/_estimate_model_memory",
	"/_ml/anomaly_detectors/_stats",
	"/_ml/anomaly_detectors/_validate",
	"/_ml/anomaly_detectors/_validate/detector",
	"/_ml/anomaly_detectors/{job_id}",
	"/_ml/anomaly_detectors/{job_id}/_close",
	"/_ml/anomaly_detectors/{job_id}/_data",
	"/_ml/anomaly_detectors/{job_id}/_flush",
	"/_ml/anomaly_detectors/{job_id}/_forecast",
	"/_ml/anomaly_detectors/{job_id}/_forecast/{forecast_id}",
	"/_ml/anomaly_detectors/{job_id}/_open",
	"/_ml/anomaly_detectors/{job_id}/_reset",
	"/_ml/anomaly_detectors/{job_id}/_stats",
	"/_ml/anomaly_detectors/{job_id}/_update",
	"/_ml/anomaly_detectors/{job_id}/model_snapshots",
	"/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}",
	"/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_revert",
	"/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_update",
	"/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_upgrade",
	"/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_upgrade/_stats",
	"/_ml/anomaly_detectors/{job_id}/results/buckets",
	"/_ml/anomaly_detectors/{job_id}/results/buckets/{timestamp}",
	"/_ml/anomaly_detectors/{job_id}/results/categories/",
	"/_ml/anomaly_detectors/{job_id}/results/categories/{category_id}",
	"/_ml/anomaly_detectors/{job_id}/results/influencers",
	"/_ml/anomaly_detectors/{job_id}/results/overall_buckets",
	"/_ml/anomaly_detectors/{job_id}/results/records",
	"/_ml/calendars",
	"/_ml/calendars/{calendar_id}",
	"/_ml/calendars/{calendar_id}/events",
	"/_ml/calendars/{calendar_id}/events/{event_id}",
	"/_ml/calendars/{calendar_id}/jobs/{job_id}",
	"/_ml/data_frame/_evaluate",
	"/_ml/data_frame/analytics",
	"/_ml/data_frame/analytics/_explain",
	"/_ml/data_frame/analytics/_preview",
	"/_ml/data_frame/analytics/_stats",
	"/_ml/data_frame/analytics/{id}",
	"/_ml/data_frame/analytics/{id}/_explain",
	"/_ml/data_frame/analytics/{id}/_preview",
	"/_ml/data_frame/analytics/{id}/_start",
	"/_ml/data_frame/analytics/{id}/_stats",
	"/_ml/data_frame/analytics/{id}/_stop",
	"/_ml/data_frame/analytics/{id}/_update",
	"/_ml/datafeeds",
	"/_ml/datafeeds/_preview",
	"/_ml/datafeeds/_stats",
	"/_ml/datafeeds/{datafeed_id}",
	"/_ml/datafeeds/{datafeed_id}/_preview",
	"/_ml/datafeeds/{datafeed_id}/_start",
	"/_ml/datafeeds/{datafeed_id}/_stats",
	"/_ml/datafeeds/{datafeed_id}/_stop",
	"/_ml/datafeeds/{datafeed_id}/_update",
	"/_ml/filters",
	"/_ml/filters/{filter_id}",
	"/_ml/filters/{filter_id}/_update",
	"/_ml/info",
	"/_ml/set_upgrade_mode",
	"/_ml/trained_models",
	"/_ml/trained_models/_stats",
	"/_ml/trained_models/{model_id}",
	"/_ml/trained_models/{model_id}/_stats",
	"/_ml/trained_models/{model_id}/definition/{part}",
	"/_ml/trained_models/{model_id}/deployment/_infer",
	"/_ml/trained_models/{model_id}/deployment/_start",
	"/_ml/trained_models/{model_id}/deployment/_stop",
	"/_ml/trained_models/{model_id}/model_aliases/{model_alias}",
	"/_ml/trained_models/{model_id}/vocabulary",
	"/_monitoring/bulk",
	"/_monitoring/{type}/bulk",
	"/_msearch",
	"/_msearch/template",
	"/_mtermvectors",
	"/_nodes",
	"/_nodes/hot_threads",
	"/_nodes/reload_secure_settings",
	"/_nodes/shutdown",
	"/_nodes/stats",
	"/_nodes/stats/{metric}",
	"/_nodes/stats/{metric}/{index_metric}",
	"/_nodes/usage",
	"/_nodes/usage/{metric}",
	"/_nodes/{metric}",
	"/_nodes/{node_id}",
	"/_nodes/{node_id}/_repositories_metering",
	"/_nodes/{node_id}/_repositories_metering/{max_archive_version}",
	"/_nodes/{node_id}/hot_threads",
	"/_nodes/{node_id}/reload_secure_settings",
	"/_nodes/{node_id}/shutdown",
	"/_nodes/{node_id}/stats",
	"/_nodes/{node_id}/stats/{metric}",
	"/_nodes/{node_id}/stats/{metric}/{index_metric}",
	"/_nodes/{node_id}/usage",
	"/_nodes/{node_id}/usage/{metric}",
	"/_nodes/{node_id}/{metric}",
	"/_pit",
	"/_rank_eval",
	"/_recovery",
	"/_refresh",
	"/_reindex",
	"/_reindex/{task_id}/_rethrottle",
	"/_remote/info",
	"/_render/template",
	"/_render/template/{id}",
	"/_resolve/index/{name}",
	"/_rollup/data/",
	"/_rollup/data/{id}",
	"/_rollup/job/",
	"/_rollup/job/{id}",
	"/_rollup/job/{id}/_start",
	"/_rollup/job/{id}/_stop",
	"/_script_context",
	"/_script_language",
	"/_scripts/painless/_execute",
	"/_scripts/{id}",
	"/_scripts/{id}/{context}",
	"/_search",
	"/_search/scroll",
	"/_search/scroll/{scroll_id}",
	"/_search/template",
	"/_search_shards",
	"/_searchable_snapshots/cache/clear",
	"/_searchable_snapshots/cache/stats",
	"/_searchable_snapshots/stats",
	"/_searchable_snapshots/{node_id}/cache/stats",
	"/_security/_authenticate",
	"/_security/_query/api_key",
	"/_security/api_key",
	"/_security/api_key/grant",
	"/_security/api_key/{ids}/_clear_cache",
	"/_security/enroll/kibana",
	"/_security/enroll/node",
	"/_security/oauth2/token",
	"/_security/privilege",
	"/_security/privilege/",
	"/_security/privilege/_builtin",
	"/_security/privilege/{application}",
	"/_security/privilege/{application}/_clear_cache",
	"/_security/privilege/{application}/{name}",
	"/_security/realm/{realms}/_clear_cache",
	"/_security/role",
	"/_security/role/{name}",
	"/_security/role/{name}/_clear_cache",
	"/_security/role_mapping",
	"/_security/role_mapping/{name}",
	"/_security/saml/authenticate",
	"/_security/saml/complete_logout",
	"/_security/saml/invalidate",
	"/_security/saml/logout",
	"/_security/saml/metadata/{realm_name}",
	"/_security/saml/prepare",
	"/_security/service",
	"/_security/service/{namespace}",
	"/_security/service/{namespace}/{service}",
	"/_security/service/{namespace}/{service}/credential",
	"/_security/service/{namespace}/{service}/credential/token",
	"/_security/service/{namespace}/{service}/credential/token/{name}",
	"/_security/service/{namespace}/{service}/credential/token/{name}/_clear_cache",
	"/_security/user",
	"/_security/user/_has_privileges",
	"/_security/user/_password",
	"/_security/user/_privileges",
	"/_security/user/{username}",
	"/_security/user/{username}/_disable",
	"/_security/user/{username}/_enable",
	"/_security/user/{username}/_password",
	"/_security/user/{user}/_has_privileges",
	"/_segments",
	"/_settings",
	"/_settings/{name}",
	"/_shard_stores",
	"/_slm/_execute_retention",
	"/_slm/policy",
	"/_slm/policy/{policy_id}",
	"/_slm/policy/{policy_id}/_execute",
	"/_slm/start",
	"/_slm/stats",
	"/_slm/status",
	"/_slm/stop",
	"/_snapshot",
	"/_snapshot/_status",
	"/_snapshot/{repository}",
	"/_snapshot/{repository}/_analyze",
	"/_snapshot/{repository}/_cleanup",
	"/_snapshot/{repository}/_status",
	"/_snapshot/{repository}/_verify",
	"/_snapshot/{repository}/{snapshot}",
	"/_snapshot/{repository}/{snapshot}/_clone/{target_snapshot}",
	"/_snapshot/{repository}/{snapshot}/_mount",
	"/_snapshot/{repository}/{snapshot}/_restore",
	"/_snapshot/{repository}/{snapshot}/_status",
	"/_sql",
	"/_sql/async/delete/{id}",
	"/_sql/async/status/{id}",
	"/_sql/async/{id}",
	"/_sql/close",
	"/_sql/translate",
	"/_ssl/certificates",
	"/_stats",
	"/_stats/{metric}",
	"/_tasks",
	"/_tasks/_cancel",
	"/_tasks/{task_id}",
	"/_tasks/{task_id}/_cancel",
	"/_template",
	"/_template/{name}",
	"/_text_structure/find_structure",
	"/_transform",
	"/_transform/_preview",
	"/_transform/_upgrade",
	"/_transform/{transform_id}",
	"/_transform/{transform_id}/_preview",
	"/_transform/{transform_id}/_reset",
	"/_transform/{transform_id}/_start",
	"/_transform/{transform_id}/_stats",
	"/_transform/{transform_id}/_stop",
	"/_transform/{transform_id}/_update",
	"/_update_by_query/{task_id}/_rethrottle",
	"/_validate/query",
	"/_watcher/_query/watches",
	"/_watcher/_start",
	"/_watcher/_stop",
	"/_watcher/stats",
	"/_watcher/stats/{metric}",
	"/_watcher/watch/_execute",
	"/_watcher/watch/{id}",
	"/_watcher/watch/{id}/_execute",
	"/_watcher/watch/{watch_id}/_ack",
	"/_watcher/watch/{watch_id}/_ack/{action_id}",
	"/_watcher/watch/{watch_id}/_activate",
	"/_watcher/watch/{watch_id}/_deactivate",
	"/_xpack",
	"/_xpack/usage",
	"/{alias}/_rollover",
	"/{alias}/_rollover/{new_index}",
	"/{index}",
	"/{index}/_alias",
	"/{index}/_alias/{name}",
	"/{index}/_aliases/{name}",
	"/{index}/_analyze",
	"/{index}/_async_search",
	"/{index}/_block/{block}",
	"/{index}/_bulk",
	"/{index}/_cache/clear",
	"/{index}/_ccr/follow",
	"/{index}/_ccr/forget_follower",
	"/{index}/_ccr/info",
	"/{index}/_ccr/pause_follow",
	"/{index}/_ccr/resume_follow",
	"/{index}/_ccr/stats",
	"/{index}/_ccr/unfollow",
	"/{index}/_clone/{target}",
	"/{index}/_close",
	"/{index}/_count",
	"/{index}/_create/{id}",
	"/{index}/_delete_by_query",
	"/{index}/_disk_usage",
	"/{index}/_doc",
	"/{index}/_doc/{id}",
	"/{index}/_eql/search",
	"/{index}/_explain/{id}",
	"/{index}/_field_caps",
	"/{index}/_field_usage_stats",
	"/{index}/_fleet/_fleet_msearch",
	"/{index}/_fleet/_fleet_search",
	"/{index}/_fleet/global_checkpoints",
	"/{index}/_flush",
	"/{index}/_forcemerge",
	"/{index}/_graph/explore",
	"/{index}/_ilm/explain",
	"/{index}/_ilm/remove",
	"/{index}/_ilm/retry",
	"/{index}/_knn_search",
	"/{index}/_mapping",
	"/{index}/_mapping/field/{fields}",
	"/{index}/_mget",
	"/{index}/_migration/deprecations",
	"/{index}/_msearch",
	"/{index}/_msearch/template",
	"/{index}/_mtermvectors",
	"/{index}/_mvt/{field}/{zoom}/{x}/{y}",
	"/{index}/_open",
	"/{index}/_pit",
	"/{index}/_rank_eval",
	"/{index}/_recovery",
	"/{index}/_refresh",
	"/{index}/_reload_search_analyzers",
	"/{index}/_rollup/data",
	"/{index}/_rollup/{rollup_index}",
	"/{index}/_rollup_search",
	"/{index}/_search",
	"/{index}/_search/template",
	"/{index}/_search_shards",
	"/{index}/_searchable_snapshots/cache/clear",
	"/{index}/_searchable_snapshots/stats",
	"/{index}/_segments",
	"/{index}/_settings",
	"/{index}/_settings/{name}",
	"/{index}/_shard_stores",
	"/{index}/_shrink/{target}",
	"/{index}/_source/{id}",
	"/{index}/_split/{target}",
	"/{index}/_stats",
	"/{index}/_stats/{metric}",
	"/{index}/_terms_enum",
	"/{index}/_termvectors",
	"/{index}/_termvectors/{id}",
	"/{index}/_unfreeze",
	"/{index}/_update/{id}",
	"/{index}/_update_by_query",
	"/{index}/_validate/query",
}

// tokenRegexp matches parts of a path that have a token wildcard.
var tokenRegexp = regexp.MustCompile(`{[^}]+}`)

// segment returns the first segment of path from start to the next occurance
// of the path separator and the ending index. When the final segment of the
// path is reached, the returned index is -1. This is all done without any
// allocations to the heap (something strings.Split would not accomplish).
func segment(path string, start int) (string, int) {
	if len(path) == 0 || start < 0 || start > len(path)-1 {
		return "", -1
	}
	end := strings.IndexRune(path[start+1:], '/')
	if end == -1 {
		return path[start:], -1
	}
	return path[start : start+end+1], start + end + 1
}

type node struct {
	value    string
	children map[string]*node
	wildcard *node
}

func newNode() *node {
	return &node{children: make(map[string]*node)}
}

func (root *node) Insert(key, value string) {
	if root == nil {
		n := newNode()
		*root = *n
	}

	n := root
	for part, i := segment(key, 0); part != ""; part, i = segment(key, i) {
		//iter := pathIterator(key)
		//for part := iter(); part != ""; part = iter() {
		var child *node
		if tokenRegexp.Match([]byte(part)) {
			child = n.wildcard
			if child == nil {
				child = newNode()
				n.wildcard = child
			}
		} else {
			var ok bool
			child, ok = n.children[part]
			if !ok {
				child = newNode()
				n.children[part] = child
			}
		}
		n = child
	}

	n.value = value
}

func (root *node) Get(key string) string {
	if root == nil {
		return ""
	}

	n := root
	for part, i := segment(key, 0); part != ""; part, i = segment(key, i) {
		//iter := pathIterator(key)
		//for part := iter(); part != ""; part = iter() {
		if child := n.children[part]; child != nil {
			n = child
		} else if n.wildcard != nil {
			n = n.wildcard
		} else {
			// Unknown path.
			return ""
		}
	}

	return n.value
}

func build() *node {
	// Fix overlap of `/_node/{metric}` and `/_node/{node_id}`. The {metric}
	// value has a limited set of known values it can be. Replace the token
	// with all known values to resolve the overlap.
	const nodesMetricPath = "/_nodes/{metric}"
	expanded := []string{
		"/_nodes/aggregations",
		"/_nodes/http",
		"/_nodes/indices",
		"/_nodes/ingest",
		"/_nodes/jvm",
		"/_nodes/os",
		"/_nodes/plugins",
		"/_nodes/process",
		"/_nodes/settings",
		"/_nodes/thread_pool",
		"/_nodes/transport",
		"/_nodes/_all",
		"/_nodes/_none",
	}
	i := sort.SearchStrings(paths, nodesMetricPath)
	fixed := make([]string, len(paths)+len(expanded)-1)
	copy(fixed[:i], paths[:i])
	copy(fixed[i:i+len(expanded)], expanded)
	copy(fixed[i+len(expanded):], paths[i+1:])

	root := newNode()
	for _, path := range fixed {
		root.Insert(path, path)
	}

	// Update all the expanded `/_node/{metric}` values to include the token.
	for _, s := range expanded {
		root.Insert(s, nodesMetricPath)
	}

	return root
}

var elasticsearchPathTrie = build()

// tokenize returns the tokenized form of path if it is known, otherwise an
// empty string is returned.
func tokenize(path string) string {
	return elasticsearchPathTrie.Get(path)
}
