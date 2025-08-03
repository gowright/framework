# Migration Guide

This document provides guidance for migrating to and between versions of the Gowright testing framework.

## Table of Contents

- [Migrating to Gowright](#migrating-to-gowright)
- [Version Migration](#version-migration)
- [Best Practices](#best-practices)
- [Common Issues](#common-issues)

## Migrating to Gowright

### From Standard Go Testing

If you're currently using Go's standard testing package, here's how to migrate:

#### Before (Standard Go Testing)

```go
func TestAPIEndpoint(t *testing.T) {
    client := &http.Client{Timeout: 10 * time.Second}
    
    req, err := http.NewRequest("GET", "https://api.example.com/users", nil)
    if err != nil {
        t.Fatal(err)
    }
    
    resp, err := client.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatal(err)
    }
    
    var users []map[string]interface{}
    if err := json.Unmarshal(body, &users); err != nil {
        t.Fatal(err)
    }
    
    if len(users) == 0 {
        t.Error("Expected users, got empty array")
    }
}
```

#### After (Gowright)

```go
func TestAPIEndpoint(t *testing.T) {
    config := &gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
    }
    
    apiTester := gowright.NewAPITester(config)
    err := apiTester.Initialize(config)
    require.NoError(t, err)
    defer apiTester.Cleanup()
    
    response, err := apiTester.Get("/users", nil)
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    var users []map[string]interface{}
    err = json.Unmarshal(response.Body, &users)
    require.NoError(t, err)
    assert.Greater(t, len(users), 0)
}
```

#### Benefits of Migration

- **Simplified API**: Less boilerplate code
- **Built-in Assertions**: Integration with testify
- **Better Error Handling**: Contextual error information
- **Resource Management**: Automatic cleanup
- **Reporting**: Built-in test reporting
- **Parallel Execution**: Easy concurrent testing

### From Selenium/WebDriver

If you're using Selenium for UI testing:

#### Before (Selenium)

```go
func TestWebPage(t *testing.T) {
    caps := selenium.Capabilities{"browserName": "chrome"}
    wd, err := selenium.NewRemote(caps, "http://localhost:4444/wd/hub")
    if err != nil {
        t.Fatal(err)
    }
    defer wd.Quit()
    
    err = wd.Get("https://example.com")
    if err != nil {
        t.Fatal(err)
    }
    
    element, err := wd.FindElement(selenium.ByID, "submit-button")
    if err != nil {
        t.Fatal(err)
    }
    
    err = element.Click()
    if err != nil {
        t.Fatal(err)
    }
    
    // Wait and verify
    time.Sleep(2 * time.Second)
    title, err := wd.Title()
    if err != nil {
        t.Fatal(err)
    }
    
    if !strings.Contains(title, "Success") {
        t.Errorf("Expected success page, got title: %s", title)
    }
}
```

#### After (Gowright)

```go
func TestWebPage(t *testing.T) {
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    require.NoError(t, err)
    defer uiTester.Cleanup()
    
    err = uiTester.Navigate("https://example.com")
    require.NoError(t, err)
    
    err = uiTester.Click("#submit-button")
    require.NoError(t, err)
    
    err = uiTester.WaitForElement(".success-message", 5*time.Second)
    require.NoError(t, err)
    
    text, err := uiTester.GetText("title")
    require.NoError(t, err)
    assert.Contains(t, text, "Success")
}
```

### From Database/SQL Testing

#### Before (Direct SQL)

```go
func TestUserCreation(t *testing.T) {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/testdb")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    
    // Setup
    _, err = db.Exec("DELETE FROM users WHERE email = 'test@example.com'")
    if err != nil {
        t.Fatal(err)
    }
    
    // Test
    _, err = db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Test User", "test@example.com")
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "test@example.com").Scan(&count)
    if err != nil {
        t.Fatal(err)
    }
    
    if count != 1 {
        t.Errorf("Expected 1 user, got %d", count)
    }
    
    // Cleanup
    _, err = db.Exec("DELETE FROM users WHERE email = 'test@example.com'")
    if err != nil {
        t.Fatal(err)
    }
}
```

#### After (Gowright)

```go
func TestUserCreation(t *testing.T) {
    dbTester := gowright.NewDatabaseTester()
    config := &gowright.DatabaseConfig{
        Connections: map[string]*gowright.DBConnection{
            "test": {
                Driver: "postgres",
                DSN:    "postgres://user:pass@localhost/testdb",
            },
        },
    }
    
    err := dbTester.Initialize(config)
    require.NoError(t, err)
    defer dbTester.Cleanup()
    
    dbTest := &gowright.DatabaseTest{
        Name:       "User Creation Test",
        Connection: "test",
        Setup: []string{
            "DELETE FROM users WHERE email = 'test@example.com'",
        },
        Query: "INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com')",
        Expected: &gowright.DatabaseExpectation{
            RowsAffected: 1,
        },
        Teardown: []string{
            "DELETE FROM users WHERE email = 'test@example.com'",
        },
    }
    
    result := dbTester.ExecuteTest(dbTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Version Migration

### From 0.x to 1.0

This is the first major release, so there are no breaking changes to migrate from.

### Future Version Migrations

When new versions are released, this section will contain:

- **Breaking Changes**: API changes that require code updates
- **Deprecated Features**: Features that will be removed in future versions
- **New Features**: New capabilities and how to use them
- **Configuration Changes**: Updates to configuration structure

## Best Practices

### 1. Configuration Management

#### Use Environment-Specific Configs

```go
// config/test.json
{
  "api_config": {
    "base_url": "https://test-api.example.com",
    "timeout": "10s"
  }
}

// config/prod.json
{
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "30s"
  }
}
```

```go
func loadConfig() *gowright.Config {
    env := os.Getenv("TEST_ENV")
    if env == "" {
        env = "test"
    }
    
    configFile := fmt.Sprintf("config/%s.json", env)
    config, err := gowright.LoadConfigFromFile(configFile)
    if err != nil {
        return gowright.DefaultConfig()
    }
    return config
}
```

#### Use Configuration Validation

```go
func validateConfig(config *gowright.Config) error {
    if config.APIConfig != nil && config.APIConfig.BaseURL == "" {
        return fmt.Errorf("API base URL is required")
    }
    
    if config.DatabaseConfig != nil && len(config.DatabaseConfig.Connections) == 0 {
        return fmt.Errorf("at least one database connection is required")
    }
    
    return nil
}
```

### 2. Test Organization

#### Group Related Tests

```go
func TestUserManagement(t *testing.T) {
    framework := setupFramework(t)
    defer framework.Close()
    
    t.Run("CreateUser", func(t *testing.T) {
        testCreateUser(t, framework)
    })
    
    t.Run("UpdateUser", func(t *testing.T) {
        testUpdateUser(t, framework)
    })
    
    t.Run("DeleteUser", func(t *testing.T) {
        testDeleteUser(t, framework)
    })
}
```

#### Use Test Helpers

```go
func setupFramework(t *testing.T) *gowright.Gowright {
    config := loadConfig()
    framework := gowright.New(config)
    
    err := framework.Initialize()
    require.NoError(t, err)
    
    return framework
}

func createTestUser(t *testing.T, apiTester gowright.APITester) map[string]interface{} {
    user := map[string]interface{}{
        "name":  "Test User",
        "email": fmt.Sprintf("test-%d@example.com", time.Now().Unix()),
    }
    
    response, err := apiTester.Post("/users", user, nil)
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, response.StatusCode)
    
    var createdUser map[string]interface{}
    err = json.Unmarshal(response.Body, &createdUser)
    require.NoError(t, err)
    
    return createdUser
}
```

### 3. Resource Management

#### Always Use Defer for Cleanup

```go
func TestWithResources(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close() // Always cleanup
    
    apiTester := gowright.NewAPITester(config)
    defer apiTester.Cleanup() // Cleanup individual testers too
    
    // Test code here
}
```

#### Monitor Resource Usage

```go
func TestWithResourceMonitoring(t *testing.T) {
    resourceManager := gowright.NewResourceManager(&gowright.ResourceLimits{
        MaxMemoryMB:     512,
        MaxCPUPercent:   70,
    })
    
    // Monitor resources during test
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            usage := resourceManager.GetCurrentUsage()
            if usage.MemoryMB > 400 {
                t.Logf("High memory usage: %d MB", usage.MemoryMB)
            }
        }
    }()
    
    // Your test code here
}
```

### 4. Error Handling

#### Use Contextual Errors

```go
func TestWithContextualErrors(t *testing.T) {
    apiTester := gowright.NewAPITester(config)
    
    response, err := apiTester.Get("/users", nil)
    if err != nil {
        if gowrightErr, ok := err.(*gowright.GowrightError); ok {
            t.Logf("Error type: %s", gowrightErr.Type)
            t.Logf("Context: %v", gowrightErr.Context)
        }
        t.Fatal(err)
    }
}
```

#### Implement Retry Logic

```go
func TestWithRetry(t *testing.T) {
    retryConfig := &gowright.RetryConfig{
        MaxRetries:   3,
        InitialDelay: time.Second,
        MaxDelay:     10 * time.Second,
    }
    
    err := gowright.RetryWithBackoff(context.Background(), retryConfig, func() error {
        response, err := apiTester.Get("/flaky-endpoint", nil)
        if err != nil {
            return err
        }
        if response.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", response.StatusCode)
        }
        return nil
    })
    
    require.NoError(t, err)
}
```

### 5. Parallel Testing

#### Configure Parallel Execution

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

#### Use Parallel-Safe Test Data

```go
func TestParallelSafe(t *testing.T) {
    // Use unique identifiers to avoid conflicts
    userEmail := fmt.Sprintf("test-%d-%d@example.com", 
        time.Now().Unix(), rand.Int63())
    
    // Test with unique data
}
```

## Common Issues

### 1. Browser Not Found

**Problem**: UI tests fail with "browser not found" error.

**Solution**: Install Chrome or Chromium:

```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS
brew install --cask google-chrome

