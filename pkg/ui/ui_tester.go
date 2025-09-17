// Package ui provides UI testing capabilities using browser automation
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/gowright/framework/pkg/assertions"
	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
)

// UITester implements the UITester interface for browser automation
type UITester struct {
	config      *config.BrowserConfig
	asserter    *assertions.Asserter
	initialized bool
	browser     *rod.Browser
	page        *rod.Page
	launcher    *launcher.Launcher
}

// NewUITester creates a new UI tester instance
func NewUITester() *UITester {
	return &UITester{
		asserter: assertions.NewAsserter(),
	}
}

// Initialize sets up the UI tester with browser configuration
func (ut *UITester) Initialize(cfg interface{}) error {
	browserConfig, ok := cfg.(*config.BrowserConfig)
	if !ok {
		return core.NewGowrightError(core.ConfigurationError, "invalid configuration type for UI tester", nil)
	}

	ut.config = browserConfig

	// Create launcher with configuration
	ut.launcher = launcher.New()

	// Configure browser type
	switch strings.ToLower(browserConfig.Browser) {
	case "chrome", "chromium", "":
		ut.launcher = ut.launcher.Bin("")
	case "firefox":
		// Rod primarily supports Chromium-based browsers
		// For Firefox support, you'd need additional setup
		return core.NewGowrightError(core.ConfigurationError, "Firefox support requires additional configuration", nil)
	default:
		return core.NewGowrightError(core.ConfigurationError, fmt.Sprintf("unsupported browser: %s", browserConfig.Browser), nil)
	}

	// Add default Chrome arguments to improve automation experience
	// These prevent Chrome setup dialogs and first-run experiences that can interfere with testing
	ut.launcher = ut.launcher.Set("no-default-browser-check") // Prevents "Set as default browser" dialog
	ut.launcher = ut.launcher.Set("no-first-run")             // Skips first run experience and setup wizard
	ut.launcher = ut.launcher.Set("disable-fre")              // Disables first run experience (additional safety)

	// Add essential arguments for CI/containerized environments
	ut.launcher = ut.launcher.Set("no-sandbox")            // Required for containerized environments (CI/CD)
	ut.launcher = ut.launcher.Set("disable-dev-shm-usage") // Prevents /dev/shm issues in containers
	ut.launcher = ut.launcher.Set("disable-gpu")           // Disable GPU acceleration for headless environments

	// Configure headless mode
	if browserConfig.Headless {
		ut.launcher = ut.launcher.Headless(true)
	} else {
		ut.launcher = ut.launcher.Headless(false)
	}

	// Configure window size
	if browserConfig.WindowSize != "" {
		parts := strings.Split(browserConfig.WindowSize, "x")
		if len(parts) == 2 {
			width, err1 := strconv.Atoi(parts[0])
			height, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				ut.launcher = ut.launcher.Set("window-size", fmt.Sprintf("%d,%d", width, height))
			}
		}
	}

	// Add custom browser arguments
	// Rod launcher expects arguments to be added differently
	// We'll skip custom args for now and focus on the built-in options
	// Custom arguments can be added by modifying the launcher after creation
	_ = browserConfig.BrowserArgs // Acknowledge the field exists

	// Configure additional options
	if browserConfig.DisableImages {
		ut.launcher = ut.launcher.Set("blink-settings", "imagesEnabled=false")
	}

	if browserConfig.UserAgent != "" {
		ut.launcher = ut.launcher.Set("user-agent", browserConfig.UserAgent)
	}

	// Launch browser
	url, err := ut.launcher.Launch()
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to launch browser", err)
	}

	// Connect to browser
	ut.browser = rod.New().ControlURL(url)
	if err := ut.browser.Connect(); err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to connect to browser", err)
	}

	// Create initial page
	ut.page, err = ut.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to create page", err)
	}

	// Set page timeout
	if browserConfig.Timeout > 0 {
		ut.page = ut.page.Timeout(browserConfig.Timeout)
	}

	// Create screenshot directory if specified
	if browserConfig.ScreenshotPath != "" {
		if err := os.MkdirAll(browserConfig.ScreenshotPath, 0750); err != nil {
			return core.NewGowrightError(core.BrowserError, "failed to create screenshot directory", err)
		}
	}

	ut.initialized = true
	return nil
}

