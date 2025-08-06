# Resource Management

Gowright provides comprehensive resource management capabilities to monitor, control, and optimize the usage of system resources during test execution. This ensures tests run efficiently without overwhelming the system or causing resource conflicts.

## Overview

Resource management in Gowright provides:

- **Memory Management**: Monitor and control memory usage
- **CPU Management**: Track and limit CPU consumption
- **Network Management**: Manage network connections and bandwidth
- **File System Management**: Control file handles and disk usage
- **Database Connection Pooling**: Optimize database resource usage
- **Browser Resource Management**: Control browser instances and memory
- **Automatic Cleanup**: Prevent resource leaks and ensure cleanup

## Memory Management

### Basic Memory Monitoring

```go
package main

import (
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestMemoryManagement(t *testing.T) {
    // Configure memory management
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            MemoryManagement: &gowright.MemoryManagementConfig{
                MaxMemoryMB:        512,  // 512MB limit
                MemoryCheckInterval: 1 * time.Second,
                EnableGCOptimization: true,
                GCTargetPercent:     50,   // Trigger GC at 50% of limit
                MemoryAlerts: &gowright.MemoryAlerts{
                    WarningThreshold: 0.8,  // 80% warning
                    CriticalThreshold: 0.95, // 95% critical
                    EnableAlerts:     true,
                },
            },
            EnableMonitoring: true,
            MonitoringInterval: 500 * time.Millisecond,
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    err := framework.Initialize()
    assert.NoError(t, err)
    
    // Create memory-intensive test suite
    testSuite := &gowright.TestSuite{
        Name: "Memory Management Tests",
        Tests: []gowright.Test{
            &gowright.MemoryIntensiveTest{
                Name: "Large Data Processing",
                TestFunc: func() error {
                    // Simulate large data processing
                    data := make([][]byte, 100)
                    for i := range data {
                        data[i] = make([]byte, 1024*1024) // 1MB each
                    }
                    
                    // Process data
                    for _, chunk := range data {
                        processData(chunk)
                    }
                    
                    // Explicit cleanup
                    data = nil
                    return nil
                },
                MemoryLimit: 200 * 1024 * 1024, // 200MB limit for this test
            },
            &gowright.MemoryIntensiveTest{
                Name: "Streaming Data Test",
                TestFunc: func() error {
                    return processStreamingData(50 * 1024 * 1024) // 50MB stream
                },
                MemoryLimit: 100 * 1024 * 1024, // 100MB limit
            },
        },
        Config: &gowright.TestSuiteConfig{
            EnableResourceManagement: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify memory management worked
    assert.LessOrEqual(t, results.PeakMemoryUsage, 512*1024*1024) // Under 512MB
    assert.Equal(t, 0, results.MemoryLimitViolations)
    
    // Check memory alerts
    if len(results.MemoryAlerts) > 0 {
        t.Logf("Memory alerts generated: %v", results.MemoryAlerts)
    }
    
    // Verify garbage collection was effective
    assert.Greater(t, results.GCStats.NumGC, uint32(0))
    assert.Less(t, results.GCStats.PauseTotal, 100*time.Millisecond)
}

func processData(data []byte) {
    // Simulate data processing
    for i := range data {
        data[i] = byte(i % 256)
    }
}

func processStreamingData(size int) error {
    // Simulate streaming data processing with controlled memory usage
    chunkSize := 1024 * 1024 // 1MB chunks
    buffer := make([]byte, chunkSize)
    
    for processed := 0; processed < size; processed += chunkSize {
        // Simulate reading chunk
        copy(buffer, generateData(chunkSize))
        
        // Process chunk
        processData(buffer)
        
        // Simulate some processing time
        time.Sleep(10 * time.Millisecond)
    }
    
    return nil
}

func generateData(size int) []byte {
    data := make([]byte, size)
    for i := range data {
        data[i] = byte(i % 256)
    }
    return data
}
```

### Advanced Memory Management

