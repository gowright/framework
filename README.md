# Gowright Testing Framework

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()
[![Documentation](https://img.shields.io/badge/docs-Docsify-blue)](https://gowright.github.io/framework/)

Gowright is a comprehensive testing framework for Go that provides unified testing capabilities across UI (browser, mobile), API, database, and integration testing scenarios. Built with a focus on simplicity, performance, and extensibility.

## Features

- **🌐 UI Testing**: Browser automation using Chrome DevTools Protocol via [go-rod/rod](https://github.com/go-rod/rod)
- **📱 Mobile Testing**: Mobile UI testing with device emulation and touch interactions
- **🔌 API Testing**: HTTP/REST API testing with [go-resty/resty](https://github.com/go-resty/resty/v2)
- **🗄️ Database Testing**: Multi-database support with transaction management
- **🔗 Integration Testing**: Complex workflows spanning multiple systems
- **📊 Flexible Reporting**: Local (JSON, HTML) and remote reporting (Jira Xray, AIOTest, Report Portal)
- **🧪 Testify Integration**: Compatible with [stretchr/testify](https://github.com/stretchr/testify)
- **⚡ Parallel Execution**: Concurrent test execution with resource management
- **🛡️ Error Recovery**: Graceful error handling and retry mechanisms

## Quick Start

### Installation

```bash
go get github/gowright/framework
```

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    
    "github/gowright/framework/pkg/gowright"
)

func main() {
    // Create framework with default configuration
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Initialize the framework
    if err := framework.Initialize(); err != nil {
        panic(err)
    }
    
    fmt.Println("Gowright framework initialized successfully!")
}
```

## Configuration

### Basic Configuration

```go
config := &gowright.Config{
    BrowserConfig: &gowright.BrowserConfig{
        Headless:   true,
        Timeout:    30 * time.Second,
        WindowSize: &gowright.WindowSize{Width: 1920, Height: 1080},
    },
    APIConfig: &gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
        Headers: map[string]string{
            "User-Agent": "Gowright-Test-Client",
        },
    },
    DatabaseConfig: &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "main": {
                Driver: "postgres",
                DSN:    "postgres://user:pass@localhost/testdb?sslmode=disable",
            },
        },
    },
    ReportConfig: &gowright.ReportConfig{
        LocalReports: gowright.LocalReportConfig{
            JSON:      true,
            HTML:      true,
            OutputDir: "./test-reports",
        },
    },
}

framework := gowright.New(config)
```

### Configuration from File

```go
config, err := gowright.LoadConfigFromFile("gowright-config.json")
if err != nil {
    panic(err)
}

framework := gowright.New(config)
```

Example `gowright-config.json`:

```json
{
  "log_level": "info",
  "parallel": true,
  "max_retries": 3,
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "window_size": {
      "width": 1920,
      "height": 1080
    }
  },
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "10s",
    "headers": {
      "User-Agent": "Gowright-Test-Client"
    }
  },
  "report_config": {
    "local_reports": {
      "json": true,
      "html": true,
      "output_dir": "./test-reports"
    }
  }
}
```

## Testing Modules

### API Testing

```go
package main

import (
    "net/http"
    "testing"
    
    "github/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestAPIEndpoint(t *testing.T) {
    // Create API tester
    config := &gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    }
    
    apiTester := gowright.NewAPITester(config)
    err := apiTester.Initialize(config)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Test GET request
    response, err := apiTester.Get("/posts/1", nil)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // Test with API test builder
    test := gowright.NewAPITestBuilder("Get Post", "GET", "/posts/1").
        WithTester(apiTester).
        ExpectStatus(http.StatusOK).
        ExpectJSONPath("$.id", 1).
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Database Testing

