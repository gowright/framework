---
title: Quick Start
description: Get up and running with Gowright in minutes
---

This guide will help you get started with Gowright testing framework quickly.

## Prerequisites

- Go 1.22 or later
- Chrome/Chromium browser (for UI testing)
- Database server (optional, for database testing)

## Installation

Install Gowright using Go modules:

```bash
go get github/gowright/framework
```

## Your First Test

Create a simple test file to verify everything works:

```go
package main

import (
    "testing"
    "time"
    
    "github/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGowrightBasic(t *testing.T) {
    // Create framework with default configuration
    framework := gowright.NewWithDefaults()
    defer framework.Close()
    
    // Initialize the framework
    err := framework.Initialize()
    require.NoError(t, err)
    
    // Framework is ready to use
    assert.NotNil(t, framework)
}
```

Run the test:

```bash
go test -v
```

## API Testing Example

Here's a simple API test to get you started:

```go
func TestAPIExample(t *testing.T) {
    config := &gowright.APIConfig{
        BaseURL: "https://jsonplaceholder.typicode.com",
        Timeout: 10 * time.Second,
    }
    
    apiTester := gowright.NewAPITester(config)
    err := apiTester.Initialize(config)
    require.NoError(t, err)
    defer apiTester.Cleanup()
    
    // Test GET request
    response, err := apiTester.Get("/posts/1", nil)
    require.NoError(t, err)
    assert.Equal(t, 200, response.StatusCode)
}
```

## UI Testing Example

Basic web page interaction:

```go
func TestUIExample(t *testing.T) {
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    require.NoError(t, err)
    defer uiTester.Cleanup()
    
    // Navigate to page
    err = uiTester.Navigate("https://example.com")
    require.NoError(t, err)
    
    // Get page title
    title, err := uiTester.GetText("title")
    require.NoError(t, err)
    assert.Contains(t, title, "Example")
}
```

## Next Steps

Now that you have Gowright running, explore these areas:

- [Configuration](/configuration/) - Learn about all configuration options
- [API Testing](/testing/api/) - Deep dive into API testing capabilities
- [UI Testing](/testing/ui/) - Master browser automation
- [Database Testing](/testing/database/) - Test your database operations
- [Examples](/examples/) - See more practical examples

## Need Help?

- Check out our [troubleshooting guide](/guides/troubleshooting/)
- Browse the [examples section](/examples/)
- Ask questions in [GitHub Discussions](https://github.com/your-org/gowright/discussions)
- Report issues on [GitHub Issues](https://github.com/your-org/gowright/issues)