```go
func TestAdvancedMemoryManagement(t *testing.T) {
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            MemoryManagement: &gowright.MemoryManagementConfig{
                MaxMemoryMB: 1024,
                MemoryPools: &gowright.MemoryPoolConfig{
                    EnablePooling:    true,
                    SmallPoolSize:    1024,      // 1KB pool
                    MediumPoolSize:   1024 * 64, // 64KB pool
                    LargePoolSize:    1024 * 1024, // 1MB pool
                    MaxPoolItems:     100,
                    PoolCleanupInterval: 30 * time.Second,
                },
                MemoryProfiling: &gowright.MemoryProfilingConfig{
                    EnableProfiling:   true,
                    ProfilingInterval: 5 * time.Second,
                    HeapDumpOnLimit:   true,
                    HeapDumpDir:       "./memory-dumps",
                },
                LeakDetection: &gowright.LeakDetectionConfig{
                    EnableDetection:    true,
                    DetectionInterval:  10 * time.Second,
                    LeakThreshold:      50 * 1024 * 1024, // 50MB
                    AutoCleanup:        true,
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Create tests that might have memory leaks
    testSuite := &gowright.TestSuite{
        Name: "Advanced Memory Management",
        Tests: []gowright.Test{
            &gowright.MemoryLeakTest{
                Name: "Potential Memory Leak Test",
                TestFunc: func() error {
                    // Simulate potential memory leak
                    var leakyData [][]byte
                    for i := 0; i < 50; i++ {
                        data := make([]byte, 1024*1024) // 1MB
                        leakyData = append(leakyData, data)
                        
                        // Simulate some work
                        time.Sleep(100 * time.Millisecond)
                    }
                    
                    // Intentionally not cleaning up to test leak detection
                    return nil
                },
                ExpectedMemoryIncrease: 50 * 1024 * 1024, // 50MB expected
                LeakDetectionEnabled:   true,
            },
            &gowright.MemoryPoolTest{
                Name: "Memory Pool Usage Test",
                TestFunc: func() error {
                    // Use memory pools for efficient allocation
                    pool := gowright.GetMemoryPool()
                    
                    for i := 0; i < 1000; i++ {
                        // Get buffer from pool
                        buffer := pool.GetBuffer(1024) // 1KB buffer
                        
                        // Use buffer
                        processData(buffer)
                        
                        // Return to pool
                        pool.ReturnBuffer(buffer)
                    }
                    
                    return nil
                },
                MemoryPoolEnabled: true,
            },
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Check memory profiling results
    if results.MemoryProfile != nil {
        assert.Greater(t, len(results.MemoryProfile.Snapshots), 0)
        
        // Check for memory leaks
        if len(results.MemoryProfile.DetectedLeaks) > 0 {
            t.Logf("Memory leaks detected: %v", results.MemoryProfile.DetectedLeaks)
        }
    }
    
    // Verify memory pool efficiency
    if results.MemoryPoolStats != nil {
        assert.Greater(t, results.MemoryPoolStats.PoolHitRate, 0.8) // 80% hit rate
        assert.Less(t, results.MemoryPoolStats.AllocationOverhead, 0.1) // 10% overhead
    }
}
```

## CPU Management

### CPU Monitoring and Limiting

