# Parallel Execution

Gowright provides sophisticated parallel execution capabilities that allow you to run tests concurrently while managing resources, dependencies, and ensuring test isolation. This significantly reduces test execution time and improves CI/CD pipeline efficiency.

## Overview

Parallel execution in Gowright provides:

- **Concurrent Test Execution**: Run multiple tests simultaneously
- **Resource Management**: Control memory, CPU, and connection usage
- **Test Isolation**: Ensure tests don't interfere with each other
- **Load Balancing**: Distribute tests across available resources
- **Dependency Management**: Handle test dependencies and ordering
- **Performance Monitoring**: Track resource usage and bottlenecks

## Basic Parallel Execution

### Simple Parallel Test Suite

```go
package main

import (
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestBasicParallelExecution(t *testing.T) {
    // Configure parallel execution
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 4, // Run up to 4 tests concurrently
            ResourceLimits: gowright.ResourceLimits{
                MaxMemoryMB:     512,  // 512MB total memory limit
                MaxCPUPercent:   80,   // 80% CPU usage limit
                MaxOpenFiles:    100,  // 100 open files limit
                MaxNetworkConns: 50,   // 50 network connections limit
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    err := framework.Initialize()
    assert.NoError(t, err)
    
    // Create test suite with independent tests
    testSuite := &gowright.TestSuite{
        Name: "Parallel API Tests",
        Description: "Independent API tests that can run in parallel",
        
        Tests: []gowright.Test{
            // Test 1: User API
            gowright.NewAPITestBuilder("Get User 1", "GET", "/api/users/1").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                ExpectJSONPath("$.id", 1).
                Build(),
            
            // Test 2: Product API
            gowright.NewAPITestBuilder("Get Product 1", "GET", "/api/products/1").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                ExpectJSONPath("$.id", 1).
                Build(),
            
            // Test 3: Category API
            gowright.NewAPITestBuilder("Get Categories", "GET", "/api/categories").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                ExpectJSONPath("$", gowright.IsArray).
                Build(),
            
            // Test 4: Health Check
            gowright.NewAPITestBuilder("Health Check", "GET", "/api/health").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                ExpectJSONPath("$.status", "healthy").
                Build(),
        },
        
        Config: &gowright.TestSuiteConfig{
            Parallel:       true,
            MaxConcurrency: 4,
            Timeout:        5 * time.Minute,
        },
    }
    
    // Execute tests in parallel
    startTime := time.Now()
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    executionTime := time.Since(startTime)
    
    assert.NoError(t, err)
    assert.Equal(t, 4, results.TotalTests)
    assert.Equal(t, 4, results.PassedTests)
    
    // Parallel execution should be faster than sequential
    // (assuming each test takes ~1 second, parallel should be ~1-2 seconds total)
    assert.Less(t, executionTime, 3*time.Second)
    
    // Verify parallel execution metrics
    assert.True(t, results.ParallelExecution)
    assert.Equal(t, 4, results.MaxConcurrency)
    assert.Greater(t, results.ConcurrencyUtilization, 0.5) // At least 50% utilization
}
```

### Resource-Aware Parallel Execution

