package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo_SuccessAfterRetry(t *testing.T) {
	attempts := 0

	operation := func() error {
		attempts++

		if attempts < 3 {
			return errors.New("temporary error")
		}

		return nil
	}

	err := Do(
		context.Background(),
		operation,
		func(error) bool {
			return true
		},
		[]time.Duration{
			time.Millisecond,
			time.Millisecond,
			time.Millisecond,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestDo_DoesNotRetryNonRetriableError(t *testing.T) {
	attempts := 0
	expectedErr := errors.New("permanent error")

	operation := func() error {
		attempts++
		return expectedErr
	}

	err := Do(
		context.Background(),
		operation,
		func(error) bool {
			return false
		},
		[]time.Duration{
			time.Millisecond,
			time.Millisecond,
			time.Millisecond,
		},
	)

	require.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 1, attempts)
}

func TestDo_ReturnsErrorAfterAllRetries(t *testing.T) {
	attempts := 0
	expectedErr := errors.New("temporary error")

	operation := func() error {
		attempts++
		return expectedErr
	}

	err := Do(
		context.Background(),
		operation,
		func(error) bool {
			return true
		},
		[]time.Duration{
			time.Millisecond,
			time.Millisecond,
			time.Millisecond,
		},
	)

	require.ErrorIs(t, err, expectedErr)

	// 1 первоначальная попытка + 3 retry
	assert.Equal(t, 4, attempts)
}
