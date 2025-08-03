---
title: Examples
description: Practical examples of using the Gowright testing framework
---

This page provides practical examples of using the Gowright testing framework for various testing scenarios.

## Basic Examples

### Simple Framework Setup

```go
package main

import (
    "fmt"
    "log"
    
    "github/gowright/framework/pkg/gowright"
)

func main() {
    // Create framework with default configuration
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Initialize the framework
    if err := framework.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Gowright framework initialized successfully!")
}
```

### Custom Configuration

```go
func setupCustomFramework() *gowright.Gowright {
    config := &gowright.Config{
        LogLevel: "debug",
        Parallel: true,
        BrowserConfig: &gowright.BrowserConfig{
            Headless: false,
            Timeout:  60 * time.Second,
        },
        APIConfig: &gowright.APIConfig{
            BaseURL: "https://jsonplaceholder.typicode.com",
            Timeout: 15 * time.Second,
        },
    }
    
    framework := gowright.New(config)
    if err := framework.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    return framework
}
```

## Quick Examples by Type

### API Testing

```go
func TestBasicAPIRequest(t *testing.T) {
    config := &gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    }
    
    apiTester := gowright.NewAPITester(config)
    err := apiTester.Initialize(config)
    require.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Test GET request
    response, err := apiTester.Get("/posts/1", nil)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
}
```

### UI Testing

```go
func TestWebPageNavigation(t *testing.T) {
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    require.NoError(t, err)
    defer uiTester.Cleanup()
    
    // Navigate to page
    err = uiTester.Navigate("https://example.com")
    require.NoError(t, err)
    
    // Verify page title
    title, err := uiTester.GetText("title")
    require.NoError(t, err)
    assert.Contains(t, title, "Example Domain")
}
```

### Database Testing

```go
func TestBasicDatabaseOperations(t *testing.T) {
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
    require.NoError(t, err)
    defer dbTester.Cleanup()
    
    // Create table and insert data
    _, err = dbTester.Execute("test", `
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL
        )
    `)
    require.NoError(t, err)
    
    result, err := dbTester.Execute("test", 
        "INSERT INTO users (name, email) VALUES (?, ?)", 
        "John Doe", "john@example.com")
    require.NoError(t, err)
    assert.Equal(t, int64(1), result.RowsAffected)
}
```

## Detailed Examples

For comprehensive examples covering specific testing scenarios, see:

- **[API Examples](/examples/api/)** - HTTP/REST API testing patterns
- **[UI Examples](/examples/ui/)** - Browser automation and interaction
- **[Database Examples](/examples/database/)** - Database testing strategies
- **[Integration Examples](/examples/integration/)** - End-to-end workflows

## Example Projects

### Complete Test Suite

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
func TestParallelExecution(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 4,
            ResourceLimits: gowright.ResourceLimits{
                MaxMemoryMB:   1024,
                MaxCPUPercent: 80,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Tests will run in parallel
}
```

## Best Practices Examples

### Resource Management

```go
func TestWithResourceMonitoring(t *testing.T) {
    resourceManager := gowright.NewResourceManager(&gowright.ResourceLimits{
        MaxMemoryMB:      512,
        MaxCPUPercent:    70,
    })
    
    // Monitor resources during test
    usage := resourceManager.GetCurrentUsage()
    t.Logf("Memory: %d MB, CPU: %.1f%%", usage.MemoryMB, usage.CPUPercent)
}
```

### Error Handling with Retry

```go
func TestWithRetryLogic(t *testing.T) {
    retryConfig := &gowright.RetryConfig{
        MaxRetries:   3,
        InitialDelay: time.Second,
        MaxDelay:     10 * time.Second,
    }
    
    err := gowright.RetryWithBackoff(context.Background(), retryConfig, func() error {
        // Your test operation here
        return someOperation()
    })
    
    require.NoError(t, err)
}
```

## Next Steps

- Explore specific testing modules:
  - [API Testing](/testing/api/)
  - [UI Testing](/testing/ui/)
  - [Database Testing](/testing/database/)
  - [Integration Testing](/testing/integration/)
- Learn about [configuration options](/configuration/)
- Check out [best practices](/guides/best-practices/)