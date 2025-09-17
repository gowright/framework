package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/ui"
)

func main() {
	fmt.Println("Testing UI Tester with default Chrome arguments...")
	fmt.Println("Default arguments applied:")
	fmt.Println("  --no-default-browser-check")
	fmt.Println("  --no-first-run")
	fmt.Println("  --disable-fre")
	fmt.Println()

	// Create UI tester
	tester := ui.NewUITester()

	// Configure browser with minimal settings
	browserConfig := &config.BrowserConfig{
		Browser:        "chrome",
		Headless:       false, // Use non-headless to see the browser behavior
		WindowSize:     "1280x720",
		Timeout:        30 * time.Second,
		ScreenshotPath: "./screenshots",
	}

	// Initialize tester
	fmt.Println("Initializing browser with default arguments...")
	err := tester.Initialize(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	fmt.Println("Browser launched successfully!")
	fmt.Println("You should notice:")
	fmt.Println("  - No default browser check dialog")
	fmt.Println("  - No first run setup wizard")
	fmt.Println("  - Clean browser startup")
	fmt.Println()

	// Navigate to a simple page
	fmt.Println("Navigating to Google...")
	err = tester.Navigate("https://www.google.com")
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Wait a moment for the page to load
	time.Sleep(3 * time.Second)

	// Take a screenshot to verify everything is working
	screenshotPath, err := tester.TakeScreenshot("default_args_test")
	if err != nil {
		log.Printf("Failed to take screenshot: %v", err)
	} else {
		fmt.Printf("Screenshot saved: %s\n", screenshotPath)
	}

	// Test basic functionality
	fmt.Println("Testing basic functionality...")

	// Try to get the page title
	title, err := tester.ExecuteScript("return document.title")
	if err != nil {
		log.Printf("Failed to get page title: %v", err)
	} else {
		fmt.Printf("Page title: %v\n", title)
	}

	// Check if search box is present
	visible, err := tester.IsElementVisible("textarea[name='q']")
	if err != nil {
		log.Printf("Failed to check search box visibility: %v", err)
	} else if visible {
		fmt.Println("Search box is visible and ready for interaction")

		// Try typing in the search box
		err = tester.Type("textarea[name='q']", "gowright testing framework")
		if err != nil {
			log.Printf("Failed to type in search box: %v", err)
		} else {
			fmt.Println("Successfully typed in search box")
		}
	}

	// Take final screenshot
	finalPath, err := tester.TakeScreenshot("final_state")
	if err != nil {
		log.Printf("Failed to take final screenshot: %v", err)
	} else {
		fmt.Printf("Final screenshot: %s\n", finalPath)
	}

	fmt.Println("\nTest completed successfully!")
	fmt.Println("The browser should have started cleanly without any setup dialogs.")

	// Keep browser open for a moment to observe
	if !browserConfig.Headless {
		fmt.Println("Browser will close in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}
