---
title: UI Testing
description: Learn how to test web applications with browser automation
---

Gowright provides powerful UI testing capabilities through browser automation using [go-rod/rod](https://github.com/go-rod/rod), which uses the Chrome DevTools Protocol.

## Getting Started

### Basic Setup

```go
func TestUIBasics(t *testing.T) {
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    require.NoError(t, err)
    defer uiTester.Cleanup()
    
    // Your UI tests here
}
```

### Configuration

The `BrowserConfig` struct provides various browser options:

```go
type BrowserConfig struct {
    Headless   bool          `json:"headless"`
    Timeout    time.Duration `json:"timeout"`
    UserAgent  string        `json:"user_agent,omitempty"`
    WindowSize *WindowSize   `json:"window_size,omitempty"`
}
```

## Basic Operations

### Navigation

```go
err := uiTester.Navigate("https://example.com")
require.NoError(t, err)
```

### Element Interaction

```go
// Click an element
err := uiTester.Click("#submit-button")
require.NoError(t, err)

// Type text into an input
err := uiTester.Type("#username", "testuser")
require.NoError(t, err)

// Get text from an element
text, err := uiTester.GetText(".welcome-message")
require.NoError(t, err)
assert.Contains(t, text, "Welcome")
```

### Waiting for Elements

```go
// Wait for element to appear
err := uiTester.WaitForElement(".loading-complete", 10*time.Second)
require.NoError(t, err)
```

## Advanced Features

### Screenshots

```go
screenshotPath, err := uiTester.TakeScreenshot("test-result.png")
require.NoError(t, err)
assert.NotEmpty(t, screenshotPath)
```

### Mobile Device Emulation

```go
config := &gowright.BrowserConfig{
    Headless: true,
    DeviceConfig: &gowright.DeviceConfig{
        Name:        "iPhone 12",
        UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
        Width:       390,
        Height:      844,
        DeviceScale: 3.0,
        IsMobile:    true,
        HasTouch:    true,
    },
}
```

### Form Testing

```go
// Fill out a form
err := uiTester.Type("input[name='email']", "test@example.com")
require.NoError(t, err)

err := uiTester.Type("input[name='password']", "password123")
require.NoError(t, err)

err := uiTester.Click("button[type='submit']")
require.NoError(t, err)

// Wait for success message
err := uiTester.WaitForElement(".success-message", 5*time.Second)
require.NoError(t, err)
```

## Best Practices

### Page Object Pattern

```go
type LoginPage struct {
    tester gowright.UITester
}

func NewLoginPage(tester gowright.UITester) *LoginPage {
    return &LoginPage{tester: tester}
}

func (p *LoginPage) Login(username, password string) error {
    if err := p.tester.Type("#username", username); err != nil {
        return err
    }
    if err := p.tester.Type("#password", password); err != nil {
        return err
    }
    return p.tester.Click("#login-button")
}
```

### Error Handling

```go
// Take screenshot on failure
if err != nil {
    screenshotPath, _ := uiTester.TakeScreenshot("failure-screenshot.png")
    t.Logf("Test failed, screenshot saved to: %s", screenshotPath)
    t.Fatal(err)
}
```

For more examples, see the [UI Examples](/examples/ui/) section.