# Gowright Testing Framework

[![Go Version](https://img.shields.io/badge/Go-1.22.2-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/gowright/framework)](https://goreportcard.com/report/github.com/gowright/framework)

Gowright is a comprehensive testing framework for Go that provides unified testing capabilities across UI (browser, mobile), API, database, and integration testing scenarios. Built with modern Go practices and designed for scalability and maintainability.

## 🚀 Features

- **🖥️ UI Testing**: Browser automation using [go-rod/rod](https://github.com/go-rod/rod) with support for mobile testing
- **🔌 API Testing**: HTTP/REST API testing using [go-resty/resty](https://github.com/go-resty/resty/v2)
- **🗄️ Database Testing**: Database operations and validations with multiple driver support
- **🔗 Integration Testing**: Complex workflows spanning multiple systems
- **📊 Flexible Reporting**: Multiple report formats (JSON, HTML) and integrations (Jira Xray, AIOTest, Report Portal)
- **🧪 Testify Integration**: Compatible with [stretchr/testify](https://github.com/stretchr/testify) for assertions and mocks
- **⚙️ Configuration Management**: Hierarchical configuration through code, files, and environment variables
- **🔧 Dependency Injection**: Modular architecture with pluggable components

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Core Concepts](#core-concepts)
- [Examples](#examples)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)

## 📦 Installation

### Prerequisites

- Go 1.22.2 or higher
- Git

### Install

```bash
# Clone the repository
git clone https://github.com/gowright/framework.git
cd framework

# Install dependencies
go mod download

# Run tests to verify installation
go test ./...
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github/gowright/framework/pkg/gowright"
)

func main() {
    // Create a new Gowright instance with default configuration
    gw := gowright.NewWithDefaults()

    // Initialize the framework
    if err := gw.Initialize(); err != nil {
        panic(err)
    }
    defer gw.Cleanup()

    // Create a test suite
    testSuite := &gowright.TestSuite{
        Name: "My Test Suite",
        SetupFunc: func() error {
            fmt.Println("Setting up test suite...")
            return nil
        },
        TeardownFunc: func() error {
            fmt.Println("Tearing down test suite...")
            return nil
        },
    }

    gw.SetTestSuite(testSuite)

    // Execute the test suite
    results, err := gw.ExecuteTestSuite()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Tests completed: %d passed, %d failed\n",
        results.PassedTests, results.FailedTests)
}
```

### UI Testing Example

```go
package main

import (
    "github/gowright/framework/pkg/gowright"
)

func main() {
    gw := gowright.NewWithDefaults()
    gw.Initialize()
    defer gw.Cleanup()

    uiTester := gw.GetUITester()

    // Navigate to a page
    err := uiTester.Navigate("https://example.com")
    if err != nil {
        panic(err)
    }

    // Click on an element
    err = uiTester.Click("#submit-button")
    if err != nil {
        panic(err)
    }

    // Assert element is visible
    visible, err := uiTester.AssertElementVisible("#success-message")
    if err != nil || !visible {
        panic("Element not visible")
    }
}
```

### API Testing Example

```go
package main

import (
    "github/gowright/framework/pkg/gowright"
)

func main() {
    gw := gowright.NewWithDefaults()
    gw.Initialize()
    defer gw.Cleanup()

    apiTester := gw.GetAPITester()

    // Make a GET request
    response, err := apiTester.Get("/api/users", nil)
    if err != nil {
        panic(err)
    }

    // Assert status code
    if response.StatusCode != 200 {
        panic("Expected status 200")
    }

    // Assert response body
    users, err := response.JSON()
    if err != nil {
        panic(err)
    }

    if len(users) == 0 {
        panic("Expected users in response")
    }
}
```

## ⚙️ Configuration

The framework supports hierarchical configuration through:

1. **Code**: Direct configuration object creation
2. **Files**: JSON configuration files
3. **Environment**: Environment variables

### Configuration Structure

```go
config := &gowright.Config{
    LogLevel: "info",
    Parallel: true,
    MaxRetries: 3,
    BrowserConfig: &gowright.BrowserConfig{
        Headless: true,
        Timeout: 30 * time.Second,
        WindowSize: &gowright.WindowSize{
            Width:  1920,
            Height: 1080,
        },
    },
    APIConfig: &gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 30 * time.Second,
    },
    DatabaseConfig: &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "default": {
                Driver:   "postgres",
                Host:     "localhost",
                Port:     5432,
                Database: "testdb",
                Username: "user",
                Password: "password",
            },
        },
    },
    ReportConfig: &gowright.ReportConfig{
        LocalReports: gowright.LocalReportConfig{
            JSON: true,
            HTML: true,
            OutputDir: "./reports",
        },
        RemoteReports: gowright.RemoteReportConfig{
            JiraXray: &gowright.JiraXrayConfig{
                BaseURL: "https://your-domain.atlassian.net",
                Username: "your-username",
                APIToken: "your-api-token",
                ProjectKey: "TEST",
            },
        },
    },
}
```

### Configuration File

Create a `gowright-config.json` file:

```json
{
  "log_level": "info",
  "parallel": false,
  "max_retries": 3,
  "browser_config": {
    "headless": true,
    "timeout": 30000000000,
    "window_size": {
      "width": 1920,
      "height": 1080
    }
  },
  "api_config": {
    "timeout": 30000000000
  },
  "database_config": {
    "connections": {}
  },
  "report_config": {
    "local_reports": {
      "json": true,
      "html": true,
      "output_dir": "./reports"
    },
    "remote_reports": {}
  }
}
```

### Environment Variables

Set environment variables for configuration:

```bash
export GOWRIGHT_LOG_LEVEL=debug
export GOWRIGHT_BROWSER_HEADLESS=true
export GOWRIGHT_API_BASE_URL=https://api.example.com
```

## 🏗️ Core Concepts

### Test Suite

A test suite is a collection of related tests with setup and teardown functions:

```go
testSuite := &gowright.TestSuite{
    Name: "User Management Tests",
    Tests: []gowright.Test{
        // Test implementations
    },
    SetupFunc: func() error {
        // Setup database, create test data
        return nil
    },
    TeardownFunc: func() error {
        // Clean up test data
        return nil
    },
}
```

### Test Types

The framework supports multiple test types:

- **UI Tests**: Browser automation and mobile testing
- **API Tests**: HTTP/REST API testing
- **Database Tests**: Database operations and validations
- **Integration Tests**: Multi-system workflows

### Assertions

Built-in assertion methods for common validations:

```go
// UI Assertions
uiTester.AssertElementVisible("#button")
uiTester.AssertTextEquals("#title", "Expected Title")
uiTester.AssertElementCount(".item", 5)

// API Assertions
apiTester.AssertStatusCode(200)
apiTester.AssertHeaderEquals("Content-Type", "application/json")
apiTester.AssertJSONPathEquals("$.user.name", "John Doe")

// Database Assertions
dbTester.AssertRowCount("SELECT * FROM users", 10)
dbTester.AssertColumnValue("SELECT name FROM users WHERE id = 1", "name", "John")
```

### Error Handling

Structured error handling with contextual information:

```go
err := gowright.NewGowrightError(
    gowright.BrowserError,
    "Failed to navigate to page",
    originalError,
).WithContext("url", "https://example.com")
```

## 📚 Examples

Check the `examples/` directory for comprehensive examples:

- `basic_usage.go` - Basic framework usage
- `api_testing_example.go` - API testing examples
- `reporting_example.go` - Reporting configuration
- `assertion_reporting_example.go` - Assertion and reporting
- `test_suite_with_assertions.go` - Complete test suite example

## 🔧 API Reference

### Core Interfaces

- `Tester`: Base interface for all testing modules
- `UITester`: Browser and mobile UI testing
- `APITester`: HTTP/REST API testing
- `DatabaseTester`: Database operations and validations
- `IntegrationTester`: Multi-system integration testing
- `Reporter`: Pluggable reporting system

### Main Types

- `Gowright`: Main framework orchestrator
- `TestSuite`: Collection of tests with setup/teardown
- `TestResults`: Test execution results
- `Config`: Framework configuration
- `GowrightError`: Structured error handling

### Key Methods

```go
// Framework initialization
gw := gowright.New(config)
gw.Initialize()
defer gw.Cleanup()

// Test suite management
gw.SetTestSuite(suite)
results, err := gw.ExecuteTestSuite()

// Tester access
uiTester := gw.GetUITester()
apiTester := gw.GetAPITester()
dbTester := gw.GetDatabaseTester()
```

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestUITester ./pkg/gowright

# Run tests with verbose output
go test -v ./...
```

## 📊 Reporting

The framework generates comprehensive reports in multiple formats:

- **JSON Reports**: Machine-readable test results
- **HTML Reports**: Human-readable test reports
- **Remote Integration**: Jira Xray, AIOTest, Report Portal

Reports include:

- Test execution status and timing
- Screenshots for UI tests
- Error details and stack traces
- Performance metrics
- Custom assertions and validations

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/gowright/framework.git
cd framework

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run

# Build the project
go build ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [go-rod/rod](https://github.com/go-rod/rod) for browser automation
- [go-resty/resty](https://github.com/go-resty/resty/v2) for HTTP client functionality
- [stretchr/testify](https://github.com/stretchr/testify) for testing utilities

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/gowright/framework/issues)
- **Documentation**: [GitHub Wiki](https://github.com/gowright/framework/wiki)
- **Discussions**: [GitHub Discussions](https://github.com/gowright/framework/discussions)

---

**Gowright** - A comprehensive testing framework for Go applications.
