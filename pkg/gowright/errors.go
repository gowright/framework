package gowright

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig defines retry behavior for different operations
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []ErrorType   `json:"retryable_errors"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []ErrorType{
			BrowserError,
			APIError,
			DatabaseError,
			ReportingError,
		},
	}
}

// IsRetryable checks if an error type is retryable
func (rc *RetryConfig) IsRetryable(errorType ErrorType) bool {
	for _, retryableType := range rc.RetryableErrors {
		if retryableType == errorType {
			return true
		}
	}
	return false
}

// CalculateDelay calculates the delay for a given retry attempt using exponential backoff
func (rc *RetryConfig) CalculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return rc.InitialDelay
	}

	delay := float64(rc.InitialDelay) * math.Pow(rc.BackoffFactor, float64(attempt-1))

	if delay > float64(rc.MaxDelay) {
		return rc.MaxDelay
	}

	return time.Duration(delay)
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func() error

// RetryWithBackoff executes an operation with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, config *RetryConfig, operation RetryableOperation) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Check if the error is retryable
		if gowrightErr, ok := err.(*GowrightError); ok {
			if !config.IsRetryable(gowrightErr.Type) {
				return err // Not retryable, return immediately
			}
		}

		// Calculate delay for next attempt
		delay := config.CalculateDelay(attempt + 1)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	// All retries exhausted, return the last error
	return NewGowrightError(
		ConfigurationError,
		fmt.Sprintf("operation failed after %d retries", config.MaxRetries),
		lastErr,
	).WithContext("max_retries", config.MaxRetries)
}

// ErrorRecoveryStrategy defines how to recover from specific error types
type ErrorRecoveryStrategy interface {
	CanRecover(err error) bool
	Recover(ctx context.Context, err error) error
}

// BrowserErrorRecovery handles browser-related error recovery
type BrowserErrorRecovery struct {
	retryConfig *RetryConfig
}

// NewBrowserErrorRecovery creates a new browser error recovery strategy
func NewBrowserErrorRecovery(config *RetryConfig) *BrowserErrorRecovery {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &BrowserErrorRecovery{retryConfig: config}
}

// CanRecover checks if the error can be recovered by this strategy
func (ber *BrowserErrorRecovery) CanRecover(err error) bool {
	if gowrightErr, ok := err.(*GowrightError); ok {
		return gowrightErr.Type == BrowserError
	}
	return false
}

// Recover attempts to recover from browser errors
func (ber *BrowserErrorRecovery) Recover(ctx context.Context, err error) error {
	gowrightErr, ok := err.(*GowrightError)
	if !ok {
		return err
	}

	// Add recovery context
	gowrightErr = gowrightErr.WithContext("recovery_attempted", true)
	gowrightErr = gowrightErr.WithContext("recovery_strategy", "browser_restart")

	// For browser errors, we typically need to restart the browser instance
	// This would be handled by the UITester when it detects a recovery error
	return gowrightErr
}

// APIErrorRecovery handles API-related error recovery
type APIErrorRecovery struct {
	retryConfig *RetryConfig
}

// NewAPIErrorRecovery creates a new API error recovery strategy
func NewAPIErrorRecovery(config *RetryConfig) *APIErrorRecovery {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &APIErrorRecovery{retryConfig: config}
}

// CanRecover checks if the error can be recovered by this strategy
func (aer *APIErrorRecovery) CanRecover(err error) bool {
	if gowrightErr, ok := err.(*GowrightError); ok {
		return gowrightErr.Type == APIError
	}
	return false
}

// Recover attempts to recover from API errors
func (aer *APIErrorRecovery) Recover(ctx context.Context, err error) error {
	gowrightErr, ok := err.(*GowrightError)
	if !ok {
		return err
	}

	// Add recovery context
	gowrightErr = gowrightErr.WithContext("recovery_attempted", true)
	gowrightErr = gowrightErr.WithContext("recovery_strategy", "api_retry")

	// For API errors, recovery usually involves retrying the request
	// The actual retry logic would be handled by the APITester
	return gowrightErr
}

// DatabaseErrorRecovery handles database-related error recovery
type DatabaseErrorRecovery struct {
	retryConfig *RetryConfig
}

// NewDatabaseErrorRecovery creates a new database error recovery strategy
func NewDatabaseErrorRecovery(config *RetryConfig) *DatabaseErrorRecovery {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &DatabaseErrorRecovery{retryConfig: config}
}

// CanRecover checks if the error can be recovered by this strategy
func (der *DatabaseErrorRecovery) CanRecover(err error) bool {
	if gowrightErr, ok := err.(*GowrightError); ok {
		return gowrightErr.Type == DatabaseError
	}
	return false
}

// Recover attempts to recover from database errors
func (der *DatabaseErrorRecovery) Recover(ctx context.Context, err error) error {
	gowrightErr, ok := err.(*GowrightError)
	if !ok {
		return err
	}

	// Add recovery context
	gowrightErr = gowrightErr.WithContext("recovery_attempted", true)
	gowrightErr = gowrightErr.WithContext("recovery_strategy", "database_reconnect")

	// For database errors, recovery might involve reconnecting or retrying transactions
	// The actual recovery logic would be handled by the DatabaseTester
	return gowrightErr
}

// ErrorRecoveryManager manages multiple recovery strategies
type ErrorRecoveryManager struct {
	strategies []ErrorRecoveryStrategy
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager() *ErrorRecoveryManager {
	return &ErrorRecoveryManager{
		strategies: []ErrorRecoveryStrategy{
			NewBrowserErrorRecovery(nil),
			NewAPIErrorRecovery(nil),
			NewDatabaseErrorRecovery(nil),
		},
	}
}

// AddStrategy adds a new recovery strategy
func (erm *ErrorRecoveryManager) AddStrategy(strategy ErrorRecoveryStrategy) {
	erm.strategies = append(erm.strategies, strategy)
}

// RecoverFromError attempts to recover from an error using available strategies
func (erm *ErrorRecoveryManager) RecoverFromError(ctx context.Context, err error) error {
	for _, strategy := range erm.strategies {
		if strategy.CanRecover(err) {
			return strategy.Recover(ctx, err)
		}
	}

	// No recovery strategy found
	if gowrightErr, ok := err.(*GowrightError); ok {
		gowrightErr = gowrightErr.WithContext("recovery_attempted", false)
		gowrightErr = gowrightErr.WithContext("recovery_reason", "no_strategy_found")
		return gowrightErr
	}

	return err
}

// WrapError wraps a generic error into a GowrightError with context
func WrapError(errorType ErrorType, message string, cause error) *GowrightError {
	return NewGowrightError(errorType, message, cause)
}

// WrapErrorWithContext wraps an error and adds context information
func WrapErrorWithContext(errorType ErrorType, message string, cause error, context map[string]interface{}) *GowrightError {
	err := NewGowrightError(errorType, message, cause)
	for key, value := range context {
		err = err.WithContext(key, value)
	}
	return err
}

// IsGowrightError checks if an error is a GowrightError
func IsGowrightError(err error) bool {
	_, ok := err.(*GowrightError)
	return ok
}

// GetErrorType extracts the error type from a GowrightError
func GetErrorType(err error) ErrorType {
	if gowrightErr, ok := err.(*GowrightError); ok {
		return gowrightErr.Type
	}
	return ConfigurationError // Default fallback
}

// GetErrorContext extracts context from a GowrightError
func GetErrorContext(err error) map[string]interface{} {
	if gowrightErr, ok := err.(*GowrightError); ok {
		return gowrightErr.Context
	}
	return nil
}
