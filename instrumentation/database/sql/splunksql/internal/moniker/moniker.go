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

// Package moniker provides consistent identifiers for telemetry data.
package moniker // import "github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql/internal/moniker"

// Span is the name of an OpenTelemetry Span.
type Span string

const (
	// Query is the span name for a query operation.
	Query Span = "Query"
	// Ping is the span name for a ping operation.
	Ping Span = "Ping"
	// Prepare is the span name for a prepare operation.
	Prepare Span = "Prepare"
	// Exec is the span name for a exec operation.
	Exec Span = "Exec"
	// Begin is the span name for a begin operation.
	Begin Span = "Begin"
	// Reset is the span name for a reset operation.
	Reset Span = "Reset"
	// Close is the span name for a close operation.
	Close Span = "Close"
	// Commit is the span name for a commit operation.
	Commit Span = "Commit"
	// Rollback is the span name for a rollback operation.
	Rollback Span = "Rollback"
)

// String returns the Span as a string.
func (n Span) String() string { return string(n) }
