# OpenAPI Testing Package

The OpenAPI package provides comprehensive testing capabilities for OpenAPI specifications using pb33f.io's libopenapi library. This package integrates seamlessly with the GoWright testing framework to validate OpenAPI specifications, detect breaking changes, and identify circular references.

## Features

- **Specification Validation**: Validates OpenAPI specifications against the OpenAPI 3.0.3 standard
- **Breaking Changes Detection**: Compares specifications across git commits to identify breaking changes
- **Circular Reference Detection**: Identifies circular references in schema definitions
- **GoWright Integration**: Seamless integration with the GoWright testing framework
- **Flexible Test Builder**: Fluent API for building customized test suites

## Installation

The package uses pb33f.io's libopenapi library. Dependencies are automatically managed through Go modules:

```bash
go mod tidy
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/gowright/framework/pkg/openapi"
)

func main() {
    // Create an OpenAPI tester
    tester, err := openapi.NewOpenAPITester("path/to/your/openapi.yaml")
    if err != nil {
        panic(err)
    }

    // Validate the specification
    result := tester.ValidateSpec()
    fmt.Printf("Validation: %s (Passed: %t)\n", result.Message, result.Passed)

    // Check for circular references
    circularResult := tester.DetectCircularReferences()
    fmt.Printf("Circular References: %s (Passed: %t)\n", circularResult.Message, circularResult.Passed)

    // Check for breaking changes (requires git)
    breakingResult := tester.CheckBreakingChanges("HEAD~1")
    fmt.Printf("Breaking Changes: %s (Passed: %t)\n", breakingResult.Message, breakingResult.Passed)
}
```

### GoWright Integration

```go
package main

import (
    "github.com/gowright/framework/pkg/core"
    "github.com/gowright/framework/pkg/openapi"
)

func main() {
    // Create OpenAPI integration
    integration, err := openapi.NewOpenAPIIntegration("openapi.yaml")
    if err != nil {
        panic(err)
    }

    // Create a full test suite
    suite := integration.CreateFullTestSuite("HEAD~1")

    // Run with GoWright
    gowright := core.NewGoWright()
    gowright.AddTestSuite(suite)
    gowright.RunTests()
}
```

### Test Builder Pattern

```go
package main

import (
    "github.com/gowright/framework/pkg/openapi"
)

func main() {
    // Build a customized test suite
    suite, err := openapi.NewOpenAPITestBuilder("openapi.yaml").
        WithValidation(true).
        WithCircularReferenceDetection(true).
        WithBreakingChangesDetection(true, "HEAD~1").
        Build()

    if err != nil {
        panic(err)
    }

    // Use the suite with GoWright or run tests individually
    for _, test := range suite.Tests {
        // Run individual tests
    }
}
```

## API Reference

### OpenAPITester

The main testing class that provides OpenAPI validation capabilities.

#### Methods

- `NewOpenAPITester(specPath string) (*OpenAPITester, error)`: Creates a new tester instance
- `ValidateSpec() *TestResult`: Validates the OpenAPI specification
- `DetectCircularReferences() *TestResult`: Detects circular references
- `CheckBreakingChanges(previousCommit string) *TestResult`: Compares with previous version
- `RunAllTests(previousCommit string) []*TestResult`: Runs all available tests
- `GetSummary(results []*TestResult) string`: Returns a summary of test results

### OpenAPIIntegration

Provides integration with the GoWright testing framework.

#### Methods

- `NewOpenAPIIntegration(specPath string) (*OpenAPIIntegration, error)`: Creates integration instance
- `CreateValidationTest() core.TestCase`: Creates a validation test case
- `CreateCircularReferenceTest() core.TestCase`: Creates a circular reference test case
- `CreateBreakingChangesTest(previousCommit string) core.TestCase`: Creates a breaking changes test case
- `CreateFullTestSuite(previousCommit string) *core.TestSuite`: Creates a complete test suite

