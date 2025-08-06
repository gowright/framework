# API Testing

Gowright provides comprehensive REST API testing capabilities with support for various HTTP methods, authentication, response validation, and detailed reporting.

## Overview

The API testing module is built on top of [go-resty/resty](https://github.com/go-resty/resty) and provides:

- HTTP method support (GET, POST, PUT, DELETE, PATCH, etc.)
- JSON and XML response validation
- JSONPath and XPath query support
- Authentication handling (Basic, Bearer, API Key)
- Request/response logging and debugging
- Performance metrics and timing
- Custom headers and middleware

## Basic Usage

### Simple API Test

```go
package main

import (
    "net/http"
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestBasicAPIRequest(t *testing.T) {
    // Create API tester
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Make GET request
    response, err := apiTester.Get("/posts/1", nil)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // Validate response body
    assert.Contains(t, string(response.Body), "userId")
}
```

### Using the Test Builder

The test builder provides a fluent API for creating structured tests:

```go
func TestWithBuilder(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Build and execute test
    test := gowright.NewAPITestBuilder("Get Post", "GET", "/posts/1").
        WithTester(apiTester).
        ExpectStatus(http.StatusOK).
        ExpectJSONPath("$.id", 1).
        ExpectJSONPath("$.userId", gowright.NotNil).
        ExpectHeader("Content-Type", "application/json; charset=utf-8").
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## HTTP Methods

### GET Requests

```go
func TestGetRequests(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Simple GET
    response, err := apiTester.Get("/get", nil)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // GET with query parameters
    params := map[string]string{
        "param1": "value1",
        "param2": "value2",
    }
    response, err = apiTester.Get("/get", params)
    assert.NoError(t, err)
    
    // GET with custom headers
    headers := map[string]string{
        "X-Custom-Header": "custom-value",
    }
    response, err = apiTester.GetWithHeaders("/get", params, headers)
    assert.NoError(t, err)
}
```

### POST Requests

```go
func TestPostRequests(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // POST with JSON body
    requestBody := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }
    
    response, err := apiTester.Post("/post", requestBody)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // POST with form data
    formData := map[string]string{
        "username": "johndoe",
        "password": "secret123",
    }
    
    response, err = apiTester.PostForm("/post", formData)
    assert.NoError(t, err)
    
    // POST with file upload
    response, err = apiTester.PostFile("/post", "file", "./test-file.txt")
    assert.NoError(t, err)
}
```

### PUT and PATCH Requests

```go
func TestPutPatchRequests(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // PUT request (full update)
    updateData := map[string]interface{}{
        "id":     1,
        "title":  "Updated Title",
        "body":   "Updated body content",
        "userId": 1,
    }
    
    response, err := apiTester.Put("/posts/1", updateData)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // PATCH request (partial update)
    patchData := map[string]interface{}{
        "title": "Partially Updated Title",
    }
    
    response, err = apiTester.Patch("/posts/1", patchData)
    assert.NoError(t, err)
}
```

### DELETE Requests

```go
func TestDeleteRequests(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // DELETE request
    response, err := apiTester.Delete("/posts/1")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
    
    // DELETE with query parameters
    params := map[string]string{
        "force": "true",
    }
    response, err = apiTester.DeleteWithParams("/posts/1", params)
    assert.NoError(t, err)
}
```

## Authentication

### Bearer Token Authentication

```go
func TestBearerAuth(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
        AuthConfig: &gowright.AuthConfig{
            Type:  "bearer",
            Token: "your-jwt-token-here",
        },
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Request will automatically include Authorization header
    response, err := apiTester.Get("/protected-endpoint", nil)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
}
```

### Basic Authentication

```go
func TestBasicAuth(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
        AuthConfig: &gowright.AuthConfig{
            Type:     "basic",
            Username: "user",
            Password: "pass",
        },
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    response, err := apiTester.Get("/basic-auth/user/pass", nil)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, response.StatusCode)
}
```

### API Key Authentication

```go
func TestAPIKeyAuth(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://api.example.com",
        Timeout: 10 * time.Second,
        AuthConfig: &gowright.AuthConfig{
            Type:   "api_key",
            APIKey: "your-api-key",
            Header: "X-API-Key", // Custom header name
        },
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    response, err := apiTester.Get("/api/data", nil)
    assert.NoError(t, err)
}
```

## Response Validation

### JSON Response Validation

```go
func TestJSONValidation(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    test := gowright.NewAPITestBuilder("Validate JSON Response", "GET", "/posts/1").
        WithTester(apiTester).
        ExpectStatus(http.StatusOK).
        ExpectJSONPath("$.id", 1).
        ExpectJSONPath("$.userId", gowright.GreaterThan(0)).
        ExpectJSONPath("$.title", gowright.NotEmpty).
        ExpectJSONPath("$.body", gowright.Contains("quia")).
        ExpectJSONSchema(`{
            "type": "object",
            "required": ["id", "title", "body", "userId"],
            "properties": {
                "id": {"type": "number"},
                "title": {"type": "string"},
                "body": {"type": "string"},
                "userId": {"type": "number"}
            }
        }`).
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Custom Validation Functions

```go
func TestCustomValidation(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    test := gowright.NewAPITestBuilder("Custom Validation", "GET", "/posts").
        WithTester(apiTester).
        ExpectStatus(http.StatusOK).
        WithCustomValidation(func(response *gowright.APIResponse) error {
            var posts []map[string]interface{}
            if err := json.Unmarshal(response.Body, &posts); err != nil {
                return err
            }
            
            if len(posts) == 0 {
                return errors.New("expected at least one post")
            }
            
            for _, post := range posts {
                if post["userId"] == nil {
                    return errors.New("userId is required for all posts")
                }
            }
            
            return nil
        }).
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

## Advanced Features

### Request/Response Logging

```go
func TestWithLogging(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL:    "https://httpbin.org",
        Timeout:    10 * time.Second,
        EnableLogs: true,
        LogLevel:   "debug",
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // This will log request and response details
    response, err := apiTester.Get("/get", nil)
    assert.NoError(t, err)
}
```

### Retry Configuration

```go
func TestWithRetry(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
        RetryConfig: &gowright.RetryConfig{
            MaxRetries:   3,
            InitialDelay: time.Second,
            MaxDelay:     10 * time.Second,
            Multiplier:   2.0,
        },
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // This will retry on failure
    response, err := apiTester.Get("/status/500", nil)
    // Will retry 3 times before failing
}
```

### Performance Testing

```go
func TestPerformance(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    test := gowright.NewAPITestBuilder("Performance Test", "GET", "/delay/1").
        WithTester(apiTester).
        ExpectStatus(http.StatusOK).
        ExpectResponseTime(2 * time.Second). // Max response time
        WithPerformanceMetrics(true).
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    
    // Access performance metrics
    metrics := result.PerformanceMetrics
    assert.True(t, metrics.ResponseTime < 2*time.Second)
    assert.True(t, metrics.DNSLookupTime > 0)
    assert.True(t, metrics.TCPConnectTime > 0)
}
```

## Test Suites

### Creating API Test Suites

```go
func TestAPITestSuite(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Create test suite
    testSuite := &gowright.TestSuite{
        Name: "JSONPlaceholder API Tests",
        SetupFunc: func() error {
            // Suite-level setup
            return nil
        },
        TeardownFunc: func() error {
            // Suite-level teardown
            return nil
        },
        Tests: []gowright.Test{
            gowright.NewAPITestBuilder("Get All Posts", "GET", "/posts").
                WithTester(apiTester).
                ExpectStatus(http.StatusOK).
                ExpectJSONPath("$", gowright.IsArray).
                ExpectJSONPath("$[0].id", gowright.NotNil).
                Build(),
            
            gowright.NewAPITestBuilder("Get Single Post", "GET", "/posts/1").
                WithTester(apiTester).
                ExpectStatus(http.StatusOK).
                ExpectJSONPath("$.id", 1).
                ExpectJSONPath("$.userId", gowright.GreaterThan(0)).
                Build(),
            
            gowright.NewAPITestBuilder("Create Post", "POST", "/posts").
                WithTester(apiTester).
                WithBody(map[string]interface{}{
                    "title":  "Test Post",
                    "body":   "Test body",
                    "userId": 1,
                }).
                ExpectStatus(http.StatusCreated).
                ExpectJSONPath("$.id", gowright.NotNil).
                Build(),
        },
    }
    
    // Execute test suite
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    framework.SetTestSuite(testSuite)
    results, err := framework.ExecuteTestSuite()
    assert.NoError(t, err)
    assert.Equal(t, 3, results.TotalTests)
    assert.Equal(t, 3, results.PassedTests)
}
```

## Error Handling

### Handling HTTP Errors

```go
func TestErrorHandling(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://httpbin.org",
        Timeout: 10 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Test 404 error
    test := gowright.NewAPITestBuilder("Test 404", "GET", "/status/404").
        WithTester(apiTester).
        ExpectStatus(http.StatusNotFound).
        ExpectJSONPath("$.error", gowright.NotNil).
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
    
    // Test server error
    test = gowright.NewAPITestBuilder("Test 500", "GET", "/status/500").
        WithTester(apiTester).
        ExpectStatus(http.StatusInternalServerError).
        WithErrorHandler(func(err error) bool {
            // Custom error handling logic
            return true // Continue with test
        }).
        Build()
    
    result = test.Execute()
    assert.Equal(t, gowright.TestStatusPassed, result.Status)
}
```

### Network Error Handling

```go
func TestNetworkErrors(t *testing.T) {
    apiTester := gowright.NewAPITester(&gowright.APIConfig{
        BaseURL: "https://invalid-domain-that-does-not-exist.com",
        Timeout: 5 * time.Second,
    })
    
    err := apiTester.Initialize(nil)
    assert.NoError(t, err)
    defer apiTester.Cleanup()
    
    // This will fail due to network error
    response, err := apiTester.Get("/test", nil)
    assert.Error(t, err)
    assert.Nil(t, response)
    
    // Test with error expectation
    test := gowright.NewAPITestBuilder("Network Error Test", "GET", "/test").
        WithTester(apiTester).
        ExpectError(true). // Expect this test to fail
        Build()
    
    result := test.Execute()
    assert.Equal(t, gowright.TestStatusFailed, result.Status)
    assert.NotNil(t, result.Error)
}
```

## Configuration Examples

### Complete API Configuration

```json
{
  "api_config": {
    "base_url": "https://api.example.com",
    "timeout": "30s",
    "headers": {
      "User-Agent": "Gowright-API-Tester/1.0",
      "Content-Type": "application/json",
      "Accept": "application/json"
    },
    "auth_config": {
      "type": "bearer",
      "token": "${API_TOKEN}"
    },
    "retry_config": {
      "max_retries": 3,
      "initial_delay": "1s",
      "max_delay": "10s",
      "multiplier": 2.0
    },
    "enable_logs": true,
    "log_level": "debug"
  }
}
```

### Environment-Specific Configuration

```go
func getAPIConfig() *gowright.APIConfig {
    env := os.Getenv("ENVIRONMENT")
    
    switch env {
    case "production":
        return &gowright.APIConfig{
            BaseURL: "https://api.production.com",
            Timeout: 30 * time.Second,
            AuthConfig: &gowright.AuthConfig{
                Type:  "bearer",
                Token: os.Getenv("PROD_API_TOKEN"),
            },
        }
    case "staging":
        return &gowright.APIConfig{
            BaseURL: "https://api.staging.com",
            Timeout: 20 * time.Second,
            AuthConfig: &gowright.AuthConfig{
                Type:  "bearer",
                Token: os.Getenv("STAGING_API_TOKEN"),
            },
        }
    default:
        return &gowright.APIConfig{
            BaseURL: "http://localhost:8080",
            Timeout: 10 * time.Second,
        }
    }
}
```

## Best Practices

### 1. Use Test Builders

Test builders provide better structure and readability:

```go
// Good
test := gowright.NewAPITestBuilder("User Creation", "POST", "/users").
    WithBody(userData).
    ExpectStatus(http.StatusCreated).
    ExpectJSONPath("$.id", gowright.NotNil).
    Build()

// Avoid
response, err := apiTester.Post("/users", userData)
// Manual validation...
```

### 2. Validate Response Structure

Always validate the structure of responses:

```go
test := gowright.NewAPITestBuilder("Get User", "GET", "/users/1").
    ExpectJSONSchema(`{
        "type": "object",
        "required": ["id", "name", "email"],
        "properties": {
            "id": {"type": "number"},
            "name": {"type": "string"},
            "email": {"type": "string", "format": "email"}
        }
    }`).
    Build()
```

### 3. Use Environment Variables

Keep sensitive data in environment variables:

```go
config := &gowright.APIConfig{
    BaseURL: os.Getenv("API_BASE_URL"),
    AuthConfig: &gowright.AuthConfig{
        Type:  "bearer",
        Token: os.Getenv("API_TOKEN"),
    },
}
```

### 4. Test Error Scenarios

Always test error conditions:

```go
// Test validation errors
test := gowright.NewAPITestBuilder("Invalid Data", "POST", "/users").
    WithBody(map[string]interface{}{
        "name": "", // Invalid empty name
    }).
    ExpectStatus(http.StatusBadRequest).
    ExpectJSONPath("$.errors", gowright.NotEmpty).
    Build()
```

### 5. Use Descriptive Test Names

Make test names descriptive and specific:

```go
// Good
func TestUserCanCreateAccountWithValidData(t *testing.T) {}
func TestAPIReturnsValidationErrorForMissingEmail(t *testing.T) {}

// Avoid
func TestCreateUser(t *testing.T) {}
func TestValidation(t *testing.T) {}
```

## Next Steps

- [UI Testing](ui-testing.md) - Learn about browser automation
- [Database Testing](database-testing.md) - Validate data persistence
- [Integration Testing](integration-testing.md) - Combine API with other modules
- [Examples](../examples/api-testing.md) - More API testing examples