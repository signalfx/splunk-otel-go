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

//go:build cgo && (linux || darwin)

package splunkkafka_test

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
)

const (
	k, v = "key", "value"
)

func TestCarrierGet(t *testing.T) {
	msg := &kafka.Message{
		Headers: []kafka.Header{
			{Key: k, Value: []byte(v)},
		},
	}
	carrier := splunkkafka.NewMessageCarrier(msg)
	assert.Equal(t, v, carrier.Get(k))
}

func TestCarrierGetEmpty(t *testing.T) {
	msg := &kafka.Message{}
	carrier := splunkkafka.NewMessageCarrier(msg)
	assert.Equal(t, "", carrier.Get("key"))
}

func TestCarrierSet(t *testing.T) {
	msg := &kafka.Message{}
	carrier := splunkkafka.NewMessageCarrier(msg)
	carrier.Set(k, v)
	var got string
	for _, h := range msg.Headers {
		if h.Key == k {
			got = string(h.Value)
		}
	}
	assert.Equal(t, v, got)
}

func TestCarrierSetOverwrites(t *testing.T) {
	msg := &kafka.Message{
		Headers: []kafka.Header{
			{Key: k, Value: []byte("not value")},
			{Key: k, Value: []byte("also not value")},
		},
	}
	carrier := splunkkafka.NewMessageCarrier(msg)
	carrier.Set(k, v)
	var got string
	for _, h := range msg.Headers {
		if h.Key == k {
			got = string(h.Value)
		}
	}
	assert.Equal(t, v, got)
}

func TestCarrierKeys(t *testing.T) {
	keys := []string{"one", "two", "three"}
	msg := &kafka.Message{
		Headers: []kafka.Header{
			{Key: keys[0], Value: []byte("")},
			{Key: keys[1], Value: []byte("")},
			{Key: keys[2], Value: []byte("")},
		},
	}
	carrier := splunkkafka.NewMessageCarrier(msg)
	assert.Equal(t, keys, carrier.Keys())
}
