package gowright

import "time"

// Test represents a generic test interface
type Test interface {
	GetName() string
	Execute() error
	GetResult() *TestCaseResult
}

// Reporter interface for all report destinations
type Reporter interface {
	GenerateReport(results *TestResults) error
	GetName() string
}

// Tester interface for different testing modules
type Tester interface {
	Initialize() error
	Cleanup() error
	GetName() string
}

// UITesterInterface defines the contract for UI testing
type UITesterInterface interface {
	Tester
	NavigateTo(url string) error
	Click(selector string) error
	Type(selector, text string) error
	GetText(selector string) (string, error)
	TakeScreenshot() ([]byte, error)
	WaitForElement(selector string, timeout time.Duration) error
}

// APITesterInterface defines the contract for API testing
type APITesterInterface interface {
	Tester
	Get(endpoint string) (*APIResponse, error)
	Post(endpoint string, body interface{}) (*APIResponse, error)
	Put(endpoint string, body interface{}) (*APIResponse, error)
	Delete(endpoint string) (*APIResponse, error)
	SetHeader(key, value string)
	SetAuth(config *AuthConfig) error
}

// DatabaseTesterInterface defines the contract for database testing
type DatabaseTesterInterface interface {
	Tester
	Connect(connectionName string) error
	ExecuteQuery(query string, args ...interface{}) (*DatabaseResult, error)
	ExecuteTransaction(queries []string) error
	ValidateResult(result *DatabaseResult, expected *DatabaseExpectation) error
}

// IntegrationTesterInterface defines the contract for integration testing
type IntegrationTesterInterface interface {
	Tester
	ExecuteStep(step *IntegrationStep) error
	ExecuteTest(test *IntegrationTest) error
	Rollback(steps []IntegrationStep) error
}