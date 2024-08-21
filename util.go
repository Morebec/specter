pack	age specter

import "context"

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
