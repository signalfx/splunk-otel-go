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

package splunkbuntdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenError(t *testing.T) {
	_, err := Open("/non/existing/file.db")
	assert.Error(t, err)
}

func TestOpenReturnType(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	assert.IsType(t, &DB{}, db)
}

func TestDBWithContext(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	ctx := context.WithValue(context.Background(), "key", "val")
	db = db.WithContext(ctx)
	assert.Equal(t, ctx, db.cfg.ctx)
}