```go
func TestCPUManagement(t *testing.T) {
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            CPUManagement: &gowright.CPUManagementConfig{
                MaxCPUPercent:      70,   // 70% CPU limit
                CPUCheckInterval:   1 * time.Second,
                EnableThrottling:   true,
                ThrottleThreshold:  80,   // Throttle at 80%
                CPUProfiling: &gowright.CPUProfilingConfig{
                    EnableProfiling:   true,
                    ProfilingDuration: 30 * time.Second,
                    ProfileDir:        "./cpu-profiles",
                },
                ProcessPriority: &gowright.ProcessPriorityConfig{
                    EnablePriorityControl: true,
                    TestProcessPriority:   "normal",
                    BackgroundPriority:    "low",
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "CPU Management Tests",
        Tests: []gowright.Test{
            &gowright.CPUIntensiveTest{
                Name: "CPU Intensive Calculation",
                TestFunc: func() error {
                    // CPU-intensive calculation
                    result := 0
                    for i := 0; i < 10000000; i++ {
                        result += i * i
                    }
                    _ = result
                    return nil
                },
                CPULimit: 50, // 50% CPU limit for this test
            },
            &gowright.CPUIntensiveTest{
                Name: "Parallel CPU Work",
                TestFunc: func() error {
                    // Parallel CPU work
                    numWorkers := 4
                    workChan := make(chan int, numWorkers)
                    doneChan := make(chan bool, numWorkers)
                    
                    // Start workers
                    for i := 0; i < numWorkers; i++ {
                        go func() {
                            for work := range workChan {
                                // CPU-intensive work
                                result := 0
                                for j := 0; j < work*1000000; j++ {
                                    result += j
                                }
                                _ = result
                            }
                            doneChan <- true
                        }()
                    }
                    
                    // Send work
                    for i := 0; i < 100; i++ {
                        workChan <- i
                    }
                    close(workChan)
                    
                    // Wait for completion
                    for i := 0; i < numWorkers; i++ {
                        <-doneChan
                    }
                    
                    return nil
                },
                CPULimit: 60, // 60% CPU limit
            },
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify CPU limits were respected
    assert.LessOrEqual(t, results.PeakCPUUsage, 70.0) // Under 70%
    assert.Equal(t, 0, results.CPULimitViolations)
    
    // Check CPU profiling results
    if results.CPUProfile != nil {
        assert.Greater(t, len(results.CPUProfile.Samples), 0)
        assert.NotEmpty(t, results.CPUProfile.HotSpots)
    }
    
    // Verify throttling worked if needed
    if results.CPUThrottlingEvents > 0 {
        t.Logf("CPU throttling events: %d", results.CPUThrottlingEvents)
    }
}
```

## Network Resource Management

### Connection Pool Management

```go
func TestNetworkResourceManagement(t *testing.T) {
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            NetworkManagement: &gowright.NetworkManagementConfig{
                MaxConnections:     100,
                ConnectionTimeout:  10 * time.Second,
                IdleTimeout:        30 * time.Second,
                KeepAliveTimeout:   60 * time.Second,
                ConnectionPooling: &gowright.ConnectionPoolingConfig{
                    EnablePooling:      true,
                    PoolSize:           50,
                    MaxIdleConnections: 25,
                    PoolCleanupInterval: 60 * time.Second,
                },
                BandwidthManagement: &gowright.BandwidthManagementConfig{
                    MaxBandwidthMBps:   10,   // 10 MB/s limit
                    EnableThrottling:   true,
                    ThrottleThreshold:  8,    // Throttle at 8 MB/s
                },
                DNSManagement: &gowright.DNSManagementConfig{
                    EnableCaching:     true,
                    CacheTTL:          300 * time.Second,
                    MaxCacheEntries:   1000,
                },
            },
        },
        APIConfig: &gowright.APIConfig{
            BaseURL: "https://httpbin.org",
            Timeout: 30 * time.Second,
            ConnectionPooling: true,
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Network Resource Management Tests",
        Tests: createNetworkIntensiveTests(framework, 50), // 50 concurrent network tests
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            MaxConcurrency: 20,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify network resource limits
    assert.LessOrEqual(t, results.PeakNetworkConnections, 100)
    assert.Equal(t, 0, results.NetworkLimitViolations)
    
    // Check connection pool efficiency
    if results.ConnectionPoolStats != nil {
        assert.Greater(t, results.ConnectionPoolStats.PoolHitRate, 0.7) // 70% hit rate
        assert.Less(t, results.ConnectionPoolStats.AverageWaitTime, 1*time.Second)
    }
    
    // Verify bandwidth management
    if results.BandwidthStats != nil {
        assert.LessOrEqual(t, results.BandwidthStats.PeakBandwidthMBps, 10.0)
        assert.Greater(t, results.BandwidthStats.AverageBandwidthMBps, 0.0)
    }
}

func createNetworkIntensiveTests(framework *gowright.Framework, count int) []gowright.Test {
    tests := make([]gowright.Test, count)
    
    for i := 0; i < count; i++ {
        tests[i] = gowright.NewAPITestBuilder(
            fmt.Sprintf("Network Test %d", i+1),
            "GET",
            "/bytes/1024", // Download 1KB
        ).
            WithTester(framework.GetAPITester()).
            ExpectStatus(200).
            WithTimeout(10 * time.Second).
            Build()
    }
    
    return tests
}
```