// Cleanup performs cleanup operations
func (ut *UITester) Cleanup() error {
	if ut.page != nil {
		if err := ut.page.Close(); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Warning: failed to close page: %v\n", err)
		}
		ut.page = nil
	}

	if ut.browser != nil {
		if err := ut.browser.Close(); err != nil {
			// Log error but continue cleanup
			fmt.Printf("Warning: failed to close browser: %v\n", err)
		}
		ut.browser = nil
	}

	if ut.launcher != nil {
		ut.launcher.Cleanup()
		ut.launcher = nil
	}

	ut.initialized = false
	return nil
}

// GetName returns the name of the tester
func (ut *UITester) GetName() string {
	return "UITester"
}

// Navigate navigates to the specified URL
func (ut *UITester) Navigate(url string) error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	err := ut.page.Navigate(url)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to navigate to %s", url), err)
	}

	// Wait for page to load
	err = ut.page.WaitLoad()
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "page failed to load", err)
	}

	return nil
}

// Click clicks on an element identified by the selector
func (ut *UITester) Click(selector string) error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	element, err := ut.page.Element(selector)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("element not found: %s", selector), err)
	}

	err = element.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to click element: %s", selector), err)
	}

	return nil
}

// Type types text into an element identified by the selector
func (ut *UITester) Type(selector, text string) error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	element, err := ut.page.Element(selector)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("element not found: %s", selector), err)
	}

	// Clear existing text first
	err = element.SelectAllText()
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to select text in element: %s", selector), err)
	}

	// Type the new text
	err = element.Input(text)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to type text in element: %s", selector), err)
	}

	return nil
}

// GetText retrieves text from an element identified by the selector
func (ut *UITester) GetText(selector string) (string, error) {
	if !ut.initialized {
		return "", core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return "", core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	element, err := ut.page.Element(selector)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, fmt.Sprintf("element not found: %s", selector), err)
	}

	text, err := element.Text()
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to get text from element: %s", selector), err)
	}

	return text, nil
}

// WaitForElement waits for an element to be present
func (ut *UITester) WaitForElement(selector string, timeout time.Duration) error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := ut.page.Context(ctx).Element(selector)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("element not found within timeout: %s", selector), err)
	}

	return nil
}

// TakeScreenshot captures a screenshot and returns the file path
func (ut *UITester) TakeScreenshot(filename string) (string, error) {
	if !ut.initialized {
		return "", core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return "", core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	// Determine the full path for the screenshot
	var fullPath string
	if ut.config.ScreenshotPath != "" {
		fullPath = filepath.Join(ut.config.ScreenshotPath, filename)
	} else {
		fullPath = filename
	}

	// Ensure the filename has a proper extension
	if !strings.HasSuffix(strings.ToLower(fullPath), ".png") {
		fullPath += ".png"
	}

	// Take screenshot
	screenshot, err := ut.page.Screenshot(true, nil)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to take screenshot", err)
	}

	// Write screenshot to file
	err = os.WriteFile(fullPath, screenshot, 0600)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to save screenshot", err)
	}

	return fullPath, nil
}

// GetPageSource returns the current page source
func (ut *UITester) GetPageSource() (string, error) {
	if !ut.initialized {
		return "", core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return "", core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	html, err := ut.page.HTML()
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, "failed to get page source", err)
	}

	return html, nil
}

// GetAttribute retrieves an attribute value from an element
func (ut *UITester) GetAttribute(selector, attribute string) (string, error) {
	if !ut.initialized {
		return "", core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return "", core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	element, err := ut.page.Element(selector)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, fmt.Sprintf("element not found: %s", selector), err)
	}

	attr, err := element.Attribute(attribute)
	if err != nil {
		return "", core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to get attribute %s from element: %s", attribute, selector), err)
	}

	if attr == nil {
		return "", nil
	}

	return *attr, nil
}

