package mobile

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// AppiumClient represents an Appium WebDriver client
type AppiumClient struct {
	serverURL  string
	sessionID  string
	httpClient *http.Client
}

// AppiumCapabilities represents the capabilities for an Appium session
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

// AppiumElement represents a mobile element
type AppiumElement struct {
	client    *AppiumClient
	elementID string
}

// AppiumResponse represents a standard Appium WebDriver response
type AppiumResponse struct {
	SessionID string      `json:"sessionId,omitempty"`
	Status    int         `json:"status"`
	Value     interface{} `json:"value"`
}

// NewAppiumClient creates a new Appium client
func NewAppiumClient(serverURL string) *AppiumClient {
	return &AppiumClient{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateSession creates a new Appium session
func (c *AppiumClient) CreateSession(ctx context.Context, capabilities *AppiumCapabilities) error {
	sessionData := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"alwaysMatch": capabilities,
		},
	}

	response, err := c.sendRequest("POST", "/session", sessionData)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to create Appium session", err)
	}

	if sessionID, ok := response.Value.(map[string]interface{})["sessionId"].(string); ok {
		c.sessionID = sessionID
	} else if response.SessionID != "" {
		c.sessionID = response.SessionID
	} else {
		return core.NewGowrightError(core.BrowserError, "failed to get session ID from response", nil)
	}

	return nil
}

// DeleteSession deletes the current Appium session
func (c *AppiumClient) DeleteSession(ctx context.Context) error {
	if c.sessionID == "" {
		return nil
	}

	_, err := c.sendRequest("DELETE", fmt.Sprintf("/session/%s", c.sessionID), nil)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to delete Appium session", err)
	}

	c.sessionID = ""
	return nil
}

// FindElement finds an element using the specified locator strategy
func (c *AppiumClient) FindElement(ctx context.Context, using, value string) (*AppiumElement, error) {
	if c.sessionID == "" {
		return nil, core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	data := map[string]string{
		"using": using,
		"value": value,
	}

	response, err := c.sendRequest("POST", fmt.Sprintf("/session/%s/element", c.sessionID), data)
	if err != nil {
		return nil, core.NewGowrightError(core.BrowserError, "failed to find element", err)
	}

	elementData, ok := response.Value.(map[string]interface{})
	if !ok {
		return nil, core.NewGowrightError(core.BrowserError, "invalid element response", nil)
	}

	var elementID string
	if id, exists := elementData["ELEMENT"]; exists {
		elementID = id.(string)
	} else if id, exists := elementData["element-6066-11e4-a52e-4f735466cecf"]; exists {
		elementID = id.(string)
	} else {
		return nil, core.NewGowrightError(core.BrowserError, "element ID not found in response", nil)
	}

	return &AppiumElement{
		client:    c,
		elementID: elementID,
	}, nil
}

// TakeScreenshot takes a screenshot and returns the base64 encoded image
func (c *AppiumClient) TakeScreenshot(ctx context.Context) ([]byte, error) {
	if c.sessionID == "" {
		return nil, core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	response, err := c.sendRequest("GET", fmt.Sprintf("/session/%s/screenshot", c.sessionID), nil)
	if err != nil {
		return nil, core.NewGowrightError(core.BrowserError, "failed to take screenshot", err)
	}

	screenshot, ok := response.Value.(string)
	if !ok {
		return nil, core.NewGowrightError(core.BrowserError, "invalid screenshot response", nil)
	}

	// Decode base64 to bytes
	data, err := base64.StdEncoding.DecodeString(screenshot)
	if err != nil {
		return nil, core.NewGowrightError(core.BrowserError, "failed to decode screenshot", err)
	}

	return data, nil
}

// GetPageSource gets the current page source
func (c *AppiumClient) GetPageSource(ctx context.Context) (string, error) {
	if c.sessionID == "" {
		return "", core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	response, err := c.sendRequest("GET", fmt.Sprintf("/session/%s/source", c.sessionID), nil)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to get page source", err)
	}

	source, ok := response.Value.(string)
	if !ok {
		return "", core.NewGowrightError(core.BrowserError, "invalid page source response", nil)
	}

	return source, nil
}

// SetOrientation sets the device orientation
func (c *AppiumClient) SetOrientation(ctx context.Context, orientation string) error {
	if c.sessionID == "" {
		return core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	data := map[string]string{
		"orientation": orientation,
	}

	_, err := c.sendRequest("POST", fmt.Sprintf("/session/%s/orientation", c.sessionID), data)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to set orientation", err)
	}

	return nil
}

// sendRequest sends an HTTP request to the Appium server
func (c *AppiumClient) sendRequest(method, path string, data interface{}) (*AppiumResponse, error) {
	url := c.serverURL + path

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error to avoid overriding main error
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var appiumResponse AppiumResponse
	if err := json.Unmarshal(responseBody, &appiumResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	return &appiumResponse, nil
}

// Element methods

// Click clicks on the element
func (e *AppiumElement) Click(ctx context.Context) error {
	_, err := e.client.sendRequest("POST", fmt.Sprintf("/session/%s/element/%s/click", e.client.sessionID, e.elementID), nil)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to click element", err)
	}
	return nil
}

// SendKeys sends text to the element
func (e *AppiumElement) SendKeys(ctx context.Context, text string) error {
	data := map[string]interface{}{
		"value": []string{text},
	}

	_, err := e.client.sendRequest("POST", fmt.Sprintf("/session/%s/element/%s/value", e.client.sessionID, e.elementID), data)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to send keys to element", err)
	}
	return nil
}

// GetText gets the text content of the element
func (e *AppiumElement) GetText(ctx context.Context) (string, error) {
	response, err := e.client.sendRequest("GET", fmt.Sprintf("/session/%s/element/%s/text", e.client.sessionID, e.elementID), nil)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to get element text", err)
	}

	text, ok := response.Value.(string)
	if !ok {
		return "", core.NewGowrightError(core.BrowserError, "invalid text response", nil)
	}

	return text, nil
}

