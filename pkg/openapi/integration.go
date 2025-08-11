package openapi

import (
	"fmt"
	"time"

	"github.com/gowright/framework/pkg/core"
)

// OpenAPIIntegration provides integration with the GoWright testing framework
type OpenAPIIntegration struct {
	tester *OpenAPITester
}

// NewOpenAPIIntegration creates a new OpenAPI integration instance
func NewOpenAPIIntegration(specPath string) (*OpenAPIIntegration, error) {
	tester, err := NewOpenAPITester(specPath)
	if err != nil {
		return nil, err
	}

	return &OpenAPIIntegration{
		tester: tester,
	}, nil
}

// OpenAPIValidationTest implements the Test interface for OpenAPI validation
type OpenAPIValidationTest struct {
	name   string
	tester *OpenAPITester
}

// GetName returns the test name
func (t *OpenAPIValidationTest) GetName() string {
	return t.name
}

// Execute runs the OpenAPI validation test
func (t *OpenAPIValidationTest) Execute() *core.TestCaseResult {
	startTime := time.Now()
	result := &core.TestCaseResult{
		Name:      t.name,
		StartTime: startTime,
	}

	testResult := t.tester.ValidateSpec()

	if testResult.Passed {
		result.Status = core.TestStatusPassed
	} else {
		result.Status = core.TestStatusFailed
		result.Error = fmt.Errorf("OpenAPI validation failed: %s", testResult.Message)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Add logs for detailed information
	result.Logs = append(result.Logs, testResult.Message)
	for _, err := range testResult.Errors {
		result.Logs = append(result.Logs, fmt.Sprintf("Error at %s: %s", err.Path, err.Message))
	}
	for _, warning := range testResult.Warnings {
		result.Logs = append(result.Logs, fmt.Sprintf("Warning at %s: %s", warning.Path, warning.Message))
	}

	return result
}

// OpenAPICircularReferenceTest implements the Test interface for circular reference detection
type OpenAPICircularReferenceTest struct {
	name   string
	tester *OpenAPITester
}

// GetName returns the test name
func (t *OpenAPICircularReferenceTest) GetName() string {
	return t.name
}

// Execute runs the circular reference detection test
func (t *OpenAPICircularReferenceTest) Execute() *core.TestCaseResult {
	startTime := time.Now()
	result := &core.TestCaseResult{
		Name:      t.name,
		StartTime: startTime,
	}

	testResult := t.tester.DetectCircularReferences()

	if testResult.Passed {
		result.Status = core.TestStatusPassed
	} else {
		result.Status = core.TestStatusFailed
		result.Error = fmt.Errorf("circular reference detection failed: %s", testResult.Message)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Add logs for detailed information
	result.Logs = append(result.Logs, testResult.Message)
	for _, detail := range testResult.Details {
		result.Logs = append(result.Logs, fmt.Sprintf("Circular reference: %s", detail))
	}

	return result
}

// OpenAPIBreakingChangesTest implements the Test interface for breaking changes detection
type OpenAPIBreakingChangesTest struct {
	name           string
	tester         *OpenAPITester
	previousCommit string
}

// GetName returns the test name
func (t *OpenAPIBreakingChangesTest) GetName() string {
	return t.name
}

// Execute runs the breaking changes detection test
func (t *OpenAPIBreakingChangesTest) Execute() *core.TestCaseResult {
	startTime := time.Now()
	result := &core.TestCaseResult{
		Name:      t.name,
		StartTime: startTime,
	}

	testResult := t.tester.CheckBreakingChanges(t.previousCommit)

	if testResult.Passed {
		result.Status = core.TestStatusPassed
	} else {
		result.Status = core.TestStatusFailed
		result.Error = fmt.Errorf("breaking changes detected: %s", testResult.Message)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Add logs for detailed information
	result.Logs = append(result.Logs, testResult.Message)
	for _, detail := range testResult.Details {
		result.Logs = append(result.Logs, fmt.Sprintf("Breaking change: %s", detail))
	}

	return result
}

// CreateValidationTest creates a GoWright test for OpenAPI validation
func (oi *OpenAPIIntegration) CreateValidationTest() core.Test {
	return &OpenAPIValidationTest{
		name:   "OpenAPI Specification Validation",
		tester: oi.tester,
	}
}

// CreateCircularReferenceTest creates a GoWright test for circular reference detection
func (oi *OpenAPIIntegration) CreateCircularReferenceTest() core.Test {
	return &OpenAPICircularReferenceTest{
		name:   "OpenAPI Circular Reference Detection",
		tester: oi.tester,
	}
}

// CreateBreakingChangesTest creates a GoWright test for breaking changes detection
func (oi *OpenAPIIntegration) CreateBreakingChangesTest(previousCommit string) core.Test {
	return &OpenAPIBreakingChangesTest{
		name:           "OpenAPI Breaking Changes Detection",
		tester:         oi.tester,
		previousCommit: previousCommit,
	}
}

// CreateFullTestSuite creates a complete test suite for OpenAPI testing
func (oi *OpenAPIIntegration) CreateFullTestSuite(previousCommit string) *core.TestSuite {
	suite := &core.TestSuite{
		Name:  "OpenAPI Comprehensive Testing",
		Tests: []core.Test{},
	}

	// Add validation test
	suite.Tests = append(suite.Tests, oi.CreateValidationTest())

	// Add circular reference test
	suite.Tests = append(suite.Tests, oi.CreateCircularReferenceTest())

	// Add breaking changes test if previous commit is provided
	if previousCommit != "" {
		suite.Tests = append(suite.Tests, oi.CreateBreakingChangesTest(previousCommit))
	}

	return suite
}

// OpenAPITestBuilder provides a fluent interface for building OpenAPI tests
type OpenAPITestBuilder struct {
	specPath               string
	previousCommit         string
	includeValidation      bool
	includeCircularRef     bool
	includeBreakingChanges bool
}

// NewOpenAPITestBuilder creates a new OpenAPI test builder
func NewOpenAPITestBuilder(specPath string) *OpenAPITestBuilder {
	return &OpenAPITestBuilder{
		specPath:               specPath,
		includeValidation:      true,
		includeCircularRef:     true,
		includeBreakingChanges: false,
	}
}

// WithPreviousCommit sets the previous commit for breaking changes detection
func (b *OpenAPITestBuilder) WithPreviousCommit(commit string) *OpenAPITestBuilder {
	b.previousCommit = commit
	b.includeBreakingChanges = true
	return b
}

// WithValidation enables or disables validation testing
func (b *OpenAPITestBuilder) WithValidation(enabled bool) *OpenAPITestBuilder {
	b.includeValidation = enabled
	return b
}

// WithCircularReferenceDetection enables or disables circular reference detection
func (b *OpenAPITestBuilder) WithCircularReferenceDetection(enabled bool) *OpenAPITestBuilder {
	b.includeCircularRef = enabled
	return b
}

// WithBreakingChangesDetection enables or disables breaking changes detection
func (b *OpenAPITestBuilder) WithBreakingChangesDetection(enabled bool, commit string) *OpenAPITestBuilder {
	b.includeBreakingChanges = enabled
	if enabled {
		b.previousCommit = commit
	}
	return b
}

// Build creates the test suite based on the builder configuration
func (b *OpenAPITestBuilder) Build() (*core.TestSuite, error) {
	integration, err := NewOpenAPIIntegration(b.specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAPI integration: %w", err)
	}

	suite := &core.TestSuite{
		Name:  "OpenAPI Testing Suite",
		Tests: []core.Test{},
	}

	if b.includeValidation {
		suite.Tests = append(suite.Tests, integration.CreateValidationTest())
	}

	if b.includeCircularRef {
		suite.Tests = append(suite.Tests, integration.CreateCircularReferenceTest())
	}

	if b.includeBreakingChanges && b.previousCommit != "" {
		suite.Tests = append(suite.Tests, integration.CreateBreakingChangesTest(b.previousCommit))
	}

	return suite, nil
}
