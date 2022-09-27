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
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logLevel is an SDK logging level.
type logLevel struct {
	name     string
	priority int8
}

// String returns the log level as a string.
func (l logLevel) String() string { return l.name }

// String returns the log level tranlated to a zap Level.
func (l logLevel) ZapLevel() zapcore.Level { return zapcore.Level(l.priority) }

var (
	// debugLevel is any verbosity 2 or higher. The zap levels are capped at
	// -127 because int8 is the underlying type. Use this as a stand-in for
	// debug so all debug levels are logged.
	debugLevel = logLevel{name: "debug", priority: -127}
	// infoLevel is verbosity equal to 1.
	infoLevel = logLevel{name: "info", priority: -1}
	// infoLevel is verbosity equal to 0.
	warnLevel = logLevel{name: "warn", priority: 0}
	// errorLevel only prints log messages made with the logr.Error function.
	errorLevel = logLevel{name: "error", priority: 1}

	logLevels = []logLevel{debugLevel, infoLevel, warnLevel, errorLevel}
)

// zapLevel returns the parsed zapcore.Level.
func zapLevel(level string) zapcore.Level {
	for _, l := range logLevels {
		if l.String() == strings.ToLower(level) {
			return l.ZapLevel()
		}
	}
	// unrecognized level, use "info" level.
	return infoLevel.ZapLevel()
}

// zapLevelEncoder translates our verbosity levels to human-meaningful terms.
func zapLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch v := int8(l); {
	case v > 0:
		enc.AppendString(errorLevel.String())
	case v == 0:
		enc.AppendString(warnLevel.String())
	case v == -1:
		enc.AppendString(infoLevel.String())
	case v <= -2:
		enc.AppendString(debugLevel.String())
	}
}

func zapConfig(level string) *zap.Config {
	zc := zap.NewProductionConfig()
	zc.Encoding = "console"
	// Human-readable timestamps for console format of logs.
	zc.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Translate our verbosity levels to logged levels.
	zc.EncoderConfig.EncodeLevel = zapLevelEncoder
	zc.Level = zap.NewAtomicLevelAt(zapLevel(level))
	return &zc
}

// logger configures and returns the default project logger.
//
// The returned logger is configured to match verbosity levels as such for any
// Info log made:
//   - warning: 0
//   - info: 1
//   - debug: 2+
func logger(zc *zap.Config) logr.Logger {
	z, err := zc.Build()
	if err != nil {
		// This should never happen because we control zc. Panic to expose the
		// bug ASAP to the developer.
		panic(err)
	}
	return zapr.NewLogger(z)
}
