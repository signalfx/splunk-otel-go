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

// Package splunkmgo provides OpenTelemetry instrumentation for the
// github.com/github.com/globalsign/mgo package.
package splunkmgo

import (
	"time"

	"github.com/globalsign/mgo"
	"go.opentelemetry.io/otel/attribute"
)

// Dial returns a new traced session to the cluster identified by the given
// seed server(s).
func Dial(url string, opts ...Option) (*Session, error) {
	s, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}
	return newSession(s, opts...), nil
}

// DialWithInfo establishes a new traced session to the cluster identified by
// info.
func DialWithInfo(dialInfo *mgo.DialInfo, opts ...Option) (*Session, error) {
	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	return newSession(s, opts...), nil
}

// DialWithTimeout works like Dial, but uses timeout as the amount of time to
// wait for a server to respond when first connecting and also on follow up
// operations in the session. If timeout is zero, the call may block forever
// waiting for a connection to be made.
func DialWithTimeout(url string, timeout time.Duration, opts ...Option) (*Session, error) {
	s, err := mgo.DialWithTimeout(url, timeout)
	if err != nil {
		return nil, err
	}
	return newSession(s, opts...), nil
}

// Session is a traced communication session with the database.
type Session struct {
	*mgo.Session

	cfg *config
}

func newSession(s *mgo.Session, opts ...Option) *Session {
	version := "unknown"
	if info, err := s.BuildInfo(); err == nil {
		version = info.Version
	}
	opts = append(opts, WithAttributes([]attribute.KeyValue{
		attribute.String("mgo.version", version),
	}))

	// FIXME: peer info.
	return &Session{
		Session: s,
		cfg:     newConfig(opts...),
	}
}
