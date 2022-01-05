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
	"fmt"
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

var pathMatcher = []pathRegex{
	{regexp.MustCompile("^/$"), "/"},
	{regexp.MustCompile("^/_alias$"), "/_alias"},
	{regexp.MustCompile("^/_alias/[^/]+$"), "/_alias/{name}"},
	{regexp.MustCompile("^/_aliases$"), "/_aliases"},
	{regexp.MustCompile("^/_analyze$"), "/_analyze"},
	{regexp.MustCompile("^/_async_search$"), "/_async_search"},
	{regexp.MustCompile("^/_async_search/status/[^/]+$"), "/_async_search/status/{id}"},
	{regexp.MustCompile("^/_async_search/[^/]+$"), "/_async_search/{id}"},
	{regexp.MustCompile("^/_autoscaling/capacity$"), "/_autoscaling/capacity"},
	{regexp.MustCompile("^/_autoscaling/policy/[^/]+$"), "/_autoscaling/policy/{name}"},
	{regexp.MustCompile("^/_bulk$"), "/_bulk"},
	{regexp.MustCompile("^/_cache/clear$"), "/_cache/clear"},
	{regexp.MustCompile("^/_cat$"), "/_cat"},
	{regexp.MustCompile("^/_cat/aliases$"), "/_cat/aliases"},
	{regexp.MustCompile("^/_cat/aliases/[^/]+$"), "/_cat/aliases/{name}"},
	{regexp.MustCompile("^/_cat/allocation$"), "/_cat/allocation"},
	{regexp.MustCompile("^/_cat/allocation/[^/]+$"), "/_cat/allocation/{node_id}"},
	{regexp.MustCompile("^/_cat/count$"), "/_cat/count"},
	{regexp.MustCompile("^/_cat/count/[^/]+$"), "/_cat/count/{index}"},
	{regexp.MustCompile("^/_cat/fielddata$"), "/_cat/fielddata"},
	{regexp.MustCompile("^/_cat/fielddata/[^/]+$"), "/_cat/fielddata/{fields}"},
	{regexp.MustCompile("^/_cat/health$"), "/_cat/health"},
	{regexp.MustCompile("^/_cat/indices$"), "/_cat/indices"},
	{regexp.MustCompile("^/_cat/indices/[^/]+$"), "/_cat/indices/{index}"},
	{regexp.MustCompile("^/_cat/master$"), "/_cat/master"},
	{regexp.MustCompile("^/_cat/ml/anomaly_detectors$"), "/_cat/ml/anomaly_detectors"},
	{regexp.MustCompile("^/_cat/ml/anomaly_detectors/[^/]+$"), "/_cat/ml/anomaly_detectors/{job_id}"},
	{regexp.MustCompile("^/_cat/ml/data_frame/analytics$"), "/_cat/ml/data_frame/analytics"},
	{regexp.MustCompile("^/_cat/ml/data_frame/analytics/[^/]+$"), "/_cat/ml/data_frame/analytics/{id}"},
	{regexp.MustCompile("^/_cat/ml/datafeeds$"), "/_cat/ml/datafeeds"},
	{regexp.MustCompile("^/_cat/ml/datafeeds/[^/]+$"), "/_cat/ml/datafeeds/{datafeed_id}"},
	{regexp.MustCompile("^/_cat/ml/trained_models$"), "/_cat/ml/trained_models"},
	{regexp.MustCompile("^/_cat/ml/trained_models/[^/]+$"), "/_cat/ml/trained_models/{model_id}"},
	{regexp.MustCompile("^/_cat/nodeattrs$"), "/_cat/nodeattrs"},
	{regexp.MustCompile("^/_cat/nodes$"), "/_cat/nodes"},
	{regexp.MustCompile("^/_cat/pending_tasks$"), "/_cat/pending_tasks"},
	{regexp.MustCompile("^/_cat/plugins$"), "/_cat/plugins"},
	{regexp.MustCompile("^/_cat/recovery$"), "/_cat/recovery"},
	{regexp.MustCompile("^/_cat/recovery/[^/]+$"), "/_cat/recovery/{index}"},
	{regexp.MustCompile("^/_cat/repositories$"), "/_cat/repositories"},
	{regexp.MustCompile("^/_cat/segments$"), "/_cat/segments"},
	{regexp.MustCompile("^/_cat/segments/[^/]+$"), "/_cat/segments/{index}"},
	{regexp.MustCompile("^/_cat/shards$"), "/_cat/shards"},
	{regexp.MustCompile("^/_cat/shards/[^/]+$"), "/_cat/shards/{index}"},
	{regexp.MustCompile("^/_cat/snapshots$"), "/_cat/snapshots"},
	{regexp.MustCompile("^/_cat/snapshots/[^/]+$"), "/_cat/snapshots/{repository}"},
	{regexp.MustCompile("^/_cat/tasks$"), "/_cat/tasks"},
	{regexp.MustCompile("^/_cat/templates$"), "/_cat/templates"},
	{regexp.MustCompile("^/_cat/templates/[^/]+$"), "/_cat/templates/{name}"},
	{regexp.MustCompile("^/_cat/thread_pool$"), "/_cat/thread_pool"},
	{regexp.MustCompile("^/_cat/thread_pool/[^/]+$"), "/_cat/thread_pool/{thread_pool_patterns}"},
	{regexp.MustCompile("^/_cat/transforms$"), "/_cat/transforms"},
	{regexp.MustCompile("^/_cat/transforms/[^/]+$"), "/_cat/transforms/{transform_id}"},
	{regexp.MustCompile("^/_ccr/auto_follow$"), "/_ccr/auto_follow"},
	{regexp.MustCompile("^/_ccr/auto_follow/[^/]+$"), "/_ccr/auto_follow/{name}"},
	{regexp.MustCompile("^/_ccr/auto_follow/[^/]+/pause$"), "/_ccr/auto_follow/{name}/pause"},
	{regexp.MustCompile("^/_ccr/auto_follow/[^/]+/resume$"), "/_ccr/auto_follow/{name}/resume"},
	{regexp.MustCompile("^/_ccr/stats$"), "/_ccr/stats"},
	{regexp.MustCompile("^/_cluster/allocation/explain$"), "/_cluster/allocation/explain"},
	{regexp.MustCompile("^/_cluster/health$"), "/_cluster/health"},
	{regexp.MustCompile("^/_cluster/health/[^/]+$"), "/_cluster/health/{index}"},
	{regexp.MustCompile("^/_cluster/pending_tasks$"), "/_cluster/pending_tasks"},
	{regexp.MustCompile("^/_cluster/reroute$"), "/_cluster/reroute"},
	{regexp.MustCompile("^/_cluster/settings$"), "/_cluster/settings"},
	{regexp.MustCompile("^/_cluster/state$"), "/_cluster/state"},
	{regexp.MustCompile("^/_cluster/state/[^/]+$"), "/_cluster/state/{metric}"},
	{regexp.MustCompile("^/_cluster/state/[^/]+/[^/]+$"), "/_cluster/state/{metric}/{index}"},
	{regexp.MustCompile("^/_cluster/stats$"), "/_cluster/stats"},
	{regexp.MustCompile("^/_cluster/stats/nodes/[^/]+$"), "/_cluster/stats/nodes/{node_id}"},
	{regexp.MustCompile("^/_cluster/voting_config_exclusions$"), "/_cluster/voting_config_exclusions"},
	{regexp.MustCompile("^/_component_template$"), "/_component_template"},
	{regexp.MustCompile("^/_component_template/[^/]+$"), "/_component_template/{name}"},
	{regexp.MustCompile("^/_count$"), "/_count"},
	{regexp.MustCompile("^/_dangling$"), "/_dangling"},
	{regexp.MustCompile("^/_dangling/[^/]+$"), "/_dangling/{index_uuid}"},
	{regexp.MustCompile("^/_data_stream$"), "/_data_stream"},
	{regexp.MustCompile("^/_data_stream/_migrate/[^/]+$"), "/_data_stream/_migrate/{name}"},
	{regexp.MustCompile("^/_data_stream/_modify$"), "/_data_stream/_modify"},
	{regexp.MustCompile("^/_data_stream/_promote/[^/]+$"), "/_data_stream/_promote/{name}"},
	{regexp.MustCompile("^/_data_stream/_stats$"), "/_data_stream/_stats"},
	{regexp.MustCompile("^/_data_stream/[^/]+$"), "/_data_stream/{name}"},
	{regexp.MustCompile("^/_data_stream/[^/]+/_stats$"), "/_data_stream/{name}/_stats"},
	{regexp.MustCompile("^/_delete_by_query/[^/]+/_rethrottle$"), "/_delete_by_query/{task_id}/_rethrottle"},
	{regexp.MustCompile("^/_enrich/_stats$"), "/_enrich/_stats"},
	{regexp.MustCompile("^/_enrich/policy$"), "/_enrich/policy"},
	{regexp.MustCompile("^/_enrich/policy/[^/]+$"), "/_enrich/policy/{name}"},
	{regexp.MustCompile("^/_enrich/policy/[^/]+/_execute$"), "/_enrich/policy/{name}/_execute"},
	{regexp.MustCompile("^/_eql/search/status/[^/]+$"), "/_eql/search/status/{id}"},
	{regexp.MustCompile("^/_eql/search/[^/]+$"), "/_eql/search/{id}"},
	{regexp.MustCompile("^/_features$"), "/_features"},
	{regexp.MustCompile("^/_features/_reset$"), "/_features/_reset"},
	{regexp.MustCompile("^/_field_caps$"), "/_field_caps"},
	{regexp.MustCompile("^/_fleet/_fleet_msearch$"), "/_fleet/_fleet_msearch"},
	{regexp.MustCompile("^/_flush$"), "/_flush"},
	{regexp.MustCompile("^/_forcemerge$"), "/_forcemerge"},
	{regexp.MustCompile("^/_ilm/migrate_to_data_tiers$"), "/_ilm/migrate_to_data_tiers"},
	{regexp.MustCompile("^/_ilm/move/[^/]+$"), "/_ilm/move/{index}"},
	{regexp.MustCompile("^/_ilm/policy$"), "/_ilm/policy"},
	{regexp.MustCompile("^/_ilm/policy/[^/]+$"), "/_ilm/policy/{policy}"},
	{regexp.MustCompile("^/_ilm/start$"), "/_ilm/start"},
	{regexp.MustCompile("^/_ilm/status$"), "/_ilm/status"},
	{regexp.MustCompile("^/_ilm/stop$"), "/_ilm/stop"},
	{regexp.MustCompile("^/_index_template$"), "/_index_template"},
	{regexp.MustCompile("^/_index_template/_simulate$"), "/_index_template/_simulate"},
	{regexp.MustCompile("^/_index_template/_simulate/[^/]+$"), "/_index_template/_simulate/{name}"},
	{regexp.MustCompile("^/_index_template/_simulate_index/[^/]+$"), "/_index_template/_simulate_index/{name}"},
	{regexp.MustCompile("^/_index_template/[^/]+$"), "/_index_template/{name}"},
	{regexp.MustCompile("^/_ingest/geoip/stats$"), "/_ingest/geoip/stats"},
	{regexp.MustCompile("^/_ingest/pipeline$"), "/_ingest/pipeline"},
	{regexp.MustCompile("^/_ingest/pipeline/_simulate$"), "/_ingest/pipeline/_simulate"},
	{regexp.MustCompile("^/_ingest/pipeline/[^/]+$"), "/_ingest/pipeline/{id}"},
	{regexp.MustCompile("^/_ingest/pipeline/[^/]+/_simulate$"), "/_ingest/pipeline/{id}/_simulate"},
	{regexp.MustCompile("^/_ingest/processor/grok$"), "/_ingest/processor/grok"},
	{regexp.MustCompile("^/_license$"), "/_license"},
	{regexp.MustCompile("^/_license/basic_status$"), "/_license/basic_status"},
	{regexp.MustCompile("^/_license/start_basic$"), "/_license/start_basic"},
	{regexp.MustCompile("^/_license/start_trial$"), "/_license/start_trial"},
	{regexp.MustCompile("^/_license/trial_status$"), "/_license/trial_status"},
	{regexp.MustCompile("^/_logstash/pipeline/[^/]+$"), "/_logstash/pipeline/{id}"},
	{regexp.MustCompile("^/_mapping$"), "/_mapping"},
	{regexp.MustCompile("^/_mapping/field/[^/]+$"), "/_mapping/field/{fields}"},
	{regexp.MustCompile("^/_mget$"), "/_mget"},
	{regexp.MustCompile("^/_migration/deprecations$"), "/_migration/deprecations"},
	{regexp.MustCompile("^/_migration/system_features$"), "/_migration/system_features"},
	{regexp.MustCompile("^/_ml/_delete_expired_data$"), "/_ml/_delete_expired_data"},
	{regexp.MustCompile("^/_ml/_delete_expired_data/[^/]+$"), "/_ml/_delete_expired_data/{job_id}"},
	{regexp.MustCompile("^/_ml/anomaly_detectors$"), "/_ml/anomaly_detectors"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/_estimate_model_memory$"), "/_ml/anomaly_detectors/_estimate_model_memory"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/_stats$"), "/_ml/anomaly_detectors/_stats"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/_validate$"), "/_ml/anomaly_detectors/_validate"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/_validate/detector$"), "/_ml/anomaly_detectors/_validate/detector"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+$"), "/_ml/anomaly_detectors/{job_id}"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_close$"), "/_ml/anomaly_detectors/{job_id}/_close"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_data$"), "/_ml/anomaly_detectors/{job_id}/_data"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_flush$"), "/_ml/anomaly_detectors/{job_id}/_flush"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_forecast$"), "/_ml/anomaly_detectors/{job_id}/_forecast"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_forecast/[^/]+$"), "/_ml/anomaly_detectors/{job_id}/_forecast/{forecast_id}"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_open$"), "/_ml/anomaly_detectors/{job_id}/_open"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_reset$"), "/_ml/anomaly_detectors/{job_id}/_reset"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_stats$"), "/_ml/anomaly_detectors/{job_id}/_stats"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/_update$"), "/_ml/anomaly_detectors/{job_id}/_update"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/model_snapshots$"), "/_ml/anomaly_detectors/{job_id}/model_snapshots"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/model_snapshots/[^/]+$"), "/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/model_snapshots/[^/]+/_revert$"), "/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_revert"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/model_snapshots/[^/]+/_update$"), "/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_update"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/model_snapshots/[^/]+/_upgrade$"), "/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_upgrade"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/model_snapshots/[^/]+/_upgrade/_stats$"), "/_ml/anomaly_detectors/{job_id}/model_snapshots/{snapshot_id}/_upgrade/_stats"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/buckets$"), "/_ml/anomaly_detectors/{job_id}/results/buckets"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/buckets/[^/]+$"), "/_ml/anomaly_detectors/{job_id}/results/buckets/{timestamp}"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/categories/$"), "/_ml/anomaly_detectors/{job_id}/results/categories/"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/categories/[^/]+$"), "/_ml/anomaly_detectors/{job_id}/results/categories/{category_id}"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/influencers$"), "/_ml/anomaly_detectors/{job_id}/results/influencers"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/overall_buckets$"), "/_ml/anomaly_detectors/{job_id}/results/overall_buckets"},
	{regexp.MustCompile("^/_ml/anomaly_detectors/[^/]+/results/records$"), "/_ml/anomaly_detectors/{job_id}/results/records"},
	{regexp.MustCompile("^/_ml/calendars$"), "/_ml/calendars"},
	{regexp.MustCompile("^/_ml/calendars/[^/]+$"), "/_ml/calendars/{calendar_id}"},
	{regexp.MustCompile("^/_ml/calendars/[^/]+/events$"), "/_ml/calendars/{calendar_id}/events"},
	{regexp.MustCompile("^/_ml/calendars/[^/]+/events/[^/]+$"), "/_ml/calendars/{calendar_id}/events/{event_id}"},
	{regexp.MustCompile("^/_ml/calendars/[^/]+/jobs/[^/]+$"), "/_ml/calendars/{calendar_id}/jobs/{job_id}"},
	{regexp.MustCompile("^/_ml/data_frame/_evaluate$"), "/_ml/data_frame/_evaluate"},
	{regexp.MustCompile("^/_ml/data_frame/analytics$"), "/_ml/data_frame/analytics"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/_explain$"), "/_ml/data_frame/analytics/_explain"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/_preview$"), "/_ml/data_frame/analytics/_preview"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/_stats$"), "/_ml/data_frame/analytics/_stats"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+$"), "/_ml/data_frame/analytics/{id}"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+/_explain$"), "/_ml/data_frame/analytics/{id}/_explain"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+/_preview$"), "/_ml/data_frame/analytics/{id}/_preview"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+/_start$"), "/_ml/data_frame/analytics/{id}/_start"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+/_stats$"), "/_ml/data_frame/analytics/{id}/_stats"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+/_stop$"), "/_ml/data_frame/analytics/{id}/_stop"},
	{regexp.MustCompile("^/_ml/data_frame/analytics/[^/]+/_update$"), "/_ml/data_frame/analytics/{id}/_update"},
	{regexp.MustCompile("^/_ml/datafeeds$"), "/_ml/datafeeds"},
	{regexp.MustCompile("^/_ml/datafeeds/_preview$"), "/_ml/datafeeds/_preview"},
	{regexp.MustCompile("^/_ml/datafeeds/_stats$"), "/_ml/datafeeds/_stats"},
	{regexp.MustCompile("^/_ml/datafeeds/[^/]+$"), "/_ml/datafeeds/{datafeed_id}"},
	{regexp.MustCompile("^/_ml/datafeeds/[^/]+/_preview$"), "/_ml/datafeeds/{datafeed_id}/_preview"},
	{regexp.MustCompile("^/_ml/datafeeds/[^/]+/_start$"), "/_ml/datafeeds/{datafeed_id}/_start"},
	{regexp.MustCompile("^/_ml/datafeeds/[^/]+/_stats$"), "/_ml/datafeeds/{datafeed_id}/_stats"},
	{regexp.MustCompile("^/_ml/datafeeds/[^/]+/_stop$"), "/_ml/datafeeds/{datafeed_id}/_stop"},
	{regexp.MustCompile("^/_ml/datafeeds/[^/]+/_update$"), "/_ml/datafeeds/{datafeed_id}/_update"},
	{regexp.MustCompile("^/_ml/filters$"), "/_ml/filters"},
	{regexp.MustCompile("^/_ml/filters/[^/]+$"), "/_ml/filters/{filter_id}"},
	{regexp.MustCompile("^/_ml/filters/[^/]+/_update$"), "/_ml/filters/{filter_id}/_update"},
	{regexp.MustCompile("^/_ml/info$"), "/_ml/info"},
	{regexp.MustCompile("^/_ml/set_upgrade_mode$"), "/_ml/set_upgrade_mode"},
	{regexp.MustCompile("^/_ml/trained_models$"), "/_ml/trained_models"},
	{regexp.MustCompile("^/_ml/trained_models/_stats$"), "/_ml/trained_models/_stats"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+$"), "/_ml/trained_models/{model_id}"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/_stats$"), "/_ml/trained_models/{model_id}/_stats"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/definition/[^/]+$"), "/_ml/trained_models/{model_id}/definition/{part}"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/deployment/_infer$"), "/_ml/trained_models/{model_id}/deployment/_infer"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/deployment/_start$"), "/_ml/trained_models/{model_id}/deployment/_start"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/deployment/_stop$"), "/_ml/trained_models/{model_id}/deployment/_stop"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/model_aliases/[^/]+$"), "/_ml/trained_models/{model_id}/model_aliases/{model_alias}"},
	{regexp.MustCompile("^/_ml/trained_models/[^/]+/vocabulary$"), "/_ml/trained_models/{model_id}/vocabulary"},
	{regexp.MustCompile("^/_monitoring/bulk$"), "/_monitoring/bulk"},
	{regexp.MustCompile("^/_monitoring/[^/]+/bulk$"), "/_monitoring/{type}/bulk"},
	{regexp.MustCompile("^/_msearch$"), "/_msearch"},
	{regexp.MustCompile("^/_msearch/template$"), "/_msearch/template"},
	{regexp.MustCompile("^/_mtermvectors$"), "/_mtermvectors"},
	{regexp.MustCompile("^/_nodes$"), "/_nodes"},
	{regexp.MustCompile("^/_nodes/hot_threads$"), "/_nodes/hot_threads"},
	{regexp.MustCompile("^/_nodes/reload_secure_settings$"), "/_nodes/reload_secure_settings"},
	{regexp.MustCompile("^/_nodes/shutdown$"), "/_nodes/shutdown"},
	{regexp.MustCompile("^/_nodes/stats$"), "/_nodes/stats"},
	{regexp.MustCompile("^/_nodes/stats/[^/]+$"), "/_nodes/stats/{metric}"},
	{regexp.MustCompile("^/_nodes/stats/[^/]+/[^/]+$"), "/_nodes/stats/{metric}/{index_metric}"},
	{regexp.MustCompile("^/_nodes/usage$"), "/_nodes/usage"},
	{regexp.MustCompile("^/_nodes/usage/[^/]+$"), "/_nodes/usage/{metric}"},
	{regexp.MustCompile("^/_nodes/(?:aggregations|http|indices|ingest|jvm|os|plugins|process|settings|thread_pool|transport|_all|_none)$"), "/_nodes/{metric}"},
	{regexp.MustCompile("^/_nodes/[^/]+$"), "/_nodes/{node_id}"},
	{regexp.MustCompile("^/_nodes/[^/]+/_repositories_metering$"), "/_nodes/{node_id}/_repositories_metering"},
	{regexp.MustCompile("^/_nodes/[^/]+/_repositories_metering/[^/]+$"), "/_nodes/{node_id}/_repositories_metering/{max_archive_version}"},
	{regexp.MustCompile("^/_nodes/[^/]+/hot_threads$"), "/_nodes/{node_id}/hot_threads"},
	{regexp.MustCompile("^/_nodes/[^/]+/reload_secure_settings$"), "/_nodes/{node_id}/reload_secure_settings"},
	{regexp.MustCompile("^/_nodes/[^/]+/shutdown$"), "/_nodes/{node_id}/shutdown"},
	{regexp.MustCompile("^/_nodes/[^/]+/stats$"), "/_nodes/{node_id}/stats"},
	{regexp.MustCompile("^/_nodes/[^/]+/stats/[^/]+$"), "/_nodes/{node_id}/stats/{metric}"},
	{regexp.MustCompile("^/_nodes/[^/]+/stats/[^/]+/[^/]+$"), "/_nodes/{node_id}/stats/{metric}/{index_metric}"},
	{regexp.MustCompile("^/_nodes/[^/]+/usage$"), "/_nodes/{node_id}/usage"},
	{regexp.MustCompile("^/_nodes/[^/]+/usage/[^/]+$"), "/_nodes/{node_id}/usage/{metric}"},
	{regexp.MustCompile("^/_nodes/[^/]+/[^/]+$"), "/_nodes/{node_id}/{metric}"},
	{regexp.MustCompile("^/_pit$"), "/_pit"},
	{regexp.MustCompile("^/_rank_eval$"), "/_rank_eval"},
	{regexp.MustCompile("^/_recovery$"), "/_recovery"},
	{regexp.MustCompile("^/_refresh$"), "/_refresh"},
	{regexp.MustCompile("^/_reindex$"), "/_reindex"},
	{regexp.MustCompile("^/_reindex/[^/]+/_rethrottle$"), "/_reindex/{task_id}/_rethrottle"},
	{regexp.MustCompile("^/_remote/info$"), "/_remote/info"},
	{regexp.MustCompile("^/_render/template$"), "/_render/template"},
	{regexp.MustCompile("^/_render/template/[^/]+$"), "/_render/template/{id}"},
	{regexp.MustCompile("^/_resolve/index/[^/]+$"), "/_resolve/index/{name}"},
	{regexp.MustCompile("^/_rollup/data/$"), "/_rollup/data/"},
	{regexp.MustCompile("^/_rollup/data/[^/]+$"), "/_rollup/data/{id}"},
	{regexp.MustCompile("^/_rollup/job/$"), "/_rollup/job/"},
	{regexp.MustCompile("^/_rollup/job/[^/]+$"), "/_rollup/job/{id}"},
	{regexp.MustCompile("^/_rollup/job/[^/]+/_start$"), "/_rollup/job/{id}/_start"},
	{regexp.MustCompile("^/_rollup/job/[^/]+/_stop$"), "/_rollup/job/{id}/_stop"},
	{regexp.MustCompile("^/_script_context$"), "/_script_context"},
	{regexp.MustCompile("^/_script_language$"), "/_script_language"},
	{regexp.MustCompile("^/_scripts/painless/_execute$"), "/_scripts/painless/_execute"},
	{regexp.MustCompile("^/_scripts/[^/]+$"), "/_scripts/{id}"},
	{regexp.MustCompile("^/_scripts/[^/]+/[^/]+$"), "/_scripts/{id}/{context}"},
	{regexp.MustCompile("^/_search$"), "/_search"},
	{regexp.MustCompile("^/_search/scroll$"), "/_search/scroll"},
	{regexp.MustCompile("^/_search/scroll/[^/]+$"), "/_search/scroll/{scroll_id}"},
	{regexp.MustCompile("^/_search/template$"), "/_search/template"},
	{regexp.MustCompile("^/_search_shards$"), "/_search_shards"},
	{regexp.MustCompile("^/_searchable_snapshots/cache/clear$"), "/_searchable_snapshots/cache/clear"},
	{regexp.MustCompile("^/_searchable_snapshots/cache/stats$"), "/_searchable_snapshots/cache/stats"},
	{regexp.MustCompile("^/_searchable_snapshots/stats$"), "/_searchable_snapshots/stats"},
	{regexp.MustCompile("^/_searchable_snapshots/[^/]+/cache/stats$"), "/_searchable_snapshots/{node_id}/cache/stats"},
	{regexp.MustCompile("^/_security/_authenticate$"), "/_security/_authenticate"},
	{regexp.MustCompile("^/_security/_query/api_key$"), "/_security/_query/api_key"},
	{regexp.MustCompile("^/_security/api_key$"), "/_security/api_key"},
	{regexp.MustCompile("^/_security/api_key/grant$"), "/_security/api_key/grant"},
	{regexp.MustCompile("^/_security/api_key/[^/]+/_clear_cache$"), "/_security/api_key/{ids}/_clear_cache"},
	{regexp.MustCompile("^/_security/enroll/kibana$"), "/_security/enroll/kibana"},
	{regexp.MustCompile("^/_security/enroll/node$"), "/_security/enroll/node"},
	{regexp.MustCompile("^/_security/oauth2/token$"), "/_security/oauth2/token"},
	{regexp.MustCompile("^/_security/privilege$"), "/_security/privilege"},
	{regexp.MustCompile("^/_security/privilege/$"), "/_security/privilege/"},
	{regexp.MustCompile("^/_security/privilege/_builtin$"), "/_security/privilege/_builtin"},
	{regexp.MustCompile("^/_security/privilege/[^/]+$"), "/_security/privilege/{application}"},
	{regexp.MustCompile("^/_security/privilege/[^/]+/_clear_cache$"), "/_security/privilege/{application}/_clear_cache"},
	{regexp.MustCompile("^/_security/privilege/[^/]+/[^/]+$"), "/_security/privilege/{application}/{name}"},
	{regexp.MustCompile("^/_security/realm/[^/]+/_clear_cache$"), "/_security/realm/{realms}/_clear_cache"},
	{regexp.MustCompile("^/_security/role$"), "/_security/role"},
	{regexp.MustCompile("^/_security/role/[^/]+$"), "/_security/role/{name}"},
	{regexp.MustCompile("^/_security/role/[^/]+/_clear_cache$"), "/_security/role/{name}/_clear_cache"},
	{regexp.MustCompile("^/_security/role_mapping$"), "/_security/role_mapping"},
	{regexp.MustCompile("^/_security/role_mapping/[^/]+$"), "/_security/role_mapping/{name}"},
	{regexp.MustCompile("^/_security/saml/authenticate$"), "/_security/saml/authenticate"},
	{regexp.MustCompile("^/_security/saml/complete_logout$"), "/_security/saml/complete_logout"},
	{regexp.MustCompile("^/_security/saml/invalidate$"), "/_security/saml/invalidate"},
	{regexp.MustCompile("^/_security/saml/logout$"), "/_security/saml/logout"},
	{regexp.MustCompile("^/_security/saml/metadata/[^/]+$"), "/_security/saml/metadata/{realm_name}"},
	{regexp.MustCompile("^/_security/saml/prepare$"), "/_security/saml/prepare"},
	{regexp.MustCompile("^/_security/service$"), "/_security/service"},
	{regexp.MustCompile("^/_security/service/[^/]+$"), "/_security/service/{namespace}"},
	{regexp.MustCompile("^/_security/service/[^/]+/[^/]+$"), "/_security/service/{namespace}/{service}"},
	{regexp.MustCompile("^/_security/service/[^/]+/[^/]+/credential$"), "/_security/service/{namespace}/{service}/credential"},
	{regexp.MustCompile("^/_security/service/[^/]+/[^/]+/credential/token$"), "/_security/service/{namespace}/{service}/credential/token"},
	{regexp.MustCompile("^/_security/service/[^/]+/[^/]+/credential/token/[^/]+$"), "/_security/service/{namespace}/{service}/credential/token/{name}"},
	{regexp.MustCompile("^/_security/service/[^/]+/[^/]+/credential/token/[^/]+/_clear_cache$"), "/_security/service/{namespace}/{service}/credential/token/{name}/_clear_cache"},
	{regexp.MustCompile("^/_security/user$"), "/_security/user"},
	{regexp.MustCompile("^/_security/user/_has_privileges$"), "/_security/user/_has_privileges"},
	{regexp.MustCompile("^/_security/user/_password$"), "/_security/user/_password"},
	{regexp.MustCompile("^/_security/user/_privileges$"), "/_security/user/_privileges"},
	{regexp.MustCompile("^/_security/user/[^/]+$"), "/_security/user/{username}"},
	{regexp.MustCompile("^/_security/user/[^/]+/_disable$"), "/_security/user/{username}/_disable"},
	{regexp.MustCompile("^/_security/user/[^/]+/_enable$"), "/_security/user/{username}/_enable"},
	{regexp.MustCompile("^/_security/user/[^/]+/_password$"), "/_security/user/{username}/_password"},
	{regexp.MustCompile("^/_security/user/[^/]+/_has_privileges$"), "/_security/user/{user}/_has_privileges"},
	{regexp.MustCompile("^/_segments$"), "/_segments"},
	{regexp.MustCompile("^/_settings$"), "/_settings"},
	{regexp.MustCompile("^/_settings/[^/]+$"), "/_settings/{name}"},
	{regexp.MustCompile("^/_shard_stores$"), "/_shard_stores"},
	{regexp.MustCompile("^/_slm/_execute_retention$"), "/_slm/_execute_retention"},
	{regexp.MustCompile("^/_slm/policy$"), "/_slm/policy"},
	{regexp.MustCompile("^/_slm/policy/[^/]+$"), "/_slm/policy/{policy_id}"},
	{regexp.MustCompile("^/_slm/policy/[^/]+/_execute$"), "/_slm/policy/{policy_id}/_execute"},
	{regexp.MustCompile("^/_slm/start$"), "/_slm/start"},
	{regexp.MustCompile("^/_slm/stats$"), "/_slm/stats"},
	{regexp.MustCompile("^/_slm/status$"), "/_slm/status"},
	{regexp.MustCompile("^/_slm/stop$"), "/_slm/stop"},
	{regexp.MustCompile("^/_snapshot$"), "/_snapshot"},
	{regexp.MustCompile("^/_snapshot/_status$"), "/_snapshot/_status"},
	{regexp.MustCompile("^/_snapshot/[^/]+$"), "/_snapshot/{repository}"},
	{regexp.MustCompile("^/_snapshot/[^/]+/_analyze$"), "/_snapshot/{repository}/_analyze"},
	{regexp.MustCompile("^/_snapshot/[^/]+/_cleanup$"), "/_snapshot/{repository}/_cleanup"},
	{regexp.MustCompile("^/_snapshot/[^/]+/_status$"), "/_snapshot/{repository}/_status"},
	{regexp.MustCompile("^/_snapshot/[^/]+/_verify$"), "/_snapshot/{repository}/_verify"},
	{regexp.MustCompile("^/_snapshot/[^/]+/[^/]+$"), "/_snapshot/{repository}/{snapshot}"},
	{regexp.MustCompile("^/_snapshot/[^/]+/[^/]+/_clone/[^/]+$"), "/_snapshot/{repository}/{snapshot}/_clone/{target_snapshot}"},
	{regexp.MustCompile("^/_snapshot/[^/]+/[^/]+/_mount$"), "/_snapshot/{repository}/{snapshot}/_mount"},
	{regexp.MustCompile("^/_snapshot/[^/]+/[^/]+/_restore$"), "/_snapshot/{repository}/{snapshot}/_restore"},
	{regexp.MustCompile("^/_snapshot/[^/]+/[^/]+/_status$"), "/_snapshot/{repository}/{snapshot}/_status"},
	{regexp.MustCompile("^/_sql$"), "/_sql"},
	{regexp.MustCompile("^/_sql/async/delete/[^/]+$"), "/_sql/async/delete/{id}"},
	{regexp.MustCompile("^/_sql/async/status/[^/]+$"), "/_sql/async/status/{id}"},
	{regexp.MustCompile("^/_sql/async/[^/]+$"), "/_sql/async/{id}"},
	{regexp.MustCompile("^/_sql/close$"), "/_sql/close"},
	{regexp.MustCompile("^/_sql/translate$"), "/_sql/translate"},
	{regexp.MustCompile("^/_ssl/certificates$"), "/_ssl/certificates"},
	{regexp.MustCompile("^/_stats$"), "/_stats"},
	{regexp.MustCompile("^/_stats/[^/]+$"), "/_stats/{metric}"},
	{regexp.MustCompile("^/_tasks$"), "/_tasks"},
	{regexp.MustCompile("^/_tasks/_cancel$"), "/_tasks/_cancel"},
	{regexp.MustCompile("^/_tasks/[^/]+$"), "/_tasks/{task_id}"},
	{regexp.MustCompile("^/_tasks/[^/]+/_cancel$"), "/_tasks/{task_id}/_cancel"},
	{regexp.MustCompile("^/_template$"), "/_template"},
	{regexp.MustCompile("^/_template/[^/]+$"), "/_template/{name}"},
	{regexp.MustCompile("^/_text_structure/find_structure$"), "/_text_structure/find_structure"},
	{regexp.MustCompile("^/_transform$"), "/_transform"},
	{regexp.MustCompile("^/_transform/_preview$"), "/_transform/_preview"},
	{regexp.MustCompile("^/_transform/_upgrade$"), "/_transform/_upgrade"},
	{regexp.MustCompile("^/_transform/[^/]+$"), "/_transform/{transform_id}"},
	{regexp.MustCompile("^/_transform/[^/]+/_preview$"), "/_transform/{transform_id}/_preview"},
	{regexp.MustCompile("^/_transform/[^/]+/_reset$"), "/_transform/{transform_id}/_reset"},
	{regexp.MustCompile("^/_transform/[^/]+/_start$"), "/_transform/{transform_id}/_start"},
	{regexp.MustCompile("^/_transform/[^/]+/_stats$"), "/_transform/{transform_id}/_stats"},
	{regexp.MustCompile("^/_transform/[^/]+/_stop$"), "/_transform/{transform_id}/_stop"},
	{regexp.MustCompile("^/_transform/[^/]+/_update$"), "/_transform/{transform_id}/_update"},
	{regexp.MustCompile("^/_update_by_query/[^/]+/_rethrottle$"), "/_update_by_query/{task_id}/_rethrottle"},
	{regexp.MustCompile("^/_validate/query$"), "/_validate/query"},
	{regexp.MustCompile("^/_watcher/_query/watches$"), "/_watcher/_query/watches"},
	{regexp.MustCompile("^/_watcher/_start$"), "/_watcher/_start"},
	{regexp.MustCompile("^/_watcher/_stop$"), "/_watcher/_stop"},
	{regexp.MustCompile("^/_watcher/stats$"), "/_watcher/stats"},
	{regexp.MustCompile("^/_watcher/stats/[^/]+$"), "/_watcher/stats/{metric}"},
	{regexp.MustCompile("^/_watcher/watch/_execute$"), "/_watcher/watch/_execute"},
	{regexp.MustCompile("^/_watcher/watch/[^/]+$"), "/_watcher/watch/{id}"},
	{regexp.MustCompile("^/_watcher/watch/[^/]+/_execute$"), "/_watcher/watch/{id}/_execute"},
	{regexp.MustCompile("^/_watcher/watch/[^/]+/_ack$"), "/_watcher/watch/{watch_id}/_ack"},
	{regexp.MustCompile("^/_watcher/watch/[^/]+/_ack/[^/]+$"), "/_watcher/watch/{watch_id}/_ack/{action_id}"},
	{regexp.MustCompile("^/_watcher/watch/[^/]+/_activate$"), "/_watcher/watch/{watch_id}/_activate"},
	{regexp.MustCompile("^/_watcher/watch/[^/]+/_deactivate$"), "/_watcher/watch/{watch_id}/_deactivate"},
	{regexp.MustCompile("^/_xpack$"), "/_xpack"},
	{regexp.MustCompile("^/_xpack/usage$"), "/_xpack/usage"},
	{regexp.MustCompile("^/[^/]+/_rollover$"), "/{alias}/_rollover"},
	{regexp.MustCompile("^/[^/]+/_rollover/[^/]+$"), "/{alias}/_rollover/{new_index}"},
	{regexp.MustCompile("^/[^/]+$"), "/{index}"},
	{regexp.MustCompile("^/[^/]+/_alias$"), "/{index}/_alias"},
	{regexp.MustCompile("^/[^/]+/_alias/[^/]+$"), "/{index}/_alias/{name}"},
	{regexp.MustCompile("^/[^/]+/_aliases/[^/]+$"), "/{index}/_aliases/{name}"},
	{regexp.MustCompile("^/[^/]+/_analyze$"), "/{index}/_analyze"},
	{regexp.MustCompile("^/[^/]+/_async_search$"), "/{index}/_async_search"},
	{regexp.MustCompile("^/[^/]+/_block/[^/]+$"), "/{index}/_block/{block}"},
	{regexp.MustCompile("^/[^/]+/_bulk$"), "/{index}/_bulk"},
	{regexp.MustCompile("^/[^/]+/_cache/clear$"), "/{index}/_cache/clear"},
	{regexp.MustCompile("^/[^/]+/_ccr/follow$"), "/{index}/_ccr/follow"},
	{regexp.MustCompile("^/[^/]+/_ccr/forget_follower$"), "/{index}/_ccr/forget_follower"},
	{regexp.MustCompile("^/[^/]+/_ccr/info$"), "/{index}/_ccr/info"},
	{regexp.MustCompile("^/[^/]+/_ccr/pause_follow$"), "/{index}/_ccr/pause_follow"},
	{regexp.MustCompile("^/[^/]+/_ccr/resume_follow$"), "/{index}/_ccr/resume_follow"},
	{regexp.MustCompile("^/[^/]+/_ccr/stats$"), "/{index}/_ccr/stats"},
	{regexp.MustCompile("^/[^/]+/_ccr/unfollow$"), "/{index}/_ccr/unfollow"},
	{regexp.MustCompile("^/[^/]+/_clone/[^/]+$"), "/{index}/_clone/{target}"},
	{regexp.MustCompile("^/[^/]+/_close$"), "/{index}/_close"},
	{regexp.MustCompile("^/[^/]+/_count$"), "/{index}/_count"},
	{regexp.MustCompile("^/[^/]+/_create/[^/]+$"), "/{index}/_create/{id}"},
	{regexp.MustCompile("^/[^/]+/_delete_by_query$"), "/{index}/_delete_by_query"},
	{regexp.MustCompile("^/[^/]+/_disk_usage$"), "/{index}/_disk_usage"},
	{regexp.MustCompile("^/[^/]+/_doc$"), "/{index}/_doc"},
	{regexp.MustCompile("^/[^/]+/_doc/[^/]+$"), "/{index}/_doc/{id}"},
	{regexp.MustCompile("^/[^/]+/_eql/search$"), "/{index}/_eql/search"},
	{regexp.MustCompile("^/[^/]+/_explain/[^/]+$"), "/{index}/_explain/{id}"},
	{regexp.MustCompile("^/[^/]+/_field_caps$"), "/{index}/_field_caps"},
	{regexp.MustCompile("^/[^/]+/_field_usage_stats$"), "/{index}/_field_usage_stats"},
	{regexp.MustCompile("^/[^/]+/_fleet/_fleet_msearch$"), "/{index}/_fleet/_fleet_msearch"},
	{regexp.MustCompile("^/[^/]+/_fleet/_fleet_search$"), "/{index}/_fleet/_fleet_search"},
	{regexp.MustCompile("^/[^/]+/_fleet/global_checkpoints$"), "/{index}/_fleet/global_checkpoints"},
	{regexp.MustCompile("^/[^/]+/_flush$"), "/{index}/_flush"},
	{regexp.MustCompile("^/[^/]+/_forcemerge$"), "/{index}/_forcemerge"},
	{regexp.MustCompile("^/[^/]+/_graph/explore$"), "/{index}/_graph/explore"},
	{regexp.MustCompile("^/[^/]+/_ilm/explain$"), "/{index}/_ilm/explain"},
	{regexp.MustCompile("^/[^/]+/_ilm/remove$"), "/{index}/_ilm/remove"},
	{regexp.MustCompile("^/[^/]+/_ilm/retry$"), "/{index}/_ilm/retry"},
	{regexp.MustCompile("^/[^/]+/_knn_search$"), "/{index}/_knn_search"},
	{regexp.MustCompile("^/[^/]+/_mapping$"), "/{index}/_mapping"},
	{regexp.MustCompile("^/[^/]+/_mapping/field/[^/]+$"), "/{index}/_mapping/field/{fields}"},
	{regexp.MustCompile("^/[^/]+/_mget$"), "/{index}/_mget"},
	{regexp.MustCompile("^/[^/]+/_migration/deprecations$"), "/{index}/_migration/deprecations"},
	{regexp.MustCompile("^/[^/]+/_msearch$"), "/{index}/_msearch"},
	{regexp.MustCompile("^/[^/]+/_msearch/template$"), "/{index}/_msearch/template"},
	{regexp.MustCompile("^/[^/]+/_mtermvectors$"), "/{index}/_mtermvectors"},
	{regexp.MustCompile("^/[^/]+/_mvt/[^/]+/[^/]+/[^/]+/[^/]+$"), "/{index}/_mvt/{field}/{zoom}/{x}/{y}"},
	{regexp.MustCompile("^/[^/]+/_open$"), "/{index}/_open"},
	{regexp.MustCompile("^/[^/]+/_pit$"), "/{index}/_pit"},
	{regexp.MustCompile("^/[^/]+/_rank_eval$"), "/{index}/_rank_eval"},
	{regexp.MustCompile("^/[^/]+/_recovery$"), "/{index}/_recovery"},
	{regexp.MustCompile("^/[^/]+/_refresh$"), "/{index}/_refresh"},
	{regexp.MustCompile("^/[^/]+/_reload_search_analyzers$"), "/{index}/_reload_search_analyzers"},
	{regexp.MustCompile("^/[^/]+/_rollup/data$"), "/{index}/_rollup/data"},
	{regexp.MustCompile("^/[^/]+/_rollup/[^/]+$"), "/{index}/_rollup/{rollup_index}"},
	{regexp.MustCompile("^/[^/]+/_rollup_search$"), "/{index}/_rollup_search"},
	{regexp.MustCompile("^/[^/]+/_search$"), "/{index}/_search"},
	{regexp.MustCompile("^/[^/]+/_search/template$"), "/{index}/_search/template"},
	{regexp.MustCompile("^/[^/]+/_search_shards$"), "/{index}/_search_shards"},
	{regexp.MustCompile("^/[^/]+/_searchable_snapshots/cache/clear$"), "/{index}/_searchable_snapshots/cache/clear"},
	{regexp.MustCompile("^/[^/]+/_searchable_snapshots/stats$"), "/{index}/_searchable_snapshots/stats"},
	{regexp.MustCompile("^/[^/]+/_segments$"), "/{index}/_segments"},
	{regexp.MustCompile("^/[^/]+/_settings$"), "/{index}/_settings"},
	{regexp.MustCompile("^/[^/]+/_settings/[^/]+$"), "/{index}/_settings/{name}"},
	{regexp.MustCompile("^/[^/]+/_shard_stores$"), "/{index}/_shard_stores"},
	{regexp.MustCompile("^/[^/]+/_shrink/[^/]+$"), "/{index}/_shrink/{target}"},
	{regexp.MustCompile("^/[^/]+/_source/[^/]+$"), "/{index}/_source/{id}"},
	{regexp.MustCompile("^/[^/]+/_split/[^/]+$"), "/{index}/_split/{target}"},
	{regexp.MustCompile("^/[^/]+/_stats$"), "/{index}/_stats"},
	{regexp.MustCompile("^/[^/]+/_stats/[^/]+$"), "/{index}/_stats/{metric}"},
	{regexp.MustCompile("^/[^/]+/_terms_enum$"), "/{index}/_terms_enum"},
	{regexp.MustCompile("^/[^/]+/_termvectors$"), "/{index}/_termvectors"},
	{regexp.MustCompile("^/[^/]+/_termvectors/[^/]+$"), "/{index}/_termvectors/{id}"},
	{regexp.MustCompile("^/[^/]+/_unfreeze$"), "/{index}/_unfreeze"},
	{regexp.MustCompile("^/[^/]+/_update/[^/]+$"), "/{index}/_update/{id}"},
	{regexp.MustCompile("^/[^/]+/_update_by_query$"), "/{index}/_update_by_query"},
	{regexp.MustCompile("^/[^/]+/_validate/query$"), "/{index}/_validate/query"},
}

