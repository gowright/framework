package gowright

import (
	"database/sql"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-resty/resty/v2"
)

// Gowright is the main framework struct
type Gowright struct {
	config    *Config
	reporter  *ReportManager
	testSuite *TestSuite
}

// TestSuite represents a collection of tests
type TestSuite struct {
	Name         string
	Tests        []Test
	SetupFunc    func() error
	TeardownFunc func() error
}

// UITester provides browser automation capabilities
type UITester struct {
	browser *rod.Browser
	page    *rod.Page
	config  *BrowserConfig
}

// APITester provides HTTP client capabilities
type APITester struct {
	client *resty.Client
	config *APIConfig
}

// DatabaseTester provides database testing capabilities
type DatabaseTester struct {
	connections map[string]*sql.DB
	config      *DatabaseConfig
}

// IntegrationTester coordinates multi-system tests
type IntegrationTester struct {
	uiTester  *UITester
	apiTester *APITester
	dbTester  *DatabaseTester
}

// ReportManager coordinates all reporting activities
type ReportManager struct {
	config    *ReportConfig
	reporters []Reporter
	results   *TestResults
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
	TestCases    []TestCaseResult `json:"test_cases"`
}

// TestCaseResult represents individual test results
type TestCaseResult struct {
	Name        string        `json:"name"`
	Status      TestStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	Error       error         `json:"error,omitempty"`
	Screenshots []string      `json:"screenshots,omitempty"`
	Logs        []string      `json:"logs,omitempty"`
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

// APIResponse represents an HTTP response
type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       []byte                 `json:"body"`
	JSON       map[string]interface{} `json:"json,omitempty"`
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

// DatabaseResult represents database query results
type DatabaseResult struct {
	Rows         []map[string]interface{} `json:"rows"`
	RowsAffected int64                    `json:"rows_affected"`
	Error        error                    `json:"error,omitempty"`
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

// IntegrationStep represents a single step in an integration test
type IntegrationStep struct {
	Type       StepType    `json:"type"`
	Action     interface{} `json:"action"`
	Validation interface{} `json:"validation,omitempty"`
}