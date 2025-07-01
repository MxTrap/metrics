package utils

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRetrySuccessFirstAttempt(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := Retry(fn, 3)
	require.NoError(t, err, "Retry should succeed on first attempt")
	assert.Equal(t, 1, callCount, "fn should be called exactly once")
}

func TestRetrySuccessAfterRetries(t *testing.T) {
	callCount := 0
	successAfter := 2
	fn := func() error {
		callCount++
		if callCount >= successAfter {
			return nil
		}
		return errors.New("temporary error")
	}

	err := Retry(fn, 3)
	require.NoError(t, err, "Retry should succeed after retries")
	assert.Equal(t, successAfter, callCount, "fn should be called expected number of times")
}

func TestRetryFailure(t *testing.T) {
	callCount := 0
	expectedError := errors.New("persistent error")
	fn := func() error {
		callCount++
		return expectedError
	}

	err := Retry(fn, 3)
	assert.Error(t, err, "Retry should return error after all attempts")
	assert.Equal(t, expectedError, err, "Retry should return the last error")
	assert.Equal(t, 3, callCount, "fn should be called exactly retryCount times")
}
