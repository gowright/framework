package gowright

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.Contains(t, config.RetryableErrors, BrowserError)
	assert.Contains(t, config.RetryableErrors, APIError)
	assert.Contains(t, config.RetryableErrors, DatabaseError)
	assert.Contains(t, config.RetryableErrors, ReportingError)
}

func TestRetryConfig_IsRetryable(t *testing.T) {
	config := DefaultRetryConfig()
	
	assert.True(t, config.IsRetryable(BrowserError))
	assert.True(t, config.IsRetryable(APIError))
	assert.True(t, config.IsRetryable(DatabaseError))
	assert.True(t, config.IsRetryable(ReportingError))
	assert.False(t, config.IsRetryable(ConfigurationError))
	assert.False(t, config.IsRetryable(AssertionError))
}

func TestRetryConfig_CalculateDelay(t *testing.T) {
	config := DefaultRetryConfig()
	
	// Test initial delay
	assert.Equal(t, 100*time.Millisecond, config.CalculateDelay(0))
	assert.Equal(t, 100*time.Millisecond, config.CalculateDelay(1))
	
	// Test exponential backoff
	assert.Equal(t, 200*time.Millisecond, config.CalculateDelay(2))
	assert.Equal(t, 400*time.Millisecond, config.CalculateDelay(3))
	
	// Test max delay cap
	config.MaxDelay = 300 * time.Millisecond
	assert.Equal(t, 300*time.Millisecond, config.CalculateDelay(3))
}

func TestRetryWithBackoff_Success(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxRetries = 2
	config.InitialDelay = 1 * time.Millisecond
	
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 2 {
			return NewGowrightError(BrowserError, "temporary failure", nil)
		}
		return nil // Success on second attempt
	}
	
	ctx := context.Background()
	err := RetryWithBackoff(ctx, config, operation)
	
	assert.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestRetryWithBackoff_MaxRetriesExceeded(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxRetries = 2
	config.InitialDelay = 1 * time.Millisecond
	
	attempts := 0
	operation := func() error {
		attempts++
		return NewGowrightError(BrowserError, "persistent failure", nil)
	}
	
	ctx := context.Background()
	err := RetryWithBackoff(ctx, config, operation)
	
	assert.Error(t, err)
	assert.Equal(t, 3, attempts) // Initial attempt + 2 retries
	
	gowrightErr, ok := err.(*GowrightError)
	require.True(t, ok)
	assert.Equal(t, ConfigurationError, gowrightErr.Type)
	assert.Contains(t, gowrightErr.Message, "operation failed after 2 retries")
}

func TestRetryWithBackoff_NonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxRetries = 2
	config.InitialDelay = 1 * time.Millisecond
	
	attempts := 0
	operation := func() error {
		attempts++
		return NewGowrightError(AssertionError, "assertion failed", nil)
	}
	
	ctx := context.Background()
	err := RetryWithBackoff(ctx, config, operation)
	
	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Should not retry
	
	gowrightErr, ok := err.(*GowrightError)
	require.True(t, ok)
	assert.Equal(t, AssertionError, gowrightErr.Type)
}