### OpenAPITestBuilder

Fluent API for building customized test suites.

#### Methods

- `NewOpenAPITestBuilder(specPath string) *OpenAPITestBuilder`: Creates a new builder
- `WithPreviousCommit(commit string) *OpenAPITestBuilder`: Sets previous commit for breaking changes
- `WithValidation(enabled bool) *OpenAPITestBuilder`: Enables/disables validation testing
- `WithCircularReferenceDetection(enabled bool) *OpenAPITestBuilder`: Enables/disables circular reference detection
- `WithBreakingChangesDetection(enabled bool, commit string) *OpenAPITestBuilder`: Enables/disables breaking changes detection
- `Build() (*core.TestSuite, error)`: Builds the test suite

## Test Results

All test methods return `TestResult` objects with the following structure:

```go
type TestResult struct {
    TestName     string                // Name of the test
    Passed       bool                  // Whether the test passed
    Message      string                // Summary message
    Details      []string              // Detailed information
    Errors       []ValidationError     // Validation errors found
    Warnings     []ValidationWarning   // Validation warnings
}
```

### Validation Errors

```go
type ValidationError struct {
    Path        string    // Path in the specification where error occurred
    Message     string    // Error message
    Severity    string    // Error severity level
    Line        int       // Line number (if available)
    Column      int       // Column number (if available)
}
```

### Breaking Changes

```go
type BreakingChange struct {
    Type        string      // Type of breaking change
    Path        string      // Path where change occurred
    OldValue    interface{} // Previous value
    NewValue    interface{} // New value
    Description string      // Description of the change
    Impact      string      // Impact on API consumers
}
```

## Validation Features

### Specification Validation

- Validates OpenAPI version
- Checks required fields (info, paths)
- Validates path items and operations
- Validates component schemas
- Checks response definitions
- Validates parameter definitions

### Breaking Changes Detection

The package detects various types of breaking changes:

- **PATH_REMOVED**: API paths that were removed
- **OPERATION_REMOVED**: HTTP operations that were removed
- **REQUIRED_PARAMETER_ADDED**: New required parameters
- **SCHEMA_REMOVED**: Schema definitions that were removed
- **RESPONSE_REMOVED**: Response definitions that were removed

### Circular Reference Detection

- Detects circular references in schema definitions
- Provides detailed reference chains
- Identifies problematic paths in the specification

## Git Integration

The breaking changes detection feature requires git to be available and the specification file to be tracked in git. The package uses git commands to retrieve previous versions of the specification for comparison.

## Examples

See the `examples/openapi_testing_example.go` file for comprehensive usage examples including:

- Basic validation
- Comprehensive testing
- Breaking changes detection
- Test builder pattern usage
- GoWright framework integration

## Error Handling

The package provides detailed error information for various scenarios:

- File not found errors
- Invalid YAML/JSON parsing errors
- OpenAPI specification parsing errors
- Git command execution errors
- Validation errors with specific paths and messages

## Best Practices

1. **Version Control**: Keep your OpenAPI specifications in version control for breaking changes detection
2. **Continuous Integration**: Integrate OpenAPI testing into your CI/CD pipeline
3. **Comprehensive Testing**: Use all three test types (validation, circular references, breaking changes) for thorough coverage
4. **Error Handling**: Always check for errors when creating tester instances
5. **Test Organization**: Use the test builder pattern for complex test configurations

## Dependencies

- `github.com/pb33f/libopenapi`: Core OpenAPI parsing and validation
- `github.com/pb33f/libopenapi-validator`: Additional validation capabilities
- `github.com/stretchr/testify`: Testing utilities (for tests)
- `github.com/gowright/framework/pkg/core`: GoWright framework integration

## Contributing

When contributing to this package:

1. Ensure all tests pass
2. Add tests for new functionality
3. Update documentation for API changes
4. Follow Go best practices and conventions
5. Test with various OpenAPI specification formats