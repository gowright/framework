package gowright

import (
	"encoding/json"
	"errors"
	"time"
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

// TestStatus represents the status of a test execution
type TestStatus int

const (
	TestStatusPassed TestStatus = iota
	TestStatusFailed
	TestStatusSkipped
	TestStatusError
)

// String returns the string representation of TestStatus
func (ts TestStatus) String() string {
	switch ts {
	case TestStatusPassed:
		return "passed"
	case TestStatusFailed:
		return "failed"
	case TestStatusSkipped:
		return "skipped"
	case TestStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// AssertionStep represents a single assertion step in a test
type AssertionStep struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      TestStatus    `json:"status"`
	Error       error         `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Expected    interface{}   `json:"expected,omitempty"`
	Actual      interface{}   `json:"actual,omitempty"`
}

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
	SuiteName    string            `json:"suite_name"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	TotalTests   int               `json:"total_tests"`
	PassedTests  int               `json:"passed_tests"`
	FailedTests  int               `json:"failed_tests"`
	SkippedTests int               `json:"skipped_tests"`
	ErrorTests   int               `json:"error_tests"`
	TestCases    []TestCaseResult  `json:"test_cases"`
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

// UITest represents a UI test case
type UITest struct {
	Name       string
	URL        string
	Actions    []UIAction
	Assertions []UIAssertion
}

// UIAction represents a UI interaction
type UIAction struct {
	Type     string      `json:"type"`
	Selector string      `json:"selector,omitempty"`
	Value    string      `json:"value,omitempty"`
	Options  interface{} `json:"options,omitempty"`
}

// UIAssertion represents a UI validation
type UIAssertion struct {
	Type     string      `json:"type"`
	Selector string      `json:"selector,omitempty"`
	Expected interface{} `json:"expected"`
	Options  interface{} `json:"options,omitempty"`
}

// APITest represents an API test case
type APITest struct {
	Name     string                 `json:"name"`
	Method   string                 `json:"method"`
	Endpoint string                 `json:"endpoint"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Body     interface{}            `json:"body,omitempty"`
	Expected *APIExpectation        `json:"expected"`
}

// APIExpectation represents expected API response
type APIExpectation struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
	JSONPath   map[string]interface{} `json:"json_path,omitempty"`
}

// DatabaseTest represents a database test case
type DatabaseTest struct {
	Name       string                `json:"name"`
	Connection string                `json:"connection"`
	Setup      []string              `json:"setup,omitempty"`
	Query      string                `json:"query"`
	Expected   *DatabaseExpectation  `json:"expected"`
	Teardown   []string              `json:"teardown,omitempty"`
}

// DatabaseExpectation represents expected database results
type DatabaseExpectation struct {
	RowCount     int                      `json:"row_count,omitempty"`
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	RowsAffected int64                    `json:"rows_affected,omitempty"`
}

// IntegrationTest represents a complex integration test
type IntegrationTest struct {
	Name     string              `json:"name"`
	Steps    []IntegrationStep   `json:"steps"`
	Rollback []IntegrationStep   `json:"rollback,omitempty"`
}

// UIStepAction represents a UI action in an integration step
type UIStepAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// APIStepAction represents an API action in an integration step
type APIStepAction struct {
	Method   string            `json:"method"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
}

// DatabaseStepAction represents a database action in an integration step
type DatabaseStepAction struct {
	Connection string        `json:"connection"`
	Query      string        `json:"query"`
	Args       []interface{} `json:"args,omitempty"`
}

// APIStepValidation represents validation criteria for API responses
type APIStepValidation struct {
	ExpectedStatusCode int                    `json:"expected_status_code,omitempty"`
	ExpectedHeaders    map[string]string      `json:"expected_headers,omitempty"`
	ExpectedBody       interface{}            `json:"expected_body,omitempty"`
	JSONPath           map[string]interface{} `json:"json_path,omitempty"`
}

// DatabaseStepValidation represents validation criteria for database results
type DatabaseStepValidation struct {
	ExpectedRowCount *int                     `json:"expected_row_count,omitempty"`
	ExpectedRows     []map[string]interface{} `json:"expected_rows,omitempty"`
	ExpectedAffected *int64                   `json:"expected_affected,omitempty"`
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