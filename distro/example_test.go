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

package distro_test

import (
	"context"

	"github.com/signalfx/splunk-otel-go/distro"
)

func Example() {
	// By default, the Run function will create a Jaeger exporter to a locally
	// running Splunk Smart Agent at http://localhost:9080 and will configure
	// the B3 context propagation format to be used in extracting and
	// injecting trace context.
	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}
	// To ensure all spans are flushed before the application exits, make sure
	// to shutdown.
	defer sdk.Shutdown(context.Background())
}
