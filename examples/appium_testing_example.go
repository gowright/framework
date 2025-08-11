package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/gowright"
)

func main() {
	// Create a new Appium client
	client := gowright.NewAppiumClient("http://localhost:4723")
	ctx := context.Background()

	// Example 1: Android App Testing
	fmt.Println("=== Android App Testing Example ===")
	if err := androidAppExample(ctx, client); err != nil {
		log.Printf("Android example failed: %v", err)
	}

	// Example 2: iOS App Testing
	fmt.Println("\n=== iOS App Testing Example ===")
	if err := iOSAppExample(ctx, client); err != nil {
		log.Printf("iOS example failed: %v", err)
	}

	// Example 3: Web App Testing on Mobile
	fmt.Println("\n=== Mobile Web Testing Example ===")
	if err := mobileWebExample(ctx, client); err != nil {
		log.Printf("Mobile web example failed: %v", err)
	}

	// Example 4: Appium with GoWright Framework
	fmt.Println("\n=== Appium with GoWright Framework Example ===")
	appiumWithGoWrightExample()
}

func androidAppExample(ctx context.Context, client *gowright.AppiumClient) error {
	// Define Android capabilities
	caps := gowright.AppiumCapabilities{
		PlatformName:      "Android",
		PlatformVersion:   "11",
		DeviceName:        "emulator-5554",
		AppPackage:        "com.android.calculator2",
		AppActivity:       ".Calculator",
		AutomationName:    "UiAutomator2",
		NoReset:           true,
		NewCommandTimeout: 60,
	}

	// Create session
	fmt.Println("Creating Android session...")
	if err := client.CreateSession(ctx, caps); err != nil {
		return fmt.Errorf("failed to create Android session: %w", err)
	}
	defer func() {
		fmt.Println("Closing Android session...")
		if err := client.DeleteSession(ctx); err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}()

	fmt.Printf("Session created with ID: %s\n", client.GetSessionID())

	// Wait for the calculator to load
	time.Sleep(2 * time.Second)

	// Find and click number buttons to perform calculation: 5 + 3 = 8
	fmt.Println("Performing calculation: 5 + 3 = 8")

	// Click number 5
	num5, err := client.WaitForElementClickable(ctx, gowright.ByID, "com.android.calculator2:id/digit_5", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to find number 5: %w", err)
	}
	if err := num5.Click(ctx); err != nil {
		return fmt.Errorf("failed to click number 5: %w", err)
	}

	// Click plus button
	plus, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/op_add")
	if err != nil {
		return fmt.Errorf("failed to find plus button: %w", err)
	}
	if err := plus.Click(ctx); err != nil {
		return fmt.Errorf("failed to click plus button: %w", err)
	}

	// Click number 3
	num3, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/digit_3")
	if err != nil {
		return fmt.Errorf("failed to find number 3: %w", err)
	}
	if err := num3.Click(ctx); err != nil {
		return fmt.Errorf("failed to click number 3: %w", err)
	}

	// Click equals button
	equals, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/eq")
	if err != nil {
		return fmt.Errorf("failed to find equals button: %w", err)
	}
	if err := equals.Click(ctx); err != nil {
		return fmt.Errorf("failed to click equals button: %w", err)
	}

	// Get the result
	result, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/result")
	if err != nil {
		return fmt.Errorf("failed to find result: %w", err)
	}

	resultText, err := result.GetText(ctx)
	if err != nil {
		return fmt.Errorf("failed to get result text: %w", err)
	}

	fmt.Printf("Calculation result: %s\n", resultText)

	// Take a screenshot
	fmt.Println("Taking screenshot...")
	screenshot, err := client.TakeScreenshot(ctx)
	if err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}
	fmt.Printf("Screenshot captured (length: %d bytes)\n", len(screenshot))

	// Test touch actions
	fmt.Println("Testing touch actions...")
	width, height, err := client.GetWindowSize(ctx)
	if err != nil {
		return fmt.Errorf("failed to get window size: %w", err)
	}
	fmt.Printf("Screen size: %dx%d\n", width, height)

	// Perform a swipe gesture
	if err := client.Swipe(ctx, width/2, height/2, width/2, height/4, 1000); err != nil {
		return fmt.Errorf("failed to perform swipe: %w", err)
	}
	fmt.Println("Swipe gesture performed")

	return nil
}

