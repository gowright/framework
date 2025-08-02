package gowright

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// RodUITester implements the UITester interface using go-rod/rod
type RodUITester struct {
	browser *rod.Browser
	page    *rod.Page
	config  *BrowserConfig
	name    string
}

// NewRodUITester creates a new RodUITester instance
func NewRodUITester(config *BrowserConfig) *RodUITester {
	if config == nil {
		config = &BrowserConfig{
			Headless: true,
			Timeout:  30 * time.Second,
			WindowSize: &WindowSize{
				Width:  1920,
				Height: 1080,
			},
		}
	}
	
	return &RodUITester{
		config: config,
		name:   "RodUITester",
	}
}

// Initialize sets up the browser with the provided configuration
func (r *RodUITester) Initialize(config interface{}) error {
	if config != nil {
		if browserConfig, ok := config.(*BrowserConfig); ok {
			r.config = browserConfig
		}
	}

	// Create launcher with configuration
	launcher := launcher.New()
	
	if r.config.Headless {
		launcher = launcher.Headless(true)
	} else {
		launcher = launcher.Headless(false)
	}
	
	if r.config.UserAgent != "" {
		launcher = launcher.UserDataDir("")
	}

	// Launch browser
	url, err := launcher.Launch()
	if err != nil {
		return NewGowrightError(BrowserError, "failed to launch browser", err)
	}

	// Connect to browser
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return NewGowrightError(BrowserError, "failed to connect to browser", err)
	}

	r.browser = browser

	// Create initial page
	page, err := r.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return NewGowrightError(BrowserError, "failed to create page", err)
	}

	r.page = page

	// Set window size if specified
	if r.config.WindowSize != nil {
		err = r.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:  r.config.WindowSize.Width,
			Height: r.config.WindowSize.Height,
		})
		if err != nil {
			return NewGowrightError(BrowserError, "failed to set window size", err)
		}
	}

	// Set user agent if specified
	if r.config.UserAgent != "" {
		err = r.page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: r.config.UserAgent,
		})
		if err != nil {
			return NewGowrightError(BrowserError, "failed to set user agent", err)
		}
	}

	return nil
}

// Cleanup performs any necessary cleanup operations
func (r *RodUITester) Cleanup() error {
	var errors []error

	if r.page != nil {
		if err := r.page.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close page: %w", err))
		}
	}

	if r.browser != nil {
		if err := r.browser.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close browser: %w", err))
		}
	}

	if len(errors) > 0 {
		return NewGowrightError(BrowserError, "cleanup failed", fmt.Errorf("multiple errors: %v", errors))
	}

	return nil
}

// GetName returns the name of the tester
func (r *RodUITester) GetName() string {
	return r.name
}

// Navigate navigates to the specified URL
func (r *RodUITester) Navigate(url string) error {
	if r.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	err := r.page.Context(ctx).Navigate(url)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to navigate to %s", url), err)
	}

	// Wait for page to load
	err = r.page.Context(ctx).WaitLoad()
	if err != nil {
		return NewGowrightError(BrowserError, "failed to wait for page load", err)
	}

	return nil
}

// Click clicks on an element identified by the selector
func (r *RodUITester) Click(selector string) error {
	if r.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	element, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	err = element.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to click element with selector %s", selector), err)
	}

	return nil
}

// Type types text into an element identified by the selector
func (r *RodUITester) Type(selector, text string) error {
	if r.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	element, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	// Clear existing text first
	err = element.SelectAllText()
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to select text in element with selector %s", selector), err)
	}

	// Type the new text
	err = element.Input(text)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to type text into element with selector %s", selector), err)
	}

	return nil
}

// GetText retrieves text from an element identified by the selector
func (r *RodUITester) GetText(selector string) (string, error) {
	if r.page == nil {
		return "", NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	element, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		return "", NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	text, err := element.Text()
	if err != nil {
		return "", NewGowrightError(BrowserError, fmt.Sprintf("failed to get text from element with selector %s", selector), err)
	}

	return text, nil
}

// WaitForElement waits for an element to be present
func (r *RodUITester) WaitForElement(selector string, timeout time.Duration) error {
	if r.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("element with selector %s not found within timeout", selector), err)
	}

	return nil
}