// IsElementVisible checks if an element is visible
func (ut *UITester) IsElementVisible(selector string) (bool, error) {
	if !ut.initialized {
		return false, core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return false, core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	element, err := ut.page.Element(selector)
	if err != nil {
		return false, nil // Element doesn't exist, so it's not visible
	}

	visible, err := element.Visible()
	if err != nil {
		return false, core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to check visibility of element: %s", selector), err)
	}

	return visible, nil
}

// ScrollToElement scrolls to make an element visible
func (ut *UITester) ScrollToElement(selector string) error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	element, err := ut.page.Element(selector)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("element not found: %s", selector), err)
	}

	err = element.ScrollIntoView()
	if err != nil {
		return core.NewGowrightError(core.BrowserError, fmt.Sprintf("failed to scroll to element: %s", selector), err)
	}

	return nil
}

// ExecuteScript executes JavaScript in the browser
func (ut *UITester) ExecuteScript(script string) (interface{}, error) {
	if !ut.initialized {
		return nil, core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return nil, core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	result, err := ut.page.Eval(script)
	if err != nil {
		return nil, core.NewGowrightError(core.BrowserError, "failed to execute script", err)
	}

	return result.Value, nil
}

// WaitForText waits for an element to contain specific text
func (ut *UITester) WaitForText(selector, expectedText string, timeout time.Duration) error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return core.NewGowrightError(core.BrowserError, fmt.Sprintf("timeout waiting for text '%s' in selector '%s'", expectedText, selector), nil)
		case <-ticker.C:
			el, err := ut.page.Element(selector)
			if err != nil {
				continue // element not found yet, keep waiting
			}
			text, err := el.Text()
			if err != nil {
				continue // unable to get text, keep waiting
			}
			if strings.Contains(text, expectedText) {
				return nil
			}
		}
	}
}

// DismissCookieNotices attempts to dismiss common cookie notices and privacy banners
func (ut *UITester) DismissCookieNotices() error {
	if !ut.initialized {
		return core.NewGowrightError(core.BrowserError, "UI tester not initialized", nil)
	}

	if ut.page == nil {
		return core.NewGowrightError(core.BrowserError, "no page available", nil)
	}

	// JavaScript to hide and dismiss common cookie banners
	script := `
		// Common cookie banner selectors
		const cookieSelectors = [
			// Generic selectors
			'[id*="cookie" i]', '[class*="cookie" i]',
			'[id*="consent" i]', '[class*="consent" i]', 
			'[id*="gdpr" i]', '[class*="gdpr" i]',
			'[id*="privacy" i]', '[class*="privacy" i]',
			'[aria-label*="cookie" i]', '[aria-label*="consent" i]',
			
			// Common class names
			'.cookie-banner', '.consent-banner', '.privacy-banner',
			'.cookie-notice', '.consent-notice', '.privacy-notice',
			'.cookie-bar', '.consent-bar', '.privacy-bar',
			'.gdpr-banner', '.gdpr-notice', '.gdpr-bar',
			
			// Common IDs
			'#cookieConsent', '#cookie-consent', '#privacy-notice',
			'#gdpr-consent', '#cookie-banner', '#consent-banner',
			
			// Button selectors for accepting
			'button[id*="accept" i]', 'button[class*="accept" i]',
			'button[id*="agree" i]', 'button[class*="agree" i]',
			'a[id*="accept" i]', 'a[class*="accept" i]',
			
			// Specific popular cookie consent tools
			'#onetrust-accept-btn-handler', // OneTrust
			'.ot-sdk-show-settings', // OneTrust settings
			'#truste-consent-button', // TrustArc
			'.trustarc-banner-container', // TrustArc
			'.cc-dismiss', // Cookie Consent by Silktide
			'.cc-allow', // Cookie Consent by Silktide
			'[data-testid*="cookie"]', // Test ID based
			'[data-cy*="cookie"]', // Cypress test selectors
		];
		
		let dismissed = 0;
		
		// Try to click accept buttons first
		cookieSelectors.forEach(selector => {
			try {
				const elements = document.querySelectorAll(selector);
				elements.forEach(el => {
					if (el && (el.tagName === 'BUTTON' || el.tagName === 'A') && 
						(el.textContent.toLowerCase().includes('accept') || 
						 el.textContent.toLowerCase().includes('agree') ||
						 el.textContent.toLowerCase().includes('allow') ||
						 el.id.toLowerCase().includes('accept') ||
						 el.className.toLowerCase().includes('accept'))) {
						el.click();
						dismissed++;
					}
				});
			} catch (e) {
				// Ignore errors for individual selectors
			}
		});
		
		// Then hide remaining banners
		cookieSelectors.forEach(selector => {
			try {
				const elements = document.querySelectorAll(selector);
				elements.forEach(el => {
					if (el && el.offsetParent !== null) { // Only visible elements
						el.style.display = 'none';
						el.remove();
						dismissed++;
					}
				});
			} catch (e) {
				// Ignore errors for individual selectors
			}
		});
		
		// Hide overlay backgrounds that might be left behind
		document.querySelectorAll('[class*="overlay"], [class*="backdrop"], [class*="modal-backdrop"]').forEach(el => {
			if (el.style.zIndex > 1000 || el.className.toLowerCase().includes('cookie') || 
				el.className.toLowerCase().includes('consent') || el.className.toLowerCase().includes('gdpr')) {
				el.style.display = 'none';
				el.remove();
				dismissed++;
			}
		});
		
		return dismissed;
	`

	result, err := ut.page.Eval(script)
	if err != nil {
		return core.NewGowrightError(core.BrowserError, "failed to execute cookie dismissal script", err)
	}

	// Log how many elements were dismissed
	if result != nil {
		if count := result.Value.Num(); count > 0 {
			fmt.Printf("Dismissed %d cookie notice elements\n", int(count))
		}
	}

	return nil
}

