# Quick Start

Get up and running with Gowright in just a few minutes. This guide will walk you through creating your first test and understanding the basic concepts.

## Your First Test

Let's create a simple test that demonstrates Gowright's capabilities across different testing modules.

### 1. Create a Test File

Create a new file called `first_test.go`:

```go title="first_test.go"
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestFirstGowrightTest(t *testing.T) {
    // Initialize the framework
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    if err := framework.Initialize(); err != nil {
        t.Fatal("Failed to initialize framework:", err)
    }
    
    fmt.Println("🚀 Running your first Gowright test!")
    
    // Test 1: API Testing
    testAPIEndpoint(t, framework)
    
    // Test 2: Database Testing
    testDatabaseOperations(t, framework)
    
    // Test 3: UI Testing (optional - requires Chrome)
    testUIInteraction(t, framework)
    
    // Test 4: Mobile Testing (optional - requires Appium server)
    testMobileInteraction(t, framework)
    
    fmt.Println("✅ All tests completed successfully!")
}

func testAPIEndpoint(t *testing.T, framework *gowright.Framework) {
    fmt.Println("📡 Testing API endpoint...")
    
    // Create API tester
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Test GET request
    response, err := apiTester.Get("/get", nil)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // Test with API test builder
    test := gowright.NewAPITestBuilder("Get Request Test", "GET", "/get").
        WithTester(apiTester).
        ExpectStatus(http.StatusOK).
        ExpectJSONPath("$.url", "https://httpbin.org/get").
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    
    fmt.Println("  ✅ API test passed")
}

func testDatabaseOperations(t *testing.T, framework *gowright.Framework) {
    fmt.Println("🗄️  Testing database operations...")
    
    // Create database tester with SQLite in-memory database
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
    
    // Create test table
    _, err = dbTester.Execute("test", `
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
    assert.NoError(t, err)
    
    // Test database operations
    dbTest := &gowright.DatabaseTest{
        Name:       "User Creation Test",
        Connection: "test",
        Setup: []string{
            "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')",
            "INSERT INTO users (name, email) VALUES ('Jane Smith', 'jane@example.com')",
        },
        Query: "SELECT COUNT(*) as count FROM users WHERE name LIKE '%Doe%'",
        Expected: &gowright.DatabaseExpectation{
            RowCount: 1,
        },
        Teardown: []string{
            "DELETE FROM users WHERE email IN ('john@example.com', 'jane@example.com')",
        },
    }
    
    result := dbTester.ExecuteTest(dbTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    
    fmt.Println("  ✅ Database test passed")
}

func testUIInteraction(t *testing.T, framework *gowright.Framework) {
    fmt.Println("🌐 Testing UI interaction...")
    
    // Create UI tester
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    if err != nil {
        fmt.Printf("  ⚠️  Skipping UI test (Chrome not available): %v\n", err)
        return
    }
    defer uiTester.Cleanup()
    
    // Navigate to a test page
    err = uiTester.Navigate("https://httpbin.org/forms/post")
    assert.NoError(t, err)
    
    // Fill out a form
    err = uiTester.Type("input[name='custname']", "Test User")
    assert.NoError(t, err)
    
    err = uiTester.Type("input[name='custtel']", "123-456-7890")
    assert.NoError(t, err)
    
    err = uiTester.Type("input[name='custemail']", "test@example.com")
    assert.NoError(t, err)
    
    // Take a screenshot
    screenshotPath, err := uiTester.TakeScreenshot("form-filled.png")
    assert.NoError(t, err)
    assert.NotEmpty(t, screenshotPath)
    
    fmt.Println("  ✅ UI test passed")
}

