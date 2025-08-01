package gowright

import (
	"time"
)

// Tester represents the base interface for all testing modules
type Tester interface {
	// Initialize sets up the tester with the provided configuration
	Initialize(config interface{}) error
	
	// Cleanup performs any necessary cleanup operations
	Cleanup() error
	
	// GetName returns the name of the tester
	GetName() string
}

// UITester interface defines methods for UI testing capabilities
type UITester interface {
	Tester
	
	// Navigate navigates to the specified URL
	Navigate(url string) error
	
	// Click clicks on an element identified by the selector
	Click(selector string) error
	
	// Type types text into an element identified by the selector
	Type(selector, text string) error
	
	// GetText retrieves text from an element identified by the selector
	GetText(selector string) (string, error)
	
	// WaitForElement waits for an element to be present
	WaitForElement(selector string, timeout time.Duration) error
	
	// TakeScreenshot captures a screenshot and returns the file path
	TakeScreenshot(filename string) (string, error)
	
	// GetPageSource returns the current page source
	GetPageSource() (string, error)
}

// APITester interface defines methods for API testing capabilities
type APITester interface {
	Tester
	
	// Get performs a GET request to the specified endpoint
	Get(endpoint string, headers map[string]string) (*APIResponse, error)
	
	// Post performs a POST request to the specified endpoint
	Post(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error)
	
	// Put performs a PUT request to the specified endpoint
	Put(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error)
	
	// Delete performs a DELETE request to the specified endpoint
	Delete(endpoint string, headers map[string]string) (*APIResponse, error)
	
	// SetAuth sets authentication for API requests
	SetAuth(auth *AuthConfig) error
}

// DatabaseTester interface defines methods for database testing capabilities
type DatabaseTester interface {
	Tester
	
	// Connect establishes a connection to the database
	Connect(connectionName string) error
	
	// Execute executes a SQL query and returns the result
	Execute(connectionName, query string, args ...interface{}) (*DatabaseResult, error)
	
	// BeginTransaction starts a new database transaction
	BeginTransaction(connectionName string) (Transaction, error)
	
	// ValidateData validates data against expected results
	ValidateData(connectionName, query string, expected interface{}) error
}

// IntegrationTester interface defines methods for integration testing
type IntegrationTester interface {
	Tester
	
	// ExecuteStep executes a single integration step
	ExecuteStep(step *IntegrationStep) error
	
	// ExecuteWorkflow executes a complete integration workflow
	ExecuteWorkflow(steps []IntegrationStep) error
	
	// Rollback performs rollback operations for failed tests
	Rollback(steps []IntegrationStep) error
}

// Reporter interface defines methods for test result reporting
type Reporter interface {
	// GenerateReport generates a report from test results
	GenerateReport(results *TestResults) error
	
	// GetName returns the name of the reporter
	GetName() string
	
	// IsEnabled returns whether this reporter is enabled
	IsEnabled() bool
}

// Transaction interface defines database transaction operations
type Transaction interface {
	// Commit commits the transaction
	Commit() error
	
	// Rollback rolls back the transaction
	Rollback() error
	
	// Execute executes a query within the transaction
	Execute(query string, args ...interface{}) (*DatabaseResult, error)
}

// APIResponse represents an HTTP response
type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string][]string    `json:"headers"`
	Body       []byte                 `json:"body"`
	Duration   time.Duration          `json:"duration"`
}

// DatabaseResult represents a database query result
type DatabaseResult struct {
	Rows         []map[string]interface{} `json:"rows"`
	RowsAffected int64                    `json:"rows_affected"`
	Duration     time.Duration            `json:"duration"`
}

// StepType represents the type of integration step
type StepType int

const (
	StepTypeUI StepType = iota
	StepTypeAPI
	StepTypeDatabase
)

// String returns the string representation of StepType
func (st StepType) String() string {
	switch st {
	case StepTypeUI:
		return "ui"
	case StepTypeAPI:
		return "api"
	case StepTypeDatabase:
		return "database"
	default:
		return "unknown"
	}
}

// IntegrationStep represents a single step in an integration test
type IntegrationStep struct {
	Type       StepType    `json:"type"`
	Action     interface{} `json:"action"`
	Validation interface{} `json:"validation"`
	Name       string      `json:"name"`
}