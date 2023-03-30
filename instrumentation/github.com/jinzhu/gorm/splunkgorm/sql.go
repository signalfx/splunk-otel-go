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

// Package splunkgorm provides instrumentation for the [github.com/jinzhu/gorm]
// package.
package splunkgorm // import "github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm"

import (
	"github.com/jinzhu/gorm"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

// openFunc allows overrides for testing.
var openFunc = splunksql.Open

// Open opens a traced gorm database connection specified by its database
// driver name and a driver-specific data source name. The driver must already
// be registered by the driver package.
func Open(driverName, dataSourceName string, opts ...splunksql.Option) (*gorm.DB, error) {
	db, err := openFunc(driverName, dataSourceName, opts...)
	if err != nil {
		return nil, err
	}
	return gorm.Open(driverName, db)
}
