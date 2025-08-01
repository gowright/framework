package main

import (
	"fmt"
	"log"
	"time"

	"github/gowright/framework/pkg/gowright"
)

// ExampleTest demonstrates a test using the assertion system
type ExampleTest struct {
	name string
}

func (et *ExampleTest) GetName() string {
	return et.name
}

func (et *ExampleTest) Execute() *gowright.TestCaseResult {
	startTime := time.Now()
	
	// Create a test assertion instance
	ta := gowright.NewTestAssertion(et.name)
	
	// Log test start
	ta.Log("Starting example test execution")
	
	// Simulate API testing with assertions
	ta.Log("Testing API response validation")
	apiResponse := map[string]interface{}{
		"status":  "success",
		"message": "Operation completed successfully",
		"data": map[string]interface{}{
			"user_id": 12345,
			"name":    "John Doe",
			"email":   "john.doe@example.com",
		},
		"items": []string{"item1", "item2", "item3"},
	}
	
	// Perform assertions with detailed logging
	ta.Equal("success", apiResponse["status"], "API should return success status")
	ta.NotNil(apiResponse["data"], "API response should contain data")
	ta.Contains(apiResponse["message"].(string), "completed", "Message should indicate completion")
	ta.Len(apiResponse["items"], 3, "Should return exactly 3 items")
	
	// Test user data validation
	ta.Log("Validating user data structure")
	userData := apiResponse["data"].(map[string]interface{})
	ta.NotEmpty(userData["name"], "User name should not be empty")
	ta.True(userData["user_id"].(int) > 0, "User ID should be positive")
	ta.Contains(userData["email"].(string), "@", "Email should contain @ symbol")
	
	// Simulate a database check
	ta.Log("Performing database validation")
	dbConnected := true
	recordCount := 5
	
	ta.True(dbConnected, "Database connection should be established")
	ta.True(recordCount > 0, "Should have records in database")
	ta.Equal(5, recordCount, "Should have exactly 5 records")
	
	// Simulate an intentional failure for demonstration
	ta.Log("Testing error handling")
	ta.Equal("expected_value", "actual_value", "This assertion will fail for demonstration")
	
	// Complete the test
	endTime := time.Now()
	ta.Log("Test execution completed")
	
	// Determine overall test status
	status := gowright.TestStatusPassed
	var testError error
	if ta.HasFailures() {
		status = gowright.TestStatusFailed
		testError = fmt.Errorf("test failed with %d assertion failures", func() int {
			_, failed := ta.GetSummary()
			return failed
		}())
	}
	
	return &gowright.TestCaseResult{
		Name:      et.name,
		Status:    status,
		Duration:  endTime.Sub(startTime),
		Error:     testError,
		StartTime: startTime,
		EndTime:   endTime,
		Logs:      ta.GetLogs(),
		Steps:     ta.GetSteps(),
	}
}

func main() {
	fmt.Println("Running example test with assertion reporting...")
	
	// Create and execute the test
	test := &ExampleTest{name: "API Integration Test with Assertions"}
	result := test.Execute()
	
	// Create test results structure
	testResults := &gowright.TestResults{
		SuiteName:    "Assertion Reporting Example",
		StartTime:    result.StartTime,
		EndTime:      result.EndTime,
		TotalTests:   1,
		PassedTests:  0,
		FailedTests:  0,
		SkippedTests: 0,
		ErrorTests:   0,
		TestCases:    []gowright.TestCaseResult{*result},
	}
	
	// Update counters based on result
	switch result.Status {
	case gowright.TestStatusPassed:
		testResults.PassedTests = 1
	case gowright.TestStatusFailed:
		testResults.FailedTests = 1
	case gowright.TestStatusSkipped:
		testResults.SkippedTests = 1
	case gowright.TestStatusError:
		testResults.ErrorTests = 1
	}
	
	// Create report configuration
	config := &gowright.ReportConfig{
		LocalReports: gowright.LocalReportConfig{
			JSON:      true,
			HTML:      true,
			OutputDir: "./assertion-reports",
		},
	}
	
	// Generate reports
	reportManager := gowright.NewReportManager(config)
	if err := reportManager.GenerateReports(testResults); err != nil {
		log.Fatalf("Failed to generate reports: %v", err)
	}
	
	// Display summary
	fmt.Printf("\nTest Execution Summary:\n")
	fmt.Printf("Test: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Assertion Steps: %d\n", len(result.Steps))
	fmt.Printf("Log Entries: %d\n", len(result.Logs))
	
	if result.Error != nil {
		fmt.Printf("Error: %s\n", result.Error.Error())
	}
	
	// Display assertion summary
	passed := 0
	failed := 0
	for _, step := range result.Steps {
		if step.Status == gowright.TestStatusPassed {
			passed++
		} else if step.Status == gowright.TestStatusFailed {
			failed++
		}
	}
	
	fmt.Printf("\nAssertion Summary:\n")
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	
	fmt.Printf("\nDetailed Assertion Steps:\n")
	for i, step := range result.Steps {
		status := "✓"
		if step.Status == gowright.TestStatusFailed {
			status = "✗"
		}
		fmt.Printf("%d. %s %s: %s (%v)\n", i+1, status, step.Name, step.Description, step.Duration)
		if step.Error != nil {
			fmt.Printf("   Error: %s\n", step.Error.Error())
		}
	}
	
	fmt.Printf("\nReports generated in: %s\n", config.LocalReports.OutputDir)
	fmt.Println("Open the HTML report to see detailed assertion steps with styling!")
}