// IsDisplayed checks if the element is displayed
func (e *AppiumElement) IsDisplayed(ctx context.Context) (bool, error) {
	response, err := e.client.sendRequest("GET", fmt.Sprintf("/session/%s/element/%s/displayed", e.client.sessionID, e.elementID), nil)
	if err != nil {
		return false, core.NewGowrightError(core.BrowserError, "failed to check if element is displayed", err)
	}

	displayed, ok := response.Value.(bool)
	if !ok {
		return false, core.NewGowrightError(core.BrowserError, "invalid displayed response", nil)
	}

	return displayed, nil
}

// IsEnabled checks if the element is enabled
func (e *AppiumElement) IsEnabled(ctx context.Context) (bool, error) {
	response, err := e.client.sendRequest("GET", fmt.Sprintf("/session/%s/element/%s/enabled", e.client.sessionID, e.elementID), nil)
	if err != nil {
		return false, core.NewGowrightError(core.BrowserError, "failed to check if element is enabled", err)
	}

	enabled, ok := response.Value.(bool)
	if !ok {
		return false, core.NewGowrightError(core.BrowserError, "invalid enabled response", nil)
	}

	return enabled, nil
}

// Additional methods used in examples

// GetSessionID returns the current session ID
func (c *AppiumClient) GetSessionID() string {
	return c.sessionID
}

