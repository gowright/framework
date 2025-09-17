package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/core"
	"github.com/gowright/framework/pkg/ui"
)

func main() {
	// Create UI tester
	tester := ui.NewUITester()

	// Configure browser to disable cookie notices and privacy popups
	browserConfig := &config.BrowserConfig{
		Browser:        "chrome",
		Headless:       false, // Set to true for headless mode
		WindowSize:     "1920x1080",
		Timeout:        30 * time.Second,
		ScreenshotPath: "./screenshots",

		// Note: BrowserArgs implementation is pending rod API integration
		// For now, we'll rely on the DismissCookieNotices() method and JavaScript

		// Additional configuration
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	// Initialize tester
	err := tester.Initialize(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	// Test with a site that typically shows cookie notices
	fmt.Println("Testing cookie notice suppression...")

	// Navigate to a site (replace with your target site)
	err = tester.Navigate("https://www.google.com")
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Wait a moment for any popups to appear
	time.Sleep(2 * time.Second)

	// Take a screenshot to verify no cookie notices
	screenshotPath, err := tester.TakeScreenshot("no_cookie_notice")
	if err != nil {
		log.Printf("Failed to take screenshot: %v", err)
	} else {
		fmt.Printf("Screenshot saved: %s\n", screenshotPath)
	}

	// Additional JavaScript to hide any remaining cookie banners
	_, err = tester.ExecuteScript(`
		// Hide common cookie banner selectors
		const cookieSelectors = [
			'[id*="cookie"]', '[class*="cookie"]',
			'[id*="consent"]', '[class*="consent"]', 
			'[id*="gdpr"]', '[class*="gdpr"]',
			'[id*="privacy"]', '[class*="privacy"]',
			'[aria-label*="cookie"]', '[aria-label*="consent"]',
			'.cookie-banner', '.consent-banner', '.privacy-banner',
			'#cookieConsent', '#cookie-consent', '#privacy-notice'
		];
		
		cookieSelectors.forEach(selector => {
			document.querySelectorAll(selector).forEach(el => {
				el.style.display = 'none';
				el.remove();
			});
		});
		
		// Also hide any overlay backgrounds
		document.querySelectorAll('[class*="overlay"], [class*="backdrop"]').forEach(el => {
			if (el.style.zIndex > 1000) {
				el.style.display = 'none';
			}
		});
		
		return 'Cookie banners hidden';
	`)

	if err != nil {
		log.Printf("Failed to execute JavaScript: %v", err)
	} else {
		fmt.Println("Executed JavaScript to hide any remaining cookie banners")
	}

	// Create a structured test that handles cookie notices
	test := &core.UITest{
		Name: "Test with Cookie Notice Handling",
		URL:  "https://example.com", // Replace with your target URL
		Actions: []core.UIAction{
			// Wait for page to load
			{Type: "wait", Value: "2s"},

			// Try to dismiss any cookie notices that might still appear
			// Note: These will fail silently if elements don't exist
			{Type: "click", Selector: "[data-testid='cookie-accept']"},
			{Type: "click", Selector: ".cookie-accept"},
			{Type: "click", Selector: "#accept-cookies"},
			{Type: "click", Selector: "button[contains(text(), 'Accept')]"},

			// Take screenshot after handling
			{Type: "screenshot", Value: "after_cookie_handling"},

			// Continue with your actual test actions
			{Type: "click", Selector: "input[name='search']"},
			{Type: "type", Selector: "input[name='search']", Value: "test query"},
		},
		Assertions: []core.UIAssertion{
			{Type: "element_exists", Selector: "input[name='search']"},
		},
	}

	fmt.Println("Running structured test with cookie handling...")
	result := tester.ExecuteTest(test)

	fmt.Printf("Test completed: %s\n", result.Status)
	if result.Error != nil {
		fmt.Printf("Test error: %v\n", result.Error)
	}

	fmt.Println("Cookie notice suppression test completed!")
}
