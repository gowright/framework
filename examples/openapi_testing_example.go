package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gowright/framework/pkg/core"
	"github.com/gowright/framework/pkg/openapi"
)

func main() {
	// Example 1: Basic OpenAPI validation
	fmt.Println("=== Example 1: Basic OpenAPI Validation ===")
	basicValidationExample()

	fmt.Println("\n=== Example 2: Comprehensive OpenAPI Testing ===")
	comprehensiveTestingExample()

	fmt.Println("\n=== Example 3: Breaking Changes Detection ===")
	breakingChangesExample()

	fmt.Println("\n=== Example 4: Using Test Builder Pattern ===")
	testBuilderExample()

	fmt.Println("\n=== Example 5: Integration with GoWright Framework ===")
	gowrightIntegrationExample()
}

func basicValidationExample() {
	// Create a sample OpenAPI spec file
	specContent := `
openapi: 3.0.3
info:
  title: Sample API
  version: 1.0.0
  description: A sample API for demonstration
paths:
  /users:
    get:
      summary: Get all users
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
          format: email
`

	// Write spec to temporary file
	tmpFile := "temp_openapi.yaml"
	if err := os.WriteFile(tmpFile, []byte(specContent), 0644); err != nil {
		log.Fatalf("Failed to write temp spec file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Create OpenAPI tester
	tester, err := openapi.NewOpenAPITester(tmpFile)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI tester: %v", err)
	}

	// Run validation
	result := tester.ValidateSpec()
	fmt.Printf("Validation Result: %s\n", result.Message)
	fmt.Printf("Passed: %t\n", result.Passed)

	if len(result.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Path, err.Message)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s: %s\n", warning.Path, warning.Message)
		}
	}
}

func comprehensiveTestingExample() {
	// Create a more complex OpenAPI spec
	specContent := `
openapi: 3.0.3
info:
  title: E-commerce API
  version: 2.0.0
  description: A comprehensive e-commerce API
servers:
  - url: https://api.example.com/v2
paths:
  /products:
    get:
      summary: List products
      parameters:
        - name: category
          in: query
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
      responses:
        '200':
          description: Products retrieved successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  products:
                    type: array
                    items:
                      $ref: '#/components/schemas/Product'
                  total:
                    type: integer
        '400':
          description: Bad request
    post:
      summary: Create a new product
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '201':
          description: Product created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Product'
        '400':
          description: Invalid input
        '401':
          description: Unauthorized
  /products/{id}:
    get:
      summary: Get product by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Product found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Product'
        '404':
          description: Product not found
components:
  schemas:
    Product:
      type: object
      required:
        - id
        - name
        - price
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        price:
          type: number
          format: float
          minimum: 0
        category:
          type: string
        inStock:
          type: boolean
        createdAt:
          type: string
          format: date-time
    CreateProductRequest:
      type: object
      required:
        - name
        - price
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 255
        description:
          type: string
          maxLength: 1000
        price:
          type: number
          format: float
          minimum: 0
        category:
          type: string
        inStock:
          type: boolean
          default: true
`

	tmpFile := "temp_comprehensive.yaml"
	if err := os.WriteFile(tmpFile, []byte(specContent), 0644); err != nil {
		log.Fatalf("Failed to write temp spec file: %v", err)
	}
	defer os.Remove(tmpFile)

	tester, err := openapi.NewOpenAPITester(tmpFile)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI tester: %v", err)
	}

	// Run all tests
	results := tester.RunAllTests("")

	fmt.Printf("Total tests run: %d\n", len(results))
	for _, result := range results {
		fmt.Printf("\nTest: %s\n", result.TestName)
		fmt.Printf("Status: %s\n", getStatusString(result.Passed))
		fmt.Printf("Message: %s\n", result.Message)

		if len(result.Details) > 0 {
			fmt.Println("Details:")
			for _, detail := range result.Details {
				fmt.Printf("  - %s\n", detail)
			}
		}
	}

	// Print summary
	summary := tester.GetSummary(results)
	fmt.Printf("\n%s\n", summary)
}