// WaitForElement waits for an element to be present
func (c *AppiumClient) WaitForElement(ctx context.Context, using, value string, timeout time.Duration) (*AppiumElement, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		element, err := c.FindElement(ctx, using, value)
		if err == nil {
			return element, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, core.NewGowrightError(core.BrowserError, "element not found within timeout", nil)
}

// WaitForElementClickable waits for an element to be clickable (present and enabled)
func (c *AppiumClient) WaitForElementClickable(ctx context.Context, using, value string, timeout time.Duration) (*AppiumElement, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		element, err := c.FindElement(ctx, using, value)
		if err == nil {
			enabled, err := element.IsEnabled(ctx)
			if err == nil && enabled {
				return element, nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, core.NewGowrightError(core.BrowserError, "element not clickable within timeout", nil)
}

// FindElements finds multiple elements using the specified locator strategy
func (c *AppiumClient) FindElements(ctx context.Context, using, value string) ([]*AppiumElement, error) {
	if c.sessionID == "" {
		return nil, core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	data := map[string]string{
		"using": using,
		"value": value,
	}

	response, err := c.sendRequest("POST", fmt.Sprintf("/session/%s/elements", c.sessionID), data)
	if err != nil {
		return nil, core.NewGowrightError(core.BrowserError, "failed to find elements", err)
	}

	elementsData, ok := response.Value.([]interface{})
	if !ok {
		return nil, core.NewGowrightError(core.BrowserError, "invalid elements response", nil)
	}

	var elements []*AppiumElement
	for _, elementData := range elementsData {
		elementMap, ok := elementData.(map[string]interface{})
		if !ok {
			continue
		}

		var elementID string
		if id, exists := elementMap["ELEMENT"]; exists {
			elementID = id.(string)
		} else if id, exists := elementMap["element-6066-11e4-a52e-4f735466cecf"]; exists {
			elementID = id.(string)
		} else {
			continue
		}

		elements = append(elements, &AppiumElement{
			client:    c,
			elementID: elementID,
		})
	}

	return elements, nil
}

// GetWindowSize returns the current window size
func (c *AppiumClient) GetWindowSize(ctx context.Context) (int, int, error) {
	if c.sessionID == "" {
		return 0, 0, core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	response, err := c.sendRequest("GET", fmt.Sprintf("/session/%s/window/rect", c.sessionID), nil)
	if err != nil {
		return 0, 0, core.NewGowrightError(core.BrowserError, "failed to get window size", err)
	}

	sizeData, ok := response.Value.(map[string]interface{})
	if !ok {
		return 0, 0, core.NewGowrightError(core.BrowserError, "invalid window size response", nil)
	}

	width, _ := sizeData["width"].(float64)
	height, _ := sizeData["height"].(float64)

	return int(width), int(height), nil
}

// Swipe performs a swipe gesture
func (c *AppiumClient) Swipe(ctx context.Context, startX, startY, endX, endY int, duration int) error {
	if c.sessionID == "" {
		return core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	actions := map[string]interface{}{
		"actions": []map[string]interface{}{
			{
				"type": "pointer",
				"id":   "finger1",
				"parameters": map[string]interface{}{
					"pointerType": "touch",
				},
				"actions": []map[string]interface{}{
					{
						"type":     "pointerMove",
						"duration": 0,
						"x":        startX,
						"y":        startY,
					},
					{
						"type": "pointerDown",
					},
					{
						"type":     "pointerMove",
						"duration": duration,
						"x":        endX,
						"y":        endY,
					},
					{
						"type": "pointerUp",
					},
				},
			},
		},
	}

	_, err := c.sendRequest("POST", fmt.Sprintf("/session/%s/actions", c.sessionID), actions)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to perform swipe", err)
	}

	return nil
}

// StartActivity starts an Android activity
func (c *AppiumClient) StartActivity(ctx context.Context, appPackage, appActivity string) error {
	if c.sessionID == "" {
		return core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	data := map[string]string{
		"appPackage":  appPackage,
		"appActivity": appActivity,
	}

	_, err := c.sendRequest("POST", fmt.Sprintf("/session/%s/appium/device/start_activity", c.sessionID), data)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to start activity", err)
	}

	return nil
}

// GetOrientation gets the current device orientation
func (c *AppiumClient) GetOrientation(ctx context.Context) (string, error) {
	if c.sessionID == "" {
		return "", core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	response, err := c.sendRequest("GET", fmt.Sprintf("/session/%s/orientation", c.sessionID), nil)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to get orientation", err)
	}

	orientation, ok := response.Value.(string)
	if !ok {
		return "", core.NewGowrightError(core.BrowserError, "invalid orientation response", nil)
	}

	return orientation, nil
}