func TestRetryWithBackoff_ContextCancellation(t *testing.T) {
	config := DefaultRetryConfig()
	config.MaxRetries = 5
	config.InitialDelay = 100 * time.Millisecond
	
	attempts := 0
	operation := func() error {
		attempts++
		return NewGowrightError(BrowserError, "failure", nil)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	
	err := RetryWithBackoff(ctx, config, operation)
	
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Equal(t, 1, attempts) // Should only attempt once before timeout
}

func TestBrowserErrorRecovery(t *testing.T) {
	recovery := NewBrowserErrorRecovery(nil)
	
	// Test CanRecover
	browserErr := NewGowrightError(BrowserError, "browser crashed", nil)
	apiErr := NewGowrightError(APIError, "api failed", nil)
	genericErr := errors.New("generic error")
	
	assert.True(t, recovery.CanRecover(browserErr))
	assert.False(t, recovery.CanRecover(apiErr))
	assert.False(t, recovery.CanRecover(genericErr))
	
	// Test Recover
	ctx := context.Background()
	recoveredErr := recovery.Recover(ctx, browserErr)
	
	require.Error(t, recoveredErr)
	gowrightErr, ok := recoveredErr.(*GowrightError)
	require.True(t, ok)
	
	assert.Equal(t, true, gowrightErr.Context["recovery_attempted"])
	assert.Equal(t, "browser_restart", gowrightErr.Context["recovery_strategy"])
}

func TestAPIErrorRecovery(t *testing.T) {
	recovery := NewAPIErrorRecovery(nil)
	
	// Test CanRecover
	apiErr := NewGowrightError(APIError, "api timeout", nil)
	browserErr := NewGowrightError(BrowserError, "browser failed", nil)
	
	assert.True(t, recovery.CanRecover(apiErr))
	assert.False(t, recovery.CanRecover(browserErr))
	
	// Test Recover
	ctx := context.Background()
	recoveredErr := recovery.Recover(ctx, apiErr)
	
	require.Error(t, recoveredErr)
	gowrightErr, ok := recoveredErr.(*GowrightError)
	require.True(t, ok)
	
	assert.Equal(t, true, gowrightErr.Context["recovery_attempted"])
	assert.Equal(t, "api_retry", gowrightErr.Context["recovery_strategy"])
}

func TestDatabaseErrorRecovery(t *testing.T) {
	recovery := NewDatabaseErrorRecovery(nil)
	
	// Test CanRecover
	dbErr := NewGowrightError(DatabaseError, "connection lost", nil)
	apiErr := NewGowrightError(APIError, "api failed", nil)
	
	assert.True(t, recovery.CanRecover(dbErr))
	assert.False(t, recovery.CanRecover(apiErr))
	
	// Test Recover
	ctx := context.Background()
	recoveredErr := recovery.Recover(ctx, dbErr)
	
	require.Error(t, recoveredErr)
	gowrightErr, ok := recoveredErr.(*GowrightError)
	require.True(t, ok)
	
	assert.Equal(t, true, gowrightErr.Context["recovery_attempted"])
	assert.Equal(t, "database_reconnect", gowrightErr.Context["recovery_strategy"])
}

func TestErrorRecoveryManager(t *testing.T) {
	manager := NewErrorRecoveryManager()
	
	// Test with recoverable error
	browserErr := NewGowrightError(BrowserError, "browser error", nil)
	ctx := context.Background()
	
	recoveredErr := manager.RecoverFromError(ctx, browserErr)
	require.Error(t, recoveredErr)
	
	gowrightErr, ok := recoveredErr.(*GowrightError)
	require.True(t, ok)
	assert.Equal(t, true, gowrightErr.Context["recovery_attempted"])
	
	// Test with non-recoverable error
	assertionErr := NewGowrightError(AssertionError, "assertion failed", nil)
	recoveredErr = manager.RecoverFromError(ctx, assertionErr)
	
	require.Error(t, recoveredErr)
	gowrightErr, ok = recoveredErr.(*GowrightError)
	require.True(t, ok)
	assert.Equal(t, false, gowrightErr.Context["recovery_attempted"])
	assert.Equal(t, "no_strategy_found", gowrightErr.Context["recovery_reason"])
}

func TestErrorRecoveryManager_AddStrategy(t *testing.T) {
	manager := NewErrorRecoveryManager()
	initialStrategies := len(manager.strategies)
	
	// Add custom strategy
	customStrategy := NewBrowserErrorRecovery(nil)
	manager.AddStrategy(customStrategy)
	
	assert.Equal(t, initialStrategies+1, len(manager.strategies))
}

func TestWrapError(t *testing.T) {
	cause := errors.New("original error")
	wrappedErr := WrapError(APIError, "API call failed", cause)
	
	assert.Equal(t, APIError, wrappedErr.Type)
	assert.Equal(t, "API call failed", wrappedErr.Message)
	assert.Equal(t, cause, wrappedErr.Cause)
	assert.Contains(t, wrappedErr.Error(), "API call failed")
	assert.Contains(t, wrappedErr.Error(), "original error")
}

func TestWrapErrorWithContext(t *testing.T) {
	cause := errors.New("original error")
	context := map[string]interface{}{
		"endpoint": "/api/users",
		"method":   "GET",
		"status":   500,
	}
	
	wrappedErr := WrapErrorWithContext(APIError, "API call failed", cause, context)
	
	assert.Equal(t, APIError, wrappedErr.Type)
	assert.Equal(t, "API call failed", wrappedErr.Message)
	assert.Equal(t, cause, wrappedErr.Cause)
	assert.Equal(t, "/api/users", wrappedErr.Context["endpoint"])
	assert.Equal(t, "GET", wrappedErr.Context["method"])
	assert.Equal(t, 500, wrappedErr.Context["status"])
}

func TestIsGowrightError(t *testing.T) {
	gowrightErr := NewGowrightError(BrowserError, "test error", nil)
	genericErr := errors.New("generic error")
	
	assert.True(t, IsGowrightError(gowrightErr))
	assert.False(t, IsGowrightError(genericErr))
}

func TestGetErrorType(t *testing.T) {
	gowrightErr := NewGowrightError(DatabaseError, "test error", nil)
	genericErr := errors.New("generic error")
	
	assert.Equal(t, DatabaseError, GetErrorType(gowrightErr))
	assert.Equal(t, ConfigurationError, GetErrorType(genericErr)) // Default fallback
}

func TestGetErrorContext(t *testing.T) {
	gowrightErr := NewGowrightError(APIError, "test error", nil)
	gowrightErr.WithContext("key1", "value1")
	gowrightErr.WithContext("key2", 42)
	
	genericErr := errors.New("generic error")
	
	context := GetErrorContext(gowrightErr)
	assert.NotNil(t, context)
	assert.Equal(t, "value1", context["key1"])
	assert.Equal(t, 42, context["key2"])
	
	context = GetErrorContext(genericErr)
	assert.Nil(t, context)
}

func TestGowrightError_WithContext(t *testing.T) {
	err := NewGowrightError(BrowserError, "test error", nil)
	
	// Test adding context
	err.WithContext("key1", "value1")
	err.WithContext("key2", 42)
	
	assert.Equal(t, "value1", err.Context["key1"])
	assert.Equal(t, 42, err.Context["key2"])
	
	// Test chaining
	err2 := NewGowrightError(APIError, "another error", nil).
		WithContext("chained1", "value1").
		WithContext("chained2", "value2")
	
	assert.Equal(t, "value1", err2.Context["chained1"])
	assert.Equal(t, "value2", err2.Context["chained2"])
}

func TestGowrightError_Error(t *testing.T) {
	// Test without cause
	err1 := NewGowrightError(BrowserError, "browser failed", nil)
	assert.Equal(t, "browser failed", err1.Error())
	
	// Test with cause
	cause := errors.New("underlying error")
	err2 := NewGowrightError(APIError, "API failed", cause)
	assert.Equal(t, "API failed: underlying error", err2.Error())
}

// Benchmark tests for performance validation
func BenchmarkRetryWithBackoff(b *testing.B) {
	config := DefaultRetryConfig()
	config.MaxRetries = 1
	config.InitialDelay = 1 * time.Nanosecond
	
	operation := func() error {
		return nil // Always succeed
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RetryWithBackoff(ctx, config, operation)
	}
}

func BenchmarkErrorRecoveryManager(b *testing.B) {
	manager := NewErrorRecoveryManager()
	err := NewGowrightError(BrowserError, "test error", nil)
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.RecoverFromError(ctx, err)
	}
}