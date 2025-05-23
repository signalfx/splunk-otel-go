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

package splunkpgx_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
)

type server struct {
	DB *sql.DB
}

func (s *server) listenAndServe() error {
	// Requests to /square/n will return the square of n.
	http.HandleFunc("/square/", s.handle)
	return http.ListenAndServe(":80", nil)
}

func (s *server) handle(w http.ResponseWriter, req *http.Request) {
	idx := strings.LastIndex(req.URL.Path, "/")
	n, err := strconv.Atoi(req.URL.Path[idx+1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "SELECT squareNumber FROM squarenum WHERE number = ?"
	var nSquared int
	// Propagate the context to ensure created spans are included in any
	// existing trace.
	if err := s.DB.QueryRowContext(req.Context(), query, n).Scan(&nSquared); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d", nSquared)
}

func Example() {
	// Create a traced connection to the Postgres database.
	db, err := splunksql.Open("pgx", "postgres://localhost:5432/dbname")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Validate DSN data by opening a connection. There is no parent context
	// to pass here so the span created from this operation will be in its own
	// trace.
	if err := db.PingContext(context.Background()); err != nil {
		panic(err)
	}

	srv := &server{DB: db}
	if err := srv.listenAndServe(); err != nil {
		panic(err)
	}
}
