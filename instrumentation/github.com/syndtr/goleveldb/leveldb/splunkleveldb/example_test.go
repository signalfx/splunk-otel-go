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

package splunkleveldb_test

import (
	"context"
	"log"

	"github.com/syndtr/goleveldb/leveldb/storage"
	"go.opentelemetry.io/otel"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/syndtr/goleveldb/leveldb/splunkleveldb"
)

func Example() {
	// Open a new database backed by memory storage.
	memstore := storage.NewMemStorage()
	// Ensure span is used as a parent for all spans the database will create.
	db, err := splunkleveldb.Open(memstore, nil)
	if err != nil {
		// Assume corruptions and attept a recover.
		db, err = splunkleveldb.Recover(memstore, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer func() {
		if cErr := db.Close(); cErr != nil {
			log.Fatal(cErr)
		}
	}()

	ctx, span := otel.Tracer("my-inst").Start(context.Background(), "main")
	defer span.End()
	// Ensure span is used as a parent for all spans the database will create.
	db = db.WithContext(ctx)

	// Write to the database. A span is created to trace this operation.
	if err = db.Put([]byte("key"), []byte("value"), nil); err != nil {
		log.Println("failed to write data", err)
	}

	// Read back the data. A span is created to trace this operation.
	data, err := db.Get([]byte("key"), nil)
	if err != nil {
		log.Println("failed to read data", err)
	} else {
		log.Println("read from database:", string(data))
	}

	// Delete the data for the key. A span is created to trace this operation.
	if err = db.Delete([]byte("key"), nil); err != nil {
		log.Println("failed to delete data", err)
	}
}

func ExampleOpen() {
	memstore := storage.NewMemStorage()
	db, err := splunkleveldb.Open(memstore, nil)
	if err != nil {
		// Assume corruptions and attept a recover.
		db, err = splunkleveldb.Recover(memstore, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()
}

func ExampleOpenFile() {
	dbPath := "/path/to/db"
	db, err := splunkleveldb.OpenFile(dbPath, nil)
	if err != nil {
		// Assume corruptions and attept a recover.
		db, err = splunkleveldb.RecoverFile(dbPath, nil)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()
}
