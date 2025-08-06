# Test Suites

Gowright's test suite functionality allows you to organize, group, and execute collections of tests with shared setup, teardown, and configuration. Test suites provide structure and enable complex testing workflows.

## Overview

Test suites in Gowright provide:

- **Test Organization**: Group related tests logically
- **Shared Setup/Teardown**: Common initialization and cleanup
- **Parallel Execution**: Run tests concurrently within suites
- **Dependency Management**: Control test execution order
- **Resource Sharing**: Share connections and data between tests
- **Comprehensive Reporting**: Suite-level and individual test reporting

## Basic Test Suite

### Simple Test Suite

```go
package main

import (
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestBasicTestSuite(t *testing.T) {
    // Create framework
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    err := framework.Initialize()
    assert.NoError(t, err)
    
    // Create test suite
    testSuite := &gowright.TestSuite{
        Name:        "User Management API Tests",
        Description: "Comprehensive tests for user management functionality",
        
        // Suite-level setup
        SetupFunc: func() error {
            // Initialize test database
            dbTester := framework.GetDatabaseTester()
            _, err := dbTester.Execute("test", `
                CREATE TABLE IF NOT EXISTS users (
                    id INTEGER PRIMARY KEY,
                    username TEXT UNIQUE NOT NULL,
                    email TEXT UNIQUE NOT NULL,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
                )
            `)
            return err
        },
        
        // Suite-level teardown
        TeardownFunc: func() error {
            // Clean up test database
            dbTester := framework.GetDatabaseTester()
            _, err := dbTester.Execute("test", "DROP TABLE IF EXISTS users")
            return err
        },
        
        // Individual tests
        Tests: []gowright.Test{
            // Test 1: Create user
            gowright.NewAPITestBuilder("Create User", "POST", "/api/users").
                WithTester(framework.GetAPITester()).
                WithBody(map[string]interface{}{
                    "username": "testuser1",
                    "email":    "test1@example.com",
                }).
                ExpectStatus(201).
                ExpectJSONPath("$.id", gowright.NotNil).
                ExpectJSONPath("$.username", "testuser1").
                Build(),
            
            // Test 2: Get user
            gowright.NewAPITestBuilder("Get User", "GET", "/api/users/1").
                WithTester(framework.GetAPITester()).
                ExpectStatus(200).
                ExpectJSONPath("$.username", "testuser1").
                ExpectJSONPath("$.email", "test1@example.com").
                Build(),
            
            // Test 3: Update user
            gowright.NewAPITestBuilder("Update User", "PUT", "/api/users/1").
                WithTester(framework.GetAPITester()).
                WithBody(map[string]interface{}{
                    "email": "updated@example.com",
                }).
                ExpectStatus(200).
                ExpectJSONPath("$.email", "updated@example.com").
                Build(),
            
            // Test 4: Delete user
            gowright.NewAPITestBuilder("Delete User", "DELETE", "/api/users/1").
                WithTester(framework.GetAPITester()).
                ExpectStatus(204).
                Build(),
        },
        
        // Suite configuration
        Config: &gowright.TestSuiteConfig{
            Parallel:    false, // Run tests sequentially
            MaxRetries:  2,
            Timeout:     5 * time.Minute,
            StopOnError: false,
        },
    }
    
    // Execute test suite
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify results
    assert.Equal(t, 4, results.TotalTests)
    assert.Equal(t, 4, results.PassedTests)
    assert.Equal(t, 0, results.FailedTests)
    assert.Equal(t, gowright.TestStatusPassed, results.OverallStatus)
}
```

## Advanced Test Suite Patterns

### Hierarchical Test Suites

