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

//go:build go1.17
// +build go1.17

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
