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

package splunkgorm_test

import (
	"log"

	"github.com/jinzhu/gorm"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/jinzhu/gorm/splunkgorm"
)

func ExampleOpen() {
	// This assumes the instrumented driver,
	// "github.com/signalfx/splunk-otel-go/instrumentation/github.com/jackc/pgx/splunkpgx",
	// is imported. That will ensure the driver and the instrumentation setup
	// for the driver are registered with the appropriate packages.
	db, err := splunkgorm.Open("pgx", "postgres://localhost/db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close db: %v", err)
		}
	}()

	user := struct {
		gorm.Model
		Name string
	}{}

	// All calls through gorm.DB are now traced.
	db.Where("name = ?", "jinzhu").First(&user)
	/* ... */
}