// GetRecommendedCookieDisablingArgs returns browser arguments to minimize cookie notices
func GetRecommendedCookieDisablingArgs() []string {
	return []string{
		// Disable cookie notices and privacy sandbox
		"--disable-features=VizDisplayCompositor,PrivacySandboxSettings4",
		"--disable-privacy-sandbox-ads-apis",

		// Disable various Chrome notifications and popups
		"--disable-notifications",
		"--disable-default-apps",
		"--disable-extensions",
		"--disable-translate",

		// Disable sync and sign-in prompts
		"--disable-sync",
		"--disable-background-networking",

		// Security and privacy settings
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-default-browser-check",
		"--disable-client-side-phishing-detection",
		"--disable-component-update",

		// Disable cookie deprecation testing
		"--disable-features=CookieDeprecationFacilitatedTesting",
		"--disable-blink-features=AutomationControlled",
	}
}

// ExecuteTest executes a UI test and returns the result
func (ut *UITester) ExecuteTest(test *core.UITest) *core.TestCaseResult {
	startTime := time.Now()
	result := &core.TestCaseResult{
		Name:      test.Name,
		StartTime: startTime,
		Status:    core.TestStatusPassed,
	}

	ut.asserter.Reset()

	// Navigate to URL if specified
	if test.URL != "" {
		if err := ut.Navigate(test.URL); err != nil {
			result.Status = core.TestStatusError
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		}
	}

	// Execute actions
	for _, action := range test.Actions {
		if err := ut.executeAction(&action); err != nil {
			result.Status = core.TestStatusError
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result
		}
	}

	// Execute assertions
	for _, assertion := range test.Assertions {
		if err := ut.executeAssertion(&assertion); err != nil {
			result.Status = core.TestStatusFailed
			result.Error = err
		}
	}

	// Check for assertion failures
	if ut.asserter.HasFailures() && result.Status == core.TestStatusPassed {
		result.Status = core.TestStatusFailed
		result.Error = core.NewGowrightError(core.AssertionError, "one or more assertions failed", nil)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Steps = ut.asserter.GetSteps()

	return result
}

// executeAction executes a UI action
func (ut *UITester) executeAction(action *core.UIAction) error {
	switch action.Type {
	case "click":
		return ut.Click(action.Selector)
	case "type":
		return ut.Type(action.Selector, action.Value)
	case "navigate":
		return ut.Navigate(action.Value)
	case "wait":
		if action.Selector != "" {
			timeout := 30 * time.Second
			if ut.config.Timeout > 0 {
				timeout = ut.config.Timeout
			}
			return ut.WaitForElement(action.Selector, timeout)
		}
		// If no selector, just wait for the specified duration
		if action.Value != "" {
			if duration, err := time.ParseDuration(action.Value); err == nil {
				time.Sleep(duration)
				return nil
			}
		}
		return core.NewGowrightError(core.BrowserError, "invalid wait action configuration", nil)
	case "scroll":
		if action.Selector != "" {
			return ut.ScrollToElement(action.Selector)
		}
		return core.NewGowrightError(core.BrowserError, "scroll action requires selector", nil)
	case "screenshot":
		filename := action.Value
		if filename == "" {
			filename = fmt.Sprintf("screenshot_%d", time.Now().Unix())
		}
		_, err := ut.TakeScreenshot(filename)
		return err
	default:
		return core.NewGowrightError(core.BrowserError, "unsupported action type: "+action.Type, nil)
	}
}

// executeAssertion executes a UI assertion
func (ut *UITester) executeAssertion(assertion *core.UIAssertion) error {
	switch assertion.Type {
	case "text_equals":
		text, err := ut.GetText(assertion.Selector)
		if err != nil {
			return err
		}
		ut.asserter.Equal(assertion.Expected, text, "Text equals assertion")
	case "text_contains":
		text, err := ut.GetText(assertion.Selector)
		if err != nil {
			return err
		}
		if expectedStr, ok := assertion.Expected.(string); ok {
			ut.asserter.Contains(text, expectedStr, "Text contains assertion")
		}
	case "element_exists":
		if ut.page == nil {
			return core.NewGowrightError(core.BrowserError, "no page available", nil)
		}
		_, err := ut.page.Element(assertion.Selector)
		ut.asserter.True(err == nil, fmt.Sprintf("Element exists assertion for selector: %s", assertion.Selector))
	case "element_visible":
		visible, err := ut.IsElementVisible(assertion.Selector)
		if err != nil {
			return err
		}
		ut.asserter.True(visible, fmt.Sprintf("Element visible assertion for selector: %s", assertion.Selector))
	case "attribute_equals":
		if assertion.Attribute == "" {
			return core.NewGowrightError(core.BrowserError, "attribute name required for attribute_equals assertion", nil)
		}
		attrValue, err := ut.GetAttribute(assertion.Selector, assertion.Attribute)
		if err != nil {
			return err
		}
		ut.asserter.Equal(assertion.Expected, attrValue, fmt.Sprintf("Attribute %s equals assertion for selector: %s", assertion.Attribute, assertion.Selector))
	case "page_title_equals":
		if ut.page == nil {
			return core.NewGowrightError(core.BrowserError, "no page available", nil)
		}
		info, err := ut.page.Info()
		if err != nil {
			return core.NewGowrightError(core.BrowserError, "failed to get page info", err)
		}
		title := info.Title
		ut.asserter.Equal(assertion.Expected, title, "Page title equals assertion")
	case "url_contains":
		if ut.page == nil {
			return core.NewGowrightError(core.BrowserError, "no page available", nil)
		}
		info, err := ut.page.Info()
		if err != nil {
			return core.NewGowrightError(core.BrowserError, "failed to get page info", err)
		}
		url := info.URL
		if expectedStr, ok := assertion.Expected.(string); ok {
			ut.asserter.Contains(url, expectedStr, "URL contains assertion")
		}
	default:
		return core.NewGowrightError(core.BrowserError, "unsupported assertion type: "+assertion.Type, nil)
	}
	return nil
}
