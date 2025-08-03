---
title: API Testing
description: Learn how to test HTTP/REST APIs with Gowright
---

Gowright provides comprehensive API testing capabilities through the `APITester` interface, built on top of the popular [go-resty/resty](https://github.com/go-resty/resty) HTTP client.

## Getting Started

### Basic Setup

```go
func TestAPIBasics(t *testing.T) {
    config := &gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    }
    
    apiTester := gowright.NewAPITester(config)
    err := apiTester.Initialize(config)
    require.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Your API tests here
}
```

### Configuration

The `APIConfig` struct provides various configuration options:

```go
type APIConfig struct {
    BaseURL    string            `json:"base_url,omitempty"`
    Timeout    time.Duration     `json:"timeout"`
    Headers    map[string]string `json:"headers,omitempty"`
    AuthConfig *AuthConfig       `json:"auth_config,omitempty"`
}
```

## HTTP Methods

### GET Requests

```go
response, err := apiTester.Get("/posts/1", nil)
require.NoError(t, err)
assert.Equal(t, http.StatusOK, response.StatusCode)
```

### POST Requests

```go
newPost := map[string]interface{}{
    "title":  "Test Post",
    "body":   "This is a test post",
    "userId": 1,
}

response, err := apiTester.Post("/posts", newPost, nil)
require.NoError(t, err)
assert.Equal(t, http.StatusCreated, response.StatusCode)
```

### PUT and DELETE

```go
// PUT request
updatedPost := map[string]interface{}{
    "id":     1,
    "title":  "Updated Post",
    "body":   "Updated content",
    "userId": 1,
}
response, err := apiTester.Put("/posts/1", updatedPost, nil)

// DELETE request
response, err := apiTester.Delete("/posts/1", nil)
```

## Authentication

### Bearer Token

```go
config := &gowright.APIConfig{
    BaseURL: "https://api.example.com",
    AuthConfig: &gowright.AuthConfig{
        Type:  "bearer",
        Token: "your-jwt-token",
    },
}
```

### Basic Authentication

```go
config := &gowright.APIConfig{
    BaseURL: "https://api.example.com",
    AuthConfig: &gowright.AuthConfig{
        Type:     "basic",
        Username: "your-username",
        Password: "your-password",
    },
}
```

## Response Validation

### Status Code Validation

```go
response, err := apiTester.Get("/posts/1", nil)
require.NoError(t, err)
assert.Equal(t, http.StatusOK, response.StatusCode)
```

### JSON Response Validation

```go
var post map[string]interface{}
err = json.Unmarshal(response.Body, &post)
require.NoError(t, err)

assert.Equal(t, float64(1), post["id"])
assert.NotEmpty(t, post["title"])
```

## Advanced Features

### Custom Headers

```go
headers := map[string]string{
    "Accept":       "application/json",
    "Content-Type": "application/json",
    "X-API-Key":    "your-api-key",
}

response, err := apiTester.Get("/protected-endpoint", headers)
```

### Test Builder Pattern

```go
test := gowright.NewAPITestBuilder("Get User Posts", "GET", "/users/1/posts").
    WithTester(apiTester).
    WithHeader("Accept", "application/json").
    ExpectStatus(http.StatusOK).
    ExpectJSONPath("$[0].userId", 1).
    Build()

result := test.Execute()
assert.Equal(t, gowright.TestStatusPassed, result.Status)
```

For more examples, see the [API Examples](/examples/api/) section.