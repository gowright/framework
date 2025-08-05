//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gowright/framework/pkg/gowright"
)

func main() {
	fmt.Println("=== Gowright UI Testing Example ===\n")

	// Create browser configuration
	config := &gowright.BrowserConfig{
		Headless:   false, // Set to true for headless mode
		Timeout:    30 * time.Second,
		UserAgent:  "Gowright-UI-Tester/1.0",
		WindowSize: &gowright.WindowSize{Width: 1920, Height: 1080},
	}

	// Create and initialize UI tester
	tester := gowright.NewUITester(config)
	if err := tester.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize UI tester: %v", err)
	}
	defer tester.Cleanup()

	// Example 1: Basic navigation and element interaction
	fmt.Println("1. Testing basic navigation and form interaction")
	navigationTest := gowright.NewUITest("Navigation and Form Test", "https://httpbin.org/forms/post")

	// Add actions to the test
	navigationTest.AddAction(gowright.NewNavigateAction("https://httpbin.org/forms/post"))
	navigationTest.AddAction(gowright.NewWaitAction(2 * time.Second))
	navigationTest.AddAction(gowright.NewTypeAction("input[name='custname']", "John Doe"))
	navigationTest.AddAction(gowright.NewTypeAction("input[name='custtel']", "555-1234"))
	navigationTest.AddAction(gowright.NewTypeAction("input[name='custemail']", "john@example.com"))
	navigationTest.AddAction(gowright.NewSelectAction("select[name='size']", "medium"))
	navigationTest.AddAction(gowright.NewClickAction("input[type='submit']"))

	// Add assertions
	navigationTest.AddAssertion(gowright.NewElementPresentAssertion("form"))
	navigationTest.AddAssertion(gowright.NewTextContentAssertion("title", "httpbin.org"))

	result := navigationTest.Execute(tester)
	printUITestResult(result)

	// Example 2: Element visibility and interaction testing
	fmt.Println("\n2. Testing element visibility and interactions")
	interactionTest := gowright.NewUITest("Element Interaction Test", "https://the-internet.herokuapp.com/dropdown")

	interactionTest.AddAction(gowright.NewNavigateAction("https://the-internet.herokuapp.com/dropdown"))
	interactionTest.AddAction(gowright.NewWaitForElementAction("select#dropdown", 10*time.Second))
	interactionTest.AddAction(gowright.NewSelectAction("select#dropdown", "Option 1"))
	interactionTest.AddAction(gowright.NewWaitAction(1 * time.Second))

	// Add visibility assertions
	interactionTest.AddAssertion(gowright.NewElementVisibleAssertion("select#dropdown"))
	interactionTest.AddAssertion(gowright.NewElementPresentAssertion("h3"))
	interactionTest.AddAssertion(gowright.NewTextContentAssertion("h3", "Dropdown"))

	result = interactionTest.Execute(tester)
	printUITestResult(result)

	// Example 3: Screenshot capture and page source validation
	fmt.Println("\n3. Testing screenshot capture and page validation")
	captureTest := gowright.NewUITest("Screenshot and Validation Test", "https://example.com")

	captureTest.AddAction(gowright.NewNavigateAction("https://example.com"))
	captureTest.AddAction(gowright.NewWaitAction(2 * time.Second))
	captureTest.AddAction(gowright.NewScreenshotAction("example_page.png"))

	// Add content assertions
	captureTest.AddAssertion(gowright.NewTextContentAssertion("h1", "Example Domain"))
	captureTest.AddAssertion(gowright.NewElementPresentAssertion("p"))
	captureTest.AddAssertion(gowright.NewPageTitleAssertion("Example Domain"))

	result = captureTest.Execute(tester)
	printUITestResult(result)

	// Example 4: Dynamic content and AJAX testing
	fmt.Println("\n4. Testing dynamic content and AJAX interactions")
	ajaxTest := gowright.NewUITest("AJAX and Dynamic Content Test", "https://the-internet.herokuapp.com/dynamic_loading/1")

	ajaxTest.AddAction(gowright.NewNavigateAction("https://the-internet.herokuapp.com/dynamic_loading/1"))
	ajaxTest.AddAction(gowright.NewClickAction("button"))
	ajaxTest.AddAction(gowright.NewWaitForElementAction("#finish", 10*time.Second))

	// Add dynamic content assertions
	ajaxTest.AddAssertion(gowright.NewElementVisibleAssertion("#finish"))
	ajaxTest.AddAssertion(gowright.NewTextContentAssertion("#finish h4", "Hello World!"))

	result = ajaxTest.Execute(tester)
	printUITestResult(result)

	// Example 5: Multi-step user workflow
	fmt.Println("\n5. Testing complete user workflow")
	workflowTest := gowright.NewUITest("User Workflow Test", "https://the-internet.herokuapp.com/login")

	// Login workflow
	workflowTest.AddAction(gowright.NewNavigateAction("https://the-internet.herokuapp.com/login"))
	workflowTest.AddAction(gowright.NewTypeAction("#username", "tomsmith"))
	workflowTest.AddAction(gowright.NewTypeAction("#password", "SuperSecretPassword!"))
	workflowTest.AddAction(gowright.NewClickAction("button[type='submit']"))
	workflowTest.AddAction(gowright.NewWaitForElementAction(".flash.success", 5*time.Second))

	// Verify successful login
	workflowTest.AddAssertion(gowright.NewElementPresentAssertion(".flash.success"))
	workflowTest.AddAssertion(gowright.NewTextContainsAssertion(".flash.success", "You logged into a secure area!"))
	workflowTest.AddAssertion(gowright.NewElementPresentAssertion("a[href='/logout']"))

	// Logout workflow
	workflowTest.AddAction(gowright.NewClickAction("a[href='/logout']"))
	workflowTest.AddAction(gowright.NewWaitForElementAction(".flash.success", 5*time.Second))

	// Verify successful logout
	workflowTest.AddAssertion(gowright.NewTextContainsAssertion(".flash.success", "You logged out of the secure area!"))

	result = workflowTest.Execute(tester)
	printUITestResult(result)

	// Example 6: Error handling and negative testing
	fmt.Println("\n6. Testing error scenarios and negative cases")
	errorTest := gowright.NewUITest("Error Handling Test", "https://the-internet.herokuapp.com/login")

	// Attempt login with invalid credentials
	errorTest.AddAction(gowright.NewNavigateAction("https://the-internet.herokuapp.com/login"))
	errorTest.AddAction(gowright.NewTypeAction("#username", "invalid_user"))
	errorTest.AddAction(gowright.NewTypeAction("#password", "invalid_password"))
	errorTest.AddAction(gowright.NewClickAction("button[type='submit']"))
	errorTest.AddAction(gowright.NewWaitForElementAction(".flash.error", 5*time.Second))

	// Verify error message
	errorTest.AddAssertion(gowright.NewElementPresentAssertion(".flash.error"))
	errorTest.AddAssertion(gowright.NewTextContainsAssertion(".flash.error", "Your username is invalid!"))

	result = errorTest.Execute(tester)
	printUITestResult(result)

	// Generate comprehensive UI test report
	fmt.Println("\nGenerating UI test reports...")

	// Create test results collection
	allResults := []*gowright.TestCaseResult{
		// Add all test results here
	}

	testResults := &gowright.TestResults{
		SuiteName:    "UI Testing Example Suite",
		StartTime:    time.Now().Add(-10 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   6,
		PassedTests:  5,
		FailedTests:  1,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases:    convertToTestCaseResults(allResults),
	}

	reportConfig := &gowright.ReportConfig{
		LocalReports: gowright.LocalReportConfig{
			JSON:      true,
			HTML:      true,
			OutputDir: "./ui-test-reports",
		},
	}

	reportManager := gowright.NewReportManager(reportConfig)
	if err := reportManager.GenerateReports(testResults); err != nil {
		log.Printf("Failed to generate reports: %v", err)
	} else {
		fmt.Printf("UI test reports generated in: %s\n", config.LocalReports.OutputDir)
	}

	fmt.Println("\n=== UI Testing Complete ===")
}

func printUITestResult(result *gowright.TestCaseResult) {
	fmt.Printf("Test: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status.String())
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	if len(result.Screenshots) > 0 {
		fmt.Printf("Screenshots: %v\n", result.Screenshots)
	}

	if len(result.Logs) > 0 {
		fmt.Println("Logs:")
		for _, logEntry := range result.Logs {
			fmt.Printf("  - %s\n", logEntry)
		}
	}

	fmt.Println("---")
}

func convertToTestCaseResults(results []*gowright.TestCaseResult) []gowright.TestCaseResult {
	converted := make([]gowright.TestCaseResult, len(results))
	for i, result := range results {
		if result != nil {
			converted[i] = *result
		}
	}
	return converted
}
