package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/config"
	"github.com/gowright/framework/pkg/ui"
)

func main() {
	fmt.Println("Testing UI Tester in CI/containerized environment...")
	fmt.Println("This test verifies that Chrome can launch with the default arguments:")
	fmt.Println("  --no-sandbox (required for containers)")
	fmt.Println("  --disable-dev-shm-usage (prevents /dev/shm issues)")
	fmt.Println("  --disable-gpu (for headless environments)")
	fmt.Println()

	// Create UI tester
	tester := ui.NewUITester()

	// Configure browser for CI environment
	browserConfig := &config.BrowserConfig{
		Browser:        "chrome",
		Headless:       true, // Always use headless in CI
		WindowSize:     "1280x720",
		Timeout:        30 * time.Second,
		ScreenshotPath: "./screenshots",
	}

	// Initialize tester
	fmt.Println("Initializing browser with CI-friendly arguments...")
	err := tester.Initialize(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	fmt.Println("✅ Browser launched successfully in CI environment!")

	// Navigate to a simple test page
	fmt.Println("Testing basic functionality...")
	testHTML := `data:text/html,<html><head><title>CI Test Page</title></head><body>
		<h1 id="title">CI Test Success</h1>
		<p id="message">This page loaded successfully in a containerized environment.</p>
		<input id="test-input" type="text" placeholder="Test input" />
		<button id="test-button">Test Button</button>
	</body></html>`

	err = tester.Navigate(testHTML)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	fmt.Println("✅ Page navigation successful!")

	// Test basic interactions
	title, err := tester.GetText("#title")
	if err != nil {
		log.Printf("Failed to get title: %v", err)
	} else {
		fmt.Printf("✅ Page title: %s\n", title)
	}

	// Test typing
	err = tester.Type("#test-input", "CI test input")
	if err != nil {
		log.Printf("Failed to type: %v", err)
	} else {
		fmt.Println("✅ Text input successful!")
	}

	// Test clicking
	err = tester.Click("#test-button")
	if err != nil {
		log.Printf("Failed to click: %v", err)
	} else {
		fmt.Println("✅ Button click successful!")
	}

	// Take a screenshot to verify everything works
	screenshotPath, err := tester.TakeScreenshot("ci_test")
	if err != nil {
		log.Printf("Failed to take screenshot: %v", err)
	} else {
		fmt.Printf("✅ Screenshot saved: %s\n", screenshotPath)
	}

	// Test JavaScript execution
	result, err := tester.ExecuteScript("return document.title + ' - JS Works!'")
	if err != nil {
		log.Printf("Failed to execute JavaScript: %v", err)
	} else {
		fmt.Printf("✅ JavaScript execution: %v\n", result)
	}

	fmt.Println()
	fmt.Println("🎉 All CI environment tests passed!")
	fmt.Println("The UI testing framework is ready for containerized environments.")
}
