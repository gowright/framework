---
title: Migration Guide
description: Migrate to Gowright from other testing frameworks
---

This document provides guidance for migrating to and between versions of the Gowright testing framework.

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
}
```

### Benefits of Migration

- **Simplified API**: Less boilerplate code
- **Built-in Assertions**: Integration with testify
- **Better Error Handling**: Contextual error information
- **Resource Management**: Automatic cleanup
- **Reporting**: Built-in test reporting
- **Parallel Execution**: Easy concurrent testing

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

### Configuration Management

Use environment-specific configs:

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

### Resource Management

Always use defer for cleanup:

```go
func TestWithResources(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close() // Always cleanup
    
    // Test code here
}
```

## Common Issues

### Browser Not Found

**Problem**: UI tests fail with "browser not found" error.

**Solution**: Install Chrome or Chromium:

```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS
brew install --cask google-chrome
```

### Database Connection Issues

**Problem**: Database tests fail to connect.

**Solution**: Check connection string format:

```go
config := &gowright.DatabaseConfig{
    Connections: map[string]*gowright.DBConnection{
        "postgres": {
            Driver: "postgres",
            DSN:    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
        },
    },
}
```

## Getting Help

If you encounter issues during migration:

1. **Check the documentation**: [API Reference](/api/)
2. **Search existing issues**: [GitHub Issues](https://github.com/your-org/gowright/issues)
3. **Ask for help**: [GitHub Discussions](https://github.com/your-org/gowright/discussions)
4. **Contact maintainers**: [Email Support](mailto:support@gowright.dev)