```go
func TestResourceAwareParallelExecution(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 8,
            ResourceLimits: gowright.ResourceLimits{
                MaxMemoryMB:     1024, // 1GB memory limit
                MaxCPUPercent:   70,   // 70% CPU limit
                MaxOpenFiles:    200,  // 200 file handles
                MaxNetworkConns: 100,  // 100 network connections
            },
            LoadBalancing: &gowright.LoadBalancingConfig{
                Strategy:           "resource_aware", // Balance based on resource usage
                ResourceWeighting:  true,
                MemoryWeight:       0.4,
                CPUWeight:          0.4,
                NetworkWeight:      0.2,
            },
            Monitoring: &gowright.MonitoringConfig{
                EnableResourceMonitoring: true,
                MonitoringInterval:       1 * time.Second,
                AlertThresholds: gowright.AlertThresholds{
                    MemoryThreshold:  0.9, // Alert at 90% memory usage
                    CPUThreshold:     0.8, // Alert at 80% CPU usage
                    NetworkThreshold: 0.9, // Alert at 90% network usage
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Create resource-intensive test suite
    testSuite := &gowright.TestSuite{
        Name: "Resource-Intensive Tests",
        Tests: createResourceIntensiveTests(framework, 20), // 20 tests
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            ResourceAware: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify resource limits were respected
    assert.LessOrEqual(t, results.PeakMemoryUsage, 1024*1024*1024) // 1GB
    assert.LessOrEqual(t, results.PeakCPUUsage, 70.0)              // 70%
    assert.Equal(t, 0, results.ResourceLimitViolations)
    
    // Verify load balancing effectiveness
    assert.Greater(t, results.LoadBalancingEfficiency, 0.8) // 80% efficiency
}

func createResourceIntensiveTests(framework *gowright.Framework, count int) []gowright.Test {
    tests := make([]gowright.Test, count)
    
    for i := 0; i < count; i++ {
        tests[i] = &gowright.ResourceIntensiveTest{
            Name: fmt.Sprintf("Resource Test %d", i+1),
            TestFunc: func() error {
                // Simulate resource-intensive operations
                time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
                
                // Simulate memory usage
                data := make([]byte, 1024*1024*10) // 10MB
                _ = data
                
                // Simulate CPU usage
                for j := 0; j < 1000000; j++ {
                    _ = j * j
                }
                
                return nil
            },
            ResourceRequirements: &gowright.ResourceRequirements{
                MinMemoryMB:     10,  // Minimum 10MB memory
                MinCPUPercent:   5,   // Minimum 5% CPU
                NetworkConnections: 2, // 2 network connections
            },
        }
    }
    
    return tests
}
```

## Advanced Parallel Patterns

### Test Dependencies and Ordering

```go
func TestParallelWithDependencies(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 6,
            DependencyManagement: &gowright.DependencyManagementConfig{
                EnableDependencies: true,
                MaxWaitTime:        30 * time.Second,
                DeadlockDetection:  true,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Dependent Tests",
        Tests: []gowright.Test{
            // Setup tests (no dependencies)
            &gowright.DependentTest{
                Test: createSetupTest("Setup Database"),
                TestID: "setup_db",
                Dependencies: []string{}, // No dependencies
            },
            &gowright.DependentTest{
                Test: createSetupTest("Setup API"),
                TestID: "setup_api",
                Dependencies: []string{}, // No dependencies
            },
            
            // User tests (depend on database and API setup)
            &gowright.DependentTest{
                Test: createUserTest("Create User"),
                TestID: "create_user",
                Dependencies: []string{"setup_db", "setup_api"},
            },
            &gowright.DependentTest{
                Test: createUserTest("Update User"),
                TestID: "update_user",
                Dependencies: []string{"create_user"}, // Depends on user creation
            },
            
            // Product tests (can run in parallel with user tests after setup)
            &gowright.DependentTest{
                Test: createProductTest("Create Product"),
                TestID: "create_product",
                Dependencies: []string{"setup_db", "setup_api"},
            },
            &gowright.DependentTest{
                Test: createProductTest("Update Product"),
                TestID: "update_product",
                Dependencies: []string{"create_product"},
            },
            
            // Order tests (depend on both user and product)
            &gowright.DependentTest{
                Test: createOrderTest("Create Order"),
                TestID: "create_order",
                Dependencies: []string{"create_user", "create_product"},
            },
            
            // Cleanup tests (depend on all other tests)
            &gowright.DependentTest{
                Test: createCleanupTest("Cleanup"),
                TestID: "cleanup",
                Dependencies: []string{"update_user", "update_product", "create_order"},
            },
        },
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            DependencyAware: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify dependency execution order
    executionOrder := results.ExecutionOrder
    
    // Setup tests should run first
    setupIndices := []int{
        findTestIndex(executionOrder, "setup_db"),
        findTestIndex(executionOrder, "setup_api"),
    }
    
    // User and product creation should run after setup
    createUserIndex := findTestIndex(executionOrder, "create_user")
    createProductIndex := findTestIndex(executionOrder, "create_product")
    
    for _, setupIndex := range setupIndices {
        assert.Less(t, setupIndex, createUserIndex)
        assert.Less(t, setupIndex, createProductIndex)
    }
    
    // Order creation should run after user and product creation
    createOrderIndex := findTestIndex(executionOrder, "create_order")
    assert.Less(t, createUserIndex, createOrderIndex)
    assert.Less(t, createProductIndex, createOrderIndex)
    
    // Cleanup should run last
    cleanupIndex := findTestIndex(executionOrder, "cleanup")
    assert.Equal(t, len(executionOrder)-1, cleanupIndex)
}
```