func iOSAppExample(ctx context.Context, client *gowright.AppiumClient) error {
	// Define iOS capabilities
	caps := gowright.AppiumCapabilities{
		PlatformName:      "iOS",
		PlatformVersion:   "15.0",
		DeviceName:        "iPhone 13 Simulator",
		BundleID:          "com.apple.calculator",
		AutomationName:    "XCUITest",
		NoReset:           true,
		NewCommandTimeout: 60,
	}

	// Create session
	fmt.Println("Creating iOS session...")
	if err := client.CreateSession(ctx, caps); err != nil {
		return fmt.Errorf("failed to create iOS session: %w", err)
	}
	defer func() {
		fmt.Println("Closing iOS session...")
		if err := client.DeleteSession(ctx); err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}()

	fmt.Printf("Session created with ID: %s\n", client.GetSessionID())

	// Wait for the calculator to load
	time.Sleep(2 * time.Second)

	// Find and click number buttons using iOS locators
	fmt.Println("Performing calculation: 7 + 2 = 9")

	// Click number 7 using accessibility ID
	num7, err := client.WaitForElementClickable(ctx, gowright.ByAccessibilityID, "7", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to find number 7: %w", err)
	}
	if err := num7.Click(ctx); err != nil {
		return fmt.Errorf("failed to click number 7: %w", err)
	}

	// Click plus button
	plus, err := client.FindElement(ctx, gowright.ByAccessibilityID, "+")
	if err != nil {
		return fmt.Errorf("failed to find plus button: %w", err)
	}
	if err := plus.Click(ctx); err != nil {
		return fmt.Errorf("failed to click plus button: %w", err)
	}

	// Click number 2
	num2, err := client.FindElement(ctx, gowright.ByAccessibilityID, "2")
	if err != nil {
		return fmt.Errorf("failed to find number 2: %w", err)
	}
	if err := num2.Click(ctx); err != nil {
		return fmt.Errorf("failed to click number 2: %w", err)
	}

	// Click equals button
	equals, err := client.FindElement(ctx, gowright.ByAccessibilityID, "=")
	if err != nil {
		return fmt.Errorf("failed to find equals button: %w", err)
	}
	if err := equals.Click(ctx); err != nil {
		return fmt.Errorf("failed to click equals button: %w", err)
	}

	// Test iOS-specific locators
	fmt.Println("Testing iOS-specific locators...")

	// Using iOS predicate locators
	by, value := gowright.IOS.Label("Calculator")
	elements, err := client.FindElements(ctx, by, value)
	if err == nil {
		fmt.Printf("Found %d elements with label 'Calculator'\n", len(elements))
	}

	// Test device orientation
	orientation, err := client.GetOrientation(ctx)
	if err == nil {
		fmt.Printf("Current orientation: %s\n", orientation)
	}

	return nil
}

func mobileWebExample(ctx context.Context, client *gowright.AppiumClient) error {
	// Define capabilities for mobile web testing
	caps := gowright.AppiumCapabilities{
		PlatformName:      "Android",
		PlatformVersion:   "11",
		DeviceName:        "emulator-5554",
		AutomationName:    "UiAutomator2",
		NoReset:           true,
		NewCommandTimeout: 60,
	}

	// Create session
	fmt.Println("Creating mobile web session...")
	if err := client.CreateSession(ctx, caps); err != nil {
		return fmt.Errorf("failed to create mobile web session: %w", err)
	}
	defer func() {
		fmt.Println("Closing mobile web session...")
		if err := client.DeleteSession(ctx); err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}()

	fmt.Printf("Session created with ID: %s\n", client.GetSessionID())

	// Open Chrome browser
	fmt.Println("Opening Chrome browser...")
	if err := client.StartActivity(ctx, "com.android.chrome", "com.google.android.apps.chrome.Main"); err != nil {
		return fmt.Errorf("failed to start Chrome: %w", err)
	}

	time.Sleep(3 * time.Second)

	// Find the address bar and navigate to a website
	fmt.Println("Navigating to example.com...")
	addressBar, err := client.WaitForElement(ctx, gowright.ByID, "com.android.chrome:id/url_bar", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to find address bar: %w", err)
	}

	if err := addressBar.Click(ctx); err != nil {
		return fmt.Errorf("failed to click address bar: %w", err)
	}

	if err := addressBar.SendKeys(ctx, "https://example.com"); err != nil {
		return fmt.Errorf("failed to type URL: %w", err)
	}

	// Press enter (using Android key event)
	// This would typically require sending a key event, but for simplicity we'll skip it

	fmt.Println("Mobile web navigation example completed")

	// Test page source retrieval
	fmt.Println("Getting page source...")
	source, err := client.GetPageSource(ctx)
	if err == nil {
		fmt.Printf("Page source length: %d characters\n", len(source))
	}

	return nil
}

// appiumWithGoWrightExample demonstrates using Appium with the GoWright testing framework
func appiumWithGoWrightExample() {
	// Create a test suite that uses Appium
	suite := gowright.NewTestSuite("Mobile App Tests")

	// Add mobile test cases
	suite.AddTestFunc("Android Calculator Test", func(t *gowright.TestContext) {
		client := gowright.NewAppiumClient("http://localhost:4723")
		ctx := context.Background()

		caps := gowright.AppiumCapabilities{
			PlatformName:   "Android",
			DeviceName:     "emulator-5554",
			AppPackage:     "com.android.calculator2",
			AppActivity:    ".Calculator",
			AutomationName: "UiAutomator2",
		}

		// Create session
		err := client.CreateSession(ctx, caps)
		t.AssertNoError(err, "Should create Appium session successfully")
		defer func() {
			if err := client.DeleteSession(ctx); err != nil {
				log.Printf("Failed to delete session: %v", err)
			}
		}()

		// Perform test actions
		element, err := client.FindElement(ctx, gowright.ByID, "com.android.calculator2:id/digit_1")
		t.AssertNoError(err, "Should find digit 1 button")

		err = element.Click(ctx)
		t.AssertNoError(err, "Should click digit 1 button")

		// Add more test assertions...
	})

	// Run the test suite
	results := suite.Run()
	fmt.Printf("Test Results: %d passed, %d failed\n", results.PassedCount, results.FailedCount)
}