### Network Monitoring

```go
func TestNetworkMonitoring(t *testing.T) {
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            NetworkManagement: &gowright.NetworkManagementConfig{
                EnableMonitoring: true,
                MonitoringInterval: 1 * time.Second,
                NetworkMetrics: &gowright.NetworkMetricsConfig{
                    TrackBandwidth:    true,
                    TrackConnections:  true,
                    TrackLatency:      true,
                    TrackErrors:       true,
                    TrackDNSLookups:   true,
                },
                NetworkAlerts: &gowright.NetworkAlerts{
                    HighLatencyThreshold:    1 * time.Second,
                    HighErrorRateThreshold:  0.05, // 5% error rate
                    ConnectionLeakThreshold: 10,
                    EnableAlerts:            true,
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Create tests with various network patterns
    testSuite := &gowright.TestSuite{
        Name: "Network Monitoring Tests",
        Tests: []gowright.Test{
            // Fast requests
            gowright.NewAPITestBuilder("Fast Request", "GET", "/get").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                Build(),
            
            // Slow requests
            gowright.NewAPITestBuilder("Slow Request", "GET", "/delay/2").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                WithTimeout(5 * time.Second).
                Build(),
            
            // Large download
            gowright.NewAPITestBuilder("Large Download", "GET", "/bytes/1048576"). // 1MB
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                Build(),
            
            // Error request
            gowright.NewAPITestBuilder("Error Request", "GET", "/status/500").
                WithTester(framework.GetAPITester()).
                ExpectStatus(500).
                Build(),
        },
        Config: &gowright.TestSuiteConfig{
            EnableNetworkMonitoring: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify network monitoring data
    assert.NotNil(t, results.NetworkMetrics)
    assert.Greater(t, len(results.NetworkMetrics.BandwidthSamples), 0)
    assert.Greater(t, len(results.NetworkMetrics.LatencySamples), 0)
    assert.Greater(t, len(results.NetworkMetrics.ConnectionSamples), 0)
    
    // Check for network alerts
    if len(results.NetworkAlerts) > 0 {
        t.Logf("Network alerts: %v", results.NetworkAlerts)
    }
    
    // Verify DNS caching effectiveness
    if results.DNSStats != nil {
        assert.Greater(t, results.DNSStats.CacheHitRate, 0.0)
        assert.Less(t, results.DNSStats.AverageLookupTime, 100*time.Millisecond)
    }
}
```

## Database Resource Management

### Connection Pool Optimization