```go
func TestHierarchicalTestSuites(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Parent test suite
    parentSuite := &gowright.TestSuite{
        Name: "E-commerce Application Tests",
        Description: "Complete test coverage for e-commerce platform",
        
        SetupFunc: func() error {
            // Global setup for all child suites
            return setupTestEnvironment()
        },
        
        TeardownFunc: func() error {
            // Global cleanup
            return cleanupTestEnvironment()
        },
        
        // Child test suites
        ChildSuites: []*gowright.TestSuite{
            // User Management Suite
            {
                Name: "User Management",
                SetupFunc: func() error {
                    return setupUserTestData()
                },
                Tests: []gowright.Test{
                    createUserRegistrationTest(),
                    createUserLoginTest(),
                    createUserProfileTest(),
                },
                Config: &gowright.TestSuiteConfig{
                    Parallel: true,
                    Tags:     []string{"user", "auth"},
                },
            },
            
            // Product Management Suite
            {
                Name: "Product Management",
                SetupFunc: func() error {
                    return setupProductTestData()
                },
                Tests: []gowright.Test{
                    createProductCRUDTest(),
                    createProductSearchTest(),
                    createProductCategoryTest(),
                },
                Config: &gowright.TestSuiteConfig{
                    Parallel: true,
                    Tags:     []string{"product", "catalog"},
                },
            },
            
            // Order Management Suite
            {
                Name: "Order Management",
                Dependencies: []string{"User Management", "Product Management"},
                SetupFunc: func() error {
                    return setupOrderTestData()
                },
                Tests: []gowright.Test{
                    createOrderCreationTest(),
                    createOrderPaymentTest(),
                    createOrderFulfillmentTest(),
                },
                Config: &gowright.TestSuiteConfig{
                    Parallel: false, // Sequential for order flow
                    Tags:     []string{"order", "payment"},
                },
            },
        },
        
        Config: &gowright.TestSuiteConfig{
            Parallel:    true,
            MaxRetries:  3,
            Timeout:     30 * time.Minute,
            StopOnError: false,
        },
    }
    
    framework.SetTestSuite(parentSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify hierarchical execution
    assert.Greater(t, results.TotalTests, 0)
    assert.Equal(t, len(parentSuite.ChildSuites), len(results.ChildSuiteResults))
}
```

### Data-Driven Test Suites

```go
func TestDataDrivenTestSuite(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Test data sets
    testDataSets := []struct {
        name     string
        userData map[string]interface{}
        expected map[string]interface{}
    }{
        {
            name: "Valid User Data",
            userData: map[string]interface{}{
                "username": "validuser",
                "email":    "valid@example.com",
                "age":      25,
            },
            expected: map[string]interface{}{
                "status": 201,
                "valid":  true,
            },
        },
        {
            name: "Invalid Email",
            userData: map[string]interface{}{
                "username": "invaliduser",
                "email":    "invalid-email",
                "age":      25,
            },
            expected: map[string]interface{}{
                "status": 400,
                "valid":  false,
            },
        },
        {
            name: "Missing Username",
            userData: map[string]interface{}{
                "email": "missing@example.com",
                "age":   25,
            },
            expected: map[string]interface{}{
                "status": 400,
                "valid":  false,
            },
        },
    }
    
    // Create test suite for each data set
    var childSuites []*gowright.TestSuite
    
    for _, testData := range testDataSets {
        suite := &gowright.TestSuite{
            Name:        fmt.Sprintf("User Validation - %s", testData.name),
            Description: fmt.Sprintf("Test user validation with %s", testData.name),
            
            Tests: []gowright.Test{
                gowright.NewAPITestBuilder("Create User", "POST", "/api/users").
                    WithTester(framework.GetAPITester()).
                    WithBody(testData.userData).
                    ExpectStatus(testData.expected["status"].(int)).
                    WithCustomValidation(func(response *gowright.APIResponse) error {
                        if testData.expected["valid"].(bool) {
                            // Validate successful creation
                            var result map[string]interface{}
                            if err := json.Unmarshal(response.Body, &result); err != nil {
                                return err
                            }
                            if result["id"] == nil {
                                return errors.New("expected user ID in response")
                            }
                        } else {
                            // Validate error response
                            var result map[string]interface{}
                            if err := json.Unmarshal(response.Body, &result); err != nil {
                                return err
                            }
                            if result["error"] == nil {
                                return errors.New("expected error in response")
                            }
                        }
                        return nil
                    }).
                    Build(),
            },
            
            Config: &gowright.TestSuiteConfig{
                Tags: []string{"validation", "data-driven"},
            },
        }
        
        childSuites = append(childSuites, suite)
    }
    
    // Main test suite
    mainSuite := &gowright.TestSuite{
        Name:        "Data-Driven User Validation Tests",
        ChildSuites: childSuites,
        Config: &gowright.TestSuiteConfig{
            Parallel: true,
        },
    }
    
    framework.SetTestSuite(mainSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    assert.Equal(t, len(testDataSets), len(results.ChildSuiteResults))
}
```

