# UI Testing with Rod

This package provides UI testing capabilities using the [rod](https://github.com/go-rod/rod) browser automation library for the Gowright testing framework.

## Features

- **Browser Automation**: Full browser automation using Chrome/Chromium via Chrome DevTools Protocol
- **Element Interactions**: Click, type, scroll, attribute access, and more
- **Assertions**: Text validation, element existence, visibility checks, attribute validation
- **Screenshots**: Capture full-page screenshots with automatic file management
- **Page Source**: Extract complete HTML source for analysis
- **JavaScript Execution**: Run custom JavaScript in the browser context
- **Wait Strategies**: Wait for elements, text content, or custom conditions
- **Configurable**: Headless/headed mode, window size, timeouts, user agents

## Dependencies

The UI testing module requires:
- `github.com/go-rod/rod v0.116.2` - Browser automation library
- Chrome/Chromium browser (automatically managed by rod if not present)

## Supported Browsers

- **Chrome/Chromium** (primary support) - Full feature support
- **Firefox** - Limited support, requires additional configuration

## Configuration

```go
browserConfig := &config.BrowserConfig{
    Browser:        "chrome",           // Browser type: "chrome", "chromium"
    Headless:       true,               // Run in headless mode
    WindowSize:     "1920x1080",        // Browser window size
    Timeout:        30 * time.Second,   // Default timeout for operations
    ScreenshotPath: "./screenshots",    // Directory for screenshots
    UserAgent:      "custom-agent",     // Custom user agent string
    DisableImages:  false,              // Disable image loading for faster tests
    DisableCSS:     false,              // Disable CSS loading
    DisableJS:      false,              // Disable JavaScript execution
    BrowserArgs:    []string{           // Custom browser arguments (pending implementation)
        "--no-sandbox", 
        "--disable-dev-shm-usage",
    },
}
```

### Default Chrome Arguments

The following Chrome arguments are automatically applied to improve the automation experience:
- `--no-default-browser-check` - Prevents default browser check dialog
- `--no-first-run` - Skips first run experience and setup wizard
- `--disable-fre` - Disables first run experience

## Basic Usage

```go
// Create and initialize tester
tester := ui.NewUITester()
err := tester.Initialize(browserConfig)
if err != nil {
    log.Fatal(err)
}
defer tester.Cleanup()

// Navigate to a page
err = tester.Navigate("https://example.com")

// Interact with elements
err = tester.Click("#button")
err = tester.Type("#input", "text")

// Get element text
text, err := tester.GetText("#element")

// Take screenshot
path, err := tester.TakeScreenshot("test")

// Wait for elements
err = tester.WaitForElement("#dynamic-element", 10*time.Second)
```

## Test Structure

```go
test := &core.UITest{
    Name: "Login Test",
    URL:  "https://example.com/login",
    Actions: []core.UIAction{
        {Type: "type", Selector: "#username", Value: "user"},
        {Type: "type", Selector: "#password", Value: "pass"},
        {Type: "click", Selector: "#login-btn"},
        {Type: "wait", Selector: ".dashboard"},
        {Type: "screenshot", Value: "after_login"},
    },
    Assertions: []core.UIAssertion{
        {Type: "element_exists", Selector: ".dashboard"},
        {Type: "text_contains", Selector: ".welcome", Expected: "Welcome"},
        {Type: "url_contains", Expected: "/dashboard"},
    },
}

result := tester.ExecuteTest(test)
```

## Supported Actions

| Action | Description | Parameters |
|--------|-------------|------------|
| `navigate` | Navigate to URL | `value`: URL to navigate to |
| `click` | Click element | `selector`: CSS selector of element |
| `type` | Type text into element | `selector`: CSS selector, `value`: text to type |
| `wait` | Wait for element or duration | `selector`: CSS selector (optional), `value`: duration string (optional) |
| `scroll` | Scroll to element | `selector`: CSS selector of element |
| `screenshot` | Take screenshot | `value`: filename (optional, auto-generated if empty) |

## Supported Assertions

| Assertion | Description | Parameters |
|-----------|-------------|------------|
| `text_equals` | Element text equals expected value | `selector`, `expected` |
| `text_contains` | Element text contains expected value | `selector`, `expected` |
| `element_exists` | Element exists in DOM | `selector` |
| `element_visible` | Element is visible on page | `selector` |
| `attribute_equals` | Element attribute equals expected value | `selector`, `attribute`, `expected` |
| `page_title_equals` | Page title equals expected value | `expected` |
| `url_contains` | Current URL contains expected string | `expected` |

## Advanced Features

### Custom JavaScript Execution

Execute custom JavaScript in the browser context:

```go
// Get page title
result, err := tester.ExecuteScript("return document.title;")

// Manipulate DOM
_, err = tester.ExecuteScript(`
    document.getElementById('myElement').style.backgroundColor = 'red';
    return 'Element highlighted';
`)

// Get complex data
data, err := tester.ExecuteScript(`
    return {
        url: window.location.href,
        userAgent: navigator.userAgent,
        cookies: document.cookie
    };
`)
```

### Element Attributes

Access and validate element attributes:

```go
// Get input value
value, err := tester.GetAttribute("#username", "value")

// Get element class
className, err := tester.GetAttribute(".button", "class")

// Get data attributes
dataValue, err := tester.GetAttribute("[data-id='123']", "data-value")
```

### Element Visibility and Interaction

```go
// Check if element is visible
visible, err := tester.IsElementVisible("#modal")

// Scroll element into view
err = tester.ScrollToElement("#bottom-element")

// Wait for specific text content
err = tester.WaitForText("#status", "Complete", 10*time.Second)
```

### Screenshot Management

```go
// Take screenshot with custom name
path, err := tester.TakeScreenshot("login_page")

// Screenshots are automatically saved as PNG files
// Path will be: "./screenshots/login_page.png" (if ScreenshotPath is configured)
```

### Cookie Notice Handling

Dismiss cookie notices and privacy banners programmatically:

```go
// Navigate to page
err = tester.Navigate("https://example.com")

// Wait for page to load
time.Sleep(2 * time.Second)

// Dismiss any cookie notices that appeared
err = tester.DismissCookieNotices()
if err != nil {
    log.Printf("Failed to dismiss cookies: %v", err)
}

// Continue with your test...
```

The `DismissCookieNotices()` method automatically:
- Finds and clicks "Accept", "Agree", "Allow" buttons
- Hides common cookie banner elements
- Removes overlay backgrounds
- Handles popular consent management platforms (OneTrust, TrustArc, etc.)

**Note**: Browser argument configuration is pending rod API integration. Currently, cookie handling relies on the JavaScript-based dismissal method.

## Error Handling

All methods return `*core.GowrightError` with specific error types:

- `core.BrowserError`: Browser automation errors
- `core.ConfigurationError`: Configuration issues

## Dependencies

- `github.com/go-rod/rod`: Browser automation library
- Chrome/Chromium browser installed on system

## Installation

Rod will automatically download and manage Chrome/Chromium if not found on the system.

## Examples

See `examples/ui-testing/` for complete examples.