### Dynamic Concurrency Adjustment

```go
func TestDynamicConcurrencyAdjustment(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 10,
            DynamicScaling: &gowright.DynamicScalingConfig{
                Enabled:              true,
                MinConcurrency:       2,
                MaxConcurrency:       10,
                ScaleUpThreshold:     0.8,  // Scale up when 80% utilized
                ScaleDownThreshold:   0.3,  // Scale down when 30% utilized
                ScaleUpFactor:        1.5,  // Increase by 50%
                ScaleDownFactor:      0.7,  // Decrease by 30%
                EvaluationInterval:   5 * time.Second,
                CooldownPeriod:       10 * time.Second,
            },
            AdaptiveScheduling: &gowright.AdaptiveSchedulingConfig{
                Enabled:                true,
                LearningEnabled:        true,
                TestComplexityAnalysis: true,
                HistoricalDataWeight:   0.7,
                RealtimeDataWeight:     0.3,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Create test suite with varying complexity
    testSuite := &gowright.TestSuite{
        Name: "Dynamic Scaling Tests",
        Tests: createVariableComplexityTests(50), // 50 tests with different complexities
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            DynamicScaling: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify dynamic scaling occurred
    assert.Greater(t, len(results.ConcurrencyAdjustments), 0)
    assert.True(t, results.DynamicScalingEffective)
    
    // Verify resource utilization improved
    assert.Greater(t, results.AverageResourceUtilization, 0.7) // 70% utilization
    
    // Verify adaptive scheduling learned from test patterns
    if results.AdaptiveSchedulingData != nil {
        assert.Greater(t, results.AdaptiveSchedulingData.LearningAccuracy, 0.8)
    }
}

func createVariableComplexityTests(count int) []gowright.Test {
    tests := make([]gowright.Test, count)
    complexities := []string{"simple", "medium", "complex"}
    
    for i := 0; i < count; i++ {
        complexity := complexities[i%len(complexities)]
        
        tests[i] = &gowright.ComplexityAwareTest{
            Name: fmt.Sprintf("Test %d (%s)", i+1, complexity),
            Complexity: complexity,
            TestFunc: func(complexity string) func() error {
                return func() error {
                    switch complexity {
                    case "simple":
                        time.Sleep(100 * time.Millisecond)
                    case "medium":
                        time.Sleep(500 * time.Millisecond)
                    case "complex":
                        time.Sleep(2 * time.Second)
                    }
                    return nil
                }
            }(complexity),
            EstimatedDuration: getEstimatedDuration(complexity),
            ResourceRequirements: getResourceRequirements(complexity),
        }
    }
    
    return tests
}
```

### Parallel UI Testing

```go
func TestParallelUITesting(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 3, // Limited by browser instances
            ResourceLimits: gowright.ResourceLimits{
                MaxMemoryMB: 2048, // 2GB for multiple browsers
            },
        },
        BrowserConfig: &gowright.BrowserConfig{
            Headless: true,
            PoolSize: 3, // Browser pool for parallel execution
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Parallel UI Tests",
        Tests: []gowright.Test{
            &gowright.UITest{
                Name: "Login Test - Chrome",
                BrowserConfig: &gowright.BrowserConfig{
                    Browser:  "chrome",
                    Headless: true,
                },
                Steps: []gowright.UIStep{
                    {Action: "navigate", Target: "https://example.com/login"},
                    {Action: "type", Target: "input[name='username']", Value: "user1"},
                    {Action: "type", Target: "input[name='password']", Value: "pass1"},
                    {Action: "click", Target: "button[type='submit']"},
                    {Action: "waitFor", Target: "div.dashboard"},
                },
            },
            &gowright.UITest{
                Name: "Registration Test - Firefox",
                BrowserConfig: &gowright.BrowserConfig{
                    Browser:  "firefox",
                    Headless: true,
                },
                Steps: []gowright.UIStep{
                    {Action: "navigate", Target: "https://example.com/register"},
                    {Action: "type", Target: "input[name='email']", Value: "test@example.com"},
                    {Action: "type", Target: "input[name='password']", Value: "password123"},
                    {Action: "click", Target: "button[type='submit']"},
                    {Action: "waitFor", Target: "div.success"},
                },
            },
            &gowright.UITest{
                Name: "Profile Test - Safari",
                BrowserConfig: &gowright.BrowserConfig{
                    Browser:  "safari",
                    Headless: true,
                },
                Steps: []gowright.UIStep{
                    {Action: "navigate", Target: "https://example.com/profile"},
                    {Action: "type", Target: "input[name='name']", Value: "Test User"},
                    {Action: "click", Target: "button.save"},
                    {Action: "waitFor", Target: "div.saved"},
                },
            },
        },
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            BrowserIsolation: true, // Ensure browser isolation
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify all UI tests passed
    assert.Equal(t, 3, results.PassedTests)
    assert.Equal(t, 0, results.FailedTests)
    
    // Verify browser isolation worked
    assert.Equal(t, 0, results.BrowserConflicts)
}
```

