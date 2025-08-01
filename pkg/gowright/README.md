# Gowright Testing Framework

Gowright is a comprehensive testing framework for Go that provides unified testing capabilities across UI (browser, mobile), API, database, and integration testing scenarios.

## Features

- **UI Testing**: Browser automation using go-rod/rod with support for mobile testing
- **API Testing**: HTTP/REST API testing using go-resty/resty
- **Database Testing**: Database operations and validations with multiple driver support
- **Integration Testing**: Complex workflows spanning multiple systems
- **Flexible Reporting**: Multiple report formats (JSON, HTML) and integrations (Jira Xray, AIOTest, Report Portal)
- **Testify Integration**: Compatible with stretchr/testify for assertions and mocks

## Quick Start

```go
package main

import (
    "github/gowright/framework/pkg/gowright"
)

func main() {
    // Create a new Gowright instance with default configuration
    gw := gowright.NewWithDefaults()
    
    // Create a test suite
    testSuite := &gowright.TestSuite{
        Name: "My Test Suite",
        SetupFunc: func() error {
            // Setup logic
            return nil
        },
        TeardownFunc: func() error {
            // Cleanup logic
            return nil
        },
    }
    
    gw.SetTestSuite(testSuite)
}
```

## Configuration

The framework supports hierarchical configuration through:

1. **Code**: Direct configuration object creation
2. **Files**: JSON configuration files
3. **Environment**: Environment variables

### Example Configuration

```go
config := &gowright.Config{
    LogLevel: "info",
    Parallel: true,
    BrowserConfig: &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    },
    APIConfig: &gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 30 * time.Second,
    },
    ReportConfig: &gowright.ReportConfig{
        LocalReports: gowright.LocalReportConfig{
            JSON: true,
            HTML: true,
            OutputDir: "./reports",
        },
    },
}

gw := gowright.New(config)
```

## Core Interfaces

The framework is built around several key interfaces:

- `Tester`: Base interface for all testing modules
- `UITester`: Browser and mobile UI testing
- `APITester`: HTTP/REST API testing  
- `DatabaseTester`: Database operations and validations
- `IntegrationTester`: Multi-system integration testing
- `Reporter`: Pluggable reporting system

## Error Handling

The framework provides structured error handling with contextual information:

```go
err := gowright.NewGowrightError(
    gowright.BrowserError,
    "Failed to navigate to page",
    originalError,
).WithContext("url", "https://example.com")
```

## Dependencies

- `github.com/go-rod/rod`: Browser automation
- `github.com/go-resty/resty/v2`: HTTP client
- `github.com/stretchr/testify`: Testing utilities and assertions

## License

This project is licensed under the MIT License.