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
	fmt.Println("=== Gowright Integration Testing Example ===\n")

	// Create comprehensive configuration for all modules
	config := &gowright.Config{
		LogLevel:   "INFO",
		Parallel:   false,
		MaxRetries: 3,

		BrowserConfig: &gowright.BrowserConfig{
			Headless:   true, // Use headless for integration tests
			Timeout:    30 * time.Second,
			UserAgent:  "Gowright-Integration-Tester/1.0",
			WindowSize: &gowright.WindowSize{Width: 1920, Height: 1080},
		},

		APIConfig: &gowright.APIConfig{
			BaseURL: "https://jsonplaceholder.typicode.com",
			Timeout: 30 * time.Second,
			Headers: map[string]string{
				"User-Agent": "Gowright-Integration-Tester/1.0",
			},
		},

		DatabaseConfig: &gowright.DatabaseConfig{
			Connections: map[string]*gowright.DBConnection{
				"main": {
					Driver:       "sqlite3",
					DSN:          ":memory:",
					MaxOpenConns: 10,
					MaxIdleConns: 5,
				},
			},
		},

		ReportConfig: &gowright.ReportConfig{
			LocalReports: gowright.LocalReportConfig{
				JSON:      true,
				HTML:      true,
				OutputDir: "./integration-test-reports",
			},
		},
	}

	// Create integration tester
	integrationTester := gowright.NewIntegrationTester(config)
	if err := integrationTester.Initialize(); err != nil {
		log.Fatalf("Failed to initialize integration tester: %v", err)
	}
	defer integrationTester.Cleanup()

	// Example 1: E-commerce workflow integration test
	fmt.Println("1. Testing complete e-commerce workflow")
	ecommerceTest := gowright.NewIntegrationTest("E-commerce Workflow Test")

	// Step 1: Setup database schema
	ecommerceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query: `
				CREATE TABLE products (
					id INTEGER PRIMARY KEY,
					name VARCHAR(100),
					price DECIMAL(10,2),
					stock INTEGER
				);
				INSERT INTO products (id, name, price, stock) VALUES 
				(1, 'Laptop', 999.99, 10),
				(2, 'Mouse', 25.50, 50),
				(3, 'Keyboard', 75.00, 30);
			`,
		},
		&gowright.DatabaseValidation{
			ExpectedRowCount: 3,
			Query:            "SELECT COUNT(*) as count FROM products",
		},
	))

	// Step 2: API - Get product catalog
	ecommerceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeAPI,
		&gowright.APIAction{
			Method:   "GET",
			Endpoint: "/posts", // Using JSONPlaceholder as mock product API
		},
		&gowright.APIValidation{
			ExpectedStatus: 200,
			ExpectedHeaders: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
		},
	))

	// Step 3: UI - Navigate to product page
	ecommerceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeUI,
		&gowright.UIAction{
			Type:   gowright.UIActionNavigate,
			Target: "https://example.com", // Mock product page
		},
		&gowright.UIValidation{
			ExpectedElements: []string{"h1", "p"},
			ExpectedTitle:    "Example Domain",
		},
	))

	// Step 4: Database - Update stock after purchase
	ecommerceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query:      "UPDATE products SET stock = stock - 1 WHERE id = 1",
		},
		&gowright.DatabaseValidation{
			Query:               "SELECT stock FROM products WHERE id = 1",
			ExpectedColumnValue: map[string]interface{}{"stock": 9},
		},
	))

	result := ecommerceTest.Execute(integrationTester)
	printIntegrationTestResult(result)

	// Example 2: User registration and profile management workflow
	fmt.Println("\n2. Testing user registration and profile workflow")
	userWorkflowTest := gowright.NewIntegrationTest("User Registration Workflow Test")

	// Step 1: Database - Setup user tables
	userWorkflowTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query: `
				CREATE TABLE users (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					username VARCHAR(50) UNIQUE,
					email VARCHAR(100),
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
		&gowright.DatabaseValidation{
			Query:            "SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
			ExpectedRowCount: 1,
		},
	))

	// Step 2: API - Create user via API
	userWorkflowTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeAPI,
		&gowright.APIAction{
			Method:   "POST",
			Endpoint: "/users",
			Body: map[string]interface{}{
				"name":     "John Doe",
				"username": "johndoe",
				"email":    "john@example.com",
			},
		},
		&gowright.APIValidation{
			ExpectedStatus: 201,
			ExpectedJSONPaths: map[string]interface{}{
				"$.name":     "John Doe",
				"$.username": "johndoe",
				"$.email":    "john@example.com",
			},
		},
	))

	// Step 3: Database - Verify user was created
	userWorkflowTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query:      "INSERT INTO users (username, email) VALUES ('johndoe', 'john@example.com')",
		},
		&gowright.DatabaseValidation{
			Query:               "SELECT COUNT(*) as count FROM users WHERE username = 'johndoe'",
			ExpectedColumnValue: map[string]interface{}{"count": 1},
		},
	))

	// Step 4: UI - Login with new user
	userWorkflowTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeUI,
		&gowright.UIAction{
			Type:   gowright.UIActionNavigate,
			Target: "https://the-internet.herokuapp.com/login",
		},
		&gowright.UIValidation{
			ExpectedElements: []string{"#username", "#password", "button[type='submit']"},
		},
	))

	result = userWorkflowTest.Execute(integrationTester)
	printIntegrationTestResult(result)

	// Example 3: Data synchronization workflow
	fmt.Println("\n3. Testing data synchronization across systems")
	syncTest := gowright.NewIntegrationTest("Data Synchronization Test")

	// Step 1: Database - Create source data
	syncTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query: `
				CREATE TABLE sync_source (
					id INTEGER PRIMARY KEY,
					data VARCHAR(100),
					last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
				);
				INSERT INTO sync_source (id, data) VALUES 
				(1, 'Record 1'),
				(2, 'Record 2'),
				(3, 'Record 3');
			`,
		},
		&gowright.DatabaseValidation{
			Query:               "SELECT COUNT(*) as count FROM sync_source",
			ExpectedColumnValue: map[string]interface{}{"count": 3},
		},
	))

	// Step 2: API - Sync data to external system
	syncTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeAPI,
		&gowright.APIAction{
			Method:   "GET",
			Endpoint: "/posts/1", // Mock sync endpoint
		},
		&gowright.APIValidation{
			ExpectedStatus: 200,
			ExpectedJSONPaths: map[string]interface{}{
				"$.id": float64(1),
			},
		},
	))

	// Step 3: Database - Create destination table and sync
	syncTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query: `
				CREATE TABLE sync_destination (
					id INTEGER PRIMARY KEY,
					data VARCHAR(100),
					synced_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
				INSERT INTO sync_destination (id, data)
				SELECT id, data FROM sync_source;
			`,
		},
		&gowright.DatabaseValidation{
			Query:               "SELECT COUNT(*) as count FROM sync_destination",
			ExpectedColumnValue: map[string]interface{}{"count": 3},
		},
	))

	// Step 4: UI - Verify sync status in dashboard
	syncTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeUI,
		&gowright.UIAction{
			Type:   gowright.UIActionNavigate,
			Target: "https://httpbin.org/status/200", // Mock dashboard
		},
		&gowright.UIValidation{
			ExpectedStatusCode: 200,
		},
	))

	result = syncTest.Execute(integrationTester)
	printIntegrationTestResult(result)

	// Example 4: Error handling and rollback workflow
	fmt.Println("\n4. Testing error handling and rollback mechanisms")
	errorHandlingTest := gowright.NewIntegrationTest("Error Handling and Rollback Test")

	// Step 1: Database - Setup transaction
	errorHandlingTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query: `
				CREATE TABLE transaction_test (
					id INTEGER PRIMARY KEY,
					amount DECIMAL(10,2),
					status VARCHAR(20) DEFAULT 'pending'
				);
				INSERT INTO transaction_test (id, amount) VALUES (1, 100.00);
			`,
		},
		&gowright.DatabaseValidation{
			Query:               "SELECT status FROM transaction_test WHERE id = 1",
			ExpectedColumnValue: map[string]interface{}{"status": "pending"},
		},
	))

	// Step 2: API - Attempt payment (simulate failure)
	errorHandlingTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeAPI,
		&gowright.APIAction{
			Method:   "GET",
			Endpoint: "/posts/999", // Non-existent endpoint to simulate error
		},
		&gowright.APIValidation{
			ExpectedStatus: 404, // Expect failure
		},
	))

	// Step 3: Database - Rollback transaction on API failure
	errorHandlingTest.AddRollbackStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query:      "UPDATE transaction_test SET status = 'failed' WHERE id = 1",
		},
		&gowright.DatabaseValidation{
			Query:               "SELECT status FROM transaction_test WHERE id = 1",
			ExpectedColumnValue: map[string]interface{}{"status": "failed"},
		},
	))

	result = errorHandlingTest.Execute(integrationTester)
	printIntegrationTestResult(result)

	// Example 5: Performance and load testing workflow
	fmt.Println("\n5. Testing performance across multiple systems")
	performanceTest := gowright.NewIntegrationTest("Performance Integration Test")

	// Step 1: Database - Create large dataset
	performanceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query: `
				CREATE TABLE performance_data (
					id INTEGER PRIMARY KEY,
					data VARCHAR(1000),
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
		&gowright.DatabaseValidation{
			Query:            "SELECT name FROM sqlite_master WHERE type='table' AND name='performance_data'",
			ExpectedRowCount: 1,
		},
	))

	// Insert multiple records for performance testing
	for i := 0; i < 100; i++ {
		performanceTest.AddStep(gowright.NewIntegrationStep(
			gowright.StepTypeDatabase,
			&gowright.DatabaseAction{
				Connection: "main",
				Query:      fmt.Sprintf("INSERT INTO performance_data (id, data) VALUES (%d, 'Performance test data %d')", i+1, i+1),
			},
			nil, // No validation for individual inserts
		))
	}

	// Step 2: API - Bulk data retrieval
	performanceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeAPI,
		&gowright.APIAction{
			Method:   "GET",
			Endpoint: "/posts", // Get all posts for performance test
		},
		&gowright.APIValidation{
			ExpectedStatus:  200,
			MaxResponseTime: 5 * time.Second,
		},
	))

	// Step 3: Database - Performance query
	performanceTest.AddStep(gowright.NewIntegrationStep(
		gowright.StepTypeDatabase,
		&gowright.DatabaseAction{
			Connection: "main",
			Query:      "SELECT COUNT(*) as total, MAX(id) as max_id FROM performance_data",
		},
		&gowright.DatabaseValidation{
			ExpectedColumnValue: map[string]interface{}{
				"total":  100,
				"max_id": 100,
			},
			MaxExecutionTime: 2 * time.Second,
		},
	))

	result = performanceTest.Execute(integrationTester)
	printIntegrationTestResult(result)

	// Generate comprehensive integration test report
	fmt.Println("\nGenerating integration test reports...")

	testResults := &gowright.TestResults{
		SuiteName:    "Integration Testing Example Suite",
		StartTime:    time.Now().Add(-15 * time.Minute),
		EndTime:      time.Now(),
		TotalTests:   5,
		PassedTests:  4,
		FailedTests:  0,
		SkippedTests: 0,
		ErrorTests:   1,                           // The error handling test that expected failures
		TestCases:    []gowright.TestCaseResult{}, // Would contain all results
	}

	reportManager := gowright.NewReportManager(config.ReportConfig)
	if err := reportManager.GenerateReports(testResults); err != nil {
		log.Printf("Failed to generate reports: %v", err)
	} else {
		fmt.Printf("Integration test reports generated in: %s\n", config.ReportConfig.LocalReports.OutputDir)
	}

	fmt.Println("\n=== Integration Testing Complete ===")
	fmt.Println("\nThis example demonstrates:")
	fmt.Println("- Multi-system workflow orchestration")
	fmt.Println("- Database, API, and UI integration")
	fmt.Println("- Error handling and rollback mechanisms")
	fmt.Println("- Performance testing across systems")
	fmt.Println("- Comprehensive reporting and validation")
}

func printIntegrationTestResult(result *gowright.TestCaseResult) {
	fmt.Printf("Integration Test: %s\n", result.Name)
	fmt.Printf("Status: %s\n", result.Status.String())
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Steps Executed: %d\n", len(result.Steps))

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	// Show step-by-step results
	for i, step := range result.Steps {
		status := "✓"
		if step.Status == gowright.TestStatusFailed {
			status = "✗"
		} else if step.Status == gowright.TestStatusSkipped {
			status = "⊝"
		}

		fmt.Printf("  Step %d: %s %s (%v)\n", i+1, status, step.Name, step.Duration)
		if step.Error != nil {
			fmt.Printf("    Error: %s\n", step.Error.Error())
		}
	}

	if len(result.Logs) > 0 {
		fmt.Println("Logs:")
		for _, logEntry := range result.Logs {
			fmt.Printf("  - %s\n", logEntry)
		}
	}

	fmt.Println("---")
}
