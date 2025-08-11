# API Reference

Complete API reference for the Gowright testing framework.

## Core API

Documentation for the core testing framework API will be added here.

## Module APIs

- [UI Testing API](#ui-testing-api)
- [Mobile Testing API](#mobile-testing-api)
- [Database Testing API](#database-testing-api)
- [Integration Testing API](#integration-testing-api)
- [Assertion API](#assertion-api)
- [Reporting API](#reporting-api)

## Mobile Testing API

### AppiumClient

The main client for mobile automation using Appium WebDriver protocol.

#### Constructor

```go
func NewAppiumClient(serverURL string) *AppiumClient
```

Creates a new Appium client instance.

**Parameters:**
- `serverURL` - The Appium server URL (e.g., "http://localhost:4723")

**Returns:**
- `*AppiumClient` - New Appium client instance

#### Session Management

##### CreateSession

```go
func (c *AppiumClient) CreateSession(ctx context.Context, caps AppiumCapabilities) error
```

Creates a new Appium session with the specified capabilities.

**Parameters:**
- `ctx` - Context for request cancellation
- `caps` - Appium capabilities configuration

**Returns:**
- `error` - Error if session creation fails

##### DeleteSession

```go
func (c *AppiumClient) DeleteSession(ctx context.Context) error
```

Deletes the current Appium session.

**Parameters:**
- `ctx` - Context for request cancellation

**Returns:**
- `error` - Error if session deletion fails

##### GetSessionID

```go
func (c *AppiumClient) GetSessionID() string
```

Returns the current session ID.

**Returns:**
- `string` - Current session ID

#### Element Finding

##### FindElement

```go
func (c *AppiumClient) FindElement(ctx context.Context, by, value string) (*AppiumElement, error)
```

Finds a single element using the specified locator strategy.

**Parameters:**
- `ctx` - Context for request cancellation
- `by` - Locator strategy (e.g., ByID, ByXPath)
- `value` - Locator value

**Returns:**
- `*AppiumElement` - Found element
- `error` - Error if element not found

##### FindElements

```go
func (c *AppiumClient) FindElements(ctx context.Context, by, value string) ([]*AppiumElement, error)
```

Finds multiple elements using the specified locator strategy.

**Parameters:**
- `ctx` - Context for request cancellation
- `by` - Locator strategy
- `value` - Locator value

**Returns:**
- `[]*AppiumElement` - Array of found elements
- `error` - Error if elements not found

#### Wait Methods

##### WaitForElement

```go
func (c *AppiumClient) WaitForElement(ctx context.Context, by, value string, timeout time.Duration) (*AppiumElement, error)
```

Waits for an element to be present with a timeout.

**Parameters:**
- `ctx` - Context for request cancellation
- `by` - Locator strategy
- `value` - Locator value
- `timeout` - Maximum wait time

**Returns:**
- `*AppiumElement` - Found element
- `error` - Error if element not found within timeout

##### WaitForElementVisible

```go
func (c *AppiumClient) WaitForElementVisible(ctx context.Context, by, value string, timeout time.Duration) (*AppiumElement, error)
```

Waits for an element to be visible with a timeout.

##### WaitForElementClickable

```go
func (c *AppiumClient) WaitForElementClickable(ctx context.Context, by, value string, timeout time.Duration) (*AppiumElement, error)
```

Waits for an element to be clickable (visible and enabled) with a timeout.

#### Touch Actions

##### Tap

```go
func (c *AppiumClient) Tap(ctx context.Context, x, y int) error
```

Performs a tap action at the specified coordinates.

**Parameters:**
- `ctx` - Context for request cancellation
- `x` - X coordinate
- `y` - Y coordinate

**Returns:**
- `error` - Error if tap fails

##### Swipe

```go
func (c *AppiumClient) Swipe(ctx context.Context, startX, startY, endX, endY int, duration int) error
```

Performs a swipe gesture from start coordinates to end coordinates.

**Parameters:**
- `ctx` - Context for request cancellation
- `startX, startY` - Starting coordinates
- `endX, endY` - Ending coordinates
- `duration` - Swipe duration in milliseconds

**Returns:**
- `error` - Error if swipe fails

##### LongPress

```go
func (c *AppiumClient) LongPress(ctx context.Context, x, y int, duration int) error
```

Performs a long press action at the specified coordinates.

##### Pinch

```go
func (c *AppiumClient) Pinch(ctx context.Context, x, y int) error
```

Performs a pinch gesture (zoom out).

##### Zoom

```go
func (c *AppiumClient) Zoom(ctx context.Context, x, y int) error
```

Performs a zoom gesture (zoom in).

##### ScrollTo

```go
func (c *AppiumClient) ScrollTo(ctx context.Context, text string) error
```

Scrolls to an element with the specified text.

#### Device Management

##### GetOrientation

```go
func (c *AppiumClient) GetOrientation(ctx context.Context) (string, error)
```

Gets the current device orientation.

**Returns:**
- `string` - Current orientation ("PORTRAIT" or "LANDSCAPE")
- `error` - Error if operation fails

##### SetOrientation

```go
func (c *AppiumClient) SetOrientation(ctx context.Context, orientation string) error
```

Sets the device orientation.

**Parameters:**
- `orientation` - Target orientation ("PORTRAIT" or "LANDSCAPE")

##### GetWindowSize

```go
func (c *AppiumClient) GetWindowSize(ctx context.Context) (width, height int, err error)
```

Gets the current window size.

**Returns:**
- `width, height` - Window dimensions
- `error` - Error if operation fails

##### HideKeyboard

```go
func (c *AppiumClient) HideKeyboard(ctx context.Context) error
```

Hides the on-screen keyboard.

##### IsKeyboardShown

```go
func (c *AppiumClient) IsKeyboardShown(ctx context.Context) (bool, error)
```

Checks if the keyboard is currently displayed.

#### App Management

##### LaunchApp

```go
func (c *AppiumClient) LaunchApp(ctx context.Context) error
```

Launches an app on the device.

##### CloseApp

```go
func (c *AppiumClient) CloseApp(ctx context.Context) error
```

Closes the current app.

##### ResetApp

```go
func (c *AppiumClient) ResetApp(ctx context.Context) error
```

Resets the current app.

##### BackgroundApp

```go
func (c *AppiumClient) BackgroundApp(ctx context.Context, seconds int) error
```

Puts the app in background for the specified duration.

##### InstallApp

```go
func (c *AppiumClient) InstallApp(ctx context.Context, appPath string) error
```

Installs an app on the device.

##### RemoveApp

```go
func (c *AppiumClient) RemoveApp(ctx context.Context, appID string) error
```

Removes an app from the device.

##### IsAppInstalled

```go
func (c *AppiumClient) IsAppInstalled(ctx context.Context, appID string) (bool, error)
```

Checks if an app is installed on the device.

#### Android-Specific Methods

##### GetCurrentActivity

```go
func (c *AppiumClient) GetCurrentActivity(ctx context.Context) (string, error)
```

Gets the current activity (Android only).

##### GetCurrentPackage

```go
func (c *AppiumClient) GetCurrentPackage(ctx context.Context) (string, error)
```

Gets the current package (Android only).

##### StartActivity

```go
func (c *AppiumClient) StartActivity(ctx context.Context, appPackage, appActivity string) error
```

Starts a new activity (Android only).

#### Utility Methods

##### TakeScreenshot

```go
func (c *AppiumClient) TakeScreenshot(ctx context.Context) (string, error)
```

Captures a screenshot and returns it as base64 encoded string.

##### SaveScreenshot

```go
func (c *AppiumClient) SaveScreenshot(ctx context.Context, filePath string) error
```

Captures a screenshot and saves it to the specified file path.

##### GetPageSource

```go
func (c *AppiumClient) GetPageSource(ctx context.Context) (string, error)
```

Retrieves the current page source/XML hierarchy.

### AppiumElement

Represents a mobile element found by the AppiumClient.

#### Element Interactions

##### Click

```go
func (e *AppiumElement) Click(ctx context.Context) error
```

Performs a click action on the element.

##### SendKeys

```go
func (e *AppiumElement) SendKeys(ctx context.Context, text string) error
```

Sends text to the element.

##### Clear

```go
func (e *AppiumElement) Clear(ctx context.Context) error
```

Clears the text content of the element.

#### Element Properties

##### GetText

```go
func (e *AppiumElement) GetText(ctx context.Context) (string, error)
```

Retrieves the text content of the element.

##### GetAttribute

```go
func (e *AppiumElement) GetAttribute(ctx context.Context, name string) (string, error)
```

Retrieves the value of the specified attribute.

##### IsDisplayed

```go
func (e *AppiumElement) IsDisplayed(ctx context.Context) (bool, error)
```

Checks if the element is displayed.

##### IsEnabled

```go
func (e *AppiumElement) IsEnabled(ctx context.Context) (bool, error)
```

Checks if the element is enabled.

##### IsSelected

```go
func (e *AppiumElement) IsSelected(ctx context.Context) (bool, error)
```

Checks if the element is selected.

##### GetSize

```go
func (e *AppiumElement) GetSize(ctx context.Context) (width, height int, err error)
```

Retrieves the size of the element.

##### GetLocation

```go
func (e *AppiumElement) GetLocation(ctx context.Context) (x, y int, err error)
```

Retrieves the location of the element.

#### Child Element Finding

##### FindChildElement

```go
func (e *AppiumElement) FindChildElement(ctx context.Context, by, value string) (*AppiumElement, error)
```

Finds a child element within this element.

##### FindChildElements

```go
func (e *AppiumElement) FindChildElements(ctx context.Context, by, value string) ([]*AppiumElement, error)
```

Finds multiple child elements within this element.

### AppiumCapabilities

Configuration structure for Appium sessions.

```go
type AppiumCapabilities struct {
    PlatformName      string `json:"platformName"`
    PlatformVersion   string `json:"platformVersion,omitempty"`
    DeviceName        string `json:"deviceName"`
    App               string `json:"app,omitempty"`
    AppPackage        string `json:"appPackage,omitempty"`
    AppActivity       string `json:"appActivity,omitempty"`
    BundleID          string `json:"bundleId,omitempty"`
    AutomationName    string `json:"automationName,omitempty"`
    NoReset           bool   `json:"noReset,omitempty"`
    FullReset         bool   `json:"fullReset,omitempty"`
    NewCommandTimeout int    `json:"newCommandTimeout,omitempty"`
}
```

### Locator Constants

#### Basic Locators

```go
const (
    ByID                = "id"
    ByXPath             = "xpath"
    ByClassName         = "class name"
    ByName              = "name"
    ByTagName           = "tag name"
    ByLinkText          = "link text"
    ByPartialLinkText   = "partial link text"
    ByCSSSelector       = "css selector"
    ByAccessibilityID   = "accessibility id"
    ByAndroidUIAutomator = "-android uiautomator"
    ByIOSPredicate      = "-ios predicate string"
    ByIOSClassChain     = "-ios class chain"
    ByImage             = "-image"
    ByCustom            = "-custom"
)
```

#### Platform-Specific Locators

##### Android Locators

```go
// Resource ID
by, value := gowright.Android.ResourceID("com.example:id/button")

// Text content
by, value := gowright.Android.Text("Click me")

// Text contains
by, value := gowright.Android.TextContains("Click")

// Content description
by, value := gowright.Android.Description("Button description")

// Class name
by, value := gowright.Android.ClassName("android.widget.Button")
```

##### iOS Locators

```go
// Label
by, value := gowright.IOS.Label("Button Label")

// Name
by, value := gowright.IOS.Name("button-name")

// Value
by, value := gowright.IOS.Value("button-value")

// Element type
by, value := gowright.IOS.Type("XCUIElementTypeButton")

// Visible elements
by, value := gowright.IOS.Visible()
```