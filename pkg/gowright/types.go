package gowright

// TestStatus represents the status of a test execution
type TestStatus int

const (
	TestStatusPassed TestStatus = iota
	TestStatusFailed
	TestStatusSkipped
)

// String returns the string representation of TestStatus
func (ts TestStatus) String() string {
	switch ts {
	case TestStatusPassed:
		return "PASSED"
	case TestStatusFailed:
		return "FAILED"
	case TestStatusSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
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

// GowrightError represents framework-specific errors
type GowrightError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *GowrightError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// StepType represents the type of integration test step
type StepType int

const (
	UIStep StepType = iota
	APIStep
	DatabaseStep
)

// WindowSize represents browser window dimensions
type WindowSize struct {
	Width  int
	Height int
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type     string // "basic", "bearer", "oauth2", etc.
	Username string
	Password string
	Token    string
}

// DBConnection represents a database connection configuration
type DBConnection struct {
	Driver       string
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
}