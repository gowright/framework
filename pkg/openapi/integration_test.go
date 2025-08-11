package openapi

import (
	"testing"

	"github.com/gowright/framework/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAPIIntegration(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)

	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)
	assert.NotNil(t, integration)
	assert.NotNil(t, integration.tester)
}

func TestNewOpenAPIIntegration_InvalidSpec(t *testing.T) {
	_, err := NewOpenAPIIntegration("nonexistent.yaml")
	assert.Error(t, err)
}

func TestCreateValidationTest(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	testCase := integration.CreateValidationTest()
	assert.Equal(t, "OpenAPI Specification Validation", testCase.GetName())
	assert.NotNil(t, testCase)
}

func TestCreateValidationTest_Execution(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	testCase := integration.CreateValidationTest()
	result := testCase.Execute()

	assert.NotNil(t, result)
	assert.Equal(t, "OpenAPI Specification Validation", result.Name)
	assert.Equal(t, core.TestStatusPassed, result.Status) // Valid spec should pass
}

func TestCreateValidationTest_ExecutionWithInvalidSpec(t *testing.T) {
	specFile := createTempSpecFile(t, invalidOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	testCase := integration.CreateValidationTest()
	result := testCase.Execute()

	assert.NotNil(t, result)
	assert.Equal(t, "OpenAPI Specification Validation", result.Name)
	// The result might still pass basic validation, but should have logs
	assert.NotEmpty(t, result.Logs)
}

func TestCreateCircularReferenceTest(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	testCase := integration.CreateCircularReferenceTest()
	assert.Equal(t, "OpenAPI Circular Reference Detection", testCase.GetName())
	assert.NotNil(t, testCase)
}

func TestCreateCircularReferenceTest_Execution(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	testCase := integration.CreateCircularReferenceTest()
	result := testCase.Execute()

	assert.NotNil(t, result)
	assert.Equal(t, "OpenAPI Circular Reference Detection", result.Name)
	assert.Equal(t, core.TestStatusPassed, result.Status) // No circular references should be found
}

func TestCreateBreakingChangesTest(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	testCase := integration.CreateBreakingChangesTest("HEAD~1")
	assert.Equal(t, "OpenAPI Breaking Changes Detection", testCase.GetName())
	assert.NotNil(t, testCase)
}

func TestCreateFullTestSuite(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	suite := integration.CreateFullTestSuite("")
	assert.NotNil(t, suite)
	assert.Equal(t, "OpenAPI Comprehensive Testing", suite.Name)
	assert.Len(t, suite.Tests, 2) // Validation + Circular Reference
}

func TestCreateFullTestSuite_WithBreakingChanges(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)
	integration, err := NewOpenAPIIntegration(specFile)
	require.NoError(t, err)

	suite := integration.CreateFullTestSuite("HEAD~1")
	assert.NotNil(t, suite)
	assert.Equal(t, "OpenAPI Comprehensive Testing", suite.Name)
	assert.Len(t, suite.Tests, 3) // Validation + Circular Reference + Breaking Changes
}

func TestNewOpenAPITestBuilder(t *testing.T) {
	builder := NewOpenAPITestBuilder("test.yaml")
	assert.NotNil(t, builder)
	assert.Equal(t, "test.yaml", builder.specPath)
	assert.True(t, builder.includeValidation)
	assert.True(t, builder.includeCircularRef)
	assert.False(t, builder.includeBreakingChanges)
}

func TestOpenAPITestBuilder_WithPreviousCommit(t *testing.T) {
	builder := NewOpenAPITestBuilder("test.yaml")
	builder = builder.WithPreviousCommit("HEAD~1")

	assert.Equal(t, "HEAD~1", builder.previousCommit)
	assert.True(t, builder.includeBreakingChanges)
}

func TestOpenAPITestBuilder_WithValidation(t *testing.T) {
	builder := NewOpenAPITestBuilder("test.yaml")
	builder = builder.WithValidation(false)

	assert.False(t, builder.includeValidation)
}

func TestOpenAPITestBuilder_WithCircularReferenceDetection(t *testing.T) {
	builder := NewOpenAPITestBuilder("test.yaml")
	builder = builder.WithCircularReferenceDetection(false)

	assert.False(t, builder.includeCircularRef)
}

func TestOpenAPITestBuilder_WithBreakingChangesDetection(t *testing.T) {
	builder := NewOpenAPITestBuilder("test.yaml")
	builder = builder.WithBreakingChangesDetection(true, "HEAD~2")

	assert.True(t, builder.includeBreakingChanges)
	assert.Equal(t, "HEAD~2", builder.previousCommit)
}

func TestOpenAPITestBuilder_Build(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)

	builder := NewOpenAPITestBuilder(specFile)
	suite, err := builder.Build()

	require.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Equal(t, "OpenAPI Testing Suite", suite.Name)
	assert.Len(t, suite.Tests, 2) // Validation + Circular Reference
}

func TestOpenAPITestBuilder_Build_WithAllFeatures(t *testing.T) {
	specFile := createTempSpecFile(t, testOpenAPISpec)

	builder := NewOpenAPITestBuilder(specFile).
		WithValidation(true).
		WithCircularReferenceDetection(true).
		WithBreakingChangesDetection(true, "HEAD~1")

	suite, err := builder.Build()

	require.NoError(t, err)
	assert.NotNil(t, suite)
	assert.Equal(t, "OpenAPI Testing Suite", suite.Name)
	assert.Len(t, suite.Tests, 3) // All three tests
}

func TestOpenAPITestBuilder_Build_InvalidSpec(t *testing.T) {
	builder := NewOpenAPITestBuilder("nonexistent.yaml")
	_, err := builder.Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create OpenAPI integration")
}