type pathRegex struct {
	re   *regexp.Regexp
	path string
}

var (
	// {metric} and {node_id} need to be distinguished. Replace {metric} with
	// all the known values it can be.
	metricRegexp      = regexp.MustCompile(`/_nodes/{metric}`)
	metricReplacement = []byte(`/_nodes/(?:aggregations|http|indices|ingest|jvm|os|plugins|process|settings|thread_pool|transport|_all|_none)`)

	tokenRegexp      = regexp.MustCompile(`{[^}]+}`)
	tokenReplacement = []byte(`[^/]+`)
)

func printPathRegexSlice() {
	fmt.Println("var pathMatcher = []pathRegex{")
	for _, p := range paths {
		reStr := metricRegexp.ReplaceAll([]byte(p), metricReplacement)
		reStr = tokenRegexp.ReplaceAll([]byte(reStr), tokenReplacement)
		re := regexp.MustCompile("^" + string(reStr) + "$")
		fmt.Printf("\t{regexp.MustCompile(%q), %q},\n", re, p)
	}
	fmt.Println("}")
}

func tokenizeFromSlice(path string) string {
	for _, pRe := range pathMatcher {
		if pRe.re.MatchString(path) {
			return pRe.path
		}
	}
	return ""
}

type node struct {
	value    string
	children map[string]*node
	wildcard *node
}