### Performance Test Suites

```go
func TestPerformanceTestSuite(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    performanceSuite := &gowright.TestSuite{
        Name: "API Performance Test Suite",
        Description: "Load and performance testing for critical API endpoints",
        
        SetupFunc: func() error {
            // Setup performance test data
            return setupPerformanceTestData()
        },
        
        Tests: []gowright.Test{
            // Load test for user creation
            &gowright.PerformanceTest{
                Name: "User Creation Load Test",
                Description: "Test user creation under load",
                
                TestFunc: func() gowright.Test {
                    return gowright.NewAPITestBuilder("Create User", "POST", "/api/users").
                        WithTester(framework.GetAPITester()).
                        WithBody(map[string]interface{}{
                            "username": fmt.Sprintf("loadtest_%d", time.Now().UnixNano()),
                            "email":    fmt.Sprintf("loadtest_%d@example.com", time.Now().UnixNano()),
                        }).
                        ExpectStatus(201).
                        ExpectResponseTime(500 * time.Millisecond).
                        Build()
                },
                
                LoadConfig: &gowright.LoadConfig{
                    ConcurrentUsers: 50,
                    Duration:        2 * time.Minute,
                    RampUpTime:      30 * time.Second,
                    RampDownTime:    30 * time.Second,
                },
                
                PerformanceThresholds: &gowright.PerformanceThresholds{
                    MaxResponseTime:    1 * time.Second,
                    MinThroughput:      100, // requests per second
                    MaxErrorRate:       0.01, // 1% error rate
                    MaxMemoryUsage:     100 * 1024 * 1024, // 100MB
                },
            },
            
            // Stress test for user search
            &gowright.PerformanceTest{
                Name: "User Search Stress Test",
                Description: "Stress test user search functionality",
                
                TestFunc: func() gowright.Test {
                    searchTerms := []string{"john", "jane", "test", "user", "admin"}
                    term := searchTerms[rand.Intn(len(searchTerms))]
                    
                    return gowright.NewAPITestBuilder("Search Users", "GET", "/api/users/search").
                        WithTester(framework.GetAPITester()).
                        WithQueryParams(map[string]string{
                            "q": term,
                            "limit": "10",
                        }).
                        ExpectStatus(200).
                        ExpectResponseTime(200 * time.Millisecond).
                        Build()
                },
                
                LoadConfig: &gowright.LoadConfig{
                    ConcurrentUsers: 100,
                    Duration:        5 * time.Minute,
                    RampUpTime:      1 * time.Minute,
                },
                
                PerformanceThresholds: &gowright.PerformanceThresholds{
                    MaxResponseTime: 500 * time.Millisecond,
                    MinThroughput:   200,
                    MaxErrorRate:    0.005, // 0.5% error rate
                },
            },
        },
        
        TeardownFunc: func() error {
            // Cleanup performance test data
            return cleanupPerformanceTestData()
        },
        
        Config: &gowright.TestSuiteConfig{
            Parallel: false, // Run performance tests sequentially
            Tags:     []string{"performance", "load", "stress"},
            Timeout:  15 * time.Minute,
        },
    }
    
    framework.SetTestSuite(performanceSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Verify performance metrics
    for _, testResult := range results.TestResults {
        if perfResult, ok := testResult.(*gowright.PerformanceTestResult); ok {
            assert.LessOrEqual(t, perfResult.AverageResponseTime, 1*time.Second)
            assert.GreaterOrEqual(t, perfResult.Throughput, 100.0)
            assert.LessOrEqual(t, perfResult.ErrorRate, 0.01)
        }
    }
}
```

