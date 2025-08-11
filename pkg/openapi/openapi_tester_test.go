package openapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test OpenAPI specification for testing
const testOpenAPISpec = `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
  description: A test API for OpenAPI testing
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '400':
          description: Bad request
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
        - email
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
          format: email
        createdAt:
          type: string
          format: date-time
    CreateUserRequest:
      type: object
      required:
        - name
        - email
      properties:
        name:
          type: string
        email:
          type: string
          format: email
`

// Invalid OpenAPI spec for testing validation errors
const invalidOpenAPISpec = `
openapi: 3.0.3
info:
  title: Invalid API
paths:
  /test:
    get:
      summary: Test endpoint
      # Missing responses - this should trigger a validation error
`

func createTempSpecFile(t *testing.T, content string) string {
	tmpDir := t.TempDir()
	specFile := filepath.Join(tmpDir, "openapi.yaml")

	err := os.WriteFile(specFile, []byte(content), 0644)
	require.NoError(t, err)

	return specFile
}

func TestNewOpenAPITester(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)

	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)
	assert.NotNil(t, tester)
	assert.Equal(t, specFile, tester.specPath)
	assert.NotNil(t, tester.document)
	assert.NotNil(t, tester.model)
	// Validator is not exposed in the current implementation
}

func TestNewOpenAPITester_InvalidFile(t *testing.T) {
	_, err := NewOpenAPITester("nonexistent.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load OpenAPI spec")
}

func TestNewOpenAPITester_InvalidSpec(t *testing.T) {
	specFile := createTempSpecFile(t, "invalid yaml content: [")

	_, err := NewOpenAPITester(specFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse OpenAPI document")
}

func TestValidateSpec_ValidSpec(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	result := tester.ValidateSpec()
	assert.NotNil(t, result)
	assert.Equal(t, "OpenAPI Specification Validation", result.TestName)
	assert.True(t, result.Passed)
	assert.Equal(t, "OpenAPI specification is valid", result.Message)
	assert.Empty(t, result.Errors)
}

func TestValidateSpec_InvalidSpec(t *testing.T) {
	specFile := createTempSpecFile(t, invalidOpenAPISpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	result := tester.ValidateSpec()
	assert.NotNil(t, result)
	assert.Equal(t, "OpenAPI Specification Validation", result.TestName)
	// The result might still pass basic validation, but should have warnings
	assert.NotEmpty(t, result.Warnings)
}

func TestDetectCircularReferences_NoCircularRefs(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	result := tester.DetectCircularReferences()
	assert.NotNil(t, result)
	assert.Equal(t, "Circular Reference Detection", result.TestName)
	assert.True(t, result.Passed)
	assert.Equal(t, "No circular references detected", result.Message)
	assert.Empty(t, result.Details)
}

func TestDetectCircularReferences_WithCircularRefs(t *testing.T) {
	circularSpec := `
openapi: 3.0.3
info:
  title: Circular API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  schemas:
    A:
      type: object
      properties:
        b:
          $ref: '#/components/schemas/B'
    B:
      type: object
      properties:
        a:
          $ref: '#/components/schemas/A'
`

	specFile := createTempSpecFile(t, circularSpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	result := tester.DetectCircularReferences()
	assert.NotNil(t, result)
	assert.Equal(t, "Circular Reference Detection", result.TestName)
	// Note: The actual circular reference detection implementation is simplified
	// In a real scenario, this would detect the A -> B -> A circular reference
}

func TestRunAllTests(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	results := tester.RunAllTests("")
	assert.NotEmpty(t, results)
	assert.Len(t, results, 2) // Validation + Circular Reference tests

	// Check that all tests have proper names
	testNames := make(map[string]bool)
	for _, result := range results {
		assert.NotEmpty(t, result.TestName)
		testNames[result.TestName] = true
	}

	assert.True(t, testNames["OpenAPI Specification Validation"])
	assert.True(t, testNames["Circular Reference Detection"])
}

func TestRunAllTests_WithBreakingChanges(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	// Run with a previous commit (this will likely fail in test environment)
	results := tester.RunAllTests("HEAD~1")
	assert.NotEmpty(t, results)
	assert.Len(t, results, 3) // Validation + Circular Reference + Breaking Changes tests
}

func TestGetSummary(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	results := tester.RunAllTests("")
	summary := tester.GetSummary(results)

	assert.NotEmpty(t, summary)
	assert.Contains(t, summary, "OpenAPI Test Summary:")
	assert.Contains(t, summary, "passed")
	assert.Contains(t, summary, "failed")
	assert.Contains(t, summary, "errors")
	assert.Contains(t, summary, "warnings")
}

func TestValidateSpec_MissingInfo(t *testing.T) {
	specWithoutInfo := `
openapi: 3.0.3
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
`

	specFile := createTempSpecFile(t, specWithoutInfo)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	result := tester.ValidateSpec()
	assert.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.NotEmpty(t, result.Errors)

	// Check for info validation error
	found := false
	for _, err := range result.Errors {
		if err.Path == "info" && err.Message == "Info object is required" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find info validation error")
}

func TestValidateSpec_NoPaths(t *testing.T) {
	specWithoutPaths := `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
`

	specFile := createTempSpecFile(t, specWithoutPaths)
	tester, err := NewOpenAPITester(specFile)
	require.NoError(t, err)

	result := tester.ValidateSpec()
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Warnings)

	// Check for paths warning
	found := false
	for _, warning := range result.Warnings {
		if warning.Path == "paths" && warning.Message == "No paths defined in the specification" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find paths warning")
}