func testMobileInteraction(t *testing.T, framework *gowright.Framework) {
    fmt.Println("📱 Testing mobile interaction...")
    
    // Create Appium client
    client := gowright.NewAppiumClient("http://localhost:4723")
    ctx := context.Background()
    
    // Define Android capabilities for calculator app
    caps := gowright.AppiumCapabilities{
        PlatformName:   "Android",
        DeviceName:     "emulator-5554",
        AppPackage:     "com.android.calculator2",
        AppActivity:    ".Calculator",
        AutomationName: "UiAutomator2",
        NoReset:        true,
    }
    
    // Create session
    err := client.CreateSession(ctx, caps)
    if err != nil {
        fmt.Printf("  ⚠️  Skipping mobile test (Appium not available): %v\n", err)
        return
    }
    defer client.DeleteSession(ctx)
    
    // Perform simple calculation: 5 + 3
    digit5, err := client.WaitForElementClickable(ctx, gowright.ByID, "com.android.calculator2:id/digit_5", 5*time.Second)
    if err != nil {
        fmt.Printf("  ⚠️  Skipping mobile test (Calculator not available): %v\n", err)
        return
    }
    
    err = digit5.Click(ctx)
    assert.NoError(t, err)
    
    plus, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/op_add")
    assert.NoError(t, err)
    err = plus.Click(ctx)
    assert.NoError(t, err)
    
    digit3, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/digit_3")
    assert.NoError(t, err)
    err = digit3.Click(ctx)
    assert.NoError(t, err)
    
    equals, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/eq")
    assert.NoError(t, err)
    err = equals.Click(ctx)
    assert.NoError(t, err)
    
    // Take screenshot
    screenshot, err := client.TakeScreenshot(ctx)
    assert.NoError(t, err)
    assert.NotEmpty(t, screenshot)
    
    fmt.Println("  ✅ Mobile test passed")
}

func main() {
    // Run the test
    testing.Main(func(pat, str string) (bool, error) { return true, nil },
        []testing.InternalTest{
            {
                Name: "TestFirstGowrightTest",
                F:    TestFirstGowrightTest,
            },
        },
        []testing.InternalBenchmark{},
        []testing.InternalExample{},
    )
}
```

### 2. Run Your First Test

```bash
# Install dependencies
go mod tidy

# Run the test
go run first_test.go
```

Expected output:
```
🚀 Running your first Gowright test!
📡 Testing API endpoint...
  ✅ API test passed
🗄️  Testing database operations...
  ✅ Database test passed
🌐 Testing UI interaction...
  ✅ UI test passed
✅ All tests completed successfully!
PASS
```

## Understanding the Test

Let's break down what happened in your first test:

### API Testing
```go
// Create an API tester with base configuration
apiTester := gowright.NewAPITester(&gowright.APIConfig{
    BaseURL: "https://httpbin.org",
    Timeout: 10 * time.Second,
})

// Use the test builder for structured testing
test := gowright.NewAPITestBuilder("Get Request Test", "GET", "/get").
    WithTester(apiTester).
    ExpectStatus(http.StatusOK).
    ExpectJSONPath("$.url", "https://httpbin.org/get").
    Build()
```

**Key Concepts:**
- **API Tester**: Handles HTTP requests and responses
- **Test Builder**: Fluent API for creating structured tests
- **Expectations**: Define what constitutes a passing test

### Database Testing
```go
// Create database test with setup, execution, and teardown
dbTest := &gowright.DatabaseTest{
    Name:       "User Creation Test",
    Connection: "test",
    Setup:      []string{/* SQL statements */},
    Query:      "SELECT COUNT(*) as count FROM users WHERE name LIKE '%Doe%'",
    Expected:   &gowright.DatabaseExpectation{RowCount: 1},
    Teardown:   []string{/* Cleanup SQL */},
}
```

**Key Concepts:**
- **Database Tester**: Manages database connections and transactions
- **Test Structure**: Setup → Execute → Validate → Teardown
- **Expectations**: Validate query results and row counts

### UI Testing
```go
// Create UI tester with browser configuration
uiTester := gowright.NewRodUITester()
err := uiTester.Initialize(&gowright.BrowserConfig{
    Headless: true,
    Timeout:  30 * time.Second,
})

