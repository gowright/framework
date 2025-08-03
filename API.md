# Gowright API Documentation

This document provides comprehensive API documentation for the Gowright testing framework.

## Table of Contents

- [Core Framework](#core-framework)
- [Configuration](#configuration)
- [Testing Modules](#testing-modules)
- [Test Types](#test-types)
- [Reporting](#reporting)
- [Utilities](#utilities)

## Core Framework

### Gowright

The main framework struct that orchestrates all testing activities.

```go
type Gowright struct {
    // Private fields
}
```

#### Constructor Functions

##### New(config *Config) *Gowright

Creates a new Gowright instance with the provided configuration.

**Parameters:**
- `config`: Framework configuration. If nil, uses default configuration.

**Returns:**
- `*Gowright`: New framework instance

**Example:**
```go
config := &gowright.Config{
    BrowserConfig: &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    },
}
framework := gowright.New(config)
```

##### NewWithDefaults() *Gowright

Creates a new Gowright instance with default configuration.

**Returns:**
- `*Gowright`: New framework instance with default settings

**Example:**
```go
framework := gowright.NewWithDefaults()
```

##### NewWithOptions(options *GowrightOptions) *Gowright

Creates a new Gowright instance with dependency injection support.

**Parameters:**
- `options`: Configuration options including custom testers

**Returns:**
- `*Gowright`: New framework instance

**Example:**
```go
options := &gowright.GowrightOptions{
    Config:     config,
    UITester:   customUITester,
    APITester:  customAPITester,
}
framework := gowright.NewWithOptions(options)
```

#### Methods

##### Initialize() error

Initializes the framework and all its components.

**Returns:**
- `error`: Error if initialization fails

**Example:**
```go
if err := framework.Initialize(); err != nil {
    log.Fatal(err)
}
```

##### Cleanup() error

Performs cleanup operations for all components.

**Returns:**
- `error`: Error if cleanup fails

**Example:**
```go
defer func() {
    if err := framework.Cleanup(); err != nil {
        log.Printf("Cleanup failed: %v", err)
    }
}()
```

##### Close() error

Alias for Cleanup(). Performs cleanup and closes the framework instance.

**Returns:**
- `error`: Error if cleanup fails

**Example:**
```go
defer framework.Close()
```

##### ExecuteTestSuite() (*TestResults, error)

Executes the current test suite and returns results.

**Returns:**
- `*TestResults`: Test execution results
- `error`: Error if execution fails

**Example:**
```go
results, err := framework.ExecuteTestSuite()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Passed: %d, Failed: %d\n", results.PassedTests, results.FailedTests)
```

## Configuration

### Config

Main configuration struct for the framework.

```go
type Config struct {
    LogLevel             string                `json:"log_level"`
    Parallel             bool                  `json:"parallel"`
    MaxRetries           int                   `json:"max_retries"`
    BrowserConfig        *BrowserConfig        `json:"browser_config,omitempty"`
    APIConfig            *APIConfig            `json:"api_config,omitempty"`
    DatabaseConfig       *DatabaseConfig       `json:"database_config,omitempty"`
    ReportConfig         *ReportConfig         `json:"report_config,omitempty"`
    ParallelRunnerConfig *ParallelRunnerConfig `json:"parallel_runner_config,omitempty"`
}
```

#### Functions

##### DefaultConfig() *Config

Returns a configuration with sensible defaults.

**Returns:**
- `*Config`: Default configuration

**Example:**
```go
config := gowright.DefaultConfig()
config.Parallel = true
```

##### LoadConfigFromFile(filename string) (*Config, error)

Loads configuration from a JSON file.

**Parameters:**
- `filename`: Path to configuration file

**Returns:**
- `*Config`: Loaded configuration
- `error`: Error if loading fails

**Example:**
```go
config, err := gowright.LoadConfigFromFile("config.json")
if err != nil {
    log.Fatal(err)
}
```

### BrowserConfig

Configuration for UI testing.

```go
type BrowserConfig struct {
    Headless   bool          `json:"headless"`
    Timeout    time.Duration `json:"timeout"`
    UserAgent  string        `json:"user_agent,omitempty"`
    WindowSize *WindowSize   `json:"window_size,omitempty"`
}
```

### APIConfig

Configuration for API testing.

```go
type APIConfig struct {
    BaseURL    string            `json:"base_url,omitempty"`
    Timeout    time.Duration     `json:"timeout"`
    Headers    map[string]string `json:"headers,omitempty"`
    AuthConfig *AuthConfig       `json:"auth_config,omitempty"`
}
```

### DatabaseConfig

Configuration for database testing.

```go
type DatabaseConfig struct {
    Connections map[string]*DBConnection `json:"connections"`
}
```

## Testing Modules

### API Testing

#### APITester Interface

```go
type APITester interface {
    Tester
    Get(endpoint string, headers map[string]string) (*APIResponse, error)
    Post(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error)
    Put(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error)
    Delete(endpoint string, headers map[string]string) (*APIResponse, error)
    SetAuth(auth *AuthConfig) error
    ExecuteTest(test *APITest) *TestCaseResult
}
```

#### APITesterImpl

Implementation of APITester using go-resty.

##### NewAPITester(config *APIConfig) *APITesterImpl

Creates a new API tester instance.

**Parameters:**
- `config`: API configuration

**Returns:**
- `*APITesterImpl`: New API tester

**Example:**
```go
config := &gowright.APIConfig{
    BaseURL: "https://api.example.com",
    Timeout: 10 * time.Second,
}
apiTester := gowright.NewAPITester(config)
```

##### Methods

###### Get(endpoint string, headers map[string]string) (*APIResponse, error)

Performs a GET request.

**Parameters:**
- `endpoint`: API endpoint
- `headers`: Request headers

**Returns:**
- `*APIResponse`: HTTP response
- `error`: Error if request fails

**Example:**
```go
response, err := apiTester.Get("/users", map[string]string{
    "Accept": "application/json",
})
```

###### Post(endpoint string, body interface{}, headers map[string]string) (*APIResponse, error)

Performs a POST request.

**Parameters:**
- `endpoint`: API endpoint
- `body`: Request body
- `headers`: Request headers

**Returns:**
- `*APIResponse`: HTTP response
- `error`: Error if request fails

**Example:**
```go
user := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
}
response, err := apiTester.Post("/users", user, nil)
```

### Database Testing

#### DatabaseTester Interface

```go
type DatabaseTester interface {
    Tester
    Connect(connectionName string) error
    Execute(connectionName, query string, args ...interface{}) (*DatabaseResult, error)
    BeginTransaction(connectionName string) (Transaction, error)
    ValidateData(connectionName, query string, expected interface{}) error
    ExecuteTest(test *DatabaseTest) *TestCaseResult
}
```

#### DatabaseTesterImpl

Implementation of DatabaseTester.

##### NewDatabaseTester() *DatabaseTesterImpl

Creates a new database tester instance.

**Returns:**
- `*DatabaseTesterImpl`: New database tester

**Example:**
```go
dbTester := gowright.NewDatabaseTester()
```

##### Methods

###### Execute(connectionName, query string, args ...interface{}) (*DatabaseResult, error)

Executes a SQL query.

**Parameters:**
- `connectionName`: Database connection name
- `query`: SQL query
- `args`: Query parameters

**Returns:**
- `*DatabaseResult`: Query result
- `error`: Error if execution fails

**Example:**
```go
result, err := dbTester.Execute("main", "SELECT * FROM users WHERE id = ?", 1)
```

### UI Testing

#### UITester Interface

```go
type UITester interface {
    Tester
    Navigate(url string) error
    Click(selector string) error
    Type(selector, text string) error
    GetText(selector string) (string, error)
    WaitForElement(selector string, timeout time.Duration) error
    TakeScreenshot(filename string) (string, error)
    GetPageSource() (string, error)
    ExecuteTest(test *UITest) *TestCaseResult
}
```

#### RodUITester

Implementation of UITester using go-rod.

##### NewRodUITester() *RodUITester

Creates a new UI tester instance.

**Returns:**
- `*RodUITester`: New UI tester

**Example:**
```go
uiTester := gowright.NewRodUITester()
```

## Test Types

### APITest

Represents an API test case.

```go
type APITest struct {
    Name     string                 `json:"name"`
    Method   string                 `json:"method"`
    Endpoint string                 `json:"endpoint"`
    Headers  map[string]string      `json:"headers,omitempty"`
    Body     interface{}            `json:"body,omitempty"`
    Expected *APIExpectation        `json:"expected"`
}
```

### DatabaseTest

Represents a database test case.

```go
type DatabaseTest struct {
    Name       string                `json:"name"`
    Connection string                `json:"connection"`
    Setup      []string              `json:"setup,omitempty"`
    Query      string                `json:"query"`
    Expected   *DatabaseExpectation  `json:"expected"`
    Teardown   []string              `json:"teardown,omitempty"`
}
```

### UITest

Represents a UI test case.

```go
type UITest struct {
    Name       string
    URL        string
    Actions    []UIAction
    Assertions []UIAssertion
}
```

### IntegrationTest

Represents an integration test case.

```go
type IntegrationTest struct {
    Name     string              `json:"name"`
    Steps    []IntegrationStep   `json:"steps"`
    Rollback []IntegrationStep   `json:"rollback,omitempty"`
}
```

## Test Results

### TestCaseResult

Represents the result of a single test case execution.

```go
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
```

### TestResults

Holds all test execution results.

```go
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
```

### TestStatus

Represents the status of a test execution.

```go
type TestStatus int

const (
    TestStatusPassed TestStatus = iota
    TestStatusFailed
    TestStatusSkipped
    TestStatusError
)
```

## Reporting

### ReportManager

Coordinates all reporting activities.

```go
type ReportManager struct {
    // Private fields
}
```

#### Methods

##### GenerateReports(results *TestResults) *ReportingSummary

Generates reports using all enabled reporters.

**Parameters:**
- `results`: Test results to report

**Returns:**
- `*ReportingSummary`: Summary of reporting operations

**Example:**
```go
reporter := framework.GetReporter()
summary := reporter.GenerateReports(testResults)
fmt.Printf("Successful reports: %d\n", summary.SuccessfulReports)
```

### Reporter Interface

Interface for all report destinations.

```go
type Reporter interface {
    GenerateReport(results *TestResults) error
    GetName() string
    IsEnabled() bool
}
```

## Utilities

### Test Builders

#### APITestBuilder

Provides a fluent interface for building API tests.

```go
test := gowright.NewAPITestBuilder("Get User", "GET", "/users/1").
    WithTester(apiTester).
    WithHeader("Accept", "application/json").
    ExpectStatus(200).
    ExpectJSONPath("$.id", 1).
    Build()
```

### Assertions

#### TestAssertion

Helper for creating custom assertions.

```go
assertion := gowright.NewTestAssertion("Custom Check")
assertion.Assert(actualValue == expectedValue, "Values should match")
assertion.AssertNotNil(someObject, "Object should not be nil")
```

### Resource Management

#### ResourceManager

Monitors and manages system resources during test execution.

```go
resourceManager := gowright.NewResourceManager(&gowright.ResourceLimits{
    MaxMemoryMB:      512,
    MaxCPUPercent:    70,
    MaxOpenFiles:     50,
    MaxNetworkConns:  25,
})

usage := resourceManager.GetCurrentUsage()
```

### Error Handling

#### RetryWithBackoff

Executes a function with retry logic and exponential backoff.

```go
err := gowright.RetryWithBackoff(context.Background(), retryConfig, func() error {
    return someOperation()
})
```

#### GowrightError

Framework-specific error type with context information.

```go
type GowrightError struct {
    Type    ErrorType              `json:"type"`
    Message string                 `json:"message"`
    Cause   error                  `json:"cause,omitempty"`
    Context map[string]interface{} `json:"context,omitempty"`
}
```

## Examples

### Complete Test Example

```go
func TestCompleteWorkflow(t *testing.T) {
    // Initialize framework
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    err := framework.Initialize()
    require.NoError(t, err)
    
    // Create API test
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    err = apiTester.Initialize(apiTester.GetConfig())
    require.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Execute API test
    response, err := apiTester.Get("/posts/1", nil)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // Create database test
    dbTester := gowright.NewDatabaseTester()
    err = dbTester.Initialize(&gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "test": {
                Driver: "sqlite3",
                DSN:    ":memory:",
            },
        },
    })
    require.NoError(t, err)
    defer dbTester.Cleanup()
    
    // Execute database operations
    _, err = dbTester.Execute("test", "CREATE TABLE test (id INTEGER, name TEXT)")
    require.NoError(t, err)
    
    result, err := dbTester.Execute("test", "INSERT INTO test (id, name) VALUES (1, 'test')")
    require.NoError(t, err)
    assert.Equal(t, int64(1), result.RowsAffected)
    
    // Generate reports
    testResults := &gowright.TestResults{
        SuiteName:    "Complete Workflow Test",
        StartTime:    time.Now().Add(-time.Minute),
        EndTime:      time.Now(),
        TotalTests:   2,
        PassedTests:  2,
        FailedTests:  0,
        SkippedTests: 0,
        TestCases: []gowright.TestCaseResult{
            {
                Name:     "API Test",
                Status:   gowright.TestStatusPassed,
                Duration: time.Second,
            },
            {
                Name:     "Database Test",
                Status:   gowright.TestStatusPassed,
                Duration: 500 * time.Millisecond,
            },
        },
    }
    
    reporter := framework.GetReporter()
    summary := reporter.GenerateReports(testResults)
    assert.Greater(t, summary.SuccessfulReports, 0)
}
```

This API documentation provides comprehensive coverage of the Gowright framework's public interface, making it easy for developers to understand and use the framework effectively.