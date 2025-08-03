---
title: API Reference
description: Complete API documentation for the Gowright testing framework
---

This document provides comprehensive API documentation for the Gowright testing framework.

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

##### Close() error

Performs cleanup and closes the framework instance.

**Returns:**
- `error`: Error if cleanup fails

**Example:**
```go
defer framework.Close()
```

## Configuration Types

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

## Testing Interfaces

### APITester Interface

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

### UITester Interface

```go
type UITester interface {
    Tester
    Navigate(url string) error
    Click(selector string) error
    Type(selector, text string) error
    GetText(selector string) (string, error)
    WaitForElement(selector string, timeout time.Duration) error
    TakeScreenshot(filename string) (string, error)
    ExecuteTest(test *UITest) *TestCaseResult
}
```

### DatabaseTester Interface

```go
type DatabaseTester interface {
    Tester
    Connect(connectionName string) error
    Execute(connectionName, query string, args ...interface{}) (*DatabaseResult, error)
    BeginTransaction(connectionName string) (Transaction, error)
    ExecuteTest(test *DatabaseTest) *TestCaseResult
}
```

## Test Types

### TestCaseResult

Represents the result of a single test case execution.

```go
type TestCaseResult struct {
    Name        string          `json:"name"`
    Status      TestStatus      `json:"status"`
    Duration    time.Duration   `json:"duration"`
    Error       error           `json:"error,omitempty"`
    Screenshots []string        `json:"screenshots,omitempty"`
    StartTime   time.Time       `json:"start_time"`
    EndTime     time.Time       `json:"end_time"`
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

## Error Handling

### GowrightError

Framework-specific error type with context information.

```go
type GowrightError struct {
    Type    ErrorType              `json:"type"`
    Message string                 `json:"message"`
    Cause   error                  `json:"cause,omitempty"`
    Context map[string]interface{} `json:"context,omitempty"`
}
```

For detailed API documentation, see the specific sections:

- [Configuration Types](/api/configuration/)
- [Testing Interfaces](/api/testing/)
- [Utilities](/api/utilities/)