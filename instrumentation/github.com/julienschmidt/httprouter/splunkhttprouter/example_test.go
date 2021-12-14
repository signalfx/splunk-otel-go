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

package splunkhttprouter_test

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/julienschmidt/httprouter/splunkhttprouter"
)

func Index(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, _ *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func Example() {
	router := splunkhttprouter.New()
	router.GET("/", Index)
	router.GET("/hello/:name", Hello)

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