func newNode() *node {
	return &node{children: make(map[string]*node)}
}

func insert(n *node, key []string, value string) {
	for _, part := range key {
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

func findNode(n *node, key []string) *node {
	if n == nil {
		return nil
	}

	for _, part := range key {
		if child, ok := n.children[part]; ok {
			n = child
			continue
		}
		if n.wildcard != nil {
			n = n.wildcard
			continue
		}
		// Unknown path.
		return nil
	}
	return n
}

func find(n *node, key []string) string {
	n = findNode(n, key)
	if n == nil {
		return ""
	}
	return n.value
}

func findNoAlloc(n *node, key string) string {
	if n == nil {
		return ""
	}

	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
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
		insert(root, strings.Split(path, "/"), path)
	}

	// Update all the expanded `/_node/{metric}` values to include the token.
	for _, s := range expanded {
		n := findNode(root, strings.Split(s, "/"))
		n.value = nodesMetricPath
	}

	return root
}

var pathTrie = build()

func tokenizeFromTrie(path string) string {
	return find(pathTrie, strings.Split(path, "/"))
	//return findNoAlloc(pathTrie, path)
}

func pathSegmenter(path string, start int) (string, int) {
	if len(path) == 0 || start < 0 || start > len(path)-1 {
		return "", -1
	}
	end := strings.IndexRune(path[start+1:], '/')
	if end == -1 {
		return path[start:], -1
	}
	return path[start : start+end+1], start + end + 1
}

func pathIterator(path string) func() string {
	return func() string {
		if len(path) == 0 {
			return ""
		}

		// Assumes path has a '/' prefix.
		idx := strings.IndexRune(path[1:], '/')
		if idx == -1 {
			defer func() { path = "" }()
			return path
		}
		defer func() { path = path[idx+1:] }()
		return path[:idx+1]
	}
}

type noAllocNode struct {
	value    string
	children map[string]*noAllocNode
	wildcard *noAllocNode
}

func newNoAllocNode() *noAllocNode {
	return &noAllocNode{children: make(map[string]*noAllocNode)}
}

func (root *noAllocNode) Insert(key, value string) {
	if root == nil {
		*root = noAllocNode{children: make(map[string]*noAllocNode)}
	}

	n := root
	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
		//iter := pathIterator(key)
		//for part := iter(); part != ""; part = iter() {
		var child *noAllocNode
		if tokenRegexp.Match([]byte(part)) {
			child = n.wildcard
			if child == nil {
				child = newNoAllocNode()
				n.wildcard = child
			}
		} else {
			var ok bool
			child, ok = n.children[part]
			if !ok {
				child = newNoAllocNode()
				n.children[part] = child
			}
		}
		n = child
	}

	n.value = value
}

func (root *noAllocNode) Get(key string) string {
	if root == nil {
		return ""
	}

	n := root

	for part, i := pathSegmenter(key, 0); part != ""; part, i = pathSegmenter(key, i) {
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

func buildNoAloc() *noAllocNode {
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

	root := newNoAllocNode()
	for _, path := range fixed {
		root.Insert(path, path)
	}

	// Update all the expanded `/_node/{metric}` values to include the token.
	for _, s := range expanded {
		root.Insert(s, nodesMetricPath)
	}

	return root
}

var pathNoAllocTrie = buildNoAloc()

func tokenizeFromNoAllocTrie(path string) string {
	return pathNoAllocTrie.Get(path)
}