## Performance Optimization

### Load Testing with Parallel Execution

```go
func TestParallelLoadTesting(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 50, // High concurrency for load testing
            ResourceLimits: gowright.ResourceLimits{
                MaxMemoryMB:     4096, // 4GB memory
                MaxNetworkConns: 1000, // 1000 connections
            },
            LoadTesting: &gowright.LoadTestingConfig{
                Enabled:         true,
                RampUpDuration:  30 * time.Second,
                SustainDuration: 2 * time.Minute,
                RampDownDuration: 30 * time.Second,
                ThinkTime:       100 * time.Millisecond,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    loadTest := &gowright.LoadTest{
        Name: "API Load Test",
        Description: "Load test for user API endpoints",
        
        Scenarios: []gowright.LoadScenario{
            {
                Name:   "Get User Scenario",
                Weight: 60, // 60% of traffic
                TestFunc: func(userID int) gowright.Test {
                    return gowright.NewAPITestBuilder("Get User", "GET", fmt.Sprintf("/api/users/%d", userID)).
                        WithTester(framework.GetAPITester()).
                        ExpectStatus(200).
                        ExpectResponseTime(500 * time.Millisecond).
                        Build()
                },
                DataProvider: func() interface{} {
                    return rand.Intn(1000) + 1 // Random user ID 1-1000
                },
            },
            {
                Name:   "Create User Scenario",
                Weight: 30, // 30% of traffic
                TestFunc: func(userData map[string]interface{}) gowright.Test {
                    return gowright.NewAPITestBuilder("Create User", "POST", "/api/users").
                        WithTester(framework.GetAPITester()).
                        WithBody(userData).
                        ExpectStatus(201).
                        ExpectResponseTime(1 * time.Second).
                        Build()
                },
                DataProvider: func() interface{} {
                    return generateRandomUserData()
                },
            },
            {
                Name:   "Update User Scenario",
                Weight: 10, // 10% of traffic
                TestFunc: func(updateData map[string]interface{}) gowright.Test {
                    userID := updateData["id"].(int)
                    return gowright.NewAPITestBuilder("Update User", "PUT", fmt.Sprintf("/api/users/%d", userID)).
                        WithTester(framework.GetAPITester()).
                        WithBody(updateData).
                        ExpectStatus(200).
                        ExpectResponseTime(800 * time.Millisecond).
                        Build()
                },
                DataProvider: func() interface{} {
                    return generateRandomUpdateData()
                },
            },
        },
        
        LoadProfile: &gowright.LoadProfile{
            VirtualUsers:     50,
            Duration:         3 * time.Minute,
            RampUpDuration:   30 * time.Second,
            RampDownDuration: 30 * time.Second,
        },
        
        PerformanceThresholds: &gowright.PerformanceThresholds{
            MaxResponseTime:    2 * time.Second,
            MinThroughput:      100, // requests per second
            MaxErrorRate:       0.01, // 1% error rate
            MaxMemoryUsage:     2 * 1024 * 1024 * 1024, // 2GB
        },
    }
    
    loadTestRunner := gowright.NewLoadTestRunner(config.ParallelRunnerConfig)
    results, err := loadTestRunner.ExecuteLoadTest(loadTest)
    assert.NoError(t, err)
    
    // Verify load test results
    assert.LessOrEqual(t, results.AverageResponseTime, 2*time.Second)
    assert.GreaterOrEqual(t, results.Throughput, 100.0)
    assert.LessOrEqual(t, results.ErrorRate, 0.01)
    
    // Verify concurrency was maintained
    assert.GreaterOrEqual(t, results.AverageConcurrency, 45.0) // 90% of target
    assert.LessOrEqual(t, results.PeakMemoryUsage, 2*1024*1024*1024)
}
```

