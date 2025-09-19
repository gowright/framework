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

	// Configure browser with comprehensive cookie notice disabling
	browserConfig := &config.BrowserConfig{
		Browser:        "chrome",
		Headless:       false, // Set to true for headless mode
		WindowSize:     "1920x1080",
		Timeout:        30 * time.Second,
		ScreenshotPath: "./screenshots",

		// Note: BrowserArgs implementation is pending rod API integration
		// For now, we'll rely on the DismissCookieNotices() method

		// Additional configuration
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	// Initialize tester
	err := tester.Initialize(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	// Test sites that commonly show cookie notices
	testSites := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.stackoverflow.com",
	}

	for i, site := range testSites {
		fmt.Printf("\n=== Testing site %d: %s ===\n", i+1, site)

		// Navigate to the site
		err = tester.Navigate(site)
		if err != nil {
			log.Printf("Failed to navigate to %s: %v", site, err)
			continue
		}

		// Wait for page to load
		time.Sleep(2 * time.Second)

		// Take screenshot before cookie dismissal
		screenshotBefore := fmt.Sprintf("before_cookie_dismissal_%d", i+1)
		_, err = tester.TakeScreenshot(screenshotBefore)
		if err != nil {
			log.Printf("Failed to take before screenshot: %v", err)
		}

		// Dismiss cookie notices using the built-in method
		err = tester.DismissCookieNotices()
		if err != nil {
			log.Printf("Failed to dismiss cookie notices: %v", err)
		} else {
			fmt.Println("Cookie dismissal script executed successfully")
		}

		// Wait a moment for any animations to complete
		time.Sleep(1 * time.Second)

		// Take screenshot after cookie dismissal
		screenshotAfter := fmt.Sprintf("after_cookie_dismissal_%d", i+1)
		_, err = tester.TakeScreenshot(screenshotAfter)
		if err != nil {
			log.Printf("Failed to take after screenshot: %v", err)
		}

		fmt.Printf("Screenshots saved: %s.png and %s.png\n", screenshotBefore, screenshotAfter)
	}

	// Demonstrate using cookie handling in a structured test
	fmt.Println("\n=== Running structured test with cookie handling ===")

	test := &core.UITest{
		Name: "Test with Automatic Cookie Handling",
		URL:  "https://www.google.com",
		Actions: []core.UIAction{
			// Wait for page to load
			{Type: "wait", Value: "2s"},

			// Take initial screenshot
			{Type: "screenshot", Value: "initial_load"},

			// The cookie dismissal will be handled programmatically after this test
			// Continue with normal test actions
			{Type: "wait", Selector: "textarea[name='q']"},
			{Type: "type", Selector: "textarea[name='q']", Value: "automated testing"},
			{Type: "screenshot", Value: "search_entered"},
		},
		Assertions: []core.UIAssertion{
			{Type: "element_exists", Selector: "textarea[name='q']"},
			{Type: "page_title_equals", Expected: "Google"},
		},
	}

	// Execute the test
	result := tester.ExecuteTest(test)

	// After the test, dismiss any cookie notices that might have appeared
	err = tester.DismissCookieNotices()
	if err != nil {
		log.Printf("Failed to dismiss cookie notices after test: %v", err)
	}

	// Take final screenshot
	_, err = tester.TakeScreenshot("final_state")
	if err != nil {
		log.Printf("Failed to take final screenshot: %v", err)
	}

	// Print test results
	fmt.Printf("\nTest Results:\n")
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	if len(result.Steps) > 0 {
		fmt.Println("Steps:")
		for _, step := range result.Steps {
			fmt.Printf("  - %s\n", step)
		}
	}

	fmt.Println("\nCookie handling test completed!")
	fmt.Println("Check the screenshots directory to see the before/after comparisons.")
}
