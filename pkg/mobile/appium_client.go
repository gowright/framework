package mobile

import (
	"bytes"
	"context"
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
func (c *AppiumClient) CreateSession(capabilities *AppiumCapabilities) error {
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
func (c *AppiumClient) DeleteSession() error {
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
func (c *AppiumClient) FindElement(using, value string) (*AppiumElement, error) {
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
func (c *AppiumClient) TakeScreenshot() (string, error) {
	if c.sessionID == "" {
		return "", core.NewGowrightError(core.BrowserError, "no active session", nil)
	}

	response, err := c.sendRequest("GET", fmt.Sprintf("/session/%s/screenshot", c.sessionID), nil)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to take screenshot", err)
	}

	screenshot, ok := response.Value.(string)
	if !ok {
		return "", core.NewGowrightError(core.BrowserError, "invalid screenshot response", nil)
	}

	return screenshot, nil
}

// GetPageSource gets the current page source
func (c *AppiumClient) GetPageSource() (string, error) {
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
func (c *AppiumClient) SetOrientation(orientation string) error {
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
func (e *AppiumElement) Click() error {
	_, err := e.client.sendRequest("POST", fmt.Sprintf("/session/%s/element/%s/click", e.client.sessionID, e.elementID), nil)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to click element", err)
	}
	return nil
}

// SendKeys sends text to the element
func (e *AppiumElement) SendKeys(text string) error {
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
func (e *AppiumElement) GetText() (string, error) {
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
func (e *AppiumElement) IsDisplayed() (bool, error) {
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
func (e *AppiumElement) IsEnabled() (bool, error) {
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
