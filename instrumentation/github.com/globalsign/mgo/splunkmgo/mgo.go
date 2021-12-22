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
	"context"
	"time"

	"github.com/globalsign/mgo"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
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
	// FIXME: extract and annotate username.
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

// Clone works just like Copy, but also reuses the same socket as the original
// session, in case it had already reserved one due to its consistency
// guarantees. This behavior ensures that writes performed in the old session
// are necessarily observed when using the new session, as long as it was a
// strong or monotonic session. That said, it also means that long operations
// may cause other goroutines using the original session to wait.
func (s *Session) Clone() *Session {
	return &Session{Session: s.Session.Clone(), cfg: s.cfg.Copy()}
}

// Copy works just like New, but preserves the exact authentication
// information from the original session.
func (s *Session) Copy() *Session {
	return &Session{Session: s.Session.Copy(), cfg: s.cfg.Copy()}
}

// New creates a new traced session with the same parameters as the original
// session, including consistency, batch size, prefetching, safety mode, etc.
// The returned session will use sockets from the pool, so there's a chance
// that writes just performed in another session may not yet be visible.
//
// Login information from the original session will not be copied over
// into the new session unless it was provided through the initial URL
// for the Dial function.
//
// See the Copy and Clone methods.
func (s *Session) New() *Session {
	return &Session{Session: s.Session.New(), cfg: s.cfg.Copy()}
}

// Ping runs a trivial ping command just to get in touch with the server.
func (s *Session) Ping() error {
	return s.Run("ping", nil)
}

// DB returns a traced *Database representing the named database. If name is
// empty, the database name provided in the dialed URL is used instead. If
// that is also empty, "test" is used as a fallback in a way equivalent to the
// mongo shell.
//
// Creating this value is a very lightweight operation, and involves no
// network communication.
func (s *Session) DB(name string) *Database {
	return &Database{Database: s.Session.DB(name), Session: s}
}

// Login authenticates with MongoDB using the provided credential.  The
// authentication is valid for the whole session and will stay valid until
// Logout is explicitly called for the same database, or the session is
// closed.
func (s *Session) Login(cred *mgo.Credential) error {
	// FIXME: extract and annoate username.
	return s.Session.Login(cred)
}

// LogoutAll removes all established authentication credentials for the session.
func (s *Session) LogoutAll() {
	// FIXME: remove all annotated usernames.
	s.Session.LogoutAll()
	return
}

// Run traces and issues the provided command on the "admin" database and and
// unmarshals its result in the respective argument.
func (s *Session) Run(cmd, result interface{}) error {
	return s.DB("admin").Run(cmd, result)
}

// Database holds collections of documents.
type Database struct {
	*mgo.Database

	Session *Session
}

// Run traces and issues the provided command on the db database and
// unmarshals its result in the respective argument. The cmd argument may be
// either a string with the command name itself, in which case an empty
// document of the form bson.M{cmd: 1} will be used, or it may be a full
// command document.
//
// Note that MongoDB considers the first marshalled key as the command name,
// so when providing a command with options, it's important to use an
// ordering-preserving document, such as a struct value or an instance of
// bson.D.  For instance:
//
//     db.Run(bson.D{{"create", "mycollection"}, {"size", 1024}})
//
// For privilleged commands typically run on the "admin" database, see the Run
// method in the Session type.
func (db *Database) Run(cmd interface{}, result interface{}) error {
	name := spanName(cmd)
	return db.Session.cfg.WithSpan(db.Session.cfg.ctx, name, func(context.Context) error {
		return db.Database.Run(cmd, result)
	}, trace.WithAttributes(semconv.DBNameKey.String(db.Name)))
}

// spanName returns a span name based on the cmd being run.
func spanName(cmd interface{}) string {
	if name, ok := cmd.(string); ok {
		return name
	}

	// FIXME: handle cmd similar to bson.MarshalBuffer()

	// Fallback to something.
	return "mongo"
}
