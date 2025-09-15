package testutil

import (
	"context"
	"time"
)

// SlowAuth simulates a slow auth service to test timeout behavior.
// It implements the AuthService interface and can be used in tests
// to verify that authentication timeouts are handled correctly.
type SlowAuth struct {
	Delay time.Duration
}

// Validate implements the AuthService interface with configurable delay.
// Returns true for token "slow", false for any other token.
// Respects context cancellation/timeout.
func (s SlowAuth) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "slow" {
		select {
		case <-time.After(s.Delay):
			return "p1", "Alice", true
		case <-ctx.Done():
			// Context was cancelled/timed out
			return "", "", false
		}
	}
	return "", "", false
}
