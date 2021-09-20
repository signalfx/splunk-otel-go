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

package splunksql

import (
	"database/sql/driver"
	"io"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func handleErr(span trace.Span, err error) {
	if span == nil {
		return
	}

	switch err {
	case nil:
		// Everything Okay.
	case io.EOF:
		// Expected at end of iteration, do not record these.
	case driver.ErrSkip:
		// Expected if method not implemented, do not record these.
	default:
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
