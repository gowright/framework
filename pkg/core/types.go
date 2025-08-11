package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gowright/framework/pkg/assertions"
)

// TestSuite represents a collection of tests with setup and teardown
type TestSuite struct {
	Name         string
	Tests        []Test
	SetupFunc    func() error
	TeardownFunc func() error
}

// Test represents a generic test interface
type Test interface {
	GetName() string
	Execute() *TestCaseResult
}

// Re-export types from assertions package
type TestStatus = assertions.TestStatus
type AssertionStep = assertions.AssertionStep

// Re-export constants from assertions package
const (
	TestStatusPassed  = assertions.TestStatusPassed
	TestStatusFailed  = assertions.TestStatusFailed
	TestStatusSkipped = assertions.TestStatusSkipped
	TestStatusError   = assertions.TestStatusError
)

// TestCaseResult represents the result of a single test case execution
type TestCaseResult struct {
	Name        string          `json:"name"`
	Status      TestStatus      `json:"status"`
	Duration    time.Duration   `json:"duration"`
	Error       error           `json:"error,omitempty"`
	Screenshots []string        `json:"screenshots,omitempty"`
	Logs        []string        `json:"logs,omitempty"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	Steps       []AssertionStep `json:"steps,omitempty"`
}

// TestResults holds all test execution results
type TestResults struct {
	SuiteName    string           `json:"suite_name"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	TotalTests   int              `json:"total_tests"`
	PassedTests  int              `json:"passed_tests"`
	FailedTests  int              `json:"failed_tests"`
	SkippedTests int              `json:"skipped_tests"`
	ErrorTests   int              `json:"error_tests"`
	TestCases    []TestCaseResult `json:"test_cases"`
}

// ErrorType represents different types of framework errors
type ErrorType int

const (
	ConfigurationError ErrorType = iota
	BrowserError
	APIError
	DatabaseError
	ReportingError
	AssertionError
	TestExecutionErrorType
	TestSetupErrorType
	ValidationError
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ConfigurationError:
		return "configuration_error"
	case BrowserError:
		return "browser_error"
	case APIError:
		return "api_error"
	case DatabaseError:
		return "database_error"
	case ReportingError:
		return "reporting_error"
	case AssertionError:
		return "assertion_error"
	case TestExecutionErrorType:
		return "test_execution_error"
	case TestSetupErrorType:
		return "test_setup_error"
	case ValidationError:
		return "validation_error"
	default:
		return "unknown_error"
	}
}

// GowrightError represents framework-specific errors
type GowrightError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (ge *GowrightError) Error() string {
	if ge.Cause != nil {
		return ge.Message + ": " + ge.Cause.Error()
	}
	return ge.Message
}

// NewGowrightError creates a new GowrightError
func NewGowrightError(errorType ErrorType, message string, cause error) *GowrightError {
	return &GowrightError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context information to the error
func (ge *GowrightError) WithContext(key string, value interface{}) *GowrightError {
	if ge.Context == nil {
		ge.Context = make(map[string]interface{})
	}
	ge.Context[key] = value
	return ge
}

// APIResponse represents an HTTP response
type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body"`
	JSON       map[string]interface{} `json:"json,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
}

// DatabaseResult represents a database query result
type DatabaseResult struct {
	Rows         []map[string]interface{} `json:"rows"`
	RowCount     int                      `json:"row_count"`
	RowsAffected int64                    `json:"rows_affected"`
	Duration     time.Duration            `json:"duration"`
}

// MarshalJSON implements custom JSON marshaling for TestCaseResult
func (tcr TestCaseResult) MarshalJSON() ([]byte, error) {
	type Alias TestCaseResult

	// Create a temporary struct with string error field
	temp := struct {
		Alias
		Error string `json:"error,omitempty"`
	}{
		Alias: Alias(tcr),
	}

	// Convert error to string if present
	if tcr.Error != nil {
		temp.Error = tcr.Error.Error()
	}

	return json.Marshal(temp)
}

// UnmarshalJSON implements custom JSON unmarshaling for TestCaseResult
func (tcr *TestCaseResult) UnmarshalJSON(data []byte) error {
	type Alias TestCaseResult

	// Create a temporary struct with string error field
	temp := struct {
		*Alias
		Error string `json:"error,omitempty"`
	}{
		Alias: (*Alias)(tcr),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Convert string error back to error type if present
	if temp.Error != "" {
		tcr.Error = errors.New(temp.Error)
	}

	return nil
}
