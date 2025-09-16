# UI Testing Quick Reference

## Setup

```go
import (
    "github.com/gowright/framework/pkg/ui"
    "github.com/gowright/framework/pkg/config"
    "github.com/gowright/framework/pkg/core"
)

// Create tester
tester := ui.NewUITester()

// Configure browser
browserConfig := &config.BrowserConfig{
    Browser:        "chrome",
    Headless:       true,
    WindowSize:     "1920x1080",
    Timeout:        30 * time.Second,
    ScreenshotPath: "./screenshots",
}

// Initialize and cleanup
err := tester.Initialize(browserConfig)
defer tester.Cleanup()
```

## Basic Operations

| Method | Usage | Example |
|--------|-------|---------|
| Navigate | `Navigate(url)` | `tester.Navigate("https://example.com")` |
| Click | `Click(selector)` | `tester.Click("#submit-btn")` |
| Type | `Type(selector, text)` | `tester.Type("#username", "admin")` |
| Get Text | `GetText(selector)` | `text, err := tester.GetText(".title")` |
| Wait | `WaitForElement(selector, timeout)` | `tester.WaitForElement(".loading", 10*time.Second)` |
| Screenshot | `TakeScreenshot(filename)` | `path, err := tester.TakeScreenshot("test")` |

## Advanced Operations

| Method | Usage | Example |
|--------|-------|---------|
| Get Attribute | `GetAttribute(selector, attr)` | `value, err := tester.GetAttribute("#input", "value")` |
| Check Visibility | `IsElementVisible(selector)` | `visible, err := tester.IsElementVisible("#modal")` |
| Scroll | `ScrollToElement(selector)` | `tester.ScrollToElement("#footer")` |
| Execute JS | `ExecuteScript(script)` | `result, err := tester.ExecuteScript("return document.title")` |
| Page Source | `GetPageSource()` | `html, err := tester.GetPageSource()` |

## Structured Testing

```go
test := &core.UITest{
    Name: "Login Flow",
    URL:  "https://app.example.com/login",
    Actions: []core.UIAction{
        {Type: "type", Selector: "#email", Value: "user@example.com"},
        {Type: "type", Selector: "#password", Value: "password123"},
        {Type: "click", Selector: "#login-button"},
        {Type: "wait", Selector: ".dashboard"},
        {Type: "screenshot", Value: "dashboard_loaded"},
    },
    Assertions: []core.UIAssertion{
        {Type: "element_exists", Selector: ".dashboard"},
        {Type: "text_contains", Selector: ".welcome", Expected: "Welcome"},
        {Type: "url_contains", Expected: "/dashboard"},
        {Type: "page_title_equals", Expected: "Dashboard - MyApp"},
    },
}

result := tester.ExecuteTest(test)
```

## Action Types

- `navigate` - Navigate to URL
- `click` - Click element
- `type` - Type text into element  
- `wait` - Wait for element or duration
- `scroll` - Scroll to element
- `screenshot` - Take screenshot

## Assertion Types

- `text_equals` - Element text equals value
- `text_contains` - Element text contains value
- `element_exists` - Element exists in DOM
- `element_visible` - Element is visible
- `attribute_equals` - Element attribute equals value
- `page_title_equals` - Page title equals value
- `url_contains` - URL contains string

## Configuration Options

```go
&config.BrowserConfig{
    Browser:        "chrome",           // Browser type
    Headless:       true,               // Headless mode
    WindowSize:     "1920x1080",        // Window dimensions
    Timeout:        30 * time.Second,   // Default timeout
    ScreenshotPath: "./screenshots",    // Screenshot directory
    UserAgent:      "custom-agent",     // Custom user agent
    DisableImages:  false,              // Disable images
    DisableCSS:     false,              // Disable CSS
    DisableJS:      false,              // Disable JavaScript
    BrowserArgs:    []string{           // Custom arguments
        "--no-sandbox",
        "--disable-dev-shm-usage",
    },
}
```

## Error Handling

All methods return `*core.GowrightError` with specific error types:
- `core.BrowserError` - Browser automation errors
- `core.ConfigurationError` - Configuration issues

```go
if err != nil {
    if gowrightErr, ok := err.(*core.GowrightError); ok {
        fmt.Printf("Error Type: %s, Message: %s\n", 
            gowrightErr.Type, gowrightErr.Message)
    }
}
```

## Best Practices

1. **Always cleanup**: Use `defer tester.Cleanup()`
2. **Use explicit waits**: Wait for elements before interacting
3. **Take screenshots**: Capture state for debugging
4. **Use structured tests**: Leverage `UITest` for complex flows
5. **Handle errors**: Check all return values
6. **Configure timeouts**: Set appropriate timeouts for your application