// Interact with web elements
err = uiTester.Type("input[name='custname']", "Test User")
```

**Key Concepts:**
- **UI Tester**: Controls browser automation
- **Element Interaction**: Type, click, wait for elements
- **Screenshot Capture**: Visual validation and debugging

## Configuration-Based Testing

For more complex scenarios, use configuration files:

### 1. Create Configuration File

```json title="gowright-config.json"
{
  "log_level": "info",
  "parallel": false,
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "window_size": {
      "width": 1920,
      "height": 1080
    }
  },
  "api_config": {
    "base_url": "https://httpbin.org",
    "timeout": "10s",
    "headers": {
      "User-Agent": "Gowright-QuickStart/1.0"
    }
  },
  "database_config": {
    "connections": {
      "main": {
        "driver": "sqlite3",
        "dsn": ":memory:",
        "max_open_conns": 5,
        "max_idle_conns": 2
      }
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

### 2. Use Configuration in Tests

```go title="config_test.go"
package main

import (
    "testing"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestWithConfiguration(t *testing.T) {
    // Load configuration from file
    config, err := gowright.LoadConfigFromFile("gowright-config.json")
    assert.NoError(t, err)
    
    // Create framework with loaded configuration
    framework := gowright.New(config)
    defer framework.Close()
    
    err = framework.Initialize()
    assert.NoError(t, err)
    
    // Your tests here...
    // The framework will use the configuration settings
}
```

## Integration Testing Example

Here's a more advanced example showing integration testing:

```go title="integration_test.go"
func TestUserRegistrationWorkflow(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Create integration tester
    integrationTester := gowright.NewIntegrationTester(
        framework.GetAPITester(),
        framework.GetUITester(),
        framework.GetDatabaseTester(),
    )
    
    // Define integration test workflow
    integrationTest := &gowright.IntegrationTest{
        Name: "User Registration Workflow",
        Steps: []gowright.IntegrationStep{
            {
                Type: gowright.StepTypeUI,
                Action: gowright.UIStepAction{
                    Navigate: "https://example.com/register",
                    Interactions: []gowright.UIInteraction{
                        {Type: "type", Selector: "#name", Value: "John Doe"},
                        {Type: "type", Selector: "#email", Value: "john@example.com"},
                        {Type: "click", Selector: "#submit"},
                    },
                },
                Name: "Fill Registration Form",
            },
            {
                Type: gowright.StepTypeAPI,
                Action: gowright.APIStepAction{
                    Method:   "GET",
                    Endpoint: "/api/users?email=john@example.com",
                },
                Validation: gowright.APIStepValidation{
                    ExpectedStatusCode: 200,
                    JSONPath: map[string]interface{}{
                        "$.email": "john@example.com",
                        "$.name":  "John Doe",
                    },
                },
                Name: "Verify User via API",
            },
            {
                Type: gowright.StepTypeDatabase,
                Action: gowright.DatabaseStepAction{
                    Connection: "main",
                    Query:      "SELECT COUNT(*) FROM users WHERE email = ?",
                    Args:       []interface{}{"john@example.com"},
                },
                Validation: gowright.DatabaseStepValidation{
                    ExpectedRowCount: &[]int{1}[0],
                },
                Name: "Confirm User in Database",
            },
        },
    }
    
    // Execute the integration test
    result := integrationTester.ExecuteTest(integrationTest)
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Test Reports

Gowright automatically generates comprehensive reports:

### JSON Report Structure
```json
{
  "test_suite": "Quick Start Tests",
  "start_time": "2024-01-15T10:30:00Z",
  "end_time": "2024-01-15T10:32:15Z",
  "duration": "2m15s",
  "total_tests": 3,
  "passed_tests": 3,
  "failed_tests": 0,
  "tests": [
    {
      "name": "API Test",
      "status": "passed",
      "duration": "1.2s",
      "steps": [...]
    }
  ]
}
```

### HTML Report Features
- Interactive test result dashboard
- Detailed step-by-step execution logs
- Screenshot attachments for UI tests
- Performance metrics and timing
- Error details with stack traces

## Best Practices from the Start

### 1. Organize Your Tests
```
tests/
├── unit/           # Unit tests
├── integration/    # Integration tests
├── e2e/           # End-to-end tests
└── config/        # Test configurations
```

### 2. Use Descriptive Test Names
```go
// Good
func TestUserCanLoginWithValidCredentials(t *testing.T) {}
func TestAPIReturns404ForNonexistentUser(t *testing.T) {}

// Avoid
func TestLogin(t *testing.T) {}
func TestAPI(t *testing.T) {}
```

### 3. Always Clean Up Resources
```go
func TestSomething(t *testing.T) {
    framework := gowright.NewWithDefaults()
    defer framework.Close() // Always cleanup
    
    // Your test code here
}
```

### 4. Use Assertions Effectively
```go
// Specific assertions provide better error messages
assert.Equal(t, http.StatusOK, response.StatusCode)
assert.Contains(t, response.Body, "success")
assert.NotNil(t, user.ID)

// Instead of generic assertions
assert.True(t, response.StatusCode == 200)
```

## Common Patterns

### Environment-Specific Testing
```go
func getConfigForEnvironment() *gowright.Config {
    env := os.Getenv("TEST_ENV")
    if env == "" {
        env = "development"
    }
    
    configFile := fmt.Sprintf("gowright-config.%s.json", env)
    config, err := gowright.LoadConfigFromFile(configFile)
    if err != nil {
        return gowright.DefaultConfig()
    }
    return config
}
```

### Parallel Test Execution
```go
func TestParallelExecution(t *testing.T) {
    config := &gowright.Config{
        Parallel: true,
        ParallelRunnerConfig: &gowright.ParallelRunnerConfig{
            MaxConcurrency: 4,
        },
    }
    
    framework := gowright.New(config)
    defer framework.Close()
    
    // Tests will run in parallel
}
```

### Custom Assertions
```go
func TestWithCustomAssertions(t *testing.T) {
    assertion := gowright.NewTestAssertion("Custom Validation")
    
    // Custom business logic validation
    assertion.Assert(user.Age >= 18, "User must be 18 or older")
    assertion.AssertContains(user.Permissions, "read", "User must have read permission")
    
    result := assertion.GetResult()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Next Steps

Now that you've created your first Gowright test, explore these areas:

1. **[Configuration](configuration.md)** - Learn about advanced configuration options
2. **[API Testing](../testing-modules/api-testing.md)** - Deep dive into API testing capabilities
3. **[UI Testing](../testing-modules/ui-testing.md)** - Master browser automation
4. **[Database Testing](../testing-modules/database-testing.md)** - Advanced database testing patterns
5. **[Integration Testing](../testing-modules/integration-testing.md)** - Complex workflow orchestration
6. **[Examples](../examples/basic-usage.md)** - More practical examples

## Troubleshooting

### Common Issues

**Chrome not found:**
```bash
# Set Chrome path explicitly
export CHROME_BIN=/usr/bin/google-chrome
```

**Database driver not found:**
```go
// Import the required driver
import _ "github.com/mattn/go-sqlite3"
```

**Permission denied for reports:**
```bash
# Create reports directory with proper permissions
mkdir -p ./test-reports
chmod 755 ./test-reports
```

### Getting Help

- 📖 [Full Documentation](../index.md)
- 🐛 [Issue Tracker](https://github.com/gowright/framework/issues)
- 💬 [Community Discussions](https://github.com/gowright/framework/discussions)
- 📧 [Email Support](mailto:support@gowright.dev)

Congratulations! You've successfully created and run your first Gowright test. The framework is now ready for more complex testing scenarios.