## Test Suite Configuration

### Comprehensive Configuration

```go
func TestAdvancedTestSuiteConfiguration(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Advanced Configuration Test Suite",
        
        Config: &gowright.TestSuiteConfig{
            // Execution settings
            Parallel:     true,
            MaxRetries:   3,
            Timeout:      10 * time.Minute,
            StopOnError:  false,
            
            // Resource limits
            ResourceLimits: &gowright.ResourceLimits{
                MaxMemoryMB:     512,
                MaxCPUPercent:   80,
                MaxOpenFiles:    100,
                MaxNetworkConns: 50,
            },
            
            // Retry configuration
            RetryConfig: &gowright.RetryConfig{
                MaxRetries:   3,
                InitialDelay: 1 * time.Second,
                MaxDelay:     10 * time.Second,
                Multiplier:   2.0,
                RetryableErrors: []string{
                    "connection refused",
                    "timeout",
                    "temporary failure",
                },
            },
            
            // Reporting configuration
            ReportConfig: &gowright.SuiteReportConfig{
                GenerateHTML:     true,
                GenerateJSON:     true,
                IncludeMetrics:   true,
                ScreenshotOnFail: true,
                DetailedLogs:     true,
                OutputDir:        "./test-reports/suites",
            },
            
            // Environment configuration
            Environment: map[string]string{
                "TEST_ENV":     "integration",
                "API_BASE_URL": "https://api.test.example.com",
                "DB_NAME":      "test_database",
            },
            
            // Tags for filtering
            Tags: []string{"integration", "api", "database"},
            
            // Hooks
            Hooks: &gowright.TestSuiteHooks{
                BeforeEach: func(testName string) error {
                    log.Printf("Starting test: %s", testName)
                    return nil
                },
                AfterEach: func(testName string, result *gowright.TestResult) error {
                    log.Printf("Completed test: %s, Status: %s", testName, result.Status)
                    return nil
                },
                OnError: func(testName string, err error) error {
                    log.Printf("Test failed: %s, Error: %v", testName, err)
                    // Could send notification, capture additional logs, etc.
                    return nil
                },
            },
        },
        
        // Test definitions...
        Tests: []gowright.Test{
            // Your tests here
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
}
```

### Environment-Specific Suites

```go
func TestEnvironmentSpecificSuites(t *testing.T) {
    environment := os.Getenv("TEST_ENVIRONMENT")
    if environment == "" {
        environment = "development"
    }
    
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    var testSuite *gowright.TestSuite
    
    switch environment {
    case "development":
        testSuite = createDevelopmentTestSuite(framework)
    case "staging":
        testSuite = createStagingTestSuite(framework)
    case "production":
        testSuite = createProductionTestSuite(framework)
    default:
        t.Fatalf("Unknown environment: %s", environment)
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
}

func createDevelopmentTestSuite(framework *gowright.Framework) *gowright.TestSuite {
    return &gowright.TestSuite{
        Name: "Development Environment Tests",
        Config: &gowright.TestSuiteConfig{
            Parallel:    false, // Sequential for easier debugging
            MaxRetries:  1,
            StopOnError: true,  // Stop on first error for quick feedback
            Environment: map[string]string{
                "API_BASE_URL": "http://localhost:8080",
                "DB_HOST":      "localhost",
            },
            Tags: []string{"development", "local"},
        },
        Tests: []gowright.Test{
            // Development-specific tests
        },
    }
}

func createStagingTestSuite(framework *gowright.Framework) *gowright.TestSuite {
    return &gowright.TestSuite{
        Name: "Staging Environment Tests",
        Config: &gowright.TestSuiteConfig{
            Parallel:    true,
            MaxRetries:  2,
            StopOnError: false,
            Environment: map[string]string{
                "API_BASE_URL": "https://api.staging.example.com",
                "DB_HOST":      "staging-db.example.com",
            },
            Tags: []string{"staging", "pre-production"},
        },
        Tests: []gowright.Test{
            // Staging-specific tests
        },
    }
}

func createProductionTestSuite(framework *gowright.Framework) *gowright.TestSuite {
    return &gowright.TestSuite{
        Name: "Production Smoke Tests",
        Config: &gowright.TestSuiteConfig{
            Parallel:    true,
            MaxRetries:  3,
            StopOnError: false,
            Timeout:     5 * time.Minute, // Shorter timeout for production
            Environment: map[string]string{
                "API_BASE_URL": "https://api.example.com",
            },
            Tags: []string{"production", "smoke"},
        },
        Tests: []gowright.Test{
            // Production smoke tests only
        },
    }
}
```