// TakeScreenshot captures a screenshot and returns the file path
func (r *RodUITester) TakeScreenshot(filename string) (string, error) {
	if r.page == nil {
		return "", NewGowrightError(BrowserError, "page not initialized", nil)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", NewGowrightError(BrowserError, "failed to create screenshot directory", err)
	}

	// Take screenshot
	screenshot, err := r.page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	if err != nil {
		return "", NewGowrightError(BrowserError, "failed to capture screenshot", err)
	}

	// Write to file
	err = os.WriteFile(filename, screenshot, 0644)
	if err != nil {
		return "", NewGowrightError(BrowserError, "failed to save screenshot", err)
	}

	return filename, nil
}

// GetPageSource returns the current page source
func (r *RodUITester) GetPageSource() (string, error) {
	if r.page == nil {
		return "", NewGowrightError(BrowserError, "page not initialized", nil)
	}

	html, err := r.page.HTML()
	if err != nil {
		return "", NewGowrightError(BrowserError, "failed to get page source", err)
	}

	return html, nil
}

// Additional helper methods for browser lifecycle management

// NewPage creates a new page in the browser
func (r *RodUITester) NewPage() error {
	if r.browser == nil {
		return NewGowrightError(BrowserError, "browser not initialized", nil)
	}

	page, err := r.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return NewGowrightError(BrowserError, "failed to create new page", err)
	}

	// Close old page if exists
	if r.page != nil {
		r.page.Close()
	}

	r.page = page

	// Apply configuration to new page
	if r.config.WindowSize != nil {
		err = r.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:  r.config.WindowSize.Width,
			Height: r.config.WindowSize.Height,
		})
		if err != nil {
			return NewGowrightError(BrowserError, "failed to set window size on new page", err)
		}
	}

	if r.config.UserAgent != "" {
		err = r.page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: r.config.UserAgent,
		})
		if err != nil {
			return NewGowrightError(BrowserError, "failed to set user agent on new page", err)
		}
	}

	return nil
}

// GetCurrentURL returns the current page URL
func (r *RodUITester) GetCurrentURL() (string, error) {
	if r.page == nil {
		return "", NewGowrightError(BrowserError, "page not initialized", nil)
	}

	info, err := r.page.Info()
	if err != nil {
		return "", NewGowrightError(BrowserError, "failed to get page info", err)
	}

	return info.URL, nil
}

// IsElementPresent checks if an element is present on the page
func (r *RodUITester) IsElementPresent(selector string) (bool, error) {
	if r.page == nil {
		return false, NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		// Element not found is not an error in this context
		return false, nil
	}

	return true, nil
}

