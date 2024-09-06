// Copyright 2024 Morébec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package specter_test

import (
	"github.com/morebec/specter/pkg/specter"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRunResult_ExecutionTime(t *testing.T) {
	r := specter.PipelineResult{}
	r.StartedAt = time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC)
	r.EndedAt = time.Date(2024, 01, 01, 1, 0, 0, 0, time.UTC)

	require.Equal(t, r.ExecutionTime(), time.Hour*1)
}
