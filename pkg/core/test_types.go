package core

import "time"

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
	Name     string            `json:"name"`
	Method   string            `json:"method"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
	Expected *APIExpectation   `json:"expected"`
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
	Name       string               `json:"name"`
	Connection string               `json:"connection"`
	Setup      []string             `json:"setup,omitempty"`
	Query      string               `json:"query"`
	Expected   *DatabaseExpectation `json:"expected"`
	Teardown   []string             `json:"teardown,omitempty"`
}

// DatabaseExpectation represents expected database results
type DatabaseExpectation struct {
	RowCount     int                      `json:"row_count,omitempty"`
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	RowsAffected int64                    `json:"rows_affected,omitempty"`
}

// IntegrationTest represents a complex integration test
type IntegrationTest struct {
	Name     string            `json:"name"`
	Steps    []IntegrationStep `json:"steps"`
	Rollback []IntegrationStep `json:"rollback,omitempty"`
}

// IntegrationStepType represents the type of integration step
type IntegrationStepType int

const (
	StepTypeUI IntegrationStepType = iota
	StepTypeAPI
	StepTypeDatabase
	StepTypeMobile
)

// String returns the string representation of IntegrationStepType
func (ist IntegrationStepType) String() string {
	switch ist {
	case StepTypeUI:
		return "UI"
	case StepTypeAPI:
		return "API"
	case StepTypeDatabase:
		return "Database"
	case StepTypeMobile:
		return "Mobile"
	default:
		return "Unknown"
	}
}

// IntegrationStep represents a single step in an integration test
type IntegrationStep struct {
	Name       string              `json:"name"`
	Type       IntegrationStepType `json:"type"`
	Action     interface{}         `json:"action"`
	Validation interface{}         `json:"validation,omitempty"`
	Timeout    time.Duration       `json:"timeout,omitempty"`
	RetryCount int                 `json:"retry_count,omitempty"`
}

// IntegrationStepAction is an interface for different types of step actions
type IntegrationStepAction interface {
	GetType() string
}

// IntegrationStepValidation is an interface for different types of step validations
type IntegrationStepValidation interface {
	GetType() string
}

// UIStepAction represents a UI action in an integration step
type UIStepAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// GetType returns the action type
func (usa *UIStepAction) GetType() string {
	return "UI"
}

// APIStepAction represents an API action in an integration step
type APIStepAction struct {
	Method   string            `json:"method"`
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
}

// GetType returns the action type
func (asa *APIStepAction) GetType() string {
	return "API"
}

// DatabaseStepAction represents a database action in an integration step
type DatabaseStepAction struct {
	Connection string        `json:"connection"`
	Query      string        `json:"query"`
	Args       []interface{} `json:"args,omitempty"`
}

// GetType returns the action type
func (dsa *DatabaseStepAction) GetType() string {
	return "Database"
}

// APIStepValidation represents validation criteria for API responses
type APIStepValidation struct {
	ExpectedStatusCode int                    `json:"expected_status_code,omitempty"`
	ExpectedHeaders    map[string]string      `json:"expected_headers,omitempty"`
	ExpectedBody       interface{}            `json:"expected_body,omitempty"`
	JSONPath           map[string]interface{} `json:"json_path,omitempty"`
}

// GetType returns the validation type
func (asv *APIStepValidation) GetType() string {
	return "API"
}

// DatabaseStepValidation represents validation criteria for database results
type DatabaseStepValidation struct {
	ExpectedRowCount *int                     `json:"expected_row_count,omitempty"`
	ExpectedRows     []map[string]interface{} `json:"expected_rows,omitempty"`
	ExpectedAffected *int64                   `json:"expected_affected,omitempty"`
}

// GetType returns the validation type
func (dsv *DatabaseStepValidation) GetType() string {
	return "Database"
}
