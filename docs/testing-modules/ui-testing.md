# UI Testing

Gowright provides powerful browser automation capabilities for testing web applications. Built on top of [go-rod/rod](https://github.com/go-rod/rod), it offers fast, reliable browser automation using Chrome DevTools Protocol.

## Overview

The UI testing module provides:

- Browser automation using Chrome DevTools Protocol
- Cross-platform support (Chrome, Chromium, Edge)
- Mobile device emulation
- Screenshot capture and visual validation
- Element interaction and form handling
- Page navigation and waiting strategies
- JavaScript execution and evaluation

## Basic Usage

### Simple UI Test

```go
package main

import (
    "testing"
    "time"
    
    "github.com/gowright/framework/pkg/gowright"
    "github.com/stretchr/testify/assert"
)

func TestBasicUIInteraction(t *testing.T) {
    // Create UI tester
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
        WindowSize: &gowright.WindowSize{
            Width:  1920,
            Height: 1080,
        },
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    assert.NoError(t, err)
    defer uiTester.Cleanup()
    
    // Navigate to page
    err = uiTester.Navigate("https://example.com")
    assert.NoError(t, err)
    
    // Interact with elements
    err = uiTester.Click("a[href='/about']")
    assert.NoError(t, err)
    
    // Verify page content
    text, err := uiTester.GetText("h1")
    assert.NoError(t, err)
    assert.Contains(t, text, "About")
}
```

## Browser Configuration

### Headless vs Headed Mode

```go
// Headless mode (default for CI/CD)
config := &gowright.BrowserConfig{
    Headless: true,
    Timeout:  30 * time.Second,
}

// Headed mode (useful for debugging)
config := &gowright.BrowserConfig{
    Headless: false,
    Timeout:  60 * time.Second,
}
```

### Window Size and Viewport

```go
config := &gowright.BrowserConfig{
    Headless: true,
    WindowSize: &gowright.WindowSize{
        Width:  1920,
        Height: 1080,
    },
    // Mobile viewport simulation
    DeviceEmulation: &gowright.DeviceEmulation{
        Width:  375,
        Height: 667,
        DeviceScaleFactor: 2,
        Mobile: true,
        UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
    },
}
```

### Chrome Arguments

```go
config := &gowright.BrowserConfig{
    Headless: true,
    ChromeArgs: []string{
        "--no-sandbox",
        "--disable-dev-shm-usage",
        "--disable-gpu",
        "--disable-web-security",
        "--allow-running-insecure-content",
    },
}
```

## Element Interaction

### Finding Elements

```go
func TestElementFinding(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // By CSS selector
    element, err := uiTester.FindElement("input[name='username']")
    assert.NoError(t, err)
    
    // By XPath
    element, err = uiTester.FindElementByXPath("//input[@placeholder='Enter username']")
    assert.NoError(t, err)
    
    // By text content
    element, err = uiTester.FindElementByText("Login")
    assert.NoError(t, err)
    
    // Multiple elements
    elements, err := uiTester.FindElements("div.card")
    assert.NoError(t, err)
    assert.True(t, len(elements) > 0)
}
```

### Clicking Elements

```go
func TestClicking(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Simple click
    err := uiTester.Click("button#submit")
    assert.NoError(t, err)
    
    // Double click
    err = uiTester.DoubleClick("div.item")
    assert.NoError(t, err)
    
    // Right click
    err = uiTester.RightClick("div.context-menu-trigger")
    assert.NoError(t, err)
    
    // Click at coordinates
    err = uiTester.ClickAt(100, 200)
    assert.NoError(t, err)
}
```

### Text Input

```go
func TestTextInput(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Type text
    err := uiTester.Type("input[name='username']", "testuser")
    assert.NoError(t, err)
    
    // Clear and type
    err = uiTester.ClearAndType("input[name='password']", "newpassword")
    assert.NoError(t, err)
    
    // Type with delay (simulate human typing)
    err = uiTester.TypeWithDelay("textarea", "Long text content", 50*time.Millisecond)
    assert.NoError(t, err)
    
    // Press keys
    err = uiTester.PressKey("Enter")
    assert.NoError(t, err)
    
    err = uiTester.PressKeys("Ctrl", "A") // Select all
    assert.NoError(t, err)
}
```

### Form Handling

```go
func TestFormHandling(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Fill form data
    formData := map[string]string{
        "input[name='firstName']": "John",
        "input[name='lastName']":  "Doe",
        "input[name='email']":     "john@example.com",
        "select[name='country']":  "US",
    }
    
    err := uiTester.FillForm(formData)
    assert.NoError(t, err)
    
    // Select dropdown option
    err = uiTester.SelectOption("select[name='state']", "California")
    assert.NoError(t, err)
    
    // Check/uncheck checkboxes
    err = uiTester.Check("input[type='checkbox'][name='newsletter']")
    assert.NoError(t, err)
    
    err = uiTester.Uncheck("input[type='checkbox'][name='marketing']")
    assert.NoError(t, err)
    
    // Select radio button
    err = uiTester.SelectRadio("input[name='gender'][value='male']")
    assert.NoError(t, err)
}
```

## Waiting Strategies

### Wait for Elements

```go
func TestWaiting(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Wait for element to be visible
    err := uiTester.WaitForElement("div.loading", 10*time.Second)
    assert.NoError(t, err)
    
    // Wait for element to disappear
    err = uiTester.WaitForElementToDisappear("div.loading", 30*time.Second)
    assert.NoError(t, err)
    
    // Wait for text to appear
    err = uiTester.WaitForText("Success!", 15*time.Second)
    assert.NoError(t, err)
    
    // Wait for page to load
    err = uiTester.WaitForPageLoad(20*time.Second)
    assert.NoError(t, err)
    
    // Custom wait condition
    err = uiTester.WaitForCondition(func() bool {
        elements, _ := uiTester.FindElements("div.item")
        return len(elements) >= 5
    }, 30*time.Second)
    assert.NoError(t, err)
}
```

### JavaScript Execution

```go
func TestJavaScript(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Execute JavaScript
    result, err := uiTester.ExecuteJS("return document.title")
    assert.NoError(t, err)
    title := result.(string)
    assert.NotEmpty(t, title)
    
    // Execute with arguments
    result, err = uiTester.ExecuteJSWithArgs(
        "return arguments[0] + arguments[1]",
        10, 20,
    )
    assert.NoError(t, err)
    assert.Equal(t, 30, int(result.(float64)))
    
    // Scroll to element
    err = uiTester.ScrollToElement("footer")
    assert.NoError(t, err)
    
    // Scroll by pixels
    err = uiTester.ScrollBy(0, 500)
    assert.NoError(t, err)
}
```

## Screenshots and Visual Testing

### Taking Screenshots

```go
func TestScreenshots(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Full page screenshot
    screenshotPath, err := uiTester.TakeScreenshot("full-page.png")
    assert.NoError(t, err)
    assert.FileExists(t, screenshotPath)
    
    // Element screenshot
    screenshotPath, err = uiTester.TakeElementScreenshot("div.header", "header.png")
    assert.NoError(t, err)
    
    // Screenshot with custom options
    options := &gowright.ScreenshotOptions{
        Quality:    90,
        FullPage:   true,
        OmitBackground: false,
    }
    screenshotPath, err = uiTester.TakeScreenshotWithOptions("custom.png", options)
    assert.NoError(t, err)
}
```

### Visual Comparison

```go
func TestVisualComparison(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Take baseline screenshot
    baselinePath, err := uiTester.TakeScreenshot("baseline.png")
    assert.NoError(t, err)
    
    // Make some changes to the page
    err = uiTester.Click("button.toggle-theme")
    assert.NoError(t, err)
    
    // Take comparison screenshot
    comparisonPath, err := uiTester.TakeScreenshot("comparison.png")
    assert.NoError(t, err)
    
    // Compare images
    similarity, err := uiTester.CompareImages(baselinePath, comparisonPath)
    assert.NoError(t, err)
    assert.True(t, similarity < 0.95) // Expect some difference
}
```

## Mobile Testing

### Device Emulation

```go
func TestMobileEmulation(t *testing.T) {
    // iPhone 12 Pro emulation
    config := &gowright.BrowserConfig{
        Headless: true,
        DeviceEmulation: &gowright.DeviceEmulation{
            Width:  390,
            Height: 844,
            DeviceScaleFactor: 3,
            Mobile: true,
            UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15",
        },
    }
    
    uiTester := gowright.NewRodUITester()
    err := uiTester.Initialize(config)
    assert.NoError(t, err)
    defer uiTester.Cleanup()
    
    err = uiTester.Navigate("https://example.com")
    assert.NoError(t, err)
    
    // Test mobile-specific interactions
    err = uiTester.TouchTap("button.mobile-menu")
    assert.NoError(t, err)
    
    err = uiTester.Swipe(100, 300, 100, 100) // Swipe up
    assert.NoError(t, err)
}
```

### Responsive Testing

```go
func TestResponsiveDesign(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    
    viewports := []gowright.WindowSize{
        {Width: 320, Height: 568},   // Mobile
        {Width: 768, Height: 1024},  // Tablet
        {Width: 1920, Height: 1080}, // Desktop
    }
    
    for _, viewport := range viewports {
        config := &gowright.BrowserConfig{
            Headless:   true,
            WindowSize: &viewport,
        }
        
        err := uiTester.Initialize(config)
        assert.NoError(t, err)
        
        err = uiTester.Navigate("https://example.com")
        assert.NoError(t, err)
        
        // Test layout at this viewport
        isVisible, err := uiTester.IsElementVisible("nav.mobile-menu")
        assert.NoError(t, err)
        
        if viewport.Width < 768 {
            assert.True(t, isVisible, "Mobile menu should be visible on small screens")
        } else {
            assert.False(t, isVisible, "Mobile menu should be hidden on large screens")
        }
        
        uiTester.Cleanup()
    }
}
```

## Advanced Features

### File Upload

```go
func TestFileUpload(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Upload single file
    err := uiTester.UploadFile("input[type='file']", "./test-file.txt")
    assert.NoError(t, err)
    
    // Upload multiple files
    files := []string{"./file1.txt", "./file2.txt", "./file3.txt"}
    err = uiTester.UploadFiles("input[type='file'][multiple]", files)
    assert.NoError(t, err)
}
```

### Drag and Drop

```go
func TestDragAndDrop(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Drag and drop between elements
    err := uiTester.DragAndDrop("div.draggable", "div.dropzone")
    assert.NoError(t, err)
    
    // Drag to coordinates
    err = uiTester.DragToCoordinates("div.draggable", 300, 400)
    assert.NoError(t, err)
}
```

### Browser Navigation

```go
func TestNavigation(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    // Navigate to URL
    err := uiTester.Navigate("https://example.com")
    assert.NoError(t, err)
    
    // Navigate to another page
    err = uiTester.Navigate("https://example.com/about")
    assert.NoError(t, err)
    
    // Go back
    err = uiTester.GoBack()
    assert.NoError(t, err)
    
    // Go forward
    err = uiTester.GoForward()
    assert.NoError(t, err)
    
    // Refresh page
    err = uiTester.Refresh()
    assert.NoError(t, err)
    
    // Get current URL
    currentURL, err := uiTester.GetCurrentURL()
    assert.NoError(t, err)
    assert.Contains(t, currentURL, "example.com")
}
```

## Test Patterns

### Page Object Model

```go
// Define page objects
type LoginPage struct {
    uiTester *gowright.RodUITester
}

func NewLoginPage(uiTester *gowright.RodUITester) *LoginPage {
    return &LoginPage{uiTester: uiTester}
}

func (p *LoginPage) Navigate() error {
    return p.uiTester.Navigate("https://example.com/login")
}

func (p *LoginPage) Login(username, password string) error {
    if err := p.uiTester.Type("input[name='username']", username); err != nil {
        return err
    }
    if err := p.uiTester.Type("input[name='password']", password); err != nil {
        return err
    }
    return p.uiTester.Click("button[type='submit']")
}

func (p *LoginPage) IsLoggedIn() (bool, error) {
    return p.uiTester.IsElementVisible("div.user-menu")
}

// Use in tests
func TestLoginFlow(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    loginPage := NewLoginPage(uiTester)
    
    err := loginPage.Navigate()
    assert.NoError(t, err)
    
    err = loginPage.Login("testuser", "password123")
    assert.NoError(t, err)
    
    isLoggedIn, err := loginPage.IsLoggedIn()
    assert.NoError(t, err)
    assert.True(t, isLoggedIn)
}
```

### Data-Driven Testing

```go
func TestLoginWithMultipleCredentials(t *testing.T) {
    testCases := []struct {
        name     string
        username string
        password string
        expected bool
    }{
        {"Valid credentials", "user1", "pass1", true},
        {"Invalid username", "invalid", "pass1", false},
        {"Invalid password", "user1", "invalid", false},
        {"Empty credentials", "", "", false},
    }
    
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            loginPage := NewLoginPage(uiTester)
            
            err := loginPage.Navigate()
            assert.NoError(t, err)
            
            err = loginPage.Login(tc.username, tc.password)
            assert.NoError(t, err)
            
            isLoggedIn, err := loginPage.IsLoggedIn()
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, isLoggedIn)
        })
    }
}
```

## Configuration Examples

### Complete UI Configuration

```json
{
  "browser_config": {
    "headless": true,
    "timeout": "30s",
    "user_agent": "Gowright-UI-Tester/1.0",
    "window_size": {
      "width": 1920,
      "height": 1080
    },
    "chrome_path": "/usr/bin/google-chrome",
    "chrome_args": [
      "--no-sandbox",
      "--disable-dev-shm-usage",
      "--disable-gpu",
      "--disable-web-security"
    ],
    "device_emulation": {
      "width": 375,
      "height": 667,
      "device_scale_factor": 2,
      "mobile": true,
      "user_agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"
    }
  }
}
```

## Best Practices

### 1. Use Explicit Waits

```go
// Good - wait for specific condition
err := uiTester.WaitForElement("div.content", 10*time.Second)

// Avoid - arbitrary sleep
time.Sleep(5 * time.Second)
```

### 2. Use Stable Selectors

```go
// Good - stable selectors
err := uiTester.Click("button[data-testid='submit-button']")
err := uiTester.Click("input[name='username']")

// Avoid - fragile selectors
err := uiTester.Click("div > div:nth-child(3) > button")
```

### 3. Handle Dynamic Content

```go
// Wait for dynamic content to load
err := uiTester.WaitForCondition(func() bool {
    elements, _ := uiTester.FindElements("div.item")
    return len(elements) > 0
}, 30*time.Second)
```

### 4. Take Screenshots on Failure

```go
func TestWithScreenshotOnFailure(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    // ... initialization
    
    defer func() {
        if t.Failed() {
            screenshotPath, _ := uiTester.TakeScreenshot("failure.png")
            t.Logf("Screenshot saved: %s", screenshotPath)
        }
    }()
    
    // Your test code here
}
```

### 5. Clean Up Resources

```go
func TestWithProperCleanup(t *testing.T) {
    uiTester := gowright.NewRodUITester()
    
    config := &gowright.BrowserConfig{
        Headless: true,
        Timeout:  30 * time.Second,
    }
    
    err := uiTester.Initialize(config)
    assert.NoError(t, err)
    defer uiTester.Cleanup() // Always cleanup
    
    // Your test code here
}
```

## Troubleshooting

### Common Issues

**Chrome not found:**
```bash
# Set Chrome path explicitly
export CHROME_BIN=/usr/bin/google-chrome
```

**Element not found:**
```go
// Add explicit wait
err := uiTester.WaitForElement("selector", 10*time.Second)
if err != nil {
    // Element might not exist or selector is wrong
}
```

**Timeout errors:**
```go
// Increase timeout for slow operations
config := &gowright.BrowserConfig{
    Timeout: 60 * time.Second, // Longer timeout
}
```

**Headless mode issues:**
```go
// Use headed mode for debugging
config := &gowright.BrowserConfig{
    Headless: false, // See what's happening
}
```

## Next Steps

- [Database Testing](database-testing.md) - Validate data persistence
- [Integration Testing](integration-testing.md) - Combine UI with other modules
- [Examples](../examples/ui-testing.md) - More UI testing examples
- [Best Practices](../reference/best-practices.md) - UI testing best practices