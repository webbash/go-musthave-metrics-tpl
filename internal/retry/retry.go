// Package retry provides retry helpers for transient operations.
package retry

import (
	"context"
	"time"
)

func Do(
	ctx context.Context,
	operation func() error,
	isRetriable func(error) bool,
	intervals []time.Duration,
) error {
	err := operation()
	if err == nil || !isRetriable(err) {
		return err
	}

	for _, interval := range intervals {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		err = operation()
		if err == nil || !isRetriable(err) {
			return err
		}
	}

	return err
}