## Test Suite Filtering and Selection

### Tag-Based Filtering

```go
func TestTagBasedFiltering(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Comprehensive Test Suite",
        
        Tests: []gowright.Test{
            &gowright.TaggedTest{
                Test: createAPITest("User API Test"),
                Tags: []string{"api", "user", "smoke"},
            },
            &gowright.TaggedTest{
                Test: createUITest("Login UI Test"),
                Tags: []string{"ui", "auth", "smoke"},
            },
            &gowright.TaggedTest{
                Test: createDatabaseTest("User DB Test"),
                Tags: []string{"database", "user", "integration"},
            },
            &gowright.TaggedTest{
                Test: createPerformanceTest("Load Test"),
                Tags: []string{"performance", "load", "stress"},
            },
        },
    }
    
    // Run only smoke tests
    smokeResults, err := framework.ExecuteTestSuiteWithTags(testSuite, []string{"smoke"})
    assert.NoError(t, err)
    assert.Equal(t, 2, smokeResults.TotalTests) // Only API and UI tests
    
    // Run only integration tests
    integrationResults, err := framework.ExecuteTestSuiteWithTags(testSuite, []string{"integration"})
    assert.NoError(t, err)
    assert.Equal(t, 1, integrationResults.TotalTests) // Only database test
    
    // Run tests with multiple tags (OR logic)
    multiTagResults, err := framework.ExecuteTestSuiteWithTags(testSuite, []string{"api", "ui"})
    assert.NoError(t, err)
    assert.Equal(t, 2, multiTagResults.TotalTests) // API and UI tests
}
```

### Conditional Test Execution

```go
func TestConditionalExecution(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Conditional Test Suite",
        
        Tests: []gowright.Test{
            &gowright.ConditionalTest{
                Test: createAPITest("API Test"),
                Condition: func() bool {
                    // Only run if API is available
                    return isAPIAvailable()
                },
            },
            &gowright.ConditionalTest{
                Test: createUITest("UI Test"),
                Condition: func() bool {
                    // Only run if browser is available
                    return isBrowserAvailable()
                },
            },
            &gowright.ConditionalTest{
                Test: createDatabaseTest("Database Test"),
                Condition: func() bool {
                    // Only run if database is available
                    return isDatabaseAvailable()
                },
            },
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Results will only include tests whose conditions were met
    assert.LessOrEqual(t, results.TotalTests, 3)
}
```

## Test Suite Reporting

### Custom Reporting