# Windows
# Download and install Chrome from https://www.google.com/chrome/
```

### 2. Database Connection Issues

**Problem**: Database tests fail to connect.

**Solutions**:

```go
// Check connection string format
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "postgres": {
            Driver: "postgres",
            DSN:    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
        },
        "mysql": {
            Driver: "mysql",
            DSN:    "user:password@tcp(localhost:3306)/dbname",
        },
        "sqlite": {
            Driver: "sqlite3",
            DSN:    "./test.db", // or ":memory:" for in-memory
        },
    },
}

// Import required drivers
import (
    _ "github.com/lib/pq"           // PostgreSQL
    _ "github.com/go-sql-driver/mysql" // MySQL
    _ "github.com/mattn/go-sqlite3"    // SQLite
)
```

### 3. Memory Issues

**Problem**: Tests consume too much memory.

**Solutions**:

```go
// Set resource limits
config.ParallelRunnerConfig = &gowright.ParallelRunnerConfig{
    ResourceLimits: gowright.ResourceLimits{
        MaxMemoryMB: 512, // Limit memory usage
    },
}

// Reduce parallel execution
config.ParallelRunnerConfig.MaxConcurrency = 2

// Use memory-efficient operations
uiTester.SetMemoryEfficientCapture(true)
```

### 4. Timeout Issues

**Problem**: Tests timeout frequently.

**Solutions**:

```go
// Increase timeouts
config.BrowserConfig.Timeout = 60 * time.Second
config.APIConfig.Timeout = 30 * time.Second

// Use appropriate wait strategies
err := uiTester.WaitForElement(".loading", 10*time.Second)

// Implement retry logic for flaky operations
```

### 5. Port Conflicts

**Problem**: Tests fail due to port conflicts.

**Solutions**:

```go
// Use dynamic ports for test servers
server := httptest.NewServer(handler)
defer server.Close()

// Use different ports for different test environments
config.APIConfig.BaseURL = fmt.Sprintf("http://localhost:%d", getAvailablePort())
```

## Getting Help

If you encounter issues during migration:

1. **Check the documentation**: [API Documentation](API.md)
2. **Search existing issues**: [GitHub Issues](https://github.com/your-org/gowright/issues)
3. **Ask for help**: [GitHub Discussions](https://github.com/your-org/gowright/discussions)
4. **Contact maintainers**: [Email Support](mailto:support@gowright.dev)

## Contributing to Migration Guide

If you find issues or have suggestions for improving this migration guide, please:

1. Open an issue describing the problem
2. Submit a pull request with improvements
3. Share your migration experience in discussions

Your feedback helps make migrations smoother for everyone!