package main

import (
	"fmt"
	"log"
	"time"

	"github/gowright/framework/pkg/gowright"
)

func main() {
	// Create a configuration with both JSON and HTML reporting enabled
	config := &gowright.Config{
		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: "./example-reports",
			},
		},
	}

	// Create a report manager
	reportManager := gowright.NewReportManager(config.ReportConfig)

	// Create sample test results
	startTime := time.Now()
	testResults := &gowright.TestResults{
		SuiteName:    "Example Test Suite",
		StartTime:    startTime,
		EndTime:      startTime.Add(time.Minute * 2),
		TotalTests:   4,
		PassedTests:  2,
		FailedTests:  1,
		SkippedTests: 1,
		ErrorTests:   0,
		TestCases: []gowright.TestCaseResult{
			{
				Name:        "Login Test",
				Status:      gowright.TestStatusPassed,
				Duration:    time.Second * 15,
				StartTime:   startTime,
				EndTime:     startTime.Add(time.Second * 15),
				Screenshots: []string{"login_success.png"},
				Logs:        []string{"User logged in successfully", "Dashboard loaded"},
			},
			{
				Name:      "API Health Check",
				Status:    gowright.TestStatusPassed,
				Duration:  time.Second * 5,
				StartTime: startTime.Add(time.Second * 15),
				EndTime:   startTime.Add(time.Second * 20),
				Logs:      []string{"API responded with 200 OK", "Health check passed"},
			},
			{
				Name:        "Database Connection Test",
				Status:      gowright.TestStatusFailed,
				Duration:    time.Second * 30,
				Error:       fmt.Errorf("connection timeout after 30 seconds"),
				StartTime:   startTime.Add(time.Second * 20),
				EndTime:     startTime.Add(time.Second * 50),
				Screenshots: []string{"db_error_screen.png"},
				Logs:        []string{"Attempting database connection", "Connection timeout", "Test failed"},
			},
			{
				Name:      "Performance Test",
				Status:    gowright.TestStatusSkipped,
				Duration:  0,
				StartTime: startTime.Add(time.Second * 50),
				EndTime:   startTime.Add(time.Second * 50),
				Logs:      []string{"Test skipped due to missing performance baseline"},
			},
		},
	}

	// Generate reports
	fmt.Println("Generating test reports...")
	if err := reportManager.GenerateReports(testResults); err != nil {
		log.Fatalf("Failed to generate reports: %v", err)
	}

	fmt.Println("Reports generated successfully!")
	fmt.Printf("- JSON report: %s\n", config.ReportConfig.LocalReports.OutputDir)
	fmt.Printf("- HTML report: %s\n", config.ReportConfig.LocalReports.OutputDir)
	
	// Display summary
	fmt.Printf("\nTest Summary:\n")
	fmt.Printf("Suite: %s\n", testResults.SuiteName)
	fmt.Printf("Total Tests: %d\n", testResults.TotalTests)
	fmt.Printf("Passed: %d\n", testResults.PassedTests)
	fmt.Printf("Failed: %d\n", testResults.FailedTests)
	fmt.Printf("Skipped: %d\n", testResults.SkippedTests)
	fmt.Printf("Duration: %v\n", testResults.EndTime.Sub(testResults.StartTime))
}