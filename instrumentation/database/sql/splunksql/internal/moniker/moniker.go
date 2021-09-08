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

type Span string

const (
	Query    Span = "Query"
	Ping     Span = "Ping"
	Prepare  Span = "Prepare"
	Exec     Span = "Exec"
	Begin    Span = "Begin"
	Reset    Span = "Reset"
	Close    Span = "Close"
	Commit   Span = "Commit"
	Rollback Span = "Rollback"
	Rows     Span = "Rows"
)

func (n Span) String() string { return string(n) }

type Event string

const (
	Next Event = "Next"
)

func (n Event) String() string { return string(n) }
