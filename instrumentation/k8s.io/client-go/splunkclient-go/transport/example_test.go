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

//go:build !(go1.1 || go1.2 || go1.3 || go1.4 || go1.5 || go1.6 || go1.7 || go1.8 || go1.9 || go1.10 || go1.11 || go1.12 || go1.13 || go1.14 || go1.15 || go1.16)
// +build !go1.1,!go1.2,!go1.3,!go1.4,!go1.5,!go1.6,!go1.7,!go1.8,!go1.9,!go1.10,!go1.11,!go1.12,!go1.13,!go1.14,!go1.15,!go1.16

package transport_test

import (
	"context"
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/signalfx/splunk-otel-go/instrumentation/k8s.io/client-go/splunkclient-go/transport"
)

func Example() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// Use this to trace all calls made to the Kubernetes API.
	cfg.WrapTransport = transport.NewWrapperFunc()

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	pods, err := client.CoreV1().Pods("default").List(ctx, meta_v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println(pods.Items)
}
