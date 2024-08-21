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