### Parallel Database Testing

```go
func TestParallelDatabaseTesting(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 10,
        },
        DatabaseConfig: &gowright.DatabaseConfig{
            Connections: map[string]*gowright.DBConnection{
                "test": {
                    Driver:       "postgres",
                    DSN:          "postgres://user:pass@localhost/testdb",
                    MaxOpenConns: 20, // Pool for parallel access
                    MaxIdleConns: 10,
                },
            },
            ConnectionPooling: &gowright.ConnectionPoolingConfig{
                EnablePooling:    true,
                PoolSize:         20,
                MaxWaitTime:      5 * time.Second,
                IdleTimeout:      10 * time.Minute,
                ConnectionReuse:  true,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Create parallel database tests
    testSuite := &gowright.TestSuite{
        Name: "Parallel Database Tests",
        
        SetupFunc: func() error {
            // Create test tables
            dbTester := framework.GetDatabaseTester()
            _, err := dbTester.Execute("test", `
                CREATE TABLE IF NOT EXISTS test_users (
                    id SERIAL PRIMARY KEY,
                    username VARCHAR(50) UNIQUE,
                    email VARCHAR(100),
                    created_at TIMESTAMP DEFAULT NOW()
                )
            `)
            return err
        },
        
        Tests: createParallelDatabaseTests(framework, 20), // 20 parallel DB tests
        
        TeardownFunc: func() error {
            // Cleanup
            dbTester := framework.GetDatabaseTester()
            _, err := dbTester.Execute("test", "DROP TABLE IF EXISTS test_users")
            return err
        },
        
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            DatabaseIsolation: true, // Ensure transaction isolation
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify all database tests passed
    assert.Equal(t, 20, results.PassedTests)
    assert.Equal(t, 0, results.FailedTests)
    
    // Verify no database conflicts occurred
    assert.Equal(t, 0, results.DatabaseConflicts)
    assert.Equal(t, 0, results.DeadlockCount)
}

func createParallelDatabaseTests(framework *gowright.Framework, count int) []gowright.Test {
    tests := make([]gowright.Test, count)
    
    for i := 0; i < count; i++ {
        testID := i + 1
        tests[i] = &gowright.DatabaseTest{
            Name:       fmt.Sprintf("Database Test %d", testID),
            Connection: "test",
            Setup: []string{
                fmt.Sprintf("INSERT INTO test_users (username, email) VALUES ('user%d', 'user%d@example.com')", testID, testID),
            },
            Query: "SELECT COUNT(*) as count FROM test_users WHERE username = ?",
            Args:  []interface{}{fmt.Sprintf("user%d", testID)},
            Expected: &gowright.DatabaseExpectation{
                RowCount: 1,
                Columns: map[string]interface{}{
                    "count": 1,
                },
            },
            Teardown: []string{
                fmt.Sprintf("DELETE FROM test_users WHERE username = 'user%d'", testID),
            },
            IsolationLevel: "READ_COMMITTED", // Ensure proper isolation
        }
    }
    
    return tests
}
```

## Monitoring and Debugging

### Real-time Monitoring

