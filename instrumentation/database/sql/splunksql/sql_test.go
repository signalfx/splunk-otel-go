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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func runConcurrently(funcs ...func()) {
	var wg sync.WaitGroup

	for _, f := range funcs {
		wg.Add(1)
		go func(f func()) {
			f()
			wg.Done()
		}(f)
	}

	wg.Wait()
}

func TestRegisterConcurrentSafe(t *testing.T) {
	instCfg := InstrumentationConfig{
		DBSystem: DBSystem(
			attribute.String("test", "database"),
		),
	}

	runConcurrently(
		func() { _ = retrieve("blank") },
		func() { Register("testSQL", instCfg) },
		func() { _ = retrieve("testSQL") },
	)

	assert.Equal(t, instCfg, retrieve("testSQL"))
}

func TestRegisterPanic(t *testing.T) {
	instCfg := InstrumentationConfig{}
	Register("TestRegisterPanic", instCfg)
	assert.Panics(t, func() { Register("TestRegisterPanic", instCfg) })
}
