package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/ui"
)

func main() {
	// Create UI tester
	tester := ui.NewUITester()

	// Configure browser
	browserConfig := &config.BrowserConfig{
		Browser:        "chrome",
		Headless:       false, // Set to true for headless mode
		WindowSize:     "1920x1080",
		Timeout:        30 * time.Second,
		ScreenshotPath: "./screenshots",
	}

	// Initialize tester
	err := tester.Initialize(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	fmt.Println("Testing cookie notice dismissal...")

	// Navigate to a site that typically shows cookie notices
	err = tester.Navigate("https://www.github.com")
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	// Take screenshot before dismissing cookies
	beforePath, err := tester.TakeScreenshot("before_cookie_dismissal")
	if err != nil {
		log.Printf("Failed to take before screenshot: %v", err)
	} else {
		fmt.Printf("Before screenshot: %s\n", beforePath)
	}

	// Dismiss cookie notices
	fmt.Println("Dismissing cookie notices...")
	err = tester.DismissCookieNotices()
	if err != nil {
		log.Printf("Failed to dismiss cookie notices: %v", err)
	} else {
		fmt.Println("Cookie dismissal completed")
	}

	// Wait for any animations to complete
	time.Sleep(2 * time.Second)

	// Take screenshot after dismissing cookies
	afterPath, err := tester.TakeScreenshot("after_cookie_dismissal")
	if err != nil {
		log.Printf("Failed to take after screenshot: %v", err)
	} else {
		fmt.Printf("After screenshot: %s\n", afterPath)
	}

	// Try interacting with the page to ensure it's functional
	fmt.Println("Testing page interaction...")

	// Look for a search input or similar element
	searchSelectors := []string{
		"input[name='q']",
		"input[type='search']",
		"[data-target='qbsearch-input.inputElement']",
		".search-input",
	}

	for _, selector := range searchSelectors {
		visible, err := tester.IsElementVisible(selector)
		if err == nil && visible {
			fmt.Printf("Found interactive element: %s\n", selector)

			// Try to click and type in it
			err = tester.Click(selector)
			if err == nil {
				err = tester.Type(selector, "test search")
				if err == nil {
					fmt.Println("Successfully interacted with search element")
					break
				}
			}
		}
	}

	// Take final screenshot
	finalPath, err := tester.TakeScreenshot("final_state")
	if err != nil {
		log.Printf("Failed to take final screenshot: %v", err)
	} else {
		fmt.Printf("Final screenshot: %s\n", finalPath)
	}

	fmt.Println("\nCookie dismissal test completed!")
	fmt.Println("Check the screenshots to see the before/after comparison.")
}
