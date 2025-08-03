---
title: Best Practices
description: Best practices for using Gowright effectively
---

This guide covers best practices for using Gowright effectively in your testing workflows.

## Test Organization

### Structure Your Tests

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

### Use Test Helpers

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

## Resource Management

### Always Use Defer for Cleanup

```go
func TestWithResources(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close() // Always cleanup
    
    apiTester := gowright.NewAPITester(config)
    defer apiTester.Cleanup() // Cleanup individual testers too
    
    // Test code here
}
```

### Monitor Resource Usage

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

## Configuration Management

### Environment-Specific Configs

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

### Configuration Validation

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

## Error Handling

### Use Contextual Errors

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

### Implement Retry Logic

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

## Parallel Testing

### Configure Parallel Execution

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

### Use Parallel-Safe Test Data

```go
func TestParallelSafe(t *testing.T) {
    // Use unique identifiers to avoid conflicts
    userEmail := fmt.Sprintf("test-%d-%d@example.com", 
        time.Now().Unix(), rand.Int63())
    
    // Test with unique data
}
```

## Performance Optimization

### Connection Pooling

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "main": {
            Driver:          "postgres",
            DSN:             "postgres://user:pass@localhost/testdb",
            MaxOpenConns:    25,
            MaxIdleConns:    5,
            ConnMaxLifetime: 5 * time.Minute,
        },
    },
}
```

### Efficient Resource Usage

```go
// Reuse testers when possible
var apiTester gowright.APITester

func setupAPITester() gowright.APITester {
    if apiTester == nil {
        config := &gowright.APIConfig{
            BaseURL: "https://api.example.com",
            Timeout: 10 * time.Second,
        }
        apiTester = gowright.NewAPITester(config)
        apiTester.Initialize(config)
    }
    return apiTester
}
```

## Security Considerations

### Secure Credential Management

```go
// Use environment variables for sensitive data
config := &gowright.APIConfig{
    BaseURL: os.Getenv("API_BASE_URL"),
    AuthConfig: &gowright.AuthConfig{
        Type:  "bearer",
        Token: os.Getenv("API_TOKEN"),
    },
}
```

### Input Validation

```go
func validateTestInput(input string) error {
    if len(input) > 1000 {
        return fmt.Errorf("input too long")
    }
    if strings.Contains(input, "<script>") {
        return fmt.Errorf("potentially malicious input")
    }
    return nil
}
```

## Testing Strategies

### Test Pyramid

1. **Unit Tests**: Test individual components
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete user workflows

### Data Management

```go
func setupTestData(dbTester *gowright.DatabaseTester) error {
    queries := []string{
        "DELETE FROM orders",
        "DELETE FROM users",
        "INSERT INTO users (name, email) VALUES ('Test User', 'test@example.com')",
    }
    
    for _, query := range queries {
        if _, err := dbTester.Execute("test", query); err != nil {
            return err
        }
    }
    return nil
}
```

### Test Isolation

```go
func TestWithIsolation(t *testing.T) {
    // Each test should be independent
    tx, err := dbTester.BeginTransaction("main")
    require.NoError(t, err)
    defer tx.Rollback() // Always rollback to maintain isolation
    
    // Test operations within transaction
}
```

These best practices will help you create maintainable, reliable, and efficient tests with Gowright.