```go
func TestCustomReporting(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    testSuite := &gowright.TestSuite{
        Name: "Custom Reporting Test Suite",
        
        Config: &gowright.TestSuiteConfig{
            ReportConfig: &gowright.SuiteReportConfig{
                CustomReporters: []gowright.TestReporter{
                    &CustomJUnitReporter{
                        OutputFile: "./reports/junit.xml",
                    },
                    &CustomSlackReporter{
                        WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
                        Channel:    "#test-results",
                    },
                    &CustomDashboardReporter{
                        DashboardURL: "https://dashboard.example.com/api/results",
                        APIKey:       os.Getenv("DASHBOARD_API_KEY"),
                    },
                },
            },
        },
        
        Tests: []gowright.Test{
            // Your tests here
        },
    }
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    
    // Custom reporters will have been called automatically
}

// Custom JUnit reporter
type CustomJUnitReporter struct {
    OutputFile string
}

func (r *CustomJUnitReporter) Report(results *gowright.TestSuiteResults) error {
    junitXML := convertToJUnitXML(results)
    return ioutil.WriteFile(r.OutputFile, []byte(junitXML), 0644)
}

// Custom Slack reporter
type CustomSlackReporter struct {
    WebhookURL string
    Channel    string
}

func (r *CustomSlackReporter) Report(results *gowright.TestSuiteResults) error {
    message := formatSlackMessage(results)
    return sendSlackMessage(r.WebhookURL, r.Channel, message)
}

// Custom dashboard reporter
type CustomDashboardReporter struct {
    DashboardURL string
    APIKey       string
}

func (r *CustomDashboardReporter) Report(results *gowright.TestSuiteResults) error {
    payload := convertToDashboardFormat(results)
    return sendToDashboard(r.DashboardURL, r.APIKey, payload)
}
```

## Best Practices

### 1. Organize Tests Logically

```go
// Good - Logical grouping
testSuite := &gowright.TestSuite{
    Name: "User Management Tests",
    ChildSuites: []*gowright.TestSuite{
        {Name: "Authentication Tests", Tests: authTests},
        {Name: "Profile Management Tests", Tests: profileTests},
        {Name: "Permission Tests", Tests: permissionTests},
    },
}
```

### 2. Use Appropriate Parallelization

```go
// Good - Parallel for independent tests
{
    Name: "Independent API Tests",
    Config: &gowright.TestSuiteConfig{
        Parallel: true, // Safe to run in parallel
    },
}

// Good - Sequential for dependent tests
{
    Name: "Order Workflow Tests",
    Config: &gowright.TestSuiteConfig{
        Parallel: false, // Must run in sequence
    },
}
```

### 3. Implement Proper Setup and Teardown

```go
testSuite := &gowright.TestSuite{
    SetupFunc: func() error {
        // Setup shared resources
        return setupSharedTestData()
    },
    TeardownFunc: func() error {
        // Always cleanup, even if tests fail
        return cleanupSharedTestData()
    },
}
```

### 4. Use Tags for Organization

```go
// Tag tests for easy filtering
tests := []gowright.Test{
    &gowright.TaggedTest{
        Test: apiTest,
        Tags: []string{"api", "smoke", "critical"},
    },
    &gowright.TaggedTest{
        Test: uiTest,
        Tags: []string{"ui", "regression", "slow"},
    },
}
```

### 5. Configure Appropriate Timeouts

```go
config := &gowright.TestSuiteConfig{
    Timeout: 30 * time.Minute, // Suite-level timeout
    RetryConfig: &gowright.RetryConfig{
        MaxRetries:   3,
        InitialDelay: 1 * time.Second,
    },
}
```

## Troubleshooting

### Common Issues

**Tests running out of order:**
```go
// Use dependencies to control execution order
{
    Name: "Dependent Test Suite",
    Dependencies: []string{"Setup Suite"},
    Config: &gowright.TestSuiteConfig{
        Parallel: false, // Ensure sequential execution
    },
}
```

**Resource conflicts in parallel execution:**
```go
// Use resource limits and proper isolation
config := &gowright.TestSuiteConfig{
    Parallel: true,
    ResourceLimits: &gowright.ResourceLimits{
        MaxMemoryMB:     256, // Per test limit
        MaxNetworkConns: 10,  // Prevent connection exhaustion
    },
}
```

**Setup/teardown failures:**
```go
// Add error handling and logging
SetupFunc: func() error {
    if err := setupDatabase(); err != nil {
        log.Printf("Database setup failed: %v", err)
        return fmt.Errorf("setup failed: %w", err)
    }
    return nil
},
```

## Next Steps

- [Assertions](assertions.md) - Advanced assertion patterns
- [Reporting](reporting.md) - Comprehensive test reporting
- [Parallel Execution](parallel-execution.md) - Optimize test performance
- [Examples](../examples/basic-usage.md) - Test suite examples