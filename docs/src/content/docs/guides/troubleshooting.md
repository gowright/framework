---
title: Troubleshooting
description: Common issues and solutions when using Gowright
---

This guide covers common issues you might encounter when using Gowright and how to resolve them.

## Installation Issues

### Browser Not Found

**Problem**: UI tests fail with "browser not found" error.

**Solutions**:

1. **Install Chrome or Chromium**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install chromium-browser
   
   # macOS
   brew install --cask google-chrome
   
   # Windows: Download from https://www.google.com/chrome/
   ```

2. **Let Gowright auto-download**:
   ```go
   config.BrowserConfig.AutoDownload = true
   ```

3. **Specify Chrome path**:
   ```go
   config.BrowserConfig.ExecutablePath = "/path/to/chrome"
   ```

### Database Connection Issues

**Problem**: Database tests fail to connect.

**Solutions**:

1. **Check connection string format**:
   ```go
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
   ```

2. **Import required drivers**:
   ```go
   import (
       _ "github.com/lib/pq"           // PostgreSQL
       _ "github.com/go-sql-driver/mysql" // MySQL
       _ "github.com/mattn/go-sqlite3"    // SQLite
   )
   ```

3. **Check database is running**:
   ```bash
   sudo systemctl status postgresql  # PostgreSQL
   sudo systemctl status mysql       # MySQL
   ```

## Runtime Issues

### Memory Issues

**Problem**: Tests consume too much memory.

**Solutions**:

1. **Set resource limits**:
   ```go
   config.ParallelRunnerConfig = &gowright.ParallelRunnerConfig{
       ResourceLimits: gowright.ResourceLimits{
           MaxMemoryMB: 512, // Limit memory usage
       },
   }
   ```

2. **Reduce parallel execution**:
   ```go
   config.ParallelRunnerConfig.MaxConcurrency = 2
   ```

3. **Use memory-efficient operations**:
   ```go
   uiTester.SetMemoryEfficientCapture(true)
   ```

### Timeout Issues

**Problem**: Tests timeout frequently.

**Solutions**:

1. **Increase timeouts**:
   ```go
   config.BrowserConfig.Timeout = 60 * time.Second
   config.APIConfig.Timeout = 30 * time.Second
   ```

2. **Use appropriate wait strategies**:
   ```go
   err := uiTester.WaitForElement(".loading", 10*time.Second)
   ```

3. **Implement retry logic**:
   ```go
   retryConfig := &gowright.RetryConfig{
       MaxRetries:   3,
       InitialDelay: time.Second,
       MaxDelay:     10 * time.Second,
   }
   
   err := gowright.RetryWithBackoff(context.Background(), retryConfig, func() error {
       return someOperation()
   })
   ```

### Port Conflicts

**Problem**: Tests fail due to port conflicts.

**Solutions**:

1. **Use dynamic ports**:
   ```go
   server := httptest.NewServer(handler)
   defer server.Close()
   ```

2. **Use different ports for different environments**:
   ```go
   config.APIConfig.BaseURL = fmt.Sprintf("http://localhost:%d", getAvailablePort())
   ```

## Configuration Issues

### Invalid Configuration

**Problem**: Framework fails to initialize due to invalid configuration.

**Solutions**:

1. **Validate configuration**:
   ```go
   if err := gowright.ValidateConfig(config); err != nil {
       log.Fatalf("Invalid configuration: %v", err)
   }
   ```

2. **Use default configuration as base**:
   ```go
   config := gowright.DefaultConfig()
   config.APIConfig.BaseURL = "https://api.example.com"
   ```

3. **Check required fields**:
   ```go
   if config.APIConfig != nil && config.APIConfig.BaseURL == "" {
       return fmt.Errorf("API base URL is required")
   }
   ```

### Environment Variable Issues

**Problem**: Environment variables not being loaded correctly.

**Solutions**:

1. **Check environment variable names**:
   ```bash
   export GOWRIGHT_LOG_LEVEL=debug
   export GOWRIGHT_API_BASE_URL=https://api.example.com
   ```

2. **Verify environment loading**:
   ```go
   config := gowright.LoadConfigFromEnv()
   fmt.Printf("Loaded config: %+v\n", config)
   ```

## Test Execution Issues

### Flaky Tests

**Problem**: Tests pass sometimes and fail other times.

**Solutions**:

1. **Add proper waits**:
   ```go
   // Instead of sleep
   time.Sleep(2 * time.Second)
   
   // Use explicit waits
   err := uiTester.WaitForElement(".element", 10*time.Second)
   ```

2. **Implement retry logic**:
   ```go
   for i := 0; i < 3; i++ {
       if err := performOperation(); err == nil {
           break
       }
       if i == 2 {
           t.Fatal("Operation failed after 3 attempts")
       }
       time.Sleep(time.Second)
   }
   ```

3. **Use unique test data**:
   ```go
   userEmail := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
   ```

### Resource Leaks

**Problem**: Tests leave resources open, causing subsequent tests to fail.

**Solutions**:

1. **Always use defer for cleanup**:
   ```go
   func TestWithCleanup(t *testing.T) {
       framework := gowright.NewWithDefaults()
       defer framework.Close() // Always cleanup
       
       // Test code here
   }
   ```

2. **Check for resource leaks**:
   ```go
   func TestResourceUsage(t *testing.T) {
       initialUsage := getResourceUsage()
       
       // Run test
       runTest()
       
       finalUsage := getResourceUsage()
       if finalUsage.OpenFiles > initialUsage.OpenFiles+10 {
           t.Error("Potential file descriptor leak detected")
       }
   }
   ```

## Performance Issues

### Slow Test Execution

**Problem**: Tests run slower than expected.

**Solutions**:

1. **Enable parallel execution**:
   ```go
   config := &gowright.Config{
       Parallel: true,
       ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
           MaxConcurrency: 4,
       },
   }
   ```

2. **Use connection pooling**:
   ```go
   config.DatabaseConfig.Connections["main"].MaxOpenConns = 25
   config.DatabaseConfig.Connections["main"].MaxIdleConns = 5
   ```

3. **Optimize wait times**:
   ```go
   // Use shorter, more specific waits
   err := uiTester.WaitForElement(".specific-element", 5*time.Second)
   ```

### High Resource Usage

**Problem**: Tests consume too many system resources.

**Solutions**:

1. **Set resource limits**:
   ```go
   config.ParallelRunnerConfig.ResourceLimits = gowright.ResourceLimits{
       MaxMemoryMB:     512,
       MaxCPUPercent:   70,
       MaxOpenFiles:    50,
       MaxNetworkConns: 25,
   }
   ```

2. **Monitor resource usage**:
   ```go
   resourceManager := gowright.NewResourceManager(&limits)
   usage := resourceManager.GetCurrentUsage()
   t.Logf("Memory: %d MB, CPU: %.1f%%", usage.MemoryMB, usage.CPUPercent)
   ```

## Platform-Specific Issues

### Linux Issues

**Problem**: Chrome sandbox issues on Linux.

**Solutions**:

1. **Disable sandbox**:
   ```go
   config.BrowserConfig.Arguments = []string{"--no-sandbox", "--disable-dev-shm-usage"}
   ```

2. **Fix kernel settings**:
   ```bash
   sudo sysctl kernel.unprivileged_userns_clone=1
   ```

### Windows Issues

**Problem**: Path separator issues on Windows.

**Solutions**:

1. **Use filepath.Join**:
   ```go
   configPath := filepath.Join("config", "test.json")
   ```

2. **Handle Windows-specific paths**:
   ```go
   if runtime.GOOS == "windows" {
       config.BrowserConfig.ExecutablePath = "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe"
   }
   ```

## Getting Help

If you're still experiencing issues:

1. **Enable debug logging**:
   ```go
   config.LogLevel = "debug"
   ```

2. **Check existing issues**: [GitHub Issues](https://github.com/your-org/gowright/issues)

3. **Ask for help**: [GitHub Discussions](https://github.com/your-org/gowright/discussions)

4. **Contact support**: [support@gowright.dev](mailto:support@gowright.dev)

When reporting issues, please include:
- Go version (`go version`)
- Gowright version
- Operating system
- Complete error messages
- Minimal code example that reproduces the issue