```go
func TestDatabaseOperations(t *testing.T) {
    // Create database tester
    dbTester := gowright.NewDatabaseTester()
    
    config := &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "test": {
                Driver: "sqlite3",
                DSN:    ":memory:",
            },
        },
    }
    
    err := dbTester.Initialize(config)
    assert.NoError(t, err)
    defer dbTester.Cleanup()
    
    // Execute setup
    _, err = dbTester.Execute("test", `
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL
        )
    `)
    assert.NoError(t, err)
    
    // Test database operations
    dbTest := &gowright.DatabaseTest{
        Name:       "User Creation Test",
        Connection: "test",
        Setup: []string{
            "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')",
        },
        Query: "SELECT COUNT(*) as count FROM users WHERE name = 'John Doe'",
        Expected: &gowright.DatabaseExpectation{
            RowCount: 1,
        },
        Teardown: []string{
            "DELETE FROM users WHERE email = 'john@example.com'",
        },
    }
    
    result := dbTester.ExecuteTest(dbTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### UI Testing

```go
func TestWebApplication(t *testing.T) {
    // Create UI tester
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    assert.NoError(t, err)
    defer uiTester.Cleanup()
    
    // Navigate to page
    err = uiTester.Navigate("https://example.com")
    assert.NoError(t, err)
    
    // Interact with elements
    err = uiTester.Click("button#submit")
    assert.NoError(t, err)
    
    // Wait for element
    err = uiTester.WaitForElement(".success-message", 5*time.Second)
    assert.NoError(t, err)
    
    // Take screenshot
    screenshotPath, err := uiTester.TakeScreenshot("test-result.png")
    assert.NoError(t, err)
    assert.NotEmpty(t, screenshotPath)
}
```

### Integration Testing

```go
func TestCompleteWorkflow(t *testing.T) {
    // Create integration tester
    integrationTester := gowright.NewIntegrationTester(nil, nil, nil)
    
    // Define integration test
    integrationTest := &gowright.IntegrationTest{
        Name: "User Registration Workflow",
        Steps: []gowright.IntegrationStep{
            {
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "POST",
                    Endpoint: "/api/users",
                    Body: map[string]interface{}{
                        "name":  "Test User",
                        "email": "test@example.com",
                    },
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: http.StatusCreated,
                },
                Name: "Create User via API",
            },
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "SELECT COUNT(*) FROM users WHERE email = ?",
                    Args:       []interface{}{"test@example.com"},
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                },
                Name: "Verify User in Database",
            },
        },
    }
    
    result := integrationTester.ExecuteTest(integrationTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Test Suites

### Creating Test Suites

```go
func TestCompleteTestSuite(t *testing.T) {
    // Create framework
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Create test suite
    testSuite := &gowright.TestSuite{
        Name: "Complete Application Test Suite",
        SetupFunc: func() error {
            // Suite-level setup
            return nil
        },
        TeardownFunc: func() error {
            // Suite-level teardown
            return nil
        },
        Tests: []gowright.Test{
            // Add your tests here
        },
    }
    
    framework.SetTestSuite(testSuite)
    
    // Execute test suite
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    assert.Greater(t, results.PassedTests, 0)
}
```

### Parallel Test Execution

```go
config := &gowright.Config{
    Parallel: true,
    ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
        MaxConcurrency: 4,
        ResourceLimits: gowright.ResourceLimits{
            MaxMemoryMB:      1024,
            MaxCPUPercent:    80,
            MaxOpenFiles:     100,
            MaxNetworkConns:  50,
        },
    },
}

framework := gowright.New(config)
```

## Reporting

### Local Reports

Gowright automatically generates local reports in JSON and HTML formats:

```go
config := &gowright.ReportConfig{
    LocalReports: gowright.LocalReportConfig{
        JSON:      true,
        HTML:      true,
        OutputDir: "./test-reports",
    },
}
```

### Remote Reporting

Configure remote reporting to popular test management platforms:

```go
config := &gowright.ReportConfig{
    RemoteReports: gowright.RemoteReportConfig{
        JiraXray: &gowright.JiraXrayConfig{
            URL:        "https://your-jira.atlassian.net",
            Username:   "your-username",
            Password:   "your-api-token",
            ProjectKey: "TEST",
        },
        AIOTest: &gowright.AIOTestConfig{
            URL:       "https://your-aiotest.com",
            APIKey:    "your-api-key",
            ProjectID: "your-project-id",
        },
        ReportPortal: &gowright.ReportPortalConfig{
            URL:     "https://your-reportportal.com",
            UUID:    "your-uuid",
            Project: "your-project",
            Launch:  "Automated Tests",
        },
    },
}
```

## Advanced Features

### Custom Assertions

```go
// Create custom assertion
assertion := gowright.NewTestAssertion("Custom Check")
assertion.Assert(actualValue == expectedValue, "Values should match")
assertion.AssertNotNil(someObject, "Object should not be nil")
assertion.AssertContains(slice, item, "Slice should contain item")

// Execute with assertions
result := gowright.ExecuteTestWithAssertions("My Test", func(a *gowright.TestAssertion) {
    a.Assert(true, "This should pass")
    a.AssertEqual(1, 1, "Numbers should be equal")
})
```

### Resource Management

```go
// Monitor resource usage
resourceManager := gowright.NewResourceManager(&gowright.ResourceLimits{
    MaxMemoryMB:      512,
    MaxCPUPercent:    70,
    MaxOpenFiles:     50,
    MaxNetworkConns:  25,
})

// Check resource usage
usage := resourceManager.GetCurrentUsage()
fmt.Printf("Memory: %d MB, CPU: %.1f%%\n", usage.MemoryMB, usage.CPUPercent)
```

### Error Recovery

```go
// Configure retry behavior
retryConfig := &gowright.RetryConfig{
    MaxRetries:   3,
    InitialDelay: time.Second,
    MaxDelay:     10 * time.Second,
    Multiplier:   2.0,
}

// Execute with retry
err := gowright.RetryWithBackoff(context.Background(), retryConfig, func() error {
    // Your test operation here
    return someOperation()
})
```

## Best Practices

### 1. Test Organization

```go
// Organize tests by feature
func TestUserManagement(t *testing.T) {
    t.Run("CreateUser", testCreateUser)
    t.Run("UpdateUser", testUpdateUser)
    t.Run("DeleteUser", testDeleteUser)
}
```

### 2. Resource Cleanup

```go
func TestWithCleanup(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close() // Always cleanup
    
    // Your test code here
}
```

### 3. Configuration Management

```go
// Use environment-specific configs
configFile := os.Getenv("GOWRIGHT_CONFIG")
if configFile == "" {
    configFile = "gowright-config.json"
}

config, err := gowright.LoadConfigFromFile(configFile)
```

### 4. Error Handling

```go
// Always check for errors
result := apiTester.ExecuteTest(test)
if result.Status != gowright.TestStatusPassed {
    t.Fatalf("Test failed: %v", result.Error)
}
```

## Performance Considerations

- **Parallel Execution**: Enable parallel testing for faster execution
- **Resource Limits**: Set appropriate resource limits to prevent system overload
- **Connection Pooling**: Reuse database connections and HTTP clients
- **Memory Management**: Use memory-efficient capture for large datasets
- **Cleanup**: Always cleanup resources to prevent leaks

## Troubleshooting

### Common Issues

1. **Browser not found**: Ensure Chrome/Chromium is installed for UI testing
2. **Database connection failed**: Check connection strings and database availability
3. **API timeout**: Increase timeout values for slow endpoints
4. **Memory issues**: Reduce parallel execution or increase resource limits

### Debug Mode

```go
config := &gowright.Config{
    LogLevel: "debug", // Enable debug logging
}
```

### Resource Monitoring

```go
// Monitor resource usage during tests
go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        usage := resourceManager.GetCurrentUsage()
        log.Printf("Resources: Memory=%dMB CPU=%.1f%%", 
            usage.MemoryMB, usage.CPUPercent)
    }
}()
```

## Documentation

📖 **[Complete Documentation](https://gowright.github.io/framework/)** - Visit our Docsify site for comprehensive documentation

### Quick Links
- [Getting Started](docs/getting-started/introduction.md) - Framework overview and setup
- [Installation Guide](docs/getting-started/installation.md) - Detailed installation instructions
- [Quick Start](docs/getting-started/quick-start.md) - Your first Gowright test
- [Configuration](docs/getting-started/configuration.md) - Configuration options and examples

### Testing Modules
- [API Testing](docs/testing-modules/api-testing.md) - REST API testing with validation
- [UI Testing](docs/testing-modules/ui-testing.md) - Browser automation and UI testing
- [Database Testing](docs/testing-modules/database-testing.md) - Database operations and validation
- [Integration Testing](docs/testing-modules/integration-testing.md) - Multi-system workflows

### Examples
- [Basic Usage](docs/examples/basic-usage.md) - Framework initialization examples
- [API Testing Examples](docs/examples/api-testing.md) - Comprehensive API testing scenarios
- [UI Testing Examples](docs/examples/ui-testing.md) - Browser automation examples
- [Database Examples](docs/examples/database-testing.md) - Database testing patterns
- [Integration Examples](docs/examples/integration-testing.md) - End-to-end workflows

### Local Documentation

To run the documentation locally with Docsify:

```bash
# Install docsify-cli globally
npm install -g docsify-cli

# Serve the documentation
cd docs
docsify serve .

# Or use Python's built-in server
python -m http.server 3000
```

Then open [http://localhost:3000](http://localhost:3000) in your browser.

### GitHub Pages Deployment

The documentation is automatically deployed to GitHub Pages when you push changes to the main branch. The workflow:

1. **Validates** the Docsify configuration
2. **Verifies** the documentation structure
3. **Deploys** to GitHub Pages using the official GitHub Actions

The deployment serves the `docs` directory directly as static files, making it perfect for Docsify.

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/contributing.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/your-org/gowright.git
cd gowright

# Install dependencies
go mod download

# Run tests
go test ./...

# Run integration tests
go run integration_test_runner.go

# Run benchmarks
go test -bench=. ./...
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-rod/rod](https://github.com/go-rod/rod) for browser automation
- [go-resty/resty](https://github.com/go-resty/resty) for HTTP client
- [stretchr/testify](https://github.com/stretchr/testify) for testing utilities

## Support

- 📖 [Documentation](https://gowright.github.io/framework/)
- 🐛 [Issue Tracker](https://github.com/gowright/framework/issues)
- 💬 [Discussions](https://github.com/gowright/framework/discussions)
- 📧 [Email Support](mailto:support@gowright.dev)

---

**Gowright** - Making Go testing comprehensive and enjoyable! 🚀