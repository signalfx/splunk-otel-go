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

package transport

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestToSpanNameUnrecognized(t *testing.T) {
	path := "/unrecognized"
	r, err := http.NewRequest("GET", path, nil)
	require.NoError(t, err)

	expected := "HTTP GET"
	assert.Equalf(t, expected, name(r), "path: %q", path)
}

func TestRequestToSpanName(t *testing.T) {
	tests := []struct {
		path string
		name string
	}{
		{
			path: "/api/v1/componentstatuses",
			name: "componentstatuses",
		},
		{
			path: "/api/v1/componentstatuses/NAME",
			name: "componentstatuses/{name}",
		},
		{
			path: "/api/v1/configmaps",
			name: "configmaps",
		},
		{
			path: "/api/v1/namespaces/default/bindings",
			name: "namespaces/{namespace}/bindings",
		},
		{
			path: "/api/v1/namespaces/someothernamespace/configmaps",
			name: "namespaces/{namespace}/configmaps",
		},
		{
			path: "/api/v1/namespaces/default/configmaps/some-config-map",
			name: "namespaces/{namespace}/configmaps/{name}",
		},
		{
			path: "/api/v1/namespaces/default/persistentvolumeclaims/pvc-abcd/status",
			name: "namespaces/{namespace}/persistentvolumeclaims/{name}/status",
		},
		{
			path: "/api/v1/namespaces/default/pods/pod-1234/proxy",
			name: "namespaces/{namespace}/pods/{name}/proxy",
		},
		{
			path: "/api/v1/namespaces/default/pods/pod-5678/proxy/some-path",
			name: "namespaces/{namespace}/pods/{name}/proxy/{path}",
		},
		{
			path: "/api/v1/watch/configmaps",
			name: "watch/configmaps",
		},
		{
			path: "/api/v1/watch/namespaces",
			name: "watch/namespaces",
		},
		{
			path: "/api/v1/watch/namespaces/default/configmaps",
			name: "watch/namespaces/{namespace}/configmaps",
		},
		{
			path: "/api/v1/watch/namespaces/someothernamespace/configmaps/another-name",
			name: "watch/namespaces/{namespace}/configmaps/{name}",
		},
	}

	for _, test := range tests {
		r, err := http.NewRequest("GET", test.path, nil)
		require.NoError(t, err)

		expected := "HTTP GET " + test.name
		assert.Equalf(t, expected, name(r), "path: %q", test.path)
	}
}