```go
func TestDatabaseResourceManagement(t *testing.T) {
    config := &gowright.Config{
        DatabaseConfig: &gowright.DatabaseConfig{
            Connections: map[string]*gowright.DBConnection{
                "test": {
                    Driver:       "sqlite3",
                    DSN:          ":memory:",
                    MaxOpenConns: 20,
                    MaxIdleConns: 10,
                    MaxLifetime:  "1h",
                },
            },
            ResourceManagement: &gowright.DBResourceManagementConfig{
                ConnectionPooling: &gowright.DBConnectionPoolingConfig{
                    EnablePooling:       true,
                    PoolSize:            20,
                    MaxWaitTime:         5 * time.Second,
                    IdleTimeout:         10 * time.Minute,
                    ConnectionReuse:     true,
                    HealthCheckInterval: 30 * time.Second,
                },
                QueryOptimization: &gowright.QueryOptimizationConfig{
                    EnableOptimization:  true,
                    QueryTimeout:        30 * time.Second,
                    SlowQueryThreshold:  1 * time.Second,
                    EnableQueryCaching:  true,
                    CacheTTL:            5 * time.Minute,
                },
                TransactionManagement: &gowright.TransactionManagementConfig{
                    MaxTransactionTime:  60 * time.Second,
                    DeadlockRetries:     3,
                    IsolationLevel:      "READ_COMMITTED",
                },
            },
        },
        ResourceManagement: &gowright.ResourceManagementConfig{
            DatabaseManagement: &gowright.DatabaseManagementConfig{
                MaxConnections:      50,
                ConnectionTimeout:   10 * time.Second,
                QueryTimeout:        30 * time.Second,
                EnableMonitoring:    true,
                MonitoringInterval:  2 * time.Second,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Setup test database
    dbTester := framework.GetDatabaseTester()
    _, err := dbTester.Execute("test", `
        CREATE TABLE test_performance (
            id INTEGER PRIMARY KEY,
            data TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    assert.NoError(t, err)
    
    testSuite := &gowright.TestSuite{
        Name: "Database Resource Management",
        Tests: createDatabaseResourceTests(framework, 30), // 30 concurrent DB tests
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            MaxConcurrency: 15,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify database resource management
    assert.LessOrEqual(t, results.PeakDatabaseConnections, 50)
    assert.Equal(t, 0, results.DatabaseTimeouts)
    assert.Equal(t, 0, results.ConnectionLeaks)
    
    // Check connection pool performance
    if results.DBConnectionPoolStats != nil {
        assert.Greater(t, results.DBConnectionPoolStats.PoolUtilization, 0.5)
        assert.Less(t, results.DBConnectionPoolStats.AverageWaitTime, 1*time.Second)
    }
    
    // Verify query optimization
    if results.QueryOptimizationStats != nil {
        assert.Greater(t, results.QueryOptimizationStats.CacheHitRate, 0.3)
        assert.Less(t, results.QueryOptimizationStats.AverageQueryTime, 100*time.Millisecond)
    }
}

func createDatabaseResourceTests(framework *gowright.Framework, count int) []gowright.Test {
    tests := make([]gowright.Test, count)
    
    for i := 0; i < count; i++ {
        testID := i + 1
        tests[i] = &gowright.DatabaseTest{
            Name:       fmt.Sprintf("DB Resource Test %d", testID),
            Connection: "test",
            Setup: []string{
                fmt.Sprintf("INSERT INTO test_performance (id, data) VALUES (%d, 'test data %d')", testID, testID),
            },
            Query: "SELECT data FROM test_performance WHERE id = ?",
            Args:  []interface{}{testID},
            Expected: &gowright.DatabaseExpectation{
                RowCount: 1,
                Columns: map[string]interface{}{
                    "data": fmt.Sprintf("test data %d", testID),
                },
            },
            Teardown: []string{
                fmt.Sprintf("DELETE FROM test_performance WHERE id = %d", testID),
            },
            ResourceLimits: &gowright.DBResourceLimits{
                MaxQueryTime:    5 * time.Second,
                MaxConnections:  2,
                MaxMemoryMB:     10,
            },
        }
    }
    
    return tests
}
```

## Browser Resource Management

### Browser Instance Management

```go
func TestBrowserResourceManagement(t *testing.T) {
    config := &gowright.Config{
        BrowserConfig: &gowright.BrowserConfig{
            Headless: true,
            PoolSize: 5, // Browser pool size
            ResourceManagement: &gowright.BrowserResourceManagementConfig{
                MaxMemoryMB:         512,  // 512MB per browser
                MaxCPUPercent:       30,   // 30% CPU per browser
                IdleTimeout:         5 * time.Minute,
                HealthCheckInterval: 30 * time.Second,
                AutoRestart:         true,
                RestartThreshold:    100, // Restart after 100 operations
            },
        },
        ResourceManagement: &gowright.ResourceManagementConfig{
            BrowserManagement: &gowright.BrowserManagementConfig{
                MaxBrowsers:        10,
                BrowserTimeout:     60 * time.Second,
                EnableMonitoring:   true,
                MonitoringInterval: 2 * time.Second,
                MemoryCleanup: &gowright.BrowserMemoryCleanupConfig{
                    EnableCleanup:     true,
                    CleanupInterval:   30 * time.Second,
                    MemoryThreshold:   400, // Cleanup at 400MB
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Browser Resource Management",
        Tests: createBrowserResourceTests(20), // 20 UI tests
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            MaxConcurrency: 5, // Limited by browser pool
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify browser resource management
    assert.LessOrEqual(t, results.PeakBrowserInstances, 10)
    assert.Equal(t, 0, results.BrowserCrashes)
    assert.Equal(t, 0, results.BrowserMemoryLeaks)
    
    // Check browser pool efficiency
    if results.BrowserPoolStats != nil {
        assert.Greater(t, results.BrowserPoolStats.PoolUtilization, 0.6)
        assert.Less(t, results.BrowserPoolStats.AverageStartupTime, 5*time.Second)
    }
    
    // Verify memory cleanup effectiveness
    if results.BrowserMemoryStats != nil {
        assert.Less(t, results.BrowserMemoryStats.AverageMemoryUsage, 400*1024*1024)
        assert.Greater(t, results.BrowserMemoryStats.CleanupEvents, 0)
    }
}

func createBrowserResourceTests(count int) []gowright.Test {
    tests := make([]gowright.Test, count)
    
    for i := 0; i < count; i++ {
        tests[i] = &gowright.UITest{
            Name: fmt.Sprintf("Browser Test %d", i+1),
            Steps: []gowright.UIStep{
                {Action: "navigate", Target: "https://httpbin.org"},
                {Action: "waitFor", Target: "body"},
                {Action: "click", Target: "a[href='/get']"},
                {Action: "waitFor", Target: "pre"},
            },
            ResourceLimits: &gowright.UIResourceLimits{
                MaxMemoryMB:   100,
                MaxCPUPercent: 20,
                Timeout:       30 * time.Second,
            },
        }
    }
    
    return tests
}
```

## Resource Cleanup and Leak Detection

### Automatic Resource Cleanup

```go
func TestResourceCleanup(t *testing.T) {
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            AutoCleanup: &gowright.AutoCleanupConfig{
                EnableAutoCleanup:   true,
                CleanupInterval:     30 * time.Second,
                ForceCleanupOnExit:  true,
                CleanupTimeout:      10 * time.Second,
                ResourceTracking: &gowright.ResourceTrackingConfig{
                    TrackMemory:      true,
                    TrackFiles:       true,
                    TrackConnections: true,
                    TrackBrowsers:    true,
                    TrackDatabases:   true,
                },
            },
            LeakDetection: &gowright.LeakDetectionConfig{
                EnableDetection:    true,
                DetectionInterval:  15 * time.Second,
                LeakThresholds: &gowright.LeakThresholds{
                    MemoryLeakMB:      50,
                    FileLeakCount:     20,
                    ConnectionLeakCount: 10,
                },
                AutoRepair:         true,
                AlertOnLeaks:       true,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Resource Cleanup Tests",
        Tests: []gowright.Test{
            &gowright.ResourceLeakTest{
                Name: "Memory Leak Test",
                TestFunc: func() error {
                    // Intentionally create potential memory leak
                    var data [][]byte
                    for i := 0; i < 100; i++ {
                        data = append(data, make([]byte, 1024*1024)) // 1MB each
                    }
                    
                    // Simulate work
                    time.Sleep(2 * time.Second)
                    
                    // Don't clean up to test leak detection
                    return nil
                },
                ExpectedLeaks: []string{"memory"},
            },
            &gowright.ResourceLeakTest{
                Name: "File Handle Leak Test",
                TestFunc: func() error {
                    // Open files without closing
                    var files []*os.File
                    for i := 0; i < 30; i++ {
                        file, err := os.CreateTemp("", "leak-test-")
                        if err != nil {
                            return err
                        }
                        files = append(files, file)
                    }
                    
                    // Don't close files to test leak detection
                    return nil
                },
                ExpectedLeaks: []string{"files"},
            },
            &gowright.ResourceLeakTest{
                Name: "Connection Leak Test",
                TestFunc: func() error {
                    // Create connections without closing
                    var conns []net.Conn
                    for i := 0; i < 15; i++ {
                        conn, err := net.Dial("tcp", "httpbin.org:80")
                        if err != nil {
                            continue // Skip if can't connect
                        }
                        conns = append(conns, conn)
                    }
                    
                    // Don't close connections to test leak detection
                    return nil
                },
                ExpectedLeaks: []string{"connections"},
            },
        },
        Config: &gowright.TestSuiteConfig{
            EnableResourceTracking: true,
            EnableLeakDetection:    true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify leak detection worked
    assert.Greater(t, len(results.DetectedLeaks), 0)
    
    // Check cleanup effectiveness
    if results.CleanupStats != nil {
        assert.Greater(t, results.CleanupStats.CleanupEvents, 0)
        assert.Greater(t, results.CleanupStats.ResourcesReclaimed, 0)
    }
    
    // Verify auto-repair worked
    if results.AutoRepairStats != nil {
        assert.Greater(t, results.AutoRepairStats.RepairAttempts, 0)
        assert.Greater(t, results.AutoRepairStats.SuccessfulRepairs, 0)
    }
}
```

## Resource Monitoring Dashboard

### Real-time Resource Monitoring

```go
func TestResourceMonitoringDashboard(t *testing.T) {
    config := &gowright.Config{
        ResourceManagement: &gowright.ResourceManagementConfig{
            EnableMonitoring: true,
            MonitoringInterval: 500 * time.Millisecond,
            Dashboard: &gowright.ResourceDashboardConfig{
                EnableDashboard:   true,
                DashboardPort:     8081,
                UpdateInterval:    1 * time.Second,
                HistoryRetention:  24 * time.Hour,
                Metrics: &gowright.DashboardMetricsConfig{
                    ShowMemoryMetrics:    true,
                    ShowCPUMetrics:       true,
                    ShowNetworkMetrics:   true,
                    ShowDatabaseMetrics:  true,
                    ShowBrowserMetrics:   true,
                    ShowCustomMetrics:    true,
                },
                Alerts: &gowright.DashboardAlertsConfig{
                    EnableAlerts:         true,
                    AlertWebhookURL:      os.Getenv("ALERT_WEBHOOK_URL"),
                    AlertEmailRecipients: []string{"admin@example.com"},
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Start monitoring dashboard
    dashboard := gowright.NewResourceDashboard(config.ResourceManagement.Dashboard)
    go dashboard.Start()
    defer dashboard.Stop()
    
    // Wait for dashboard to start
    time.Sleep(2 * time.Second)
    
    testSuite := &gowright.TestSuite{
        Name: "Resource Monitoring Dashboard Test",
        Tests: createResourceIntensiveTests(framework, 10),
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            EnableResourceMonitoring: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify dashboard collected metrics
    dashboardData := dashboard.GetMetrics()
    assert.NotNil(t, dashboardData)
    assert.Greater(t, len(dashboardData.MemoryMetrics), 0)
    assert.Greater(t, len(dashboardData.CPUMetrics), 0)
    
    // Check if dashboard is accessible
    resp, err := http.Get("http://localhost:8081/metrics")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    resp.Body.Close()
    
    // Verify alerts were generated if thresholds exceeded
    alerts := dashboard.GetAlerts()
    if len(alerts) > 0 {
        t.Logf("Resource alerts generated: %v", alerts)
    }
}
```

## Configuration Examples

### Complete Resource Management Configuration

```json
{
  "resource_management": {
    "enable_monitoring": true,
    "monitoring_interval": "1s",
    "memory_management": {
      "max_memory_mb": 2048,
      "memory_check_interval": "1s",
      "enable_gc_optimization": true,
      "gc_target_percent": 50,
      "memory_alerts": {
        "warning_threshold": 0.8,
        "critical_threshold": 0.95,
        "enable_alerts": true
      },
      "memory_pools": {
        "enable_pooling": true,
        "small_pool_size": 1024,
        "medium_pool_size": 65536,
        "large_pool_size": 1048576,
        "max_pool_items": 100
      },
      "leak_detection": {
        "enable_detection": true,
        "detection_interval": "10s",
        "leak_threshold": 52428800,
        "auto_cleanup": true
      }
    },
    "cpu_management": {
      "max_cpu_percent": 80,
      "cpu_check_interval": "1s",
      "enable_throttling": true,
      "throttle_threshold": 90,
      "cpu_profiling": {
        "enable_profiling": true,
        "profiling_duration": "30s",
        "profile_dir": "./cpu-profiles"
      }
    },
    "network_management": {
      "max_connections": 200,
      "connection_timeout": "10s",
      "idle_timeout": "30s",
      "connection_pooling": {
        "enable_pooling": true,
        "pool_size": 100,
        "max_idle_connections": 50
      },
      "bandwidth_management": {
        "max_bandwidth_mbps": 50,
        "enable_throttling": true,
        "throttle_threshold": 40
      }
    },
    "database_management": {
      "max_connections": 100,
      "connection_timeout": "10s",
      "query_timeout": "30s",
      "enable_monitoring": true
    },
    "browser_management": {
      "max_browsers": 20,
      "browser_timeout": "60s",
      "enable_monitoring": true,
      "memory_cleanup": {
        "enable_cleanup": true,
        "cleanup_interval": "30s",
        "memory_threshold": 419430400
      }
    },
    "auto_cleanup": {
      "enable_auto_cleanup": true,
      "cleanup_interval": "30s",
      "force_cleanup_on_exit": true,
      "cleanup_timeout": "10s"
    },
    "dashboard": {
      "enable_dashboard": true,
      "dashboard_port": 8081,
      "update_interval": "1s",
      "history_retention": "24h"
    }
  }
}
```

## Best Practices

### 1. Set Appropriate Resource Limits

```go
// Good - Reasonable limits based on system capacity
config := &gowright.ResourceManagementConfig{
    MemoryManagement: &gowright.MemoryManagementConfig{
        MaxMemoryMB: getAvailableMemory() / 2, // Use half available memory
    },
    CPUManagement: &gowright.CPUManagementConfig{
        MaxCPUPercent: 70, // Leave 30% for system
    },
}
```

### 2. Enable Monitoring for Production

```go
// Good - Comprehensive monitoring
config := &gowright.ResourceManagementConfig{
    EnableMonitoring: true,
    MonitoringInterval: 1 * time.Second,
    Dashboard: &gowright.ResourceDashboardConfig{
        EnableDashboard: true,
        EnableAlerts:    true,
    },
}
```

### 3. Implement Proper Cleanup

```go
// Good - Always cleanup resources
defer func() {
    if framework != nil {
        framework.Close() // This triggers resource cleanup
    }
}()
```

### 4. Use Resource Pools

```go
// Good - Use pools for frequently allocated resources
config := &gowright.MemoryManagementConfig{
    MemoryPools: &gowright.MemoryPoolConfig{
        EnablePooling: true,
        SmallPoolSize: 1024,
        MediumPoolSize: 64 * 1024,
        LargePoolSize: 1024 * 1024,
    },
}
```

### 5. Monitor for Leaks

```go
// Good - Enable leak detection
config := &gowright.LeakDetectionConfig{
    EnableDetection: true,
    DetectionInterval: 10 * time.Second,
    AutoRepair: true,
    AlertOnLeaks: true,
}
```

## Troubleshooting

### Common Issues

**Memory leaks:**
```go
// Enable memory profiling and leak detection
config := &gowright.MemoryManagementConfig{
    MemoryProfiling: &gowright.MemoryProfilingConfig{
        EnableProfiling: true,
        HeapDumpOnLimit: true,
    },
    LeakDetection: &gowright.LeakDetectionConfig{
        EnableDetection: true,
        AutoCleanup: true,
    },
}
```

**High CPU usage:**
```go
// Enable CPU throttling and profiling
config := &gowright.CPUManagementConfig{
    EnableThrottling: true,
    ThrottleThreshold: 80,
    CPUProfiling: &gowright.CPUProfilingConfig{
        EnableProfiling: true,
    },
}
```

**Connection pool exhaustion:**
```go
// Increase pool size and enable monitoring
config := &gowright.ConnectionPoolingConfig{
    PoolSize: 100, // Increase pool size
    MaxWaitTime: 10 * time.Second,
    HealthCheckInterval: 30 * time.Second,
}
```

## Next Steps

- [Examples](../examples/basic-usage.md) - Resource management examples
- [Best Practices](../reference/best-practices.md) - Resource management best practices
- [API Reference](../reference/api.md) - Complete API documentation
- [Troubleshooting](../reference/troubleshooting.md) - Common issues and solutions