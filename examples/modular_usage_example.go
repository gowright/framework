package main

import (
	"fmt"
	"log"

	"github.com/gowright/framework/pkg/gowright"
)

func main() {
	// Example 1: Using the main package with all testers
	fmt.Println("=== Example 1: Complete Framework Setup ===")

	// Create configuration
	config := gowright.DefaultConfig()
	config.BrowserConfig.Headless = true
	config.APIConfig.BaseURL = "https://api.example.com"

	// Create Gowright instance with all testers
	gw := gowright.NewGowrightWithAllTesters(config)

	// Initialize the framework
	if err := gw.Initialize(); err != nil {
		log.Fatalf("Failed to initialize framework: %v", err)
	}
	defer gw.Close()

	fmt.Printf("Framework initialized with UI, API, Database, and Integration testers\n")
	fmt.Printf("UI Tester: %s\n", gw.GetUITester().GetName())
	fmt.Printf("API Tester: %s\n", gw.GetAPITester().GetName())
	fmt.Printf("Database Tester: %s\n", gw.GetDatabaseTester().GetName())
	fmt.Printf("Integration Tester: %s\n", gw.GetIntegrationTester().GetName())

	// Example 2: Using individual testers
	fmt.Println("\n=== Example 2: Individual Tester Usage ===")

	// Create individual testers
	uiTester := gowright.NewUITester()
	apiTester := gowright.NewAPITester()
	dbTester := gowright.NewDatabaseTester()

	// Initialize them individually
	if err := uiTester.Initialize(config.BrowserConfig); err != nil {
		log.Printf("Failed to initialize UI tester: %v", err)
	} else {
		fmt.Printf("UI Tester initialized: %s\n", uiTester.GetName())
	}

	if err := apiTester.Initialize(config.APIConfig); err != nil {
		log.Printf("Failed to initialize API tester: %v", err)
	} else {
		fmt.Printf("API Tester initialized: %s\n", apiTester.GetName())
	}

	if err := dbTester.Initialize(config.DatabaseConfig); err != nil {
		log.Printf("Failed to initialize Database tester: %v", err)
	} else {
		fmt.Printf("Database Tester initialized: %s\n", dbTester.GetName())
	}

	// Example 3: Integration Testing
	fmt.Println("\n=== Example 3: Integration Test Example ===")

	// Create an integration test
	integrationTest := &gowright.IntegrationTest{
		Name: "User Registration Flow",
		Steps: []gowright.IntegrationStep{
			{
				Name: "Navigate to registration page",
				Type: gowright.StepTypeUI,
				Action: &gowright.UIStepAction{
					Type: "navigate",
					Parameters: map[string]interface{}{
						"url": "https://example.com/register",
					},
				},
			},
			{
				Name: "Check API health",
				Type: gowright.StepTypeAPI,
				Action: &gowright.APIStepAction{
					Method:   "GET",
					Endpoint: "/health",
				},
				Validation: &gowright.APIStepValidation{
					ExpectedStatusCode: 200,
				},
			},
			{
				Name: "Verify user table is ready",
				Type: gowright.StepTypeDatabase,
				Action: &gowright.DatabaseStepAction{
					Connection: "default",
					Query:      "SELECT COUNT(*) FROM users",
				},
			},
		},
	}

	// Execute the integration test
	result := gw.ExecuteIntegrationTest(integrationTest)
	fmt.Printf("Integration test '%s' completed with status: %s\n",
		result.Name, result.Status.String())

	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Steps executed: %d\n", len(result.Steps))

	// Example 4: Individual Test Types
	fmt.Println("\n=== Example 4: Individual Test Types ===")

	// UI Test
	uiTest := &gowright.UITest{
		Name: "Login Test",
		URL:  "https://example.com/login",
		Actions: []gowright.UIAction{
			{Type: "type", Selector: "#username", Value: "testuser"},
			{Type: "type", Selector: "#password", Value: "testpass"},
			{Type: "click", Selector: "#login-btn"},
		},
		Assertions: []gowright.UIAssertion{
			{Type: "text_contains", Selector: ".welcome", Expected: "Welcome"},
		},
	}

	uiResult := gw.ExecuteUITest(uiTest)
	fmt.Printf("UI test '%s' status: %s\n", uiResult.Name, uiResult.Status.String())

	// API Test
	apiTest := &gowright.APITest{
		Name:     "Get User API Test",
		Method:   "GET",
		Endpoint: "/api/users/1",
		Expected: &gowright.APIExpectation{
			StatusCode: 200,
		},
	}

	apiResult := gw.ExecuteAPITest(apiTest)
	fmt.Printf("API test '%s' status: %s\n", apiResult.Name, apiResult.Status.String())

	// Database Test
	dbTest := &gowright.DatabaseTest{
		Name:       "User Count Test",
		Connection: "default",
		Query:      "SELECT COUNT(*) as count FROM users",
		Expected: &gowright.DatabaseExpectation{
			RowCount: 1,
		},
	}

	dbResult := gw.ExecuteDatabaseTest(dbTest)
	fmt.Printf("Database test '%s' status: %s\n", dbResult.Name, dbResult.Status.String())

	fmt.Println("\n=== Modular Structure Benefits ===")
	fmt.Println("✓ Separated concerns into focused packages")
	fmt.Println("✓ Integration package orchestrates all other packages")
	fmt.Println("✓ Each package can be used independently")
	fmt.Println("✓ Easier to maintain and extend")
	fmt.Println("✓ Better testability and modularity")
	fmt.Println("✓ Backward compatibility maintained through main package")
}
