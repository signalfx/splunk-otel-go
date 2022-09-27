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

package distro

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type stringEncoder struct {
	zapcore.PrimitiveArrayEncoder
	got string
}

func (e *stringEncoder) AppendString(s string) {
	e.got = s
}

func (e *stringEncoder) assert(t *testing.T, want string) {
	assert.Equal(t, want, e.got, "encoded wrong level")
}

func TestZapLevelEncoder(t *testing.T) {
	levelMap := map[int8]string{
		// Not that we use it, but 5 is the "fatal" level in zap.
		5:  "error",
		4:  "error",
		3:  "error",
		2:  "error",
		1:  "error",
		0:  "warn",
		-1: "info",
	}

	enc := new(stringEncoder)
	for level, want := range levelMap {
		zapLevelEncoder(zapcore.Level(level), enc)
		enc.assert(t, want)
	}

	// Debug should be for all verbosity between -2 and the end of the of the
	// zap level range (-127).
	for i := -2; i >= -127; i-- {
		zapLevelEncoder(zapcore.Level(i), enc)
		enc.assert(t, "debug")
	}
}

func TestZapLevel(t *testing.T) {
	testcases := []struct {
		in   string
		want int8
	}{
		{in: "debug", want: -127},
		{in: "info", want: -1},
		{in: "warn", want: 0},
		{in: "error", want: 1},
		{in: "invalid", want: -1},
		{in: "DEBUG", want: -127}, // case insensitive
	}

	for _, tc := range testcases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, zapcore.Level(tc.want), zapLevel(tc.in))
		})
	}
}

func TestLoggerPanic(t *testing.T) {
	zc := zapConfig("info")
	// Set an invalid level so the zap logger build will error. This error
	// should be panic-ed.
	zc.Level = zap.AtomicLevel{}
	assert.Panics(t, func() { _ = logger(zc) })
}