```go
func TestParallelExecutionMonitoring(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 8,
            Monitoring: &gowright.MonitoringConfig{
                EnableResourceMonitoring: true,
                EnablePerformanceMetrics: true,
                MonitoringInterval:       500 * time.Millisecond,
                MetricsCollection: &gowright.MetricsCollectionConfig{
                    CollectCPUMetrics:     true,
                    CollectMemoryMetrics:  true,
                    CollectNetworkMetrics: true,
                    CollectDiskMetrics:    true,
                    CollectCustomMetrics:  true,
                },
                RealTimeReporting: &gowright.RealTimeReportingConfig{
                    Enabled:        true,
                    WebSocketPort:  8080,
                    UpdateInterval: 1 * time.Second,
                },
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Start monitoring dashboard
    monitoringServer := gowright.NewMonitoringServer(config.ParallelRunnerConfig.Monitoring)
    go monitoringServer.Start()
    defer monitoringServer.Stop()
    
    testSuite := &gowright.TestSuite{
        Name: "Monitored Parallel Tests",
        Tests: createMonitoredTests(framework, 15),
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            EnableMonitoring: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify monitoring data was collected
    assert.NotNil(t, results.MonitoringData)
    assert.Greater(t, len(results.MonitoringData.CPUMetrics), 0)
    assert.Greater(t, len(results.MonitoringData.MemoryMetrics), 0)
    assert.Greater(t, len(results.MonitoringData.NetworkMetrics), 0)
    
    // Verify performance metrics
    assert.Greater(t, results.MonitoringData.AverageConcurrency, 0.0)
    assert.Greater(t, results.MonitoringData.ThroughputMetrics.RequestsPerSecond, 0.0)
    
    // Check for performance bottlenecks
    bottlenecks := results.MonitoringData.PerformanceBottlenecks
    if len(bottlenecks) > 0 {
        t.Logf("Performance bottlenecks detected: %v", bottlenecks)
    }
}
```

### Debugging Parallel Issues

```go
func TestParallelDebugging(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 5,
            Debugging: &gowright.DebuggingConfig{
                EnableDebugging:      true,
                LogLevel:             "debug",
                TraceExecution:       true,
                CaptureStackTraces:   true,
                DeadlockDetection:    true,
                RaceConditionDetection: true,
                ResourceLeakDetection: true,
            },
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Create tests that might have parallel issues
    testSuite := &gowright.TestSuite{
        Name: "Debug Parallel Tests",
        Tests: []gowright.Test{
            &gowright.SharedResourceTest{
                Name: "Shared Resource Test 1",
                TestFunc: func() error {
                    // Simulate shared resource access
                    return accessSharedResource("resource1", 1*time.Second)
                },
            },
            &gowright.SharedResourceTest{
                Name: "Shared Resource Test 2",
                TestFunc: func() error {
                    return accessSharedResource("resource1", 800*time.Millisecond)
                },
            },
            &gowright.SharedResourceTest{
                Name: "Shared Resource Test 3",
                TestFunc: func() error {
                    return accessSharedResource("resource2", 500*time.Millisecond)
                },
            },
        },
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
            EnableDebugging: true,
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Check debugging information
    debugInfo := results.DebuggingInfo
    assert.NotNil(t, debugInfo)
    
    // Check for detected issues
    if len(debugInfo.RaceConditions) > 0 {
        t.Logf("Race conditions detected: %v", debugInfo.RaceConditions)
    }
    
    if len(debugInfo.DeadlockWarnings) > 0 {
        t.Logf("Deadlock warnings: %v", debugInfo.DeadlockWarnings)
    }
    
    if len(debugInfo.ResourceLeaks) > 0 {
        t.Logf("Resource leaks detected: %v", debugInfo.ResourceLeaks)
    }
    
    // Verify execution traces are available
    assert.Greater(t, len(debugInfo.ExecutionTraces), 0)
}

var sharedResources = make(map[string]*sync.Mutex)
var resourceMutex sync.Mutex

func accessSharedResource(resourceName string, duration time.Duration) error {
    resourceMutex.Lock()
    if sharedResources[resourceName] == nil {
        sharedResources[resourceName] = &sync.Mutex{}
    }
    resourceMutex.Unlock()
    
    // Access shared resource
    sharedResources[resourceName].Lock()
    defer sharedResources[resourceName].Unlock()
    
    // Simulate work
    time.Sleep(duration)
    
    return nil
}
```

## Configuration Examples

### Complete Parallel Configuration

