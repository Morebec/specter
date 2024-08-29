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

package specter

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCheckContextDone(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected error
	}{
		{
			name:     "Canceled context",
			ctx:      func() context.Context { ctx, cancel := context.WithCancel(context.Background()); cancel(); return ctx }(),
			expected: context.Canceled,
		},
		{
			name: "Timed out context",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()
				time.Sleep(10 * time.Nanosecond)
				return ctx
			}(),
			expected: context.DeadlineExceeded,
		},
		{
			name:     "Active context",
			ctx:      context.Background(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckContextDone(tt.ctx)
			assert.Equal(t, tt.expected, err)
		})
	}
}