func breakingChangesExample() {
	// This example demonstrates breaking changes detection
	// Note: In a real scenario, you would have actual git history

	fmt.Println("Breaking changes detection requires git history.")
	fmt.Println("This example shows how to use the API:")

	specContent := `
openapi: 3.0.3
info:
  title: API with Changes
  version: 2.0.0
paths:
  /users:
    get:
      summary: Get users
      parameters:
        - name: limit
          in: query
          required: true  # This could be a breaking change if it was optional before
          schema:
            type: integer
      responses:
        '200':
          description: Success
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
        - email  # This could be a breaking change if it was optional before
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
`

	tmpFile := "temp_breaking_changes.yaml"
	if err := os.WriteFile(tmpFile, []byte(specContent), 0644); err != nil {
		log.Fatalf("Failed to write temp spec file: %v", err)
	}
	defer os.Remove(tmpFile)

	tester, err := openapi.NewOpenAPITester(tmpFile)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI tester: %v", err)
	}

	// Attempt to check breaking changes (will likely fail without proper git setup)
	result := tester.CheckBreakingChanges("HEAD~1")
	fmt.Printf("Breaking changes check: %s\n", result.Message)
	fmt.Printf("Status: %s\n", getStatusString(result.Passed))

	if len(result.Details) > 0 {
		fmt.Println("Breaking changes found:")
		for _, detail := range result.Details {
			fmt.Printf("  - %s\n", detail)
		}
	}
}

func testBuilderExample() {
	specContent := `
openapi: 3.0.3
info:
  title: Builder Pattern API
  version: 1.0.0
paths:
  /health:
    get:
      responses:
        '200':
          description: Health check
`

	tmpFile := "temp_builder.yaml"
	if err := os.WriteFile(tmpFile, []byte(specContent), 0644); err != nil {
		log.Fatalf("Failed to write temp spec file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Use the builder pattern to create a customized test suite
	suite, err := openapi.NewOpenAPITestBuilder(tmpFile).
		WithValidation(true).
		WithCircularReferenceDetection(true).
		WithBreakingChangesDetection(false, ""). // Disable breaking changes for this example
		Build()

	if err != nil {
		log.Fatalf("Failed to build test suite: %v", err)
	}

	fmt.Printf("Created test suite: %s\n", suite.Name)
	fmt.Printf("Number of tests: %d\n", len(suite.Tests))

	for i, test := range suite.Tests {
		fmt.Printf("  %d. %s\n", i+1, test.GetName())
	}
}

func gowrightIntegrationExample() {
	specContent := `
openapi: 3.0.3
info:
  title: GoWright Integration API
  version: 1.0.0
paths:
  /api/test:
    get:
      responses:
        '200':
          description: Test endpoint
components:
  schemas:
    TestResponse:
      type: object
      properties:
        message:
          type: string
`

	tmpFile := "temp_integration.yaml"
	if err := os.WriteFile(tmpFile, []byte(specContent), 0644); err != nil {
		log.Fatalf("Failed to write temp spec file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Create OpenAPI integration
	integration, err := openapi.NewOpenAPIIntegration(tmpFile)
	if err != nil {
		log.Fatalf("Failed to create OpenAPI integration: %v", err)
	}

	// Create a full test suite
	suite := integration.CreateFullTestSuite("")

	fmt.Printf("Running OpenAPI test suite: %s\n", suite.Name)

	// Run each test in the suite
	for _, test := range suite.Tests {
		fmt.Printf("\nRunning test: %s\n", test.GetName())

		result := test.Execute()

		if result.Error != nil {
			fmt.Printf("Test execution error: %v\n", result.Error)
		}

		if result.Status == core.TestStatusPassed {
			fmt.Println("Test passed successfully!")
		} else {
			fmt.Printf("Test failed with status: %v\n", result.Status)
		}

		if len(result.Logs) > 0 {
			fmt.Println("Test logs:")
			for _, log := range result.Logs {
				fmt.Printf("  - %s\n", log)
			}
		}
	}

	fmt.Printf("\nOpenAPI test suite execution completed\n")
}

func getStatusString(passed bool) string {
	if passed {
		return "✅ PASSED"
	}
	return "❌ FAILED"
}