```json
{
  "parallel": true,
  "parallel_runner_config": {
    "max_concurrency": 8,
    "resource_limits": {
      "max_memory_mb": 2048,
      "max_cpu_percent": 80,
      "max_open_files": 200,
      "max_network_conns": 100
    },
    "load_balancing": {
      "strategy": "resource_aware",
      "resource_weighting": true,
      "memory_weight": 0.4,
      "cpu_weight": 0.4,
      "network_weight": 0.2
    },
    "dynamic_scaling": {
      "enabled": true,
      "min_concurrency": 2,
      "max_concurrency": 16,
      "scale_up_threshold": 0.8,
      "scale_down_threshold": 0.3,
      "evaluation_interval": "5s",
      "cooldown_period": "10s"
    },
    "dependency_management": {
      "enable_dependencies": true,
      "max_wait_time": "30s",
      "deadlock_detection": true
    },
    "monitoring": {
      "enable_resource_monitoring": true,
      "enable_performance_metrics": true,
      "monitoring_interval": "1s",
      "alert_thresholds": {
        "memory_threshold": 0.9,
        "cpu_threshold": 0.8,
        "network_threshold": 0.9
      }
    }
  }
}
```

## Best Practices

### 1. Design for Parallelism

```go
// Good - Independent tests
func TestIndependentAPIEndpoints(t *testing.T) {
    tests := []gowright.Test{
        createGetUserTest(1),
        createGetUserTest(2),
        createGetProductTest(1),
        createGetCategoryTest(),
    }
    // These can run in parallel safely
}

// Avoid - Tests with shared state
func TestWithSharedState(t *testing.T) {
    // These tests modify the same data and can't run in parallel
    createUserTest("testuser"),
    updateUserTest("testuser"),
    deleteUserTest("testuser"),
}
```

### 2. Manage Resources Appropriately

```go
// Good - Resource limits
config := &gowright.ParallelRunnerConfig{
    MaxConcurrency: runtime.NumCPU(), // Scale with available CPUs
    ResourceLimits: gowright.ResourceLimits{
        MaxMemoryMB: getAvailableMemory() / 2, // Use half available memory
    },
}
```

### 3. Handle Dependencies Correctly

```go
// Good - Explicit dependencies
tests := []gowright.Test{
    &gowright.DependentTest{
        Test: setupTest,
        TestID: "setup",
        Dependencies: []string{},
    },
    &gowright.DependentTest{
        Test: mainTest,
        TestID: "main",
        Dependencies: []string{"setup"},
    },
}
```

### 4. Monitor Performance

```go
// Good - Enable monitoring for parallel tests
config := &gowright.ParallelRunnerConfig{
    Monitoring: &gowright.MonitoringConfig{
        EnableResourceMonitoring: true,
        EnablePerformanceMetrics: true,
    },
}
```

### 5. Test Isolation

```go
// Good - Isolated test data
func createIsolatedTest(testID int) gowright.Test {
    return &gowright.DatabaseTest{
        Setup: []string{
            fmt.Sprintf("INSERT INTO test_data (id, value) VALUES (%d, 'test%d')", testID, testID),
        },
        Query: "SELECT value FROM test_data WHERE id = ?",
        Args:  []interface{}{testID},
        Teardown: []string{
            fmt.Sprintf("DELETE FROM test_data WHERE id = %d", testID),
        },
    }
}
```

## Troubleshooting

### Common Issues

**Resource exhaustion:**
```go
// Monitor and adjust resource limits
config := &gowright.ResourceLimits{
    MaxMemoryMB:     1024, // Reduce if hitting memory limits
    MaxNetworkConns: 50,   // Reduce if hitting connection limits
}
```

**Test dependencies not working:**
```go
// Ensure proper dependency declaration
&gowright.DependentTest{
    TestID: "dependent_test",
    Dependencies: []string{"prerequisite_test"}, // Must match exactly
}
```

**Deadlocks in parallel execution:**
```go
// Enable deadlock detection
config := &gowright.DependencyManagementConfig{
    DeadlockDetection: true,
    MaxWaitTime:       30 * time.Second, // Prevent infinite waits
}
```

## Next Steps

- [Resource Management](resource-management.md) - Monitor and control resource usage
- [Examples](../examples/basic-usage.md) - Parallel execution examples
- [Best Practices](../reference/best-practices.md) - Parallel testing best practices
- [Performance Testing](../testing-modules/integration-testing.md) - Advanced performance patterns