// IsElementVisible checks if an element is visible on the page
func (r *RodUITester) IsElementVisible(selector string) (bool, error) {
	if r.page == nil {
		return false, NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	element, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		return false, NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	visible, err := element.Visible()
	if err != nil {
		return false, NewGowrightError(BrowserError, fmt.Sprintf("failed to check visibility of element with selector %s", selector), err)
	}

	return visible, nil
}

// ScrollToElement scrolls to an element on the page
func (r *RodUITester) ScrollToElement(selector string) error {
	if r.page == nil {
		return NewGowrightError(BrowserError, "page not initialized", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.Timeout)
	defer cancel()

	element, err := r.page.Context(ctx).Element(selector)
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to find element with selector %s", selector), err)
	}

	err = element.ScrollIntoView()
	if err != nil {
		return NewGowrightError(BrowserError, fmt.Sprintf("failed to scroll to element with selector %s", selector), err)
	}

	return nil
}

// ExecuteTest executes a UI test and returns the result
func (r *RodUITester) ExecuteTest(test *UITest) *TestCaseResult {
	startTime := time.Now()
	result := &TestCaseResult{
		Name:      test.Name,
		StartTime: startTime,
		Status:    TestStatusPassed,
		Logs:      make([]string, 0),
		Steps:     make([]AssertionStep, 0),
	}

	// Navigate to the test URL
	if test.URL != "" {
		if err := r.Navigate(test.URL); err != nil {
			result.Status = TestStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(startTime)
			return result
		}
		result.Logs = append(result.Logs, fmt.Sprintf("Navigated to: %s", test.URL))
	}

	// Execute UI actions
	for i, action := range test.Actions {
		actionStart := time.Now()
		var err error

		switch action.Type {
		case "click":
			err = r.Click(action.Selector)
		case "type":
			err = r.Type(action.Selector, action.Value)
		case "navigate":
			err = r.Navigate(action.Value)
		case "wait":
			if timeout, ok := action.Options.(time.Duration); ok {
				err = r.WaitForElement(action.Selector, timeout)
			} else {
				err = r.WaitForElement(action.Selector, 10*time.Second)
			}
		default:
			err = NewGowrightError(BrowserError, fmt.Sprintf("unsupported action type: %s", action.Type), nil)
		}

		actionEnd := time.Now()
		step := AssertionStep{
			Name:        fmt.Sprintf("Action %d: %s", i+1, action.Type),
			Description: fmt.Sprintf("Execute %s action", action.Type),
			StartTime:   actionStart,
			EndTime:     actionEnd,
			Duration:    actionEnd.Sub(actionStart),
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			result.Status = TestStatusFailed
			result.Error = err
			result.Logs = append(result.Logs, fmt.Sprintf("Action %d failed: %v", i+1, err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Action %d completed: %s", i+1, action.Type))
		}

		result.Steps = append(result.Steps, step)

		if result.Status == TestStatusFailed {
			break
		}
	}

	// Execute UI assertions
	for i, assertion := range test.Assertions {
		assertionStart := time.Now()
		var success bool
		var err error

		switch assertion.Type {
		case "text_equals":
			text, getErr := r.GetText(assertion.Selector)
			if getErr != nil {
				err = getErr
			} else {
				success = text == assertion.Expected.(string)
				if !success {
					err = fmt.Errorf("expected text '%s', got '%s'", assertion.Expected, text)
				}
			}
		case "element_present":
			present, getErr := r.IsElementPresent(assertion.Selector)
			if getErr != nil {
				err = getErr
			} else {
				success = present == assertion.Expected.(bool)
				if !success {
					err = fmt.Errorf("expected element presence %v, got %v", assertion.Expected, present)
				}
			}
		case "element_visible":
			visible, getErr := r.IsElementVisible(assertion.Selector)
			if getErr != nil {
				err = getErr
			} else {
				success = visible == assertion.Expected.(bool)
				if !success {
					err = fmt.Errorf("expected element visibility %v, got %v", assertion.Expected, visible)
				}
			}
		default:
			err = NewGowrightError(AssertionError, fmt.Sprintf("unsupported assertion type: %s", assertion.Type), nil)
		}

		assertionEnd := time.Now()
		step := AssertionStep{
			Name:        fmt.Sprintf("Assertion %d: %s", i+1, assertion.Type),
			Description: fmt.Sprintf("Verify %s", assertion.Type),
			StartTime:   assertionStart,
			EndTime:     assertionEnd,
			Duration:    assertionEnd.Sub(assertionStart),
			Expected:    assertion.Expected,
		}

		if err != nil {
			step.Status = TestStatusFailed
			step.Error = err
			result.Status = TestStatusFailed
			if result.Error == nil {
				result.Error = err
			}
			result.Logs = append(result.Logs, fmt.Sprintf("Assertion %d failed: %v", i+1, err))
		} else {
			step.Status = TestStatusPassed
			result.Logs = append(result.Logs, fmt.Sprintf("Assertion %d passed: %s", i+1, assertion.Type))
		}

		result.Steps = append(result.Steps, step)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	return result
}