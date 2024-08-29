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
)

// CheckContextDone checks if the context has been canceled or timed out.
// If the context is done, it returns the context error, which can be either
// a cancellation error or a deadline exceeded error. If the context is not
// done, it returns nil.
//
// This function is useful for early exits in long-running or blocking
// operations when you then  to respond to context cancellations in a clean
// and consistent manner.
